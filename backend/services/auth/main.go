package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	_ "github.com/lib/pq"

	pb "aycom/backend/services/auth/proto"

	"golang.org/x/net/context"
)

type authServer struct {
	pb.UnimplementedAuthServiceServer
	db *sql.DB
}

func main() {
	log.Println("Starting Auth Service...")

	dbHost := getEnv("DATABASE_HOST", "localhost")
	dbPort := getEnv("DATABASE_PORT", "5432")
	dbUser := getEnv("DATABASE_USER", "kolin")
	dbPassword := getEnv("DATABASE_PASSWORD", "kolin")
	dbName := getEnv("DATABASE_NAME", "auth_db")

	dbURI := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	var db *sql.DB
	var err error
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		log.Printf("Attempting to connect to database (attempt %d/%d)...", i+1, maxRetries)
		db, err = sql.Open("postgres", dbURI)
		if err == nil {
			if err = db.Ping(); err == nil {
				log.Println("Successfully connected to the database")
				break
			}
			log.Printf("Failed to ping database: %v", err)
			db.Close()
		} else {
			log.Printf("Failed to open database connection: %v", err)
		}

		if i < maxRetries-1 {
			log.Println("Retrying in 5 seconds...")
			time.Sleep(5 * time.Second)
		}
	}

	if err != nil {
		log.Fatalf("Failed to connect to database after %d attempts: %v", maxRetries, err)
	}
	defer db.Close()

	port := getEnv("AUTH_SERVICE_PORT", "9090")
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	server := grpc.NewServer()
	pb.RegisterAuthServiceServer(server, &authServer{db: db})
	log.Printf("Auth Service listening on port %s", port)
	if err := server.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

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
	}, nil
}

func (s *authServer) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.RefreshTokenResponse, error) {
	return &pb.RefreshTokenResponse{
		Success: true,
		Message: "Token refreshed",
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
		Success: true,
		Message: "Logged in with Google",
	}, nil
}
