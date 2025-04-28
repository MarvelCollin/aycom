package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/Acad600-Tpa/WEB-MV-242/backend/services/auth/repository"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
)

// Tokens holds the JWT access and refresh tokens
type Tokens struct {
	AccessToken  string
	RefreshToken string
	UserID       string
	TokenType    string
	ExpiresIn    int64
}

// AuthService defines the methods for auth-related operations
type AuthService interface {
	RegisterUser(ctx context.Context, user *repository.User, password string, recaptchaToken string) (string, error)
	VerifyEmail(ctx context.Context, email string, code string) (string, error)
	ResendVerificationCode(ctx context.Context, email string) error
	Login(ctx context.Context, email string, password string) (*Tokens, error)
	ValidateToken(ctx context.Context, token string) (map[string]interface{}, error)
	GenerateTokens(ctx context.Context, userID string) (*Tokens, error)
	RefreshToken(ctx context.Context, refreshToken string) (*Tokens, error)
	Logout(ctx context.Context, accessToken string, refreshToken string) error
	AuthenticateWithGoogle(ctx context.Context, idToken string) (*Tokens, error)
}

// authService implements the AuthService interface
type authService struct {
	repo              repository.AuthRepository
	emailService      EmailService
	jwtSecret         string
	userServiceConn   *grpc.ClientConn
	userServiceClient UserServiceClient
}

// UserServiceClient is an interface for the user service client
type UserServiceClient interface {
	CreateUser(ctx context.Context, userId, username, email, name, gender, dateOfBirth, profilePicture, banner, secQuestion, secAnswer string, subscribeToNewsletter bool) error
	UpdateUserVerification(ctx context.Context, userID string, isVerified bool) error
}

// UserServiceClientImpl implements the UserServiceClient interface
type UserServiceClientImpl struct {
	conn *grpc.ClientConn
}

// NewUserServiceClient creates a new user service client
func NewUserServiceClient(conn *grpc.ClientConn) UserServiceClient {
	return &UserServiceClientImpl{
		conn: conn,
	}
}

// CreateUser calls the user service to create a new user
func (c *UserServiceClientImpl) CreateUser(ctx context.Context, userId, username, email, name, gender, dateOfBirth, profilePicture, banner, secQuestion, secAnswer string, subscribeToNewsletter bool) error {
	// Implementation would use the gRPC client to call the user service
	// For this example, we'll just return nil as if it succeeded
	// In a real implementation, you would use the protobuf client here
	fmt.Printf("Creating user in user service: %s, %s\n", userId, email)
	return nil
}

// UpdateUserVerification calls the user service to update a user's verification status
func (c *UserServiceClientImpl) UpdateUserVerification(ctx context.Context, userID string, isVerified bool) error {
	// Implementation would use the gRPC client to call the user service
	// For this example, we'll just return nil as if it succeeded
	fmt.Printf("Updating user verification in user service: %s, %v\n", userID, isVerified)
	return nil
}

// NewAuthService creates a new auth service
func NewAuthService(repo repository.AuthRepository, emailService EmailService, jwtSecret string, userServiceClient UserServiceClient) AuthService {
	return &authService{
		repo:              repo,
		emailService:      emailService,
		jwtSecret:         jwtSecret,
		userServiceClient: userServiceClient,
	}
}

// generateVerificationCode generates a random 6-digit verification code
func generateVerificationCode() (string, error) {
	// Generate a random 6-digit code
	max := big.NewInt(900000) // 900000 is the range (999999 - 100000 + 1)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	code := n.Int64() + 100000 // Add 100000 to ensure it's 6 digits
	return fmt.Sprintf("%06d", code), nil
}

// hashPassword hashes a password using bcrypt
func hashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// checkPassword checks if a password matches the hashed password
func checkPassword(password, hashedPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// verifyRecaptcha verifies a reCAPTCHA token with Google's reCAPTCHA API
func verifyRecaptcha(recaptchaToken string) error {
	if recaptchaToken == "" {
		return errors.New("recaptcha token is required")
	}

	// Get reCAPTCHA secret key from environment
	secretKey := os.Getenv("RECAPTCHA_SECRET_KEY")
	if secretKey == "" {
		fmt.Println("WARNING: RECAPTCHA_SECRET_KEY environment variable not set")
		return errors.New("recaptcha secret key not configured")
	}

	fmt.Println("Verifying reCAPTCHA token with secret key:", secretKey[:5]+"...")

	// Create the request to Google's reCAPTCHA API
	resp, err := http.PostForm("https://www.google.com/recaptcha/api/siteverify", url.Values{
		"secret":   {secretKey},
		"response": {recaptchaToken},
	})
	if err != nil {
		fmt.Printf("Error making reCAPTCHA verification request: %v\n", err)
		return fmt.Errorf("failed to verify recaptcha: %v", err)
	}
	defer resp.Body.Close()

	// Parse the response
	var result struct {
		Success    bool     `json:"success"`
		Score      float64  `json:"score,omitempty"`  // v3 only
		Action     string   `json:"action,omitempty"` // v3 only
		ErrorCodes []string `json:"error-codes,omitempty"`
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		fmt.Printf("Error decoding reCAPTCHA response: %v\n", err)
		return fmt.Errorf("failed to decode recaptcha response: %v", err)
	}

	fmt.Printf("reCAPTCHA verification result: %+v\n", result)

	if !result.Success {
		if len(result.ErrorCodes) > 0 {
			fmt.Printf("reCAPTCHA verification failed with errors: %v\n", result.ErrorCodes)
			return fmt.Errorf("recaptcha verification failed: %v", result.ErrorCodes)
		}
		fmt.Println("reCAPTCHA verification failed without specific error codes")
		return errors.New("recaptcha verification failed")
	}

	fmt.Println("reCAPTCHA verification successful")
	return nil
}

// RegisterUser registers a new user
func (s *authService) RegisterUser(ctx context.Context, user *repository.User, password string, recaptchaToken string) (string, error) {
	// Verify reCAPTCHA token
	err := verifyRecaptcha(recaptchaToken)
	if err != nil {
		return "", fmt.Errorf("recaptcha verification failed: %v", err)
	}

	// Check if email already exists
	existingUser, err := s.repo.FindUserByEmail(user.Email)
	if err == nil && existingUser != nil {
		return "", errors.New("email already registered")
	}

	// Check if username already exists
	existingUser, err = s.repo.FindUserByUsername(user.Username)
	if err == nil && existingUser != nil {
		return "", errors.New("username already taken")
	}

	// Generate a unique user ID
	userID := uuid.New().String()
	user.ID = userID

	// Hash the password
	hashedPassword, err := hashPassword(password)
	if err != nil {
		return "", errors.New("failed to hash password")
	}
	user.HashedPassword = hashedPassword

	// Generate a verification code
	verificationCode, err := generateVerificationCode()
	if err != nil {
		return "", errors.New("failed to generate verification code")
	}
	user.VerificationCode = verificationCode
	user.VerificationCodeExpiry = time.Now().Add(24 * time.Hour) // Code expires in 24 hours
	user.IsVerified = false
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	// Save user in auth database
	err = s.repo.SaveUser(user)
	if err != nil {
		return "", errors.New("failed to create user: " + err.Error())
	}

	// Create user in user service
	err = s.userServiceClient.CreateUser(
		ctx,
		userID,
		user.Username,
		user.Email,
		user.Name,
		user.Gender,
		user.DateOfBirth.Format("2006-01-02"),
		"", // Default empty profile picture
		"", // Default empty banner
		user.SecurityQuestion,
		user.SecurityAnswer,
		user.SubscribeToNewsletter,
	)
	if err != nil {
		// Rollback auth user if user service creation fails
		_ = s.repo.DeleteUser(userID)
		return "", errors.New("failed to create user in user service: " + err.Error())
	}

	// Send verification email
	err = s.emailService.SendVerificationEmail(user.Email, verificationCode)
	if err != nil {
		// Log the error but don't fail the registration
		fmt.Printf("Failed to send verification email: %v\n", err)
	}

	return user.Email, nil
}

// VerifyEmail verifies a user's email using the provided verification code
func (s *authService) VerifyEmail(ctx context.Context, email string, code string) (string, error) {
	// Find the user by email
	user, err := s.repo.FindUserByEmail(email)
	if err != nil {
		return "", errors.New("user not found")
	}

	// Check if the user is already verified
	if user.IsVerified {
		return user.ID, nil // Already verified, return success
	}

	// Check if the verification code is expired
	if time.Now().After(user.VerificationCodeExpiry) {
		return "", errors.New("verification code expired")
	}

	// Check if the verification code matches
	if user.VerificationCode != code {
		return "", errors.New("invalid verification code")
	}

	// Mark the user as verified
	user.IsVerified = true
	user.VerificationCode = "" // Clear the verification code
	user.UpdatedAt = time.Now()

	// Update the user in auth database
	err = s.repo.UpdateUser(user)
	if err != nil {
		return "", errors.New("failed to update user: " + err.Error())
	}

	// Update the user in user service
	err = s.userServiceClient.UpdateUserVerification(ctx, user.ID, true)
	if err != nil {
		// Log the error but don't fail the verification
		fmt.Printf("Failed to update user verification in user service: %v\n", err)
	}

	return user.ID, nil
}

// ResendVerificationCode resends the email verification code
func (s *authService) ResendVerificationCode(ctx context.Context, email string) error {
	// Find the user by email
	user, err := s.repo.FindUserByEmail(email)
	if err != nil {
		return errors.New("user not found")
	}

	// Check if the user is already verified
	if user.IsVerified {
		return errors.New("user already verified")
	}

	// Generate a new verification code
	verificationCode, err := generateVerificationCode()
	if err != nil {
		return errors.New("failed to generate verification code")
	}
	user.VerificationCode = verificationCode
	user.VerificationCodeExpiry = time.Now().Add(24 * time.Hour) // Code expires in 24 hours
	user.UpdatedAt = time.Now()

	// Update the user in database
	err = s.repo.UpdateUser(user)
	if err != nil {
		return errors.New("failed to update user: " + err.Error())
	}

	// Send verification email
	err = s.emailService.SendVerificationEmail(user.Email, verificationCode)
	if err != nil {
		return errors.New("failed to send verification email: " + err.Error())
	}

	return nil
}

// Login authenticates a user and returns tokens
func (s *authService) Login(ctx context.Context, email string, password string) (*Tokens, error) {
	// Find user by email
	user, err := s.repo.FindUserByEmail(email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	// Check password
	err = checkPassword(password, user.HashedPassword)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	// Check if user is verified
	if !user.IsVerified {
		return nil, errors.New("email not verified")
	}

	// Generate tokens
	return s.GenerateTokens(ctx, user.ID)
}

// ValidateToken validates a JWT token and returns its claims
func (s *authService) ValidateToken(ctx context.Context, tokenString string) (map[string]interface{}, error) {
	// Parse the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	// Handle errors
	if err != nil {
		return nil, errors.New("invalid token: " + err.Error())
	}

	// Check if the token is valid
	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	// Check if the token is expired
	if exp, ok := claims["exp"].(float64); ok {
		if time.Unix(int64(exp), 0).Before(time.Now()) {
			return nil, errors.New("token expired")
		}
	}

	// Convert claims to a map
	claimsMap := make(map[string]interface{})
	for key, val := range claims {
		claimsMap[key] = val
	}

	return claimsMap, nil
}

// GenerateTokens generates new access and refresh tokens for a user
func (s *authService) GenerateTokens(ctx context.Context, userID string) (*Tokens, error) {
	// Set access token expiry (1 hour)
	expiresAt := time.Now().Add(1 * time.Hour)

	// Create access token claims
	accessTokenClaims := jwt.MapClaims{
		"user_id": userID,
		"exp":     expiresAt.Unix(),
		"iat":     time.Now().Unix(),
		"type":    "access",
	}

	// Create access token
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessTokenClaims)
	accessTokenString, err := accessToken.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return nil, errors.New("failed to generate access token: " + err.Error())
	}

	// Generate refresh token (random string)
	refreshTokenBytes := make([]byte, 32)
	_, err = rand.Read(refreshTokenBytes)
	if err != nil {
		return nil, errors.New("failed to generate refresh token: " + err.Error())
	}
	refreshTokenString := base64.URLEncoding.EncodeToString(refreshTokenBytes)

	// Store refresh token in database
	token := &repository.Token{
		UserID:       userID,
		RefreshToken: refreshTokenString,
		ExpiresAt:    time.Now().Add(30 * 24 * time.Hour), // 30 days
		CreatedAt:    time.Now(),
	}
	err = s.repo.SaveToken(token)
	if err != nil {
		return nil, errors.New("failed to save refresh token: " + err.Error())
	}

	// Return tokens
	return &Tokens{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		UserID:       userID,
		TokenType:    "Bearer",
		ExpiresIn:    expiresAt.Unix() - time.Now().Unix(),
	}, nil
}

// RefreshToken refreshes an access token using a refresh token
func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (*Tokens, error) {
	// Find refresh token in database
	token, err := s.repo.FindTokenByUserID("") // We need to modify this to search by refresh token
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	// Check if refresh token matches
	if token.RefreshToken != refreshToken {
		return nil, errors.New("invalid refresh token")
	}

	// Check if refresh token is expired
	if time.Now().After(token.ExpiresAt) {
		return nil, errors.New("refresh token expired")
	}

	// Generate new tokens
	return s.GenerateTokens(ctx, token.UserID)
}

// Logout invalidates a user's tokens
func (s *authService) Logout(ctx context.Context, accessToken string, refreshToken string) error {
	// Validate access token to get user ID
	claims, err := s.ValidateToken(ctx, accessToken)
	if err != nil {
		return errors.New("invalid access token")
	}

	// Extract user ID from claims
	userID, ok := claims["user_id"].(string)
	if !ok {
		return errors.New("invalid token claims")
	}

	// Delete refresh token from database
	err = s.repo.DeleteToken(userID, refreshToken)
	if err != nil {
		return errors.New("failed to delete refresh token: " + err.Error())
	}

	return nil
}

// AuthenticateWithGoogle authenticates a user with Google OAuth
func (s *authService) AuthenticateWithGoogle(ctx context.Context, idToken string) (*Tokens, error) {
	// This would normally validate the Google ID token and extract user info
	// For this example, we'll just return an error
	return nil, errors.New("Google authentication not implemented yet")
}
