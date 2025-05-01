package handler

import (
	"context"

	"time"

	"github.com/Acad600-Tpa/WEB-MV-242/backend/services/auth/model"
	proto "github.com/Acad600-Tpa/WEB-MV-242/backend/services/auth/proto"

	"github.com/Acad600-Tpa/WEB-MV-242/backend/services/auth/service"
)

// AuthServiceServer implements the AuthService gRPC server
type AuthServiceServer struct {
	proto.UnimplementedAuthServiceServer
	authService service.AuthService
}

// NewAuthServiceServer creates a new auth service server
func NewAuthServiceServer(authService service.AuthService) *AuthServiceServer {
	return &AuthServiceServer{
		authService: authService,
	}
}

// Placeholder types for testing
type RegisterRequest struct {
	Name                  string
	Username              string
	Email                 string
	Password              string
	ConfirmPassword       string
	Gender                string
	DateOfBirth           string
	SecurityQuestion      string
	SecurityAnswer        string
	SubscribeToNewsletter bool
	ProfilePictureUrl     string
	BannerUrl             string
	RecaptchaToken        string
}

type RegisterResponse struct {
	Success bool
	Message string
	Email   string
}

// Register handles user registration
func (s *AuthServiceServer) Register(ctx context.Context, req *proto.RegisterRequest) (*proto.RegisterResponse, error) {
	// Validate request
	if req.Password != req.ConfirmPassword {
		return &proto.RegisterResponse{
			Success: false,
			Message: "Passwords do not match",
		}, nil
	}

	// Parse date of birth
	dob, _ := time.Parse("2006-01-02", req.DateOfBirth)

	// Create user from request
	user := &model.User{
		Name:                   req.Name,
		Username:               req.Username,
		Email:                  req.Email,
		Gender:                 req.Gender,
		DateOfBirth:            dob,
		SecurityQuestion:       req.SecurityQuestion,
		SecurityAnswer:         req.SecurityAnswer,
		NewsletterSubscription: req.SubscribeToNewsletter,
		// Model doesn't have these fields, but repository does
		// ProfilePictureURL and BannerURL will be handled by the user service
	}

	// Register user using auth service with reCAPTCHA token
	email, err := s.authService.RegisterUser(ctx, user, req.Password, req.RecaptchaToken)
	if err != nil {
		return &proto.RegisterResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	// Return success response
	return &proto.RegisterResponse{
		Success: true,
		Message: "Registration successful. Please check your email for verification code.",
		Email:   email,
	}, nil
}

// VerifyEmail handles email verification
func (s *AuthServiceServer) VerifyEmail(ctx context.Context, req *proto.VerifyEmailRequest) (*proto.VerifyEmailResponse, error) {
	// Return a stub implementation for now
	return &proto.VerifyEmailResponse{
		Success:      true,
		Message:      "Email verified successfully",
		AccessToken:  "sample-access-token",
		RefreshToken: "sample-refresh-token",
		UserId:       "sample-user-id",
		TokenType:    "Bearer",
		ExpiresIn:    3600,
	}, nil
}

// ResendVerificationCode resends the email verification code
func (s *AuthServiceServer) ResendVerificationCode(ctx context.Context, req *proto.ResendVerificationCodeRequest) (*proto.ResendVerificationCodeResponse, error) {
	// Return a stub implementation for now
	return &proto.ResendVerificationCodeResponse{
		Success: true,
		Message: "Verification code has been resent. Please check your email.",
	}, nil
}

// Login handles user login
func (s *AuthServiceServer) Login(ctx context.Context, req *proto.LoginRequest) (*proto.LoginResponse, error) {
	// Return a stub implementation for now
	return &proto.LoginResponse{
		Success:      true,
		Message:      "Login successful",
		AccessToken:  "sample-access-token",
		RefreshToken: "sample-refresh-token",
		UserId:       "sample-user-id",
		TokenType:    "Bearer",
		ExpiresIn:    3600,
	}, nil
}

// ValidateToken validates a JWT token
func (s *AuthServiceServer) ValidateToken(ctx context.Context, req *proto.ValidateTokenRequest) (*proto.ValidateTokenResponse, error) {
	// Return a stub implementation for now
	return &proto.ValidateTokenResponse{
		Valid:   true,
		Message: "Token is valid",
		UserId:  "sample-user-id",
	}, nil
}

// RefreshToken refreshes an access token using a refresh token
func (s *AuthServiceServer) RefreshToken(ctx context.Context, req *proto.RefreshTokenRequest) (*proto.RefreshTokenResponse, error) {
	// Return a stub implementation for now
	return &proto.RefreshTokenResponse{
		Success:      true,
		Message:      "Token refreshed successfully",
		AccessToken:  "new-access-token",
		RefreshToken: "new-refresh-token",
		UserId:       "sample-user-id",
		TokenType:    "Bearer",
		ExpiresIn:    3600,
	}, nil
}

// Logout handles user logout
func (s *AuthServiceServer) Logout(ctx context.Context, req *proto.LogoutRequest) (*proto.LogoutResponse, error) {
	// Return a stub implementation for now
	return &proto.LogoutResponse{
		Success: true,
		Message: "Logged out successfully",
	}, nil
}

// GoogleLogin handles Google OAuth login
func (s *AuthServiceServer) GoogleLogin(ctx context.Context, req *proto.GoogleLoginRequest) (*proto.GoogleLoginResponse, error) {
	// Return a stub implementation for now
	return &proto.GoogleLoginResponse{
		Success:      true,
		Message:      "Google login successful",
		AccessToken:  "sample-access-token",
		RefreshToken: "sample-refresh-token",
		UserId:       "sample-user-id",
		TokenType:    "Bearer",
		ExpiresIn:    3600,
	}, nil
}
