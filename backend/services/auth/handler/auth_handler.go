package handler

import (
	"context"

	"github.com/AYCOM/backend/services/auth/proto"
	"github.com/AYCOM/backend/services/auth/repository"
	"github.com/AYCOM/backend/services/auth/service"
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

// Register handles user registration
func (s *AuthServiceServer) Register(ctx context.Context, req *proto.RegisterRequest) (*proto.RegisterResponse, error) {
	// Validate request
	if req.Password != req.ConfirmPassword {
		return &proto.RegisterResponse{
			Success: false,
			Message: "Passwords do not match",
		}, nil
	}

	// Create user from request
	user := &repository.User{
		Name:                  req.Name,
		Username:              req.Username,
		Email:                 req.Email,
		Gender:                req.Gender,
		DateOfBirth:           req.DateOfBirth,
		SecurityQuestion:      req.SecurityQuestion,
		SecurityAnswer:        req.SecurityAnswer,
		SubscribeToNewsletter: req.SubscribeToNewsletter,
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
	// Verify email using auth service
	userID, err := s.authService.VerifyEmail(ctx, req.Email, req.VerificationCode)
	if err != nil {
		return &proto.VerifyEmailResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	// Generate tokens for the verified user
	tokens, err := s.authService.GenerateTokens(ctx, userID)
	if err != nil {
		return &proto.VerifyEmailResponse{
			Success: false,
			Message: "Failed to generate tokens: " + err.Error(),
		}, nil
	}

	// Return success response with tokens
	return &proto.VerifyEmailResponse{
		Success:      true,
		Message:      "Email verified successfully",
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		UserId:       tokens.UserID,
		TokenType:    tokens.TokenType,
		ExpiresIn:    int32(tokens.ExpiresIn),
	}, nil
}

// ResendVerificationCode resends the email verification code
func (s *AuthServiceServer) ResendVerificationCode(ctx context.Context, req *proto.ResendVerificationCodeRequest) (*proto.ResendVerificationCodeResponse, error) {
	// Resend verification code using auth service
	err := s.authService.ResendVerificationCode(ctx, req.Email)
	if err != nil {
		return &proto.ResendVerificationCodeResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	// Return success response
	return &proto.ResendVerificationCodeResponse{
		Success: true,
		Message: "Verification code has been resent. Please check your email.",
	}, nil
}

// Login handles user login
func (s *AuthServiceServer) Login(ctx context.Context, req *proto.LoginRequest) (*proto.LoginResponse, error) {
	// Login user using auth service
	tokens, err := s.authService.Login(ctx, req.Email, req.Password)
	if err != nil {
		return &proto.LoginResponse{
			Success:      false,
			Message:      err.Error(),
			AccessToken:  "",
			RefreshToken: "",
			UserId:       "",
			TokenType:    "",
			ExpiresIn:    0,
		}, nil
	}

	// Return success response with tokens
	return &proto.LoginResponse{
		Success:      true,
		Message:      "Login successful",
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		UserId:       tokens.UserID,
		TokenType:    tokens.TokenType,
		ExpiresIn:    int32(tokens.ExpiresIn),
	}, nil
}

// ValidateToken validates a JWT token
func (s *AuthServiceServer) ValidateToken(ctx context.Context, req *proto.ValidateTokenRequest) (*proto.ValidateTokenResponse, error) {
	// Validate token using auth service
	claims, err := s.authService.ValidateToken(ctx, req.Token)
	if err != nil {
		return &proto.ValidateTokenResponse{
			Valid:   false,
			Message: err.Error(),
		}, nil
	}

	// Extract user ID from claims
	userID, ok := claims["user_id"].(string)
	if !ok {
		return &proto.ValidateTokenResponse{
			Valid:   false,
			Message: "Invalid token claims",
		}, nil
	}

	// Return success response
	return &proto.ValidateTokenResponse{
		Valid:   true,
		Message: "Token is valid",
		UserId:  userID,
	}, nil
}

// RefreshToken refreshes an access token using a refresh token
func (s *AuthServiceServer) RefreshToken(ctx context.Context, req *proto.RefreshTokenRequest) (*proto.RefreshTokenResponse, error) {
	// Refresh token using auth service
	tokens, err := s.authService.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return &proto.RefreshTokenResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	// Return success response with new tokens
	return &proto.RefreshTokenResponse{
		Success:      true,
		Message:      "Token refreshed successfully",
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		UserId:       tokens.UserID,
		TokenType:    tokens.TokenType,
		ExpiresIn:    int32(tokens.ExpiresIn),
	}, nil
}

// Logout handles user logout
func (s *AuthServiceServer) Logout(ctx context.Context, req *proto.LogoutRequest) (*proto.LogoutResponse, error) {
	// Logout user using auth service
	err := s.authService.Logout(ctx, req.AccessToken, req.RefreshToken)
	if err != nil {
		return &proto.LogoutResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	// Return success response
	return &proto.LogoutResponse{
		Success: true,
		Message: "Logged out successfully",
	}, nil
}

// GoogleLogin handles Google OAuth login
func (s *AuthServiceServer) GoogleLogin(ctx context.Context, req *proto.GoogleLoginRequest) (*proto.GoogleLoginResponse, error) {
	// Authenticate with Google using auth service
	tokens, err := s.authService.AuthenticateWithGoogle(ctx, req.IdToken)
	if err != nil {
		return &proto.GoogleLoginResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	// Return success response with tokens
	return &proto.GoogleLoginResponse{
		Success:      true,
		Message:      "Google login successful",
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		UserId:       tokens.UserID,
		TokenType:    tokens.TokenType,
		ExpiresIn:    int32(tokens.ExpiresIn),
	}, nil
}
