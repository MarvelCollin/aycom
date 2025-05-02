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

// HealthCheck godoc
// @Summary Check API health
// @Description Returns health status of the API
// @Tags health
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /health [get]
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

// Login godoc
// @Summary User login
// @Description Authenticates a user and returns access token
// @Tags auth
// @Accept  json
// @Produce  json
// @Param   login body LoginRequest true "Login Credentials"
// @Success 200 {object} AuthServiceResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/auth/login [post]
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

// Register godoc
// @Summary Register a new user
// @Description Create a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param register body RegisterRequest true "Registration Information"
// @Success 200 {object} RegisterResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/auth/register [post]
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

// RefreshToken godoc
// @Summary Refresh access token
// @Description Issues a new access token using a refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param refreshToken body RefreshTokenRequest true "Refresh Token"
// @Success 200 {object} AuthServiceResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/auth/refresh-token [post]
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

// GoogleAuth godoc
// @Summary Authenticate with Google
// @Description Authenticates a user using Google OAuth
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} AuthServiceResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/auth/google [post]
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

// VerifyEmail godoc
// @Summary Verify email address
// @Description Verifies a user's email using a verification code
// @Tags auth
// @Accept json
// @Produce json
// @Param verification body VerifyEmailRequest true "Email Verification Information"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/auth/verify-email [post]
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

// ResendVerificationCode godoc
// @Summary Resend verification code
// @Description Resends a verification code to the user's email
// @Tags auth
// @Accept json
// @Produce json
// @Param resendRequest body ResendCodeRequest true "Email to resend verification code"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/auth/resend-verification [post]
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

// RegisterWithMedia godoc
// @Summary Register with media files
// @Description Register a new user with profile picture and banner image
// @Tags auth
// @Accept multipart/form-data
// @Produce json
// @Param profile_picture formData file false "Profile picture"
// @Param banner_image formData file false "Banner image"
// @Param name formData string true "Full name"
// @Param username formData string true "Username"
// @Param email formData string true "Email address"
// @Param password formData string true "Password"
// @Param confirm_password formData string true "Confirm password"
// @Param gender formData string true "Gender"
// @Param date_of_birth formData string true "Date of birth"
// @Param security_question formData string true "Security question"
// @Param security_answer formData string true "Security answer"
// @Param subscribe_to_newsletter formData boolean false "Subscribe to newsletter"
// @Param recaptcha_token formData string true "reCAPTCHA token"
// @Success 200 {object} RegisterResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/auth/register-with-media [post]
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

// Logout godoc
// @Summary Logout user
// @Description Invalidate user tokens and logout
// @Tags auth
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param logout body LogoutRequest true "Logout Request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/auth/logout [post]
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

// GetOAuthConfig godoc
// @Summary Get OAuth configuration
// @Description Returns OAuth client IDs and configuration
// @Tags auth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/auth/oauth-config [get]
func GetOAuthConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"google_client_id": Config.OAuth.GoogleClientID,
	})
}

// The following are Swagger documentation comments for handlers defined
// in other files. The actual implementations are in their respective files.

// GetUserProfile godoc
// @Summary Get user profile
// @Description Get the profile of the authenticated user
// @Tags users
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/users/profile [get]

// UpdateUserProfile godoc
// @Summary Update user profile
// @Description Update the profile of the authenticated user
// @Tags users
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param profileUpdate body UpdateUserRequest true "User profile data to update"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/users/profile [put]

// CreateThread godoc
// @Summary Create a new thread
// @Description Create a new thread/post
// @Tags threads
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param thread body map[string]interface{} true "Thread data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/threads [post]

// GetThread godoc
// @Summary Get thread by ID
// @Description Get a specific thread by its ID
// @Tags threads
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Thread ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/threads/{id} [get]

// GetThreadsByUser godoc
// @Summary Get threads by user
// @Description Get all threads created by a specific user
// @Tags threads
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "User ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/threads/user/{id} [get]

// UpdateThread godoc
// @Summary Update thread
// @Description Update an existing thread
// @Tags threads
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Thread ID"
// @Param thread body map[string]interface{} true "Thread data to update"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/threads/{id} [put]

// DeleteThread godoc
// @Summary Delete thread
// @Description Delete an existing thread
// @Tags threads
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Thread ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/threads/{id} [delete]

// UploadThreadMedia godoc
// @Summary Upload thread media
// @Description Upload media files for a thread
// @Tags threads
// @Accept multipart/form-data
// @Produce json
// @Security ApiKeyAuth
// @Param file formData file true "Media file"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/threads/media [post]

// ListProducts godoc
// @Summary List products
// @Description Get a list of all products
// @Tags products
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/products [get]

// GetProduct godoc
// @Summary Get product by ID
// @Description Get a specific product by its ID
// @Tags products
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Product ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/products/{id} [get]

// CreateProduct godoc
// @Summary Create a new product
// @Description Create a new product
// @Tags products
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param product body map[string]interface{} true "Product data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/products [post]

// UpdateProduct godoc
// @Summary Update product
// @Description Update an existing product
// @Tags products
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Product ID"
// @Param product body map[string]interface{} true "Product data to update"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/products/{id} [put]

// DeleteProduct godoc
// @Summary Delete product
// @Description Delete an existing product
// @Tags products
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Product ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/products/{id} [delete]

// CreatePayment godoc
// @Summary Create a new payment
// @Description Create a new payment
// @Tags payments
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param payment body map[string]interface{} true "Payment data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/payments [post]

// GetPayment godoc
// @Summary Get payment by ID
// @Description Get a specific payment by its ID
// @Tags payments
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path string true "Payment ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/payments/{id} [get]

// GetPaymentHistory godoc
// @Summary Get payment history
// @Description Get the payment history for the authenticated user
// @Tags payments
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/payments/history [get]
