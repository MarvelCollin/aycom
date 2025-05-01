package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

// LoginRequest defines the login request structure
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse defines the login response structure
type LoginResponse struct {
	Success      bool   `json:"success"`
	Message      string `json:"message"`
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	UserID       string `json:"user_id,omitempty"`
	TokenType    string `json:"token_type,omitempty"`
	ExpiresIn    int64  `json:"expires_in,omitempty"`
}

// ErrorResponse defines a standard error response
type ErrorResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

func main() {
	log.Println("Starting Auth Service...")

	// Get database connection parameters from environment variables
	dbHost := getEnv("DATABASE_HOST", "localhost")
	dbPort := getEnv("DATABASE_PORT", "5432")
	dbUser := getEnv("DATABASE_USER", "kolin")
	dbPassword := getEnv("DATABASE_PASSWORD", "kolin")
	dbName := getEnv("DATABASE_NAME", "auth_db")

	// Create database connection string
	dbURI := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	// Connect to the database with retry logic
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

	// Set up HTTP handlers
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	http.HandleFunc("/auth/login", func(w http.ResponseWriter, r *http.Request) {
		handleLogin(w, r, db)
	})

	// Get the port to listen on
	port := getEnv("AUTH_SERVICE_PORT", "9090")
	log.Printf("Auth Service listening on port %s", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func handleLogin(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Only allow POST requests
	if r.Method != http.MethodPost {
		errorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var request LoginRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		errorResponse(w, "Invalid request: "+err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Login request received for email: %s", request.Email)

	// Find user by email
	var id uuid.UUID
	var username, passwordHash, passwordSalt string
	var isActivated, isBanned, isDeactivated bool

	// Query the database
	query := `SELECT id, username, password_hash, password_salt, is_activated, is_banned, is_deactivated 
              FROM users WHERE email = $1`
	err = db.QueryRow(query, request.Email).Scan(
		&id, &username, &passwordHash, &passwordSalt, &isActivated, &isBanned, &isDeactivated,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("User not found: %s", request.Email)
			errorResponse(w, "Invalid email or password", http.StatusUnauthorized)
			return
		}
		log.Printf("Database error: %v", err)
		errorResponse(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Check if account is active
	if !isActivated {
		errorResponse(w, "Account not activated. Please verify your email.", http.StatusForbidden)
		return
	}

	// Check if account is banned
	if isBanned {
		errorResponse(w, "Account is banned. Please contact support.", http.StatusForbidden)
		return
	}

	// Check if account is deactivated
	if isDeactivated {
		errorResponse(w, "Account is deactivated. Please reactivate your account.", http.StatusForbidden)
		return
	}

	// Check password (for now, directly compare since we're using plain text in the DB for development)
	if passwordHash != request.Password {
		log.Println("Password mismatch")
		errorResponse(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	// Generate tokens
	accessToken, refreshToken, expiresIn, err := generateTokens(id.String())
	if err != nil {
		log.Printf("Failed to generate tokens: %v", err)
		errorResponse(w, "Failed to generate authentication tokens", http.StatusInternalServerError)
		return
	}

	// Store session
	sessionID, err := uuid.NewRandom()
	if err != nil {
		log.Printf("Failed to generate session ID: %v", err)
		errorResponse(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	_, err = db.Exec(`
        INSERT INTO sessions (id, user_id, access_token, refresh_token, expires_at)
        VALUES ($1, $2, $3, $4, $5)
    `, sessionID, id, accessToken, refreshToken, time.Now().Add(time.Second*time.Duration(expiresIn)))

	if err != nil {
		log.Printf("Failed to store session: %v", err)
		errorResponse(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	// Update last login time
	_, err = db.Exec(`
        UPDATE users SET last_login_at = $1 WHERE id = $2
    `, time.Now(), id)

	if err != nil {
		log.Printf("Failed to update last login time: %v", err)
		// Non-critical error, continue
	}

	log.Printf("Login successful for user: %s", username)

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(LoginResponse{
		Success:      true,
		Message:      "Login successful",
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		UserID:       id.String(),
		TokenType:    "Bearer",
		ExpiresIn:    expiresIn,
	})
}

func errorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{
		Success: false,
		Message: message,
	})
}

func generateTokens(userID string) (string, string, int64, error) {
	// Get JWT secret from environment or use default
	jwtSecret := getEnv("JWT_SECRET", "wompwomp123")

	// Set token expiration times
	accessExpiry := time.Now().Add(time.Hour)           // 1 hour
	refreshExpiry := time.Now().Add(7 * 24 * time.Hour) // 7 days

	// Create access token
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

	// Create refresh token
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

	// Calculate expires_in seconds
	expiresIn := accessExpiry.Unix() - time.Now().Unix()

	return accessTokenString, refreshTokenString, expiresIn, nil
}
