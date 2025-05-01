package service

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/Acad600-Tpa/WEB-MV-242/backend/services/auth/config"
	"github.com/Acad600-Tpa/WEB-MV-242/backend/services/auth/model"
	"github.com/Acad600-Tpa/WEB-MV-242/backend/services/auth/repository"

	//"github.com/Acad600-Tpa/WEB-MV-242/backend/services/user/proto" // TODO: Fix missing proto
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

// Temporary struct to replace the missing proto package
type CreateUserProfileRequest struct {
	UserId                string
	Name                  string
	Username              string
	Email                 string
	Gender                string
	DateOfBirth           string
	ProfilePictureUrl     string
	BannerUrl             string
	SecurityQuestion      string
	SecurityAnswer        string
	SubscribeToNewsletter bool
}

type UpdateUserVerificationStatusRequest struct {
	UserId     string
	IsVerified bool
}

// AuthService defines the methods for auth-related operations
type AuthService interface {
	RegisterUser(ctx context.Context, user *model.User, password string, recaptchaToken string) (string, error)
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
	// Changed method name and signature to match updated user.proto
	CreateUserProfile(ctx context.Context, req *CreateUserProfileRequest) error
	// Changed method name and signature to match updated user.proto
	UpdateUserVerificationStatus(ctx context.Context, req *UpdateUserVerificationStatusRequest) error
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

// CreateUserProfile calls the user service to create a new user profile
func (c *UserServiceClientImpl) CreateUserProfile(ctx context.Context, req *CreateUserProfileRequest) error {
	// TODO: Re-implement when proto dependency is fixed
	log.Printf("Mock implementation: Would create user profile for ID: %s", req.UserId)
	return nil
}

// UpdateUserVerificationStatus calls the user service to update a user's verification status
func (c *UserServiceClientImpl) UpdateUserVerificationStatus(ctx context.Context, req *UpdateUserVerificationStatusRequest) error {
	// TODO: Re-implement when proto dependency is fixed
	log.Printf("Mock implementation: Would update verification status for user ID: %s to %t", req.UserId, req.IsVerified)
	return nil
}

// AuthServiceImpl is a concrete implementation of AuthService
// It now holds the database connection.
// We use *sql.DB because the current repository uses it.
// Consider standardizing to GORM later if needed.
type AuthServiceImpl struct {
	DB           *sql.DB // Changed from gorm.DB to sql.DB to match repository
	repo         repository.AuthRepository
	emailService EmailService
	jwtSecret    string
	// Add other fields like userServiceClient if needed for full functionality
}

// NewAuthService creates a new auth service and initializes the DB connection.
func NewAuthService() (*AuthServiceImpl, error) {
	// Load database configuration from environment variables
	cfg := config.LoadDatabaseConfig()

	// Connect to the database with retry mechanism
	var db *sql.DB
	var err error
	maxRetries := 5
	retryInterval := time.Second * 5

	for i := 0; i < maxRetries; i++ {
		db, err = repository.NewPostgresConnection(cfg)
		if err == nil {
			// Test the connection
			if pingErr := db.Ping(); pingErr == nil {
				log.Println("Successfully connected to the database")
				break // Success
			} else {
				err = fmt.Errorf("failed to ping database: %w", pingErr)
				db.Close() // Close the potentially invalid connection
				db = nil
			}
		}
		log.Printf("Failed to connect to database (attempt %d/%d): %v", i+1, maxRetries, err)
		if i < maxRetries-1 {
			log.Printf("Retrying in %v...", retryInterval)
			time.Sleep(retryInterval)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("could not connect to database after %d attempts: %w", maxRetries, err)
	}

	// Create the repository (assuming NewSQLAuthRepository exists or needs creating)
	// repo := repository.NewSQLAuthRepository(db)

	// Load JWT secret
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "default_secret_key" // Provide a default for safety
		log.Println("Warning: JWT_SECRET environment variable not set. Using default.")
	}

	// Create the service implementation
	svc := &AuthServiceImpl{
		DB: db,
		// repo: repo, // Uncomment when repository is ready
		// emailService: emailSvc, // Initialize email service if needed
		jwtSecret: jwtSecret,
		// Initialize userServiceClient if needed
	}

	return svc, nil
}

// GetMigrationStatus checks migration status. Placeholder implementation.
// TODO: Implement actual migration logic if needed here or separately.
func (s *AuthServiceImpl) GetMigrationStatus() error {
	if s.DB == nil {
		return errors.New("database connection is not initialized")
	}
	log.Println("Checking migration status (placeholder)... Database connection is available.")
	// Attempt to run the table creation logic again, just in case
	err := repository.CreateTables(s.DB) // Use the exported CreateTables function
	if err != nil {
		log.Printf("Warning: Failed to ensure tables exist during status check: %v", err)
		// Don't fail the status check, just log the warning
	}
	return nil
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

// convertModelToRepoUser converts a model.User to a repository.User
func convertModelToRepoUser(modelUser *model.User) *repository.User {
	verificationCode := ""
	if modelUser.VerificationCode != nil {
		verificationCode = *modelUser.VerificationCode
	}

	verificationExpiry := time.Time{}
	if modelUser.VerificationCodeExpiresAt != nil {
		verificationExpiry = *modelUser.VerificationCodeExpiresAt
	}

	return &repository.User{
		ID:                    modelUser.ID,
		Name:                  modelUser.Name,
		Username:              modelUser.Username,
		Email:                 modelUser.Email,
		PasswordHash:          modelUser.PasswordHash,
		Gender:                modelUser.Gender,
		DateOfBirth:           modelUser.DateOfBirth.Format("2006-01-02"),
		SecurityQuestion:      modelUser.SecurityQuestion,
		SecurityAnswer:        modelUser.SecurityAnswer,
		EmailVerified:         modelUser.IsActivated,
		VerificationCode:      verificationCode,
		VerificationExpiresAt: verificationExpiry,
		SubscribeToNewsletter: modelUser.NewsletterSubscription,
		CreatedAt:             modelUser.CreatedAt,
		UpdatedAt:             modelUser.UpdatedAt,
	}
}

// convertRepoToModelUser converts a repository.User to a model.User
func convertRepoToModelUser(repoUser *repository.User) *model.User {
	var verificationCode *string
	if repoUser.VerificationCode != "" {
		vc := repoUser.VerificationCode
		verificationCode = &vc
	}

	var expiresAt *time.Time
	if !repoUser.VerificationExpiresAt.IsZero() {
		exp := repoUser.VerificationExpiresAt
		expiresAt = &exp
	}

	dateOfBirth, _ := time.Parse("2006-01-02", repoUser.DateOfBirth)

	return &model.User{
		ID:                        repoUser.ID,
		Name:                      repoUser.Name,
		Username:                  repoUser.Username,
		Email:                     repoUser.Email,
		PasswordHash:              repoUser.PasswordHash,
		Gender:                    repoUser.Gender,
		DateOfBirth:               dateOfBirth,
		SecurityQuestion:          repoUser.SecurityQuestion,
		SecurityAnswer:            repoUser.SecurityAnswer,
		IsActivated:               repoUser.EmailVerified,
		VerificationCode:          verificationCode,
		VerificationCodeExpiresAt: expiresAt,
		NewsletterSubscription:    repoUser.SubscribeToNewsletter,
		CreatedAt:                 repoUser.CreatedAt,
		UpdatedAt:                 repoUser.UpdatedAt,
	}
}

// RegisterUser registers a new user
func (s *AuthServiceImpl) RegisterUser(ctx context.Context, user *model.User, password string, recaptchaToken string) (string, error) {
	// Basic validation
	if user == nil {
		return "", errors.New("user is required")
	}
	if user.Email == "" {
		return "", errors.New("email is required")
	}
	if password == "" {
		return "", errors.New("password is required")
	}

	// In a real implementation we would:
	// 1. Verify the reCAPTCHA token
	// 2. Check if the user already exists
	// 3. Hash the password
	// 4. Create a verification code
	// 5. Save the user to the database
	// 6. Send a verification email

	log.Printf("Would register user with email: %s", user.Email)

	return user.Email, nil
}

// VerifyEmail verifies a user's email with a verification code
func (s *AuthServiceImpl) VerifyEmail(ctx context.Context, email string, code string) (string, error) {
	// Basic validation
	if email == "" {
		return "", errors.New("email is required")
	}
	if code == "" {
		return "", errors.New("verification code is required")
	}

	// In a real implementation we would:
	// 1. Find the user by email
	// 2. Check if the verification code matches and is not expired
	// 3. Mark the user as verified
	// 4. Return the user ID

	log.Printf("Would verify email: %s with code: %s", email, code)

	// For testing, return a dummy user ID
	return uuid.New().String(), nil
}

// ResendVerificationCode resends a verification code to a user's email
func (s *AuthServiceImpl) ResendVerificationCode(ctx context.Context, email string) error {
	// Basic validation
	if email == "" {
		return errors.New("email is required")
	}

	// In a real implementation we would:
	// 1. Find the user by email
	// 2. Generate a new verification code
	// 3. Update the user in the database
	// 4. Send a new verification email

	log.Printf("Would resend verification code to email: %s", email)

	return nil
}

// Login authenticates a user and returns JWT tokens
func (s *AuthServiceImpl) Login(ctx context.Context, email string, password string) (*Tokens, error) {
	// Basic validation
	if email == "" {
		return nil, errors.New("email is required")
	}
	if password == "" {
		return nil, errors.New("password is required")
	}

	// In a real implementation we would:
	// 1. Find the user by email
	// 2. Verify the password
	// 3. Check if the user is activated, not banned, etc.
	// 4. Generate tokens

	log.Printf("Would login user with email: %s", email)

	// For testing, return tokens for a dummy user ID
	return s.GenerateTokens(ctx, uuid.New().String())
}

// ValidateToken validates a JWT token and returns its claims
func (s *AuthServiceImpl) ValidateToken(ctx context.Context, tokenString string) (map[string]interface{}, error) {
	// Basic validation
	if tokenString == "" {
		return nil, errors.New("token is required")
	}

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

	// Check if the token is valid
	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	// Convert claims to map[string]interface{}
	result := make(map[string]interface{})
	for key, value := range claims {
		result[key] = value
	}

	return result, nil
}

// GenerateTokens generates JWT access and refresh tokens for a user
func (s *AuthServiceImpl) GenerateTokens(ctx context.Context, userID string) (*Tokens, error) {
	if userID == "" {
		return nil, errors.New("user ID is required")
	}

	// Default expiration times
	accessExpiry := time.Now().Add(time.Hour)           // 1 hour
	refreshExpiry := time.Now().Add(7 * 24 * time.Hour) // 7 days

	// Create claims for the access token
	accessClaims := jwt.MapClaims{
		"sub": userID,
		"exp": accessExpiry.Unix(),
		"iat": time.Now().Unix(),
		"typ": "access",
	}

	// Create and sign the access token
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	// Create claims for the refresh token
	refreshClaims := jwt.MapClaims{
		"sub": userID,
		"exp": refreshExpiry.Unix(),
		"iat": time.Now().Unix(),
		"typ": "refresh",
	}

	// Create and sign the refresh token
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	// Calculate token expiration duration in seconds
	expiresIn := accessExpiry.Unix() - time.Now().Unix()

	// Return the tokens
	return &Tokens{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		UserID:       userID,
		TokenType:    "Bearer",
		ExpiresIn:    expiresIn,
	}, nil
}

// RefreshToken refreshes an access token using a refresh token
func (s *AuthServiceImpl) RefreshToken(ctx context.Context, refreshTokenString string) (*Tokens, error) {
	// Basic validation
	if refreshTokenString == "" {
		return nil, errors.New("refresh token is required")
	}

	// In a real implementation we would:
	// 1. Validate the refresh token
	// 2. Check if the token exists in the database and is not revoked
	// 3. Generate new tokens

	// Parse the token to extract user ID
	token, err := jwt.Parse(refreshTokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	// Extract user ID
	userID, ok := claims["sub"].(string)
	if !ok {
		return nil, errors.New("invalid user ID in token")
	}

	// Generate new tokens
	return s.GenerateTokens(ctx, userID)
}

// Logout invalidates a user's tokens
func (s *AuthServiceImpl) Logout(ctx context.Context, accessToken string, refreshToken string) error {
	// Basic validation
	if accessToken == "" || refreshToken == "" {
		return errors.New("both access token and refresh token are required")
	}

	// In a real implementation we would:
	// 1. Add the tokens to a blacklist or remove them from the database
	// 2. Set a blacklist expiry time to match the token expiry

	log.Printf("Would logout user with access token: %s...", accessToken[:10])

	return nil
}

// AuthenticateWithGoogle implements Google OAuth authentication
func (s *AuthServiceImpl) AuthenticateWithGoogle(ctx context.Context, idToken string) (*Tokens, error) {
	// This is a placeholder implementation until the Google OAuth integration is fully implemented
	if idToken == "" {
		return nil, errors.New("Google ID token is required")
	}

	// In a real implementation, we would:
	// 1. Verify the ID token with Google's OAuth API
	// 2. Extract user information from the token
	// 3. Check if the user exists in our database
	// 4. If not, create a new user
	// 5. Generate and return JWT tokens

	log.Println("Google authentication requested with token:", idToken[:10]+"...")

	// For testing purposes, generate a random user ID
	userID := uuid.New().String()

	// Generate tokens for the user
	tokens := &Tokens{
		AccessToken:  "google_mock_access_token",
		RefreshToken: "google_mock_refresh_token",
		UserID:       userID,
		TokenType:    "Bearer",
		ExpiresIn:    3600, // 1 hour
	}

	return tokens, nil
}

func Add(a, b int) int {
	return a + b
}
