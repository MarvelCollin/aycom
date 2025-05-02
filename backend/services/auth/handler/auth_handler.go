package handler

import (
	"context"
	"errors"
	"log"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

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

// Helper to map service errors to gRPC status codes
func mapErrorToGrpcStatus(err error) error {
	if err == nil {
		return nil
	}
	log.Printf("Service Error: %v", err)

	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return status.Error(codes.NotFound, "resource not found")
	case err.Error() == "invalid credentials":
		return status.Error(codes.Unauthenticated, err.Error())
	case err.Error() == "email not verified":
		return status.Error(codes.FailedPrecondition, err.Error())
	case err.Error() == "email already exists", err.Error() == "username already exists":
		return status.Error(codes.AlreadyExists, err.Error())
	case err.Error() == "invalid or expired verification code":
		return status.Error(codes.InvalidArgument, err.Error())
	case err.Error() == "invalid or expired refresh token":
		return status.Error(codes.Unauthenticated, err.Error())
	case err.Error() == "invalid token", err.Error() == "token expired or not valid yet", err.Error() == "malformed token":
		return status.Error(codes.Unauthenticated, err.Error())
	default:
		return status.Error(codes.Internal, "an internal error occurred")
	}
}

// Register handles user registration
func (s *AuthServiceServer) Register(ctx context.Context, req *proto.RegisterRequest) (*proto.RegisterResponse, error) {
	if req.Password != req.ConfirmPassword {
		return nil, status.Error(codes.InvalidArgument, "Passwords do not match")
	}
	if req.Email == "" || req.Username == "" || req.Password == "" || req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "Name, Username, Email, and Password are required")
	}

	dob, err := time.Parse("2006-01-02", req.DateOfBirth)
	if err != nil {
		log.Printf("Warning: Could not parse DateOfBirth '%s': %v. Proceeding without it.", req.DateOfBirth, err)
		dob = time.Time{}
	}

	user := &model.User{
		Name:                   req.Name,
		Username:               req.Username,
		Email:                  req.Email,
		Gender:                 req.Gender,
		DateOfBirth:            dob,
		SecurityQuestion:       req.SecurityQuestion,
		SecurityAnswer:         req.SecurityAnswer,
		NewsletterSubscription: req.SubscribeToNewsletter,
	}

	registeredEmail, err := s.authService.RegisterUser(ctx, user, req.Password, req.RecaptchaToken)
	if err != nil {
		return nil, mapErrorToGrpcStatus(err)
	}

	return &proto.RegisterResponse{
		Success: true,
		Message: "Registration successful. Please check your email for verification code.",
		Email:   registeredEmail,
	}, nil
}

// VerifyEmail handles email verification
func (s *AuthServiceServer) VerifyEmail(ctx context.Context, req *proto.VerifyEmailRequest) (*proto.VerifyEmailResponse, error) {
	if req.Email == "" || req.Code == "" {
		return nil, status.Error(codes.InvalidArgument, "Email and verification code are required")
	}

	userID, err := s.authService.VerifyEmail(ctx, req.Email, req.Code)
	if err != nil {
		return nil, mapErrorToGrpcStatus(err)
	}

	tokens, err := s.authService.GenerateTokens(ctx, userID)
	if err != nil {
		log.Printf("VerifyEmail Error: Failed to generate tokens for user %s after verification: %v", userID, err)
		return nil, status.Error(codes.Internal, "Email verified, but failed to generate session tokens")
	}

	return &proto.VerifyEmailResponse{
		Success:      true,
		Message:      "Email verified successfully",
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		UserId:       tokens.UserID,
		TokenType:    tokens.TokenType,
		ExpiresIn:    tokens.ExpiresIn,
	}, nil
}

// ResendVerificationCode resends the email verification code
func (s *AuthServiceServer) ResendVerificationCode(ctx context.Context, req *proto.ResendVerificationCodeRequest) (*proto.ResendVerificationCodeResponse, error) {
	if req.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "Email is required")
	}

	err := s.authService.ResendVerificationCode(ctx, req.Email)
	if err != nil {
		if err.Error() == "failed to process request" {
			return &proto.ResendVerificationCodeResponse{
				Success: true,
				Message: "If an account with that email exists and is not verified, a new code has been sent.",
			}, nil
		}
		return nil, mapErrorToGrpcStatus(err)
	}

	return &proto.ResendVerificationCodeResponse{
		Success: true,
		Message: "Verification code has been resent. Please check your email.",
	}, nil
}

// Login handles user login
func (s *AuthServiceServer) Login(ctx context.Context, req *proto.LoginRequest) (*proto.LoginResponse, error) {
	if req.Email == "" || req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "Email and password are required")
	}

	tokens, err := s.authService.Login(ctx, req.Email, req.Password)
	if err != nil {
		return nil, mapErrorToGrpcStatus(err)
	}

	return &proto.LoginResponse{
		Success:      true,
		Message:      "Login successful",
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		UserId:       tokens.UserID,
		TokenType:    tokens.TokenType,
		ExpiresIn:    tokens.ExpiresIn,
	}, nil
}

func (s *AuthServiceServer) ValidateToken(ctx context.Context, req *proto.ValidateTokenRequest) (*proto.ValidateTokenResponse, error) {
	if req.Token == "" {
		return nil, status.Error(codes.InvalidArgument, "Token is required")
	}

	claims, err := s.authService.ValidateToken(ctx, req.Token)
	if err != nil {
		return &proto.ValidateTokenResponse{
			Valid:   false,
			Message: err.Error(),
			UserId:  "",
		}, nil
	}

	userID, ok := claims["sub"].(string)
	if !ok {
		log.Printf("ValidateToken Error: Missing or invalid 'sub' (UserID) claim")
		return nil, status.Error(codes.Internal, "Invalid token claims")
	}

	return &proto.ValidateTokenResponse{
		Valid:   true,
		Message: "Token is valid",
		UserId:  userID,
	}, nil
}

// RefreshToken refreshes an access token using a refresh token
func (s *AuthServiceServer) RefreshToken(ctx context.Context, req *proto.RefreshTokenRequest) (*proto.RefreshTokenResponse, error) {
	if req.RefreshToken == "" {
		return nil, status.Error(codes.InvalidArgument, "Refresh token is required")
	}

	newTokens, err := s.authService.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, mapErrorToGrpcStatus(err)
	}

	return &proto.RefreshTokenResponse{
		Success:      true,
		Message:      "Token refreshed successfully",
		AccessToken:  newTokens.AccessToken,
		RefreshToken: newTokens.RefreshToken,
		UserId:       newTokens.UserID,
		TokenType:    newTokens.TokenType,
		ExpiresIn:    newTokens.ExpiresIn,
	}, nil
}

// Logout handles user logout
func (s *AuthServiceServer) Logout(ctx context.Context, req *proto.LogoutRequest) (*proto.LogoutResponse, error) {
	if req.RefreshToken == "" {
		return nil, status.Error(codes.InvalidArgument, "Refresh token is required for logout")
	}

	err := s.authService.Logout(ctx, req.AccessToken, req.RefreshToken)
	if err != nil {
		log.Printf("Logout Error: Service layer returned error: %v", err)
	}

	return &proto.LogoutResponse{
		Success: true,
		Message: "Logged out successfully",
	}, nil
}

// GoogleLogin handles Google OAuth login
func (s *AuthServiceServer) GoogleLogin(ctx context.Context, req *proto.GoogleLoginRequest) (*proto.GoogleLoginResponse, error) {
	if req.IdToken == "" {
		return nil, status.Error(codes.InvalidArgument, "Google ID token is required")
	}

	tokens, err := s.authService.AuthenticateWithGoogle(ctx, req.IdToken)
	if err != nil {
		return nil, mapErrorToGrpcStatus(err)
	}

	return &proto.GoogleLoginResponse{
		Success:      true,
		Message:      "Google login successful",
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		UserId:       tokens.UserID,
		TokenType:    tokens.TokenType,
		ExpiresIn:    tokens.ExpiresIn,
	}, nil
}
