package service

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/Acad600-Tpa/WEB-MV-242/services/auth/repository"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// TokenDetails contains the generated token information
type TokenDetails struct {
	AccessToken  string
	RefreshToken string
	UserID       string
	TokenType    string
	ExpiresIn    int64
}

// EmailService is an interface for sending emails
type EmailService interface {
	SendVerificationEmail(email, code string) error
	SendWelcomeEmail(email, name string) error
}

// AuthService defines the interface for authentication business logic
type AuthService interface {
	GenerateTokens(ctx context.Context, userID string) (*TokenDetails, error)
	ValidateToken(ctx context.Context, tokenString string) (map[string]interface{}, error)
	RefreshToken(ctx context.Context, refreshToken string) (*TokenDetails, error)
	Logout(ctx context.Context, accessToken, refreshToken string) error

	// User registration and verification
	RegisterUser(ctx context.Context, user *repository.User, plainPassword string) (string, error)
	VerifyEmail(ctx context.Context, email, code string) (string, error)
	ResendVerificationCode(ctx context.Context, email string) error

	// Google authentication
	AuthenticateWithGoogle(ctx context.Context, idToken string) (*TokenDetails, error)

	// Validation utilities
	ValidateName(name string) error
	ValidateUsername(ctx context.Context, username string) error
	ValidateEmail(ctx context.Context, email string) error
	ValidatePassword(password string) error
	ValidateAge(dateOfBirth time.Time) error
}

// authService implements the AuthService interface
type authService struct {
	authRepo     repository.AuthRepository
	emailService EmailService
	jwtSecret    string
	accessTTL    int // in minutes
	refreshTTL   int // in days
}

// defaultEmailService is a simple implementation that prints emails to console
type defaultEmailService struct{}

func (s *defaultEmailService) SendVerificationEmail(email, code string) error {
	fmt.Printf("Sending verification email to %s with code %s\n", email, code)
	return nil
}

func (s *defaultEmailService) SendWelcomeEmail(email, name string) error {
	fmt.Printf("Sending welcome email to %s (%s)\n", name, email)
	return nil
}

// NewAuthService creates a new auth service
func NewAuthService(authRepo repository.AuthRepository, jwtSecret string) AuthService {
	return &authService{
		authRepo:     authRepo,
		emailService: NewSMTPEmailService(), // Use real email service
		jwtSecret:    jwtSecret,
		accessTTL:    15, // 15 minutes
		refreshTTL:   7,  // 7 days
	}
}

// GenerateTokens creates new access and refresh tokens for a user
func (s *authService) GenerateTokens(ctx context.Context, userID string) (*TokenDetails, error) {
	td := &TokenDetails{
		UserID:    userID,
		TokenType: "Bearer",
	}

	// Set token expiration times
	accessExpiration := time.Now().Add(time.Minute * time.Duration(s.accessTTL))
	refreshExpiration := time.Now().Add(time.Hour * 24 * time.Duration(s.refreshTTL))

	// Calculate expiration in seconds from now
	td.ExpiresIn = int64(s.accessTTL * 60)

	// Create access token claims
	accessClaims := jwt.MapClaims{
		"sub": userID,
		"exp": accessExpiration.Unix(),
		"iat": time.Now().Unix(),
		"jti": uuid.New().String(),
	}

	// Create the access token
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)

	// Sign the access token
	var err error
	td.AccessToken, err = accessToken.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	// Create refresh token claims
	refreshClaims := jwt.MapClaims{
		"sub": userID,
		"exp": refreshExpiration.Unix(),
		"iat": time.Now().Unix(),
		"jti": uuid.New().String(),
	}

	// Create the refresh token
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)

	// Sign the refresh token
	td.RefreshToken, err = refreshToken.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	// Store the refresh token in the database
	if err := s.authRepo.StoreRefreshToken(ctx, userID, td.RefreshToken, refreshExpiration); err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	return td, nil
}

// ValidateToken validates a JWT token
func (s *authService) ValidateToken(ctx context.Context, tokenString string) (map[string]interface{}, error) {
	// Parse the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	// Validate the token is valid
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Convert claims to map[string]interface{}
	result := make(map[string]interface{})
	for key, value := range claims {
		result[key] = value
	}

	return result, nil
}

// RefreshToken generates new tokens using a valid refresh token
func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (*TokenDetails, error) {
	// Validate the refresh token
	claims, err := s.ValidateToken(ctx, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// Check if token is revoked
	isRevoked, err := s.authRepo.IsTokenRevoked(ctx, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to check token status: %w", err)
	}

	if isRevoked {
		return nil, fmt.Errorf("token has been revoked")
	}

	// Get user ID from claims
	userID, ok := claims["sub"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid user ID in token")
	}

	// Verify that the user ID in the token matches the one in the database
	storedUserID, err := s.authRepo.GetUserIDByRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to verify refresh token: %w", err)
	}

	if storedUserID != userID {
		return nil, fmt.Errorf("token mismatch")
	}

	// Revoke the old refresh token
	if err := s.authRepo.RevokeRefreshToken(ctx, refreshToken); err != nil {
		return nil, fmt.Errorf("failed to revoke old token: %w", err)
	}

	// Generate new tokens
	return s.GenerateTokens(ctx, userID)
}

// Logout revokes tokens
func (s *authService) Logout(ctx context.Context, accessToken, refreshToken string) error {
	// Validate refresh token
	_, err := s.ValidateToken(ctx, refreshToken)
	if err == nil {
		// Only try to revoke if the token is still valid
		if err := s.authRepo.RevokeRefreshToken(ctx, refreshToken); err != nil {
			// Log the error but don't return it - we still want to complete the logout
			fmt.Printf("Error revoking refresh token: %v\n", err)
		}
	}

	return nil
}

// RegisterUser handles the first step of user registration
func (s *authService) RegisterUser(ctx context.Context, user *repository.User, plainPassword string) (string, error) {
	// Validate user information
	if err := s.ValidateName(user.Name); err != nil {
		return "", err
	}

	if err := s.ValidateUsername(ctx, user.Username); err != nil {
		return "", err
	}

	if err := s.ValidateEmail(ctx, user.Email); err != nil {
		return "", err
	}

	if err := s.ValidatePassword(plainPassword); err != nil {
		return "", err
	}

	if err := s.ValidateAge(user.DateOfBirth); err != nil {
		return "", err
	}

	// Generate a unique ID for the user
	user.ID = uuid.New().String()

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	user.PasswordHash = string(hashedPassword)

	// Hash the security answer
	hashedAnswer, err := bcrypt.GenerateFromPassword([]byte(strings.ToLower(user.SecurityAnswerHash)), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash security answer: %w", err)
	}
	user.SecurityAnswerHash = string(hashedAnswer)

	// Set default values
	user.EmailVerified = false
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	// Save the user to the database
	if err := s.authRepo.CreateUser(ctx, user); err != nil {
		return "", fmt.Errorf("failed to create user: %w", err)
	}

	// Generate and send verification code
	verificationCode := s.generateVerificationCode()
	if err := s.authRepo.StoreVerificationCode(ctx, user.Email, verificationCode); err != nil {
		return "", fmt.Errorf("failed to store verification code: %w", err)
	}

	// Send the verification email
	if err := s.emailService.SendVerificationEmail(user.Email, verificationCode); err != nil {
		// Log the error but continue - don't want to fail registration if email sending fails
		fmt.Printf("Failed to send verification email: %v\n", err)
	}

	return user.Email, nil
}

// VerifyEmail verifies a user's email with the provided verification code
func (s *authService) VerifyEmail(ctx context.Context, email, code string) (string, error) {
	// Get the verification code from the database
	storedCode, err := s.authRepo.GetVerificationCode(ctx, email)
	if err != nil {
		return "", fmt.Errorf("verification code not found or expired: %w", err)
	}

	// Compare the provided code with the stored code
	if storedCode.Code != code {
		return "", fmt.Errorf("invalid verification code")
	}

	// Get the user by email
	user, err := s.authRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return "", fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		return "", fmt.Errorf("user not found")
	}

	// Mark the user's email as verified
	if err := s.authRepo.VerifyUserEmail(ctx, email); err != nil {
		return "", fmt.Errorf("failed to verify user email: %w", err)
	}

	// Delete the verification code
	if err := s.authRepo.DeleteVerificationCode(ctx, email); err != nil {
		// Log the error but continue - verification was successful
		fmt.Printf("Failed to delete verification code: %v\n", err)
	}

	// Send welcome email
	if err := s.emailService.SendWelcomeEmail(email, user.Name); err != nil {
		// Log the error but continue - don't want to fail verification if email sending fails
		fmt.Printf("Failed to send welcome email: %v\n", err)
	}

	// Generate tokens for the user
	tokens, err := s.GenerateTokens(ctx, user.ID)
	if err != nil {
		return "", fmt.Errorf("failed to generate tokens: %w", err)
	}

	return user.ID, nil
}

// ResendVerificationCode generates a new verification code and sends it to the user's email
func (s *authService) ResendVerificationCode(ctx context.Context, email string) error {
	// Check if the user exists
	user, err := s.authRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		return fmt.Errorf("user not found")
	}

	// If the user's email is already verified, return an error
	if user.EmailVerified {
		return fmt.Errorf("email already verified")
	}

	// Generate a new verification code
	verificationCode := s.generateVerificationCode()

	// Store the new verification code
	if err := s.authRepo.StoreVerificationCode(ctx, email, verificationCode); err != nil {
		return fmt.Errorf("failed to store verification code: %w", err)
	}

	// Send the verification email
	if err := s.emailService.SendVerificationEmail(email, verificationCode); err != nil {
		return fmt.Errorf("failed to send verification email: %w", err)
	}

	return nil
}

// AuthenticateWithGoogle authenticates a user with Google OAuth
func (s *authService) AuthenticateWithGoogle(ctx context.Context, idToken string) (*TokenDetails, error) {
	// Verify Google ID token
	googleUserInfo, err := s.verifyGoogleToken(idToken)
	if err != nil {
		return nil, fmt.Errorf("failed to verify Google token: %w", err)
	}

	// Check if user exists by email
	email := googleUserInfo.Email
	user, err := s.authRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to check user existence: %w", err)
	}

	var userID string

	if user == nil {
		// User doesn't exist - create a new account
		newUser := &repository.User{
			ID:                  uuid.New().String(),
			Name:                googleUserInfo.Name,
			Username:            s.generateUniqueUsername(ctx, email),
			Email:               email,
			PasswordHash:        "",          // No password for OAuth users
			Gender:              "",          // Will need to collect this information later
			DateOfBirth:         time.Time{}, // Will need to collect this information later
			ProfilePicturePath:  "",          // Could save Google profile pic if needed
			BannerPath:          "",
			SecurityQuestion:    "", // Will need to collect this information later
			SecurityAnswerHash:  "",
			SubscribeNewsletter: false,
			EmailVerified:       true, // Email is already verified through Google
			CreatedAt:           time.Now(),
			UpdatedAt:           time.Now(),
		}

		// Save the new user
		if err := s.authRepo.CreateUser(ctx, newUser); err != nil {
			return nil, fmt.Errorf("failed to create user from Google account: %w", err)
		}

		userID = newUser.ID
	} else {
		// User already exists
		userID = user.ID
	}

	// Generate tokens for the user
	return s.GenerateTokens(ctx, userID)
}

// Add a helper method to verify Google tokens
func (s *authService) verifyGoogleToken(idToken string) (*GoogleUserInfo, error) {
	// In production, you should use the Google API Client Library
	// For simplicity, we'll use a direct HTTP call to Google's token verification endpoint
	resp, err := http.Get("https://oauth2.googleapis.com/tokeninfo?id_token=" + idToken)
	if err != nil {
		return nil, fmt.Errorf("failed to contact Google API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("invalid Google token")
	}

	var userInfo GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to parse Google response: %w", err)
	}

	// Validate the issuer and audience (client ID)
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	if userInfo.Aud != clientID {
		return nil, fmt.Errorf("token was not issued for this application")
	}

	return &userInfo, nil
}

// Add a struct for Google user info
type GoogleUserInfo struct {
	Sub           string `json:"sub"`            // Google user ID
	Email         string `json:"email"`          // User's email
	Name          string `json:"name"`           // User's full name
	Picture       string `json:"picture"`        // URL to profile picture
	Iss           string `json:"iss"`            // Token issuer
	Aud           string `json:"aud"`            // Intended audience (our app's client ID)
	Exp           string `json:"exp"`            // Expiration time
	EmailVerified bool   `json:"email_verified"` // Whether email is verified
}

// Add a method to generate a unique username
func (s *authService) generateUniqueUsername(ctx context.Context, email string) string {
	// Extract username part from email
	username := strings.Split(email, "@")[0]

	// Clean up the username - remove any special characters
	username = regexp.MustCompile(`[^a-zA-Z0-9]`).ReplaceAllString(username, "")

	// Check if username already exists
	existingUser, err := s.authRepo.GetUserByUsername(ctx, username)
	if err != nil || existingUser != nil {
		// If error or user exists, add a unique suffix
		return username + "_" + uuid.New().String()[0:6]
	}

	return username
}

// generateVerificationCode generates a random 6-digit verification code
func (s *authService) generateVerificationCode() string {
	// Generate 6 random digits
	var code strings.Builder
	for i := 0; i < 6; i++ {
		// Generate a random number between 0 and 9
		n, _ := rand.Int(rand.Reader, big.NewInt(10))
		code.WriteString(n.String())
	}
	return code.String()
}

// ValidateName validates the user's name
func (s *authService) ValidateName(name string) error {
	// Check name length
	if len(name) < 4 {
		return errors.New("name must be at least 4 characters long")
	}

	// Check for symbols or numbers
	if regexp.MustCompile(`[0-9!@#$%^&*(),.?":{}|<>]`).MatchString(name) {
		return errors.New("name cannot contain numbers or symbols")
	}

	return nil
}

// ValidateUsername checks if the username is valid and unique
func (s *authService) ValidateUsername(ctx context.Context, username string) error {
	// Check username length
	if len(username) < 4 {
		return errors.New("username must be at least 4 characters long")
	}

	// Check if username is already taken
	existingUser, err := s.authRepo.GetUserByUsername(ctx, username)
	if err != nil {
		return fmt.Errorf("failed to check username uniqueness: %w", err)
	}

	if existingUser != nil {
		return errors.New("username is already taken")
	}

	return nil
}

// ValidateEmail checks if the email is valid and unique
func (s *authService) ValidateEmail(ctx context.Context, email string) error {
	// Check email format
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return errors.New("invalid email format")
	}

	// Check if email is already taken
	existingUser, err := s.authRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("failed to check email uniqueness: %w", err)
	}

	if existingUser != nil {
		return errors.New("email is already registered")
	}

	return nil
}

// ValidatePassword checks if the password meets the security requirements
func (s *authService) ValidatePassword(password string) error {
	// Password should be at least 8 characters long
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}

	// Password should contain at least one uppercase letter
	if !regexp.MustCompile(`[A-Z]`).MatchString(password) {
		return errors.New("password must contain at least one uppercase letter")
	}

	// Password should contain at least one lowercase letter
	if !regexp.MustCompile(`[a-z]`).MatchString(password) {
		return errors.New("password must contain at least one lowercase letter")
	}

	// Password should contain at least one digit
	if !regexp.MustCompile(`[0-9]`).MatchString(password) {
		return errors.New("password must contain at least one number")
	}

	// Password should contain at least one special character
	if !regexp.MustCompile(`[!@#$%^&*(),.?":{}|<>]`).MatchString(password) {
		return errors.New("password must contain at least one special character")
	}

	return nil
}

// ValidateAge checks if the user is at least 13 years old
func (s *authService) ValidateAge(dateOfBirth time.Time) error {
	// Calculate age
	now := time.Now()
	age := now.Year() - dateOfBirth.Year()

	// Adjust age if birthday hasn't occurred yet this year
	if now.Month() < dateOfBirth.Month() || (now.Month() == dateOfBirth.Month() && now.Day() < dateOfBirth.Day()) {
		age--
	}

	// Check if user is at least 13 years old
	if age < 13 {
		return errors.New("you must be at least 13 years old to register")
	}

	return nil
}
