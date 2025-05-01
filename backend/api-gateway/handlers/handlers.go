package handlers

import (
	"context"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"sync"
	"time"

	"github.com/Acad600-Tpa/WEB-MV-242/backend/api-gateway/config"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/golang/groupcache/lru"
	supabase "github.com/supabase-community/storage-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var Config *config.Config

var (
	authConnPool      *ConnectionPool
	userConnPool      *ConnectionPool
	threadConnPool    *ConnectionPool
	connPoolInitOnce  sync.Once
	responseCache     *lru.Cache
	requestRateLimits = make(map[string]*RateLimiter)
	rateLimiterMutex  sync.RWMutex
	supabaseClient    *supabase.Client
	supabaseInitOnce  sync.Once
)

type ConnectionPool struct {
	connections chan *grpc.ClientConn
	serviceAddr string
	maxIdle     int
	maxOpen     int
	timeout     time.Duration
	mu          sync.Mutex
}

type RateLimiter struct {
	tokens         float64
	maxTokens      float64
	tokensPerSec   float64
	lastRefillTime time.Time
	mu             sync.Mutex
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

func NewConnectionPool(serviceAddr string, maxIdle, maxOpen int, timeout time.Duration) *ConnectionPool {
	return &ConnectionPool{
		connections: make(chan *grpc.ClientConn, maxIdle),
		serviceAddr: serviceAddr,
		maxIdle:     maxIdle,
		maxOpen:     maxOpen,
		timeout:     timeout,
	}
}

func InitConnectionPools() {
	connPoolInitOnce.Do(func() {
		authConnPool = NewConnectionPool(Config.GetAuthServiceAddr(), 5, 20, 10*time.Second)
		userConnPool = NewConnectionPool(Config.GetUserServiceAddr(), 5, 20, 10*time.Second)
		if Config.GetThreadServiceAddr() != "" {
			threadConnPool = NewConnectionPool(Config.GetThreadServiceAddr(), 5, 20, 10*time.Second)
		}
		responseCache = lru.New(100)
	})
}

func (p *ConnectionPool) Get() (*grpc.ClientConn, error) {
	select {
	case conn := <-p.connections:
		return conn, nil
	default:
		ctx, cancel := context.WithTimeout(context.Background(), p.timeout)
		defer cancel()

		conn, err := grpc.DialContext(ctx, p.serviceAddr,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithBlock())

		if err != nil {
			return nil, err
		}
		return conn, nil
	}
}

func (p *ConnectionPool) Put(conn *grpc.ClientConn) {
	select {
	case p.connections <- conn:
		// Connection successfully returned to pool
	default:
		// Pool is full, close the connection
		conn.Close()
	}
}

func (p *ConnectionPool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	close(p.connections)
	for conn := range p.connections {
		conn.Close()
	}
}

func NewRateLimiter(maxTokens, tokensPerSec float64) *RateLimiter {
	return &RateLimiter{
		tokens:         maxTokens,
		maxTokens:      maxTokens,
		tokensPerSec:   tokensPerSec,
		lastRefillTime: time.Now(),
	}
}

func (r *RateLimiter) Allow() bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(r.lastRefillTime).Seconds()
	r.lastRefillTime = now

	r.tokens += elapsed * r.tokensPerSec
	if r.tokens > r.maxTokens {
		r.tokens = r.maxTokens
	}

	if r.tokens < 1 {
		return false
	}

	r.tokens--
	return true
}

func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		rateLimiterMutex.RLock()
		limiter, exists := requestRateLimits[ip]
		rateLimiterMutex.RUnlock()

		if !exists {
			rateLimiterMutex.Lock()
			// Create a new RateLimiter and store its pointer
			newLimiter := NewRateLimiter(20, 0.33)
			requestRateLimits[ip] = newLimiter
			limiter = newLimiter
			rateLimiterMutex.Unlock()
		}

		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, ErrorResponse{
				Success: false,
				Message: "Rate limit exceeded. Please try again later.",
				Code:    "RATE_LIMIT_EXCEEDED",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func GetOAuthConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"google_client_id": Config.OAuth.GoogleClientID,
	})
}

func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	})
}

func InitServices() {
	// Initialize connection pools
	connPoolInitOnce.Do(func() {
		authConnPool = NewConnectionPool(Config.GetAuthServiceAddr(), 5, 20, 10*time.Second)
		userConnPool = NewConnectionPool(Config.GetUserServiceAddr(), 5, 20, 10*time.Second)
		if Config.GetThreadServiceAddr() != "" {
			threadConnPool = NewConnectionPool(Config.GetThreadServiceAddr(), 5, 20, 10*time.Second)
		}
		responseCache = lru.New(100)
	})

	// Initialize Supabase client
	supabaseInitOnce.Do(func() {
		if Config != nil && Config.Supabase.URL != "" && Config.Supabase.AnonKey != "" {
			supabaseClient = supabase.NewClient(Config.Supabase.URL, Config.Supabase.AnonKey, nil)
		}
	})
}

type RegisterRequest struct {
	Name                  string `json:"name" binding:"required"`
	Username              string `json:"username" binding:"required"`
	Email                 string `json:"email" binding:"required,email"`
	Password              string `json:"password" binding:"required,min=8"`
	ConfirmPassword       string `json:"confirm_password" binding:"required,eqfield=Password"`
	Gender                string `json:"gender" binding:"required"`
	DateOfBirth           string `json:"date_of_birth" binding:"required"`
	SecurityQuestion      string `json:"securityQuestion" binding:"required"`
	SecurityAnswer        string `json:"securityAnswer" binding:"required"`
	SubscribeToNewsletter bool   `json:"subscribeToNewsletter"`
	RecaptchaToken        string `json:"recaptcha_token" binding:"required"`
	ProfilePictureUrl     string `json:"profile_picture_url,omitempty"`
	BannerUrl             string `json:"banner_url,omitempty"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type VerifyEmailRequest struct {
	Email            string `json:"email" binding:"required,email"`
	VerificationCode string `json:"verification_code" binding:"required"`
}

type ResendCodeRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type LogoutRequest struct {
	AccessToken  string `json:"access_token" binding:"required"`
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type RegisterResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type UpdateUserRequest struct {
	Id                string `json:"id,omitempty"`
	Name              string `json:"name,omitempty"`
	Username          string `json:"username,omitempty"`
	Email             string `json:"email,omitempty"`
	Gender            string `json:"gender,omitempty"`
	DateOfBirth       string `json:"date_of_birth,omitempty"`
	Bio               string `json:"bio,omitempty"`
	Location          string `json:"location,omitempty"`
	Website           string `json:"website,omitempty"`
	ProfilePictureUrl string `json:"profile_picture_url,omitempty"`
	BannerUrl         string `json:"banner_url,omitempty"`
}

// Generic auth and user service response types for HTTP responses
type AuthServiceResponse struct {
	Success      bool   `json:"success"`
	Message      string `json:"message"`
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	UserId       string `json:"user_id,omitempty"`
	TokenType    string `json:"token_type,omitempty"`
	ExpiresIn    int64  `json:"expires_in,omitempty"`
}

// Login handles user authentication
func Login(c *gin.Context) {
	var request LoginRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Message: "Invalid request: " + err.Error(),
			Code:    "INVALID_REQUEST",
		})
		return
	}

	// Forward the request to the auth service through HTTP
	// Simplified implementation for now
	c.JSON(http.StatusOK, AuthServiceResponse{
		Success:      true,
		Message:      "Login successful",
		AccessToken:  "sample-access-token",
		RefreshToken: "sample-refresh-token",
		UserId:       "sample-user-id",
		TokenType:    "Bearer",
		ExpiresIn:    3600,
	})
}

// Register creates a new user account
func Register(c *gin.Context) {
	var request RegisterRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Message: "Invalid request: " + err.Error(),
			Code:    "INVALID_REQUEST",
		})
		return
	}

	// Forward the request to the auth service through HTTP
	// Simplified implementation for now
	c.JSON(http.StatusOK, RegisterResponse{
		Success: true,
		Message: "Registration successful, please verify your email",
	})
}

// RefreshToken issues a new access token using a refresh token
func RefreshToken(c *gin.Context) {
	var request RefreshTokenRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Message: "Invalid request: " + err.Error(),
			Code:    "INVALID_REQUEST",
		})
		return
	}

	// Forward the request to the auth service through HTTP
	// Simplified implementation for now
	c.JSON(http.StatusOK, AuthServiceResponse{
		Success:      true,
		Message:      "Token refreshed successfully",
		AccessToken:  "new-sample-access-token",
		RefreshToken: "new-sample-refresh-token",
		UserId:       "sample-user-id",
		TokenType:    "Bearer",
		ExpiresIn:    3600,
	})
}

// GoogleAuth authenticates a user using Google OAuth
func GoogleAuth(c *gin.Context) {
	// Implementation placeholder - using HTTP auth service
	c.JSON(http.StatusOK, AuthServiceResponse{
		Success:      true,
		Message:      "Google authentication successful",
		AccessToken:  "google-auth-access-token",
		RefreshToken: "google-auth-refresh-token",
		UserId:       "google-user-id",
		TokenType:    "Bearer",
		ExpiresIn:    3600,
	})
}

// VerifyEmail verifies a user's email using a verification code
func VerifyEmail(c *gin.Context) {
	var request VerifyEmailRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Message: "Invalid request: " + err.Error(),
			Code:    "INVALID_REQUEST",
		})
		return
	}

	// Simplified implementation for now
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Email verified successfully",
	})
}

// ResendVerificationCode resends a verification code to the user's email
func ResendVerificationCode(c *gin.Context) {
	var request ResendCodeRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Message: "Invalid request: " + err.Error(),
			Code:    "INVALID_REQUEST",
		})
		return
	}

	// Simplified implementation for now
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Verification code resent successfully",
	})
}

// User profile handlers
func UserProfileHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDAny, exists := c.Get("userId")
		if !exists {
			c.JSON(http.StatusUnauthorized, ErrorResponse{
				Success: false,
				Message: "User ID not found in token",
				Code:    "UNAUTHORIZED",
			})
			return
		}

		// Handle different HTTP methods
		switch c.Request.Method {
		case http.MethodGet:
			// Get user profile
			c.JSON(http.StatusOK, gin.H{
				"id":              userIDAny,
				"username":        "sample_user",
				"name":            "Sample User",
				"email":           "user@example.com",
				"bio":             "Sample bio",
				"location":        "Sample location",
				"website":         "https://example.com",
				"profile_picture": "https://example.com/profile.jpg",
				"banner_url":      "https://example.com/banner.jpg",
				"created_at":      time.Now().AddDate(0, -1, 0).Format(time.RFC3339),
				"updated_at":      time.Now().Format(time.RFC3339),
			})
		case http.MethodPut, http.MethodPatch:
			// Update user profile
			var request UpdateUserRequest
			if err := c.ShouldBindJSON(&request); err != nil {
				c.JSON(http.StatusBadRequest, ErrorResponse{
					Success: false,
					Message: "Invalid request body: " + err.Error(),
					Code:    "INVALID_REQUEST",
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "Profile updated successfully",
				"user_id": userIDAny,
				"user": gin.H{
					"id":              userIDAny,
					"username":        request.Username,
					"name":            request.Name,
					"email":           request.Email,
					"bio":             request.Bio,
					"location":        request.Location,
					"website":         request.Website,
					"profile_picture": request.ProfilePictureUrl,
					"banner_url":      request.BannerUrl,
					"updated_at":      time.Now().Format(time.RFC3339),
				},
			})
		default:
			c.JSON(http.StatusMethodNotAllowed, ErrorResponse{
				Success: false,
				Message: "Method not allowed",
				Code:    "METHOD_NOT_ALLOWED",
			})
		}
	}
}

// Product handler (combines all product-related operations)
func ProductHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Handle different HTTP methods
		switch c.Request.Method {
		case http.MethodGet:
			// Check if we're getting a single product or listing products
			if c.Param("id") != "" {
				// Get single product
				c.JSON(http.StatusOK, gin.H{"message": "get product endpoint", "id": c.Param("id")})
			} else {
				// List products
				c.JSON(http.StatusOK, gin.H{"message": "list products endpoint"})
			}
		case http.MethodPost:
			// Create product
			c.JSON(http.StatusOK, gin.H{"message": "create product endpoint"})
		case http.MethodPut, http.MethodPatch:
			// Update product
			c.JSON(http.StatusOK, gin.H{"message": "update product endpoint", "id": c.Param("id")})
		case http.MethodDelete:
			// Delete product
			c.JSON(http.StatusOK, gin.H{"message": "delete product endpoint", "id": c.Param("id")})
		default:
			c.JSON(http.StatusMethodNotAllowed, ErrorResponse{
				Success: false,
				Message: "Method not allowed",
				Code:    "METHOD_NOT_ALLOWED",
			})
		}
	}
}

// Payment handler (combines all payment-related operations)
func PaymentHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Handle different HTTP methods
		switch c.Request.Method {
		case http.MethodGet:
			// Check if we're getting a single payment or payment history
			if c.Param("id") != "" {
				// Get single payment
				c.JSON(http.StatusOK, gin.H{"message": "get payment endpoint", "id": c.Param("id")})
			} else {
				// Get payment history
				c.JSON(http.StatusOK, gin.H{"message": "get payment history endpoint"})
			}
		case http.MethodPost:
			// Create payment
			c.JSON(http.StatusOK, gin.H{"message": "create payment endpoint"})
		default:
			c.JSON(http.StatusMethodNotAllowed, ErrorResponse{
				Success: false,
				Message: "Method not allowed",
				Code:    "METHOD_NOT_ALLOWED",
			})
		}
	}
}

func uploadToSupabase(fileHeader *multipart.FileHeader, bucketName string, destinationPath string) (string, error) {
	if fileHeader == nil {
		return "", nil
	}

	file, err := fileHeader.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer file.Close()

	contentType := fileHeader.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	upsert := false
	fileOptions := supabase.FileOptions{
		ContentType: &contentType,
		Upsert:      &upsert,
	}

	_, err = supabaseClient.UploadFile(bucketName, destinationPath, file, fileOptions)

	if err != nil {
		return "", fmt.Errorf("failed to upload to supabase: %w", err)
	}

	publicURL := supabaseClient.GetPublicUrl(bucketName, destinationPath)

	return publicURL.SignedURL, nil
}

func RegisterWithMedia(c *gin.Context) {
	if supabaseClient == nil {
		InitServices()
	}
	if supabaseClient == nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Success: false, Message: "Supabase client not initialized", Code: "CONFIG_ERROR"})
		return
	}

	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Success: false, Message: "Invalid form data: " + err.Error(), Code: "INVALID_REQUEST"})
		return
	}

	values := form.Value
	// Parse form values (using them in a simplified implementation)
	password := values["password"][0]
	confirmPassword := values["confirm_password"][0]

	if password != confirmPassword {
		c.JSON(http.StatusBadRequest, ErrorResponse{Success: false, Message: "Passwords do not match", Code: "VALIDATION_ERROR"})
		return
	}

	profilePicFileHeader, _ := c.FormFile("profile_picture")
	bannerFileHeader, _ := c.FormFile("banner_image")

	uuidVal, _ := uuid.NewV4()
	userPathPrefix := uuidVal.String()

	// Upload files - we'd use these URLs in a real implementation
	if profilePicFileHeader != nil {
		_, err := uploadToSupabase(profilePicFileHeader, "profile-pictures", userPathPrefix+"/"+filepath.Base(profilePicFileHeader.Filename))
		if err != nil {
			log.Printf("Failed to upload profile picture: %v", err)
		}
	}

	if bannerFileHeader != nil {
		_, err := uploadToSupabase(bannerFileHeader, "banner-images", userPathPrefix+"/"+filepath.Base(bannerFileHeader.Filename))
		if err != nil {
			log.Printf("Failed to upload banner image: %v", err)
		}
	}

	// Simplified implementation
	c.JSON(http.StatusOK, RegisterResponse{
		Success: true,
		Message: "Registration with media successful",
	})
}

func Logout(c *gin.Context) {
	var request LogoutRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Message: "Invalid request: " + err.Error(),
			Code:    "INVALID_REQUEST",
		})
		return
	}

	// Simplified implementation
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Logout successful",
	})
}
