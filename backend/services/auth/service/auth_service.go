package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Acad600-Tpa/WEB-MV-242/services/auth/repository"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// TokenDetails contains the generated token information
type TokenDetails struct {
	AccessToken  string
	RefreshToken string
	UserID       string
	TokenType    string
	ExpiresIn    int64
}

// AuthService defines the interface for authentication business logic
type AuthService interface {
	GenerateTokens(ctx context.Context, userID string) (*TokenDetails, error)
	ValidateToken(ctx context.Context, tokenString string) (map[string]interface{}, error)
	RefreshToken(ctx context.Context, refreshToken string) (*TokenDetails, error)
	Logout(ctx context.Context, accessToken, refreshToken string) error
}

// authService implements the AuthService interface
type authService struct {
	authRepo   repository.AuthRepository
	jwtSecret  string
	accessTTL  int // in minutes
	refreshTTL int // in days
}

// NewAuthService creates a new auth service
func NewAuthService(authRepo repository.AuthRepository, jwtSecret string) AuthService {
	return &authService{
		authRepo:   authRepo,
		jwtSecret:  jwtSecret,
		accessTTL:  15, // 15 minutes
		refreshTTL: 7,  // 7 days
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
