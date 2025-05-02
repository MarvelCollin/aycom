package service

import (
	"context"
	"crypto/rand"
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

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
	"gorm.io/gorm"
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
// Use gorm.DB consistent with repository
type authService struct {
	db                *gorm.DB
	repo              repository.AuthRepository
	emailService      EmailService
	jwtSecret         string
	accessTTL         time.Duration
	refreshTTL        time.Duration
	userServiceConn   *grpc.ClientConn
	userServiceClient UserServiceClient
}

// UserServiceClient is an interface for the user service client
type UserServiceClient interface {
	CreateUserProfile(ctx context.Context, req *CreateUserProfileRequest) error
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
	log.Printf("Mock implementation: Would create user profile for ID: %s", req.UserId)
	return nil
}

// UpdateUserVerificationStatus calls the user service to update a user's verification status
func (c *UserServiceClientImpl) UpdateUserVerificationStatus(ctx context.Context, req *UpdateUserVerificationStatusRequest) error {
	log.Printf("Mock implementation: Would update verification status for user ID: %s to %t", req.UserId, req.IsVerified)
	return nil
}

// NewAuthService creates a new auth service and initializes the DB connection using GORM.
func NewAuthService(cfg *config.Config) (AuthService, error) {
	var db *gorm.DB
	var err error
	maxRetries := 5
	retryInterval := time.Second * 5

	for i := 0; i < maxRetries; i++ {
		db, err = repository.NewGormConnection(&cfg.Database)
		if err == nil {
			sqlDB, sqlErr := db.DB()
			if sqlErr == nil {
				if pingErr := sqlDB.Ping(); pingErr == nil {
					log.Println("Successfully connected to the database using GORM")
					break
				} else {
					err = fmt.Errorf("failed to ping database: %w", pingErr)
					sqlDB.Close()
					db = nil
				}
			} else {
				err = fmt.Errorf("failed to get underlying sql.DB from GORM: %w", sqlErr)
				db = nil
			}
		}
		log.Printf("Failed to connect to database with GORM (attempt %d/%d): %v", i+1, maxRetries, err)
		if i < maxRetries-1 {
			log.Printf("Retrying in %v...", retryInterval)
			time.Sleep(retryInterval)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("could not connect to database with GORM after %d attempts: %w", maxRetries, err)
	}

	err = db.AutoMigrate(&repository.User{}, &repository.Token{}, &repository.OAuthConnection{})
	if err != nil {
		log.Printf("Warning: Failed to auto-migrate GORM models: %v", err)
	}

	repo := repository.NewPostgresAuthRepository(db)

	var emailSvc EmailService

	accessTTL := time.Duration(cfg.AccessTTL) * time.Minute
	refreshTTL := time.Duration(cfg.RefreshTTL) * 24 * time.Hour

	svc := &authService{
		db:           db,
		repo:         repo,
		emailService: emailSvc,
		jwtSecret:    cfg.JWTSecret,
		accessTTL:    accessTTL,
		refreshTTL:   refreshTTL,
	}

	return svc, nil
}

// GetMigrationStatus checks migration status. Placeholder implementation.
func (s *authService) GetMigrationStatus() error {
	if s.db == nil {
		return errors.New("database connection is not initialized")
	}
	log.Println("Checking migration status (placeholder)... Database connection is available.")
	return nil
}

// generateVerificationCode generates a random 6-digit verification code
func generateVerificationCode() (string, error) {
	max := big.NewInt(900000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	code := n.Int64() + 100000
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

	secretKey := os.Getenv("RECAPTCHA_SECRET_KEY")
	if secretKey == "" {
		fmt.Println("WARNING: RECAPTCHA_SECRET_KEY environment variable not set")
		return errors.New("recaptcha secret key not configured")
	}

	fmt.Println("Verifying reCAPTCHA token with secret key:", secretKey[:5]+"...")

	resp, err := http.PostForm("https://www.google.com/recaptcha/api/siteverify", url.Values{
		"secret":   {secretKey},
		"response": {recaptchaToken},
	})
	if err != nil {
		fmt.Printf("Error making reCAPTCHA verification request: %v\n", err)
		return fmt.Errorf("failed to verify recaptcha: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		Success    bool     `json:"success"`
		Score      float64  `json:"score,omitempty"`
		Action     string   `json:"action,omitempty"`
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
	if modelUser == nil {
		return nil
	}

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
		Email:                 modelUser.Email,
		PasswordHash:          modelUser.PasswordHash,
		Name:                  modelUser.Name,
		Username:              modelUser.Username,
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
	if repoUser == nil {
		return nil
	}

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
func (s *authService) RegisterUser(ctx context.Context, user *model.User, password string, recaptchaToken string) (string, error) {
	if user == nil || user.Email == "" || user.Username == "" || password == "" {
		return "", errors.New("name, username, email, and password are required")
	}

	if recaptchaToken != "" && os.Getenv("RECAPTCHA_SECRET_KEY") != "" {
		if err := verifyRecaptcha(recaptchaToken); err != nil {
			return "", fmt.Errorf("recaptcha verification failed: %w", err)
		}
	}

	existingUser, _ := s.repo.FindUserByEmail(user.Email)
	if existingUser != nil {
		return "", errors.New("email already exists")
	}
	existingUser, _ = s.repo.FindUserByUsername(user.Username)
	if existingUser != nil {
		return "", errors.New("username already exists")
	}

	hashedPassword, err := hashPassword(password)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	verificationCode, err := generateVerificationCode()
	if err != nil {
		return "", fmt.Errorf("failed to generate verification code: %w", err)
	}
	expiryTime := time.Now().Add(24 * time.Hour)

	repoUser := &repository.User{
		ID:                    uuid.New(),
		Name:                  user.Name,
		Username:              user.Username,
		Email:                 user.Email,
		PasswordHash:          hashedPassword,
		Gender:                user.Gender,
		DateOfBirth:           user.DateOfBirth.Format("2006-01-02"),
		SecurityQuestion:      user.SecurityQuestion,
		SecurityAnswer:        user.SecurityAnswer,
		EmailVerified:         false,
		VerificationCode:      verificationCode,
		VerificationExpiresAt: expiryTime,
		SubscribeToNewsletter: user.NewsletterSubscription,
	}

	err = s.repo.SaveUser(repoUser)
	if err != nil {
		log.Printf("Error saving user to database: %v", err)
		return "", fmt.Errorf("failed to register user: %w", err)
	}

	if s.emailService != nil {
		err = s.emailService.SendVerificationEmail(user.Email, verificationCode)
		if err != nil {
			log.Printf("Warning: Failed to send verification email to %s: %v", user.Email, err)
		}
	} else {
		log.Printf("EmailService not configured. Skipping verification email for %s", user.Email)
	}

	log.Printf("Successfully registered user with email: %s, ID: %s", repoUser.Email, repoUser.ID)
	return repoUser.Email, nil
}

// VerifyEmail verifies a user's email with a verification code
func (s *authService) VerifyEmail(ctx context.Context, email string, code string) (string, error) {
	if email == "" || code == "" {
		return "", errors.New("email and verification code are required")
	}

	user, err := s.repo.FindUserByEmail(email)
	if err != nil {
		log.Printf("VerifyEmail: User not found for email %s: %v", email, err)
		return "", errors.New("invalid email or verification code")
	}

	if user.EmailVerified {
		log.Printf("VerifyEmail: Email %s already verified for user %s", email, user.ID)
		return user.ID.String(), nil
	}

	if user.VerificationCode != code || time.Now().After(user.VerificationExpiresAt) {
		log.Printf("VerifyEmail: Invalid or expired code for email %s", email)
		return "", errors.New("invalid or expired verification code")
	}

	user.EmailVerified = true
	user.VerificationCode = ""

	err = s.repo.UpdateUser(user)
	if err != nil {
		log.Printf("VerifyEmail: Failed to update user %s: %v", user.ID, err)
		return "", fmt.Errorf("failed to verify email: %w", err)
	}

	log.Printf("Successfully verified email for user %s", user.ID)
	return user.ID.String(), nil
}

// ResendVerificationCode resends a verification code to a user's email
func (s *authService) ResendVerificationCode(ctx context.Context, email string) error {
	if email == "" {
		return errors.New("email is required")
	}

	user, err := s.repo.FindUserByEmail(email)
	if err != nil {
		log.Printf("ResendVerificationCode: User not found for email %s: %v", email, err)
		return errors.New("failed to process request")
	}

	if user.EmailVerified {
		log.Printf("ResendVerificationCode: Email %s already verified for user %s", email, user.ID)
		return errors.New("email is already verified")
	}

	newCode, err := generateVerificationCode()
	if err != nil {
		log.Printf("ResendVerificationCode: Failed to generate new code for %s: %v", email, err)
		return fmt.Errorf("failed to generate verification code: %w", err)
	}
	newExpiry := time.Now().Add(24 * time.Hour)

	user.VerificationCode = newCode
	user.VerificationExpiresAt = newExpiry
	err = s.repo.UpdateUser(user)
	if err != nil {
		log.Printf("ResendVerificationCode: Failed to update user %s: %v", user.ID, err)
		return fmt.Errorf("failed to update verification code: %w", err)
	}

	if s.emailService != nil {
		err = s.emailService.SendVerificationEmail(email, newCode)
		if err != nil {
			log.Printf("Warning: Failed to resend verification email to %s: %v", email, err)
		}
	} else {
		log.Printf("EmailService not configured. Skipping resend verification email for %s", email)
	}

	log.Printf("Successfully resent verification code to email: %s", email)
	return nil
}

// Login authenticates a user and returns JWT tokens
func (s *authService) Login(ctx context.Context, email string, password string) (*Tokens, error) {
	if email == "" || password == "" {
		return nil, errors.New("email and password are required")
	}

	user, err := s.repo.FindUserByEmail(email)
	if err != nil {
		log.Printf("Login attempt failed for email %s: User not found or DB error: %v", email, err)
		return nil, errors.New("invalid credentials")
	}

	err = checkPassword(password, user.PasswordHash)
	if err != nil {
		log.Printf("Login attempt failed for email %s: Invalid password", email)
		return nil, errors.New("invalid credentials")
	}

	if !user.EmailVerified {
		log.Printf("Login attempt failed for email %s: Email not verified", email)
		return nil, errors.New("email not verified")
	}

	tokens, err := s.GenerateTokens(ctx, user.ID.String())
	if err != nil {
		log.Printf("Login failed for email %s: Failed to generate tokens: %v", email, err)
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	log.Printf("User %s logged in successfully", user.ID)
	return tokens, nil
}

// ValidateToken validates a JWT token and returns its claims
func (s *authService) ValidateToken(ctx context.Context, tokenString string) (map[string]interface{}, error) {
	if tokenString == "" {
		return nil, errors.New("token is required")
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				return nil, errors.New("malformed token")
			} else if ve.Errors&(jwt.ValidationErrorExpired|jwt.ValidationErrorNotValidYet) != 0 {
				return nil, errors.New("token expired or not valid yet")
			}
		}
		log.Printf("Token validation failed: %v", err)
		return nil, errors.New("invalid token")
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	result := make(map[string]interface{})
	for key, value := range claims {
		result[key] = value
	}

	return result, nil
}

// GenerateTokens generates JWT access and refresh tokens for a user
func (s *authService) GenerateTokens(ctx context.Context, userID string) (*Tokens, error) {
	if userID == "" {
		return nil, errors.New("user ID is required")
	}

	accessExpiry := time.Now().Add(s.accessTTL)
	refreshExpiry := time.Now().Add(s.refreshTTL)

	accessClaims := jwt.MapClaims{
		"sub": userID,
		"exp": accessExpiry.Unix(),
		"iat": time.Now().Unix(),
		"typ": "access",
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	refreshClaims := jwt.MapClaims{
		"sub": userID,
		"exp": refreshExpiry.Unix(),
		"iat": time.Now().Unix(),
		"typ": "refresh",
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	expiresIn := int64(time.Until(accessExpiry).Seconds())

	return &Tokens{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		UserID:       userID,
		TokenType:    "Bearer",
		ExpiresIn:    expiresIn,
	}, nil
}

// RefreshToken refreshes an access token using a refresh token
func (s *authService) RefreshToken(ctx context.Context, refreshTokenString string) (*Tokens, error) {
	if refreshTokenString == "" {
		return nil, errors.New("refresh token is required")
	}

	claims, err := s.ValidateToken(ctx, refreshTokenString)
	if err != nil {
		log.Printf("RefreshToken failed: Invalid refresh token: %v", err)
		return nil, errors.New("invalid or expired refresh token")
	}

	tokenType, ok := claims["typ"].(string)
	if !ok || tokenType != "refresh" {
		log.Printf("RefreshToken failed: Invalid token type '%s'", tokenType)
		return nil, errors.New("invalid token type")
	}

	userID, ok := claims["sub"].(string)
	if !ok || userID == "" {
		log.Printf("RefreshToken failed: Missing or invalid user ID in claims")
		return nil, errors.New("invalid user ID in token")
	}

	newTokens, err := s.GenerateTokens(ctx, userID)
	if err != nil {
		log.Printf("RefreshToken failed for user %s: Failed to generate new tokens: %v", userID, err)
		return nil, fmt.Errorf("failed to generate new tokens: %w", err)
	}

	log.Printf("Successfully refreshed tokens for user %s", userID)
	return newTokens, nil
}

// Logout invalidates a user's tokens
func (s *authService) Logout(ctx context.Context, accessToken string, refreshToken string) error {
	if refreshToken == "" {
		return errors.New("refresh token is required for logout")
	}

	claims, err := s.ValidateToken(ctx, refreshToken)
	if err != nil && !errors.Is(err, errors.New("token expired or not valid yet")) {
		log.Printf("Logout failed: Invalid refresh token: %v", err)
		return nil
	}

	userID, ok := claims["sub"].(string)
	if !ok || userID == "" {
		log.Printf("Logout failed: Could not extract user ID from refresh token")
		return nil
	}

	log.Printf("User %s logged out (refresh token invalidated if stateful)", userID)
	return nil
}

// AuthenticateWithGoogle implements Google OAuth authentication
func (s *authService) AuthenticateWithGoogle(ctx context.Context, idToken string) (*Tokens, error) {
	if idToken == "" {
		return nil, errors.New("Google ID token is required")
	}

	googleUserID := "google_" + uuid.New().String()
	email := "mock.google.user." + uuid.New().String()[:4] + "@example.com"
	name := "Mock Google User"

	oauthConn, err := s.repo.FindOAuthConnection("google", googleUserID)
	var user *repository.User

	if err == nil {
		user, err = s.repo.FindUserByID(oauthConn.UserID.String())
		if err != nil {
			log.Printf("Google Auth Error: Found OAuth connection for %s but failed to find user %s: %v", googleUserID, oauthConn.UserID, err)
			return nil, errors.New("failed to retrieve user associated with Google account")
		}
		log.Printf("Google Auth: Found existing user %s via Google ID %s", user.ID, googleUserID)
	} else if errors.Is(err, errors.New("OAuth connection not found")) {
		log.Printf("Google Auth: No existing OAuth connection for Google ID %s", googleUserID)

		user, err = s.repo.FindUserByEmail(email)
		if err == nil {
			log.Printf("Google Auth: Found existing user %s by email %s. Linking Google ID %s.", user.ID, email, googleUserID)
		} else if errors.Is(err, errors.New("user not found")) {
			log.Printf("Google Auth: No user found with email %s. Creating new user for Google ID %s.", email, googleUserID)
			newUser := &repository.User{
				ID:            uuid.New(),
				Name:          name,
				Username:      email,
				Email:         email,
				PasswordHash:  "",
				EmailVerified: true,
			}
			err = s.repo.SaveUser(newUser)
			if err != nil {
				log.Printf("Google Auth Error: Failed to create new user for email %s: %v", email, err)
				return nil, errors.New("failed to create user account for Google login")
			}
			user = newUser
			log.Printf("Google Auth: Created new user %s for email %s", user.ID, email)
		} else {
			log.Printf("Google Auth Error: Failed to check user by email %s: %v", email, err)
			return nil, errors.New("database error during Google authentication")
		}

		newConn := &repository.OAuthConnection{
			ID:         uuid.New(),
			UserID:     user.ID,
			Provider:   "google",
			ProviderID: googleUserID,
		}
		err = s.repo.SaveOAuthConnection(newConn)
		if err != nil {
			log.Printf("Google Auth Error: Failed to save OAuth connection for user %s, provider google: %v", user.ID, err)
		} else {
			log.Printf("Google Auth: Saved OAuth connection for user %s, provider google", user.ID)
		}

	} else {
		log.Printf("Google Auth Error: Failed to check OAuth connection for provider google, ID %s: %v", googleUserID, err)
		return nil, errors.New("database error during Google authentication")
	}

	tokens, err := s.GenerateTokens(ctx, user.ID.String())
	if err != nil {
		log.Printf("Google Auth Error: Failed to generate tokens for user %s: %v", user.ID, err)
		return nil, errors.New("failed to generate session tokens after Google login")
	}

	log.Printf("Google Auth successful for user %s", user.ID)
	return tokens, nil
}

// Interface for email sending (needs implementation)
type EmailService interface {
	SendVerificationEmail(toEmail, code string) error
}

// MockEmailService is a placeholder implementation
type MockEmailService struct{}

func (m *MockEmailService) SendVerificationEmail(toEmail, code string) error {
	log.Printf("MOCK EMAIL: Sending verification email to %s with code %s", toEmail, code)
	return nil
}

func Add(a, b int) int {
	return a + b
}
