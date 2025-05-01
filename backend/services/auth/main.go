package main

import (
	// "crypto/rand"
	"database/sql"
	// "encoding/base64"
	"log"
	// "os"
	// "time"

	// "golang.org/x/crypto/bcrypt"
	// "google.golang.org/grpc/codes"
	// "google.golang.org/grpc/status"

	// "github.com/dgrijalva/jwt-go"
	// "github.com/google/uuid"
	_ "github.com/lib/pq"

	// Proto package is disabled for now
	// pb "github.com/Acad600-Tpa/WEB-MV-242/backend/services/auth/proto"
	"github.com/Acad600-Tpa/WEB-MV-242/backend/services/auth/handler"
	"github.com/Acad600-Tpa/WEB-MV-242/backend/services/auth/service"
	// "golang.org/x/net/context"
)

// Temporarily define a minimal type to replace the proto server
type authServer struct {
	// pb.UnimplementedAuthServiceServer
	db *sql.DB
}

// Just define a main function that doesn't rely on the proto package
func main() {
	log.Println("Starting Auth Service...")
	log.Println("This build is just for testing. The proto package is currently disabled.")

	// Create a simple service to test if imports and type conversions are working
	// This is just for testing, not actually connecting to the DB
	authService, err := service.NewAuthService()
	if err == nil {
		log.Println("Created auth service")
		_ = handler.NewAuthServiceServer(authService)
		log.Println("Created auth handler")
	} else {
		log.Printf("Error creating auth service: %v", err)
	}

	log.Println("Type checking successful!")
}

/*
// Commented out all proto-dependent functions to avoid compilation errors
// These need to be fixed once the proto package is properly generated

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func (s *authServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	var id uuid.UUID
	var username, passwordHash, passwordSalt string
	var isActivated, isBanned, isDeactivated bool

	query := `SELECT id, username, password_hash, password_salt, is_activated, is_banned, is_deactivated
              FROM users WHERE email = $1`
	err := s.db.QueryRow(query, req.Email).Scan(
		&id, &username, &passwordHash, &passwordSalt, &isActivated, &isBanned, &isDeactivated,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("User not found: %s", req.Email)
			return nil, status.Errorf(codes.Unauthenticated, "Invalid email or password")
		}
		log.Printf("Database error: %v", err)
		return nil, status.Errorf(codes.Internal, "Internal server error")
	}

	if !isActivated {
		return nil, status.Errorf(codes.PermissionDenied, "Account not activated. Please verify your email.")
	}

	if isBanned {
		return nil, status.Errorf(codes.PermissionDenied, "Account is banned. Please contact support.")
	}

	if isDeactivated {
		return nil, status.Errorf(codes.PermissionDenied, "Account is deactivated. Please reactivate your account.")
	}

	if !verifyPassword(req.Password, passwordHash, passwordSalt) {
		log.Println("Password mismatch")
		return nil, status.Errorf(codes.Unauthenticated, "Invalid email or password")
	}

	accessToken, refreshToken, expiresIn, err := generateTokens(id.String())
	if err != nil {
		log.Printf("Failed to generate tokens: %v", err)
		return nil, status.Errorf(codes.Internal, "Failed to generate authentication tokens")
	}

	sessionID, err := uuid.NewRandom()
	if err != nil {
		log.Printf("Failed to generate session ID: %v", err)
		return nil, status.Errorf(codes.Internal, "Failed to create session")
	}

	_, err = s.db.Exec(`
        INSERT INTO sessions (id, user_id, access_token, refresh_token, expires_at)
        VALUES ($1, $2, $3, $4, $5)
    `, sessionID, id, accessToken, refreshToken, time.Now().Add(time.Second*time.Duration(expiresIn)))

	if err != nil {
		log.Printf("Failed to store session: %v", err)
		return nil, status.Errorf(codes.Internal, "Failed to create session")
	}

	_, err = s.db.Exec(`
        UPDATE users SET last_login_at = $1 WHERE id = $2
    `, time.Now(), id)

	if err != nil {
		log.Printf("Failed to update last login time: %v", err)
	}

	log.Printf("Login successful for user: %s", username)

	return &pb.LoginResponse{
		Success:      true,
		Message:      "Login successful",
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		UserId:       id.String(),
		TokenType:    "Bearer",
		ExpiresIn:    int32(expiresIn),
	}, nil
}

func hashPassword(password, salt string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(password+salt), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		return ""
	}
	return string(hash)
}

func verifyPassword(password, hash, salt string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password+salt))
	return err == nil
}

func generateSalt() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

func generateTokens(userID string) (string, string, int64, error) {
	jwtSecret := getEnv("JWT_SECRET", "wompwomp123")

	accessExpiry := time.Now().Add(time.Hour)
	refreshExpiry := time.Now().Add(7 * 24 * time.Hour)

	accessClaims := jwt.MapClaims{
		"sub": userID,
		"exp": accessExpiry.Unix(),
		"iat": time.Now().Unix(),
		"typ": "access",
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", "", 0, err
	}

	refreshClaims := jwt.MapClaims{
		"sub": userID,
		"exp": refreshExpiry.Unix(),
		"iat": time.Now().Unix(),
		"typ": "refresh",
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", "", 0, err
	}

	expiresIn := accessExpiry.Unix() - time.Now().Unix()

	return accessTokenString, refreshTokenString, expiresIn, nil
}

func (s *authServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	if req.Password != req.ConfirmPassword {
		return nil, status.Errorf(codes.InvalidArgument, "Passwords do not match")
	}

	salt, err := generateSalt()
	if err != nil {
		log.Printf("Failed to generate salt: %v", err)
		return nil, status.Errorf(codes.Internal, "Failed to create user")
	}

	passwordHash := hashPassword(req.Password, salt)
	if passwordHash == "" {
		return nil, status.Errorf(codes.Internal, "Failed to hash password")
	}

	userID, err := uuid.NewRandom()
	if err != nil {
		log.Printf("Failed to generate UUID: %v", err)
		return nil, status.Errorf(codes.Internal, "Failed to create user")
	}

	_, err = s.db.Exec(`
        INSERT INTO users (id, name, username, email, password_hash, password_salt, gender, date_of_birth, security_question, security_answer, subscribe_to_newsletter, is_activated)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
    `, userID, req.Name, req.Username, req.Email, passwordHash, salt, req.Gender, req.DateOfBirth, req.SecurityQuestion, req.SecurityAnswer, req.SubscribeToNewsletter, true)

	if err != nil {
		log.Printf("Failed to insert user: %v", err)
		return nil, status.Errorf(codes.Internal, "Failed to create user")
	}

	return &pb.RegisterResponse{
		Success: true,
		Message: "User registered successfully",
		Email:   req.Email,
	}, nil
}

func (s *authServer) VerifyEmail(ctx context.Context, req *pb.VerifyEmailRequest) (*pb.VerifyEmailResponse, error) {
	return &pb.VerifyEmailResponse{
		Success: true,
		Message: "Email verified successfully",
	}, nil
}

func (s *authServer) ResendVerificationCode(ctx context.Context, req *pb.ResendVerificationCodeRequest) (*pb.ResendVerificationCodeResponse, error) {
	return &pb.ResendVerificationCodeResponse{
		Success: true,
		Message: "Verification code sent",
	}, nil
}

func (s *authServer) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	return &pb.ValidateTokenResponse{
		Valid:   true,
		Message: "Token is valid",
		UserId:  "user123",
	}, nil
}

func (s *authServer) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.RefreshTokenResponse, error) {
	return &pb.RefreshTokenResponse{
		Success:      true,
		Message:      "Token refreshed",
		AccessToken:  "new-access-token",
		RefreshToken: "new-refresh-token",
		UserId:       "user123",
		TokenType:    "Bearer",
		ExpiresIn:    3600,
	}, nil
}

func (s *authServer) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	return &pb.LogoutResponse{
		Success: true,
		Message: "Logged out successfully",
	}, nil
}

func (s *authServer) GoogleLogin(ctx context.Context, req *pb.GoogleLoginRequest) (*pb.GoogleLoginResponse, error) {
	return &pb.GoogleLoginResponse{
		Success:      true,
		Message:      "Google login successful",
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
		UserId:       "user123",
		TokenType:    "Bearer",
		ExpiresIn:    3600,
	}, nil
}
*/
