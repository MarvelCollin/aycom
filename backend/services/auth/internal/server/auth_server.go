package server

import (
	"context"
	"time"

	"github.com/Acad600-TPA/WEB-MV-242/auth/internal/models"
	"github.com/Acad600-TPA/WEB-MV-242/auth/internal/repository"
	pb "github.com/Acad600-TPA/WEB-MV-242/auth/proto"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthServer struct {
	pb.UnimplementedAuthServiceServer
	userRepo repository.UserRepository
	jwtKey   []byte
}

func NewAuthServer(userRepo repository.UserRepository, jwtKey string) *AuthServer {
	return &AuthServer{
		userRepo: userRepo,
		jwtKey:   []byte(jwtKey),
	}
}

func (s *AuthServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	// Find user by email
	user, err := s.userRepo.FindByEmail(ctx, req.GetEmail())
	if err != nil {
		return &pb.LoginResponse{
			Success: false,
			Message: "Invalid credentials",
		}, nil
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.GetPassword())); err != nil {
		return &pb.LoginResponse{
			Success: false,
			Message: "Invalid credentials",
		}, nil
	}

	// Generate JWT token
	token, err := s.generateToken(user)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to generate token: %v", err)
	}

	return &pb.LoginResponse{
		Success: true,
		Token:   token,
		User: &pb.User{
			Id:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			Role:     user.Role,
		},
	}, nil
}

func (s *AuthServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	// Check if user with email already exists
	existing, _ := s.userRepo.FindByEmail(ctx, req.GetEmail())
	if existing != nil {
		return &pb.RegisterResponse{
			Success: false,
			Message: "Email already in use",
		}, nil
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.GetPassword()), bcrypt.DefaultCost)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to hash password: %v", err)
	}

	// Create user
	user := &models.User{
		Username: req.GetUsername(),
		Email:    req.GetEmail(),
		Password: string(hashedPassword),
		Role:     "user", // Default role
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to create user: %v", err)
	}

	return &pb.RegisterResponse{
		Success: true,
		Message: "User registered successfully",
	}, nil
}

func (s *AuthServer) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	// Parse and validate token
	token, err := jwt.Parse(req.GetToken(), func(token *jwt.Token) (interface{}, error) {
		return s.jwtKey, nil
	})

	if err != nil || !token.Valid {
		return &pb.ValidateTokenResponse{
			Valid:   false,
			Message: "Invalid token",
		}, nil
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return &pb.ValidateTokenResponse{
			Valid:   false,
			Message: "Invalid token claims",
		}, nil
	}

	// Get user ID from claims
	userID, ok := claims["sub"].(string)
	if !ok {
		return &pb.ValidateTokenResponse{
			Valid:   false,
			Message: "Invalid user ID in token",
		}, nil
	}

	// Get user from repository
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to find user: %v", err)
	}

	return &pb.ValidateTokenResponse{
		Valid: true,
		User: &pb.User{
			Id:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			Role:     user.Role,
		},
	}, nil
}

func (s *AuthServer) generateToken(user *models.User) (string, error) {
	// Set expiration time - 24 hours from now
	expirationTime := time.Now().Add(24 * time.Hour)

	// Create claims
	claims := jwt.MapClaims{
		"sub":   user.ID,               // Subject (user ID)
		"name":  user.Username,         // Username
		"email": user.Email,            // Email
		"role":  user.Role,             // Role
		"exp":   expirationTime.Unix(), // Expiration time
		"iat":   time.Now().Unix(),     // Issued at
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token with secret key
	tokenString, err := token.SignedString(s.jwtKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
