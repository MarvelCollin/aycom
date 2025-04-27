package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/Acad600-Tpa/WEB-MV-242/services/auth/proto"
	"github.com/Acad600-Tpa/WEB-MV-242/services/auth/repository"
	"github.com/Acad600-Tpa/WEB-MV-242/services/auth/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AuthServiceServer implements the gRPC Auth service server
type AuthServiceServer struct {
	proto.UnimplementedAuthServiceServer
	authService service.AuthService
}

// NewAuthServiceServer creates a new auth service gRPC server
func NewAuthServiceServer(authService service.AuthService) proto.AuthServiceServer {
	return &AuthServiceServer{
		authService: authService,
	}
}

// Login handles user authentication and token generation
func (s *AuthServiceServer) Login(ctx context.Context, req *proto.LoginRequest) (*proto.TokenResponse, error) {
	// In a real implementation, this would validate credentials against user service
	// For now, we'll just simulate successful login with a fake user ID
	userID := "user-123" // This would come from user service authentication

	// Generate tokens
	tokens, err := s.authService.GenerateTokens(ctx, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate tokens: %v", err)
	}

	return &proto.TokenResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		UserId:       tokens.UserID,
		TokenType:    tokens.TokenType,
		ExpiresIn:    tokens.ExpiresIn,
	}, nil
}

// Register handles user registration (Step 1)
func (s *AuthServiceServer) Register(ctx context.Context, req *proto.RegisterRequest) (*proto.RegistrationResponse, error) {
	// Verify reCAPTCHA token if provided
	if req.RecaptchaToken != "" {
		valid, err := verifyRecaptchaToken(req.RecaptchaToken)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to verify reCAPTCHA: %v", err)
		}
		if !valid {
			return nil, status.Errorf(codes.InvalidArgument, "reCAPTCHA verification failed, please try again")
		}
	}

	// Parse date of birth
	dob, err := time.Parse("2006-01-02", req.DateOfBirth)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid date of birth format: %v", err)
	}

	// Check if passwords match
	if req.Password != req.ConfirmPassword {
		return nil, status.Errorf(codes.InvalidArgument, "passwords do not match")
	}

	// Create user model
	user := &repository.User{
		Name:                req.Name,
		Username:            req.Username,
		Email:               req.Email,
		Gender:              req.Gender,
		DateOfBirth:         dob,
		SecurityQuestion:    req.SecurityQuestion,
		SecurityAnswerHash:  req.SecurityAnswer, // Will be hashed in service
		SubscribeNewsletter: req.SubscribeToNewsletter,
	}

	// Handle profile picture and banner if provided
	// In a real implementation, these would be saved to a storage service
	// and the paths would be stored in the user model
	if len(req.ProfilePicture) > 0 {
		user.ProfilePicturePath = fmt.Sprintf("users/%s/profile.jpg", user.Email)
	}

	if len(req.Banner) > 0 {
		user.BannerPath = fmt.Sprintf("users/%s/banner.jpg", user.Email)
	}

	// Register the user
	email, err := s.authService.RegisterUser(ctx, user, req.Password)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to register user: %v", err)
	}

	return &proto.RegistrationResponse{
		Success: true,
		Message: "Registration successful. Please check your email for verification code.",
		Email:   email,
	}, nil
}

// verifyRecaptchaToken verifies the reCAPTCHA token with Google
func verifyRecaptchaToken(token string) (bool, error) {
	if token == "" {
		return false, fmt.Errorf("empty reCAPTCHA token")
	}

	// Get secret key from environment
	secretKey := os.Getenv("GOOGLE_SECRET")
	if secretKey == "" {
		return false, fmt.Errorf("reCAPTCHA secret key not configured")
	}

	// Prepare verification request to Google
	resp, err := http.PostForm(
		"https://www.google.com/recaptcha/api/siteverify",
		url.Values{
			"secret":   {secretKey},
			"response": {token},
		},
	)

	if err != nil {
		return false, fmt.Errorf("failed to contact reCAPTCHA API: %w", err)
	}
	defer resp.Body.Close()

	// Parse response
	var result struct {
		Success    bool     `json:"success"`
		Score      float64  `json:"score,omitempty"`
		ErrorCodes []string `json:"error-codes,omitempty"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, fmt.Errorf("failed to decode reCAPTCHA response: %w", err)
	}

	return result.Success, nil
}

// VerifyEmail verifies a user's email with a verification code (Step 2)
func (s *AuthServiceServer) VerifyEmail(ctx context.Context, req *proto.VerifyEmailRequest) (*proto.TokenResponse, error) {
	// Verify the email with the provided code
	userID, err := s.authService.VerifyEmail(ctx, req.Email, req.VerificationCode)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "email verification failed: %v", err)
	}

	// Generate tokens for the verified user
	tokens, err := s.authService.GenerateTokens(ctx, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate tokens: %v", err)
	}

	return &proto.TokenResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		UserId:       tokens.UserID,
		TokenType:    tokens.TokenType,
		ExpiresIn:    tokens.ExpiresIn,
	}, nil
}

// ResendVerificationCode resends a verification code to the user's email
func (s *AuthServiceServer) ResendVerificationCode(ctx context.Context, req *proto.ResendCodeRequest) (*proto.ResendCodeResponse, error) {
	// Resend the verification code
	err := s.authService.ResendVerificationCode(ctx, req.Email)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to resend verification code: %v", err)
	}

	return &proto.ResendCodeResponse{
		Success: true,
		Message: "Verification code sent successfully.",
	}, nil
}

// GoogleAuth handles Google OAuth authentication
func (s *AuthServiceServer) GoogleAuth(ctx context.Context, req *proto.GoogleAuthRequest) (*proto.TokenResponse, error) {
	// Authenticate with Google
	tokens, err := s.authService.AuthenticateWithGoogle(ctx, req.TokenId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Google authentication failed: %v", err)
	}

	return &proto.TokenResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		UserId:       tokens.UserID,
		TokenType:    tokens.TokenType,
		ExpiresIn:    tokens.ExpiresIn,
	}, nil
}

// ValidateToken validates an access token
func (s *AuthServiceServer) ValidateToken(ctx context.Context, req *proto.ValidateRequest) (*proto.ValidateResponse, error) {
	// Validate the token
	claims, err := s.authService.ValidateToken(ctx, req.Token)
	if err != nil {
		return &proto.ValidateResponse{
			Valid:  false,
			UserId: "",
			Claims: nil,
		}, nil
	}

	// Extract user ID from claims
	userID, ok := claims["sub"].(string)
	if !ok {
		return &proto.ValidateResponse{
			Valid:  false,
			UserId: "",
			Claims: nil,
		}, nil
	}

	// Convert claims to string map
	stringClaims := make(map[string]string)
	for k, v := range claims {
		if str, ok := v.(string); ok {
			stringClaims[k] = str
		} else {
			// Convert non-string values to string as needed
			stringClaims[k] = fmt.Sprintf("%v", v)
		}
	}

	return &proto.ValidateResponse{
		Valid:  true,
		UserId: userID,
		Claims: stringClaims,
	}, nil
}

// RefreshToken generates new tokens using a refresh token
func (s *AuthServiceServer) RefreshToken(ctx context.Context, req *proto.RefreshRequest) (*proto.TokenResponse, error) {
	// Refresh tokens
	tokens, err := s.authService.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid refresh token: %v", err)
	}

	return &proto.TokenResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		UserId:       tokens.UserID,
		TokenType:    tokens.TokenType,
		ExpiresIn:    tokens.ExpiresIn,
	}, nil
}

// Logout invalidates tokens
func (s *AuthServiceServer) Logout(ctx context.Context, req *proto.LogoutRequest) (*proto.LogoutResponse, error) {
	// Logout
	err := s.authService.Logout(ctx, req.AccessToken, req.RefreshToken)
	if err != nil {
		return &proto.LogoutResponse{
			Success: false,
			Message: err.Error(),
		}, status.Errorf(codes.Internal, "failed to logout: %v", err)
	}

	return &proto.LogoutResponse{
		Success: true,
		Message: "successfully logged out",
	}, nil
}
