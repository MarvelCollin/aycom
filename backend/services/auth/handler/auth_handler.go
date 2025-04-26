package handler

import (
	"context"
	"fmt"

	"github.com/Acad600-Tpa/WEB-MV-242/services/auth/proto"
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

// Register handles user registration and token generation
func (s *AuthServiceServer) Register(ctx context.Context, req *proto.RegisterRequest) (*proto.TokenResponse, error) {
	// In a real implementation, this would create a user in user service
	// For now, we'll just simulate successful registration with a fake user ID
	userID := "user-456" // This would come from user service registration

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
