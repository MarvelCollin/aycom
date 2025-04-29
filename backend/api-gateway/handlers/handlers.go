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
	authProto "github.com/Acad600-Tpa/WEB-MV-242/backend/services/auth/proto"
	userProto "github.com/Acad600-Tpa/WEB-MV-242/backend/services/user/proto"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/golang/groupcache/lru"
	supabase "github.com/supabase-community/storage-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

// Global config for the handlers
var Config *config.Config

// gRPC connection pool
var (
	authConnPool      *ConnectionPool
	userConnPool      *ConnectionPool
	connPoolInitOnce  sync.Once
	responseCache     *lru.Cache
	requestRateLimits = make(map[string]RateLimiter)
	rateLimiterMutex  sync.RWMutex
	supabaseClient    *supabase.Client
	supabaseInitOnce  sync.Once
)

// ConnectionPool manages a pool of gRPC connections
type ConnectionPool struct {
	connections chan *grpc.ClientConn
	serviceAddr string
	maxIdle     int
	maxOpen     int
	timeout     time.Duration
	mu          sync.Mutex
}

// RateLimiter implements a simple token bucket rate limiter
type RateLimiter struct {
	tokens         float64
	maxTokens      float64
	tokensPerSec   float64
	lastRefillTime time.Time
	mu             sync.Mutex
}

// NewConnectionPool creates a new connection pool
func NewConnectionPool(serviceAddr string, maxIdle, maxOpen int, timeout time.Duration) *ConnectionPool {
	return &ConnectionPool{
		connections: make(chan *grpc.ClientConn, maxIdle),
		serviceAddr: serviceAddr,
		maxIdle:     maxIdle,
		maxOpen:     maxOpen,
		timeout:     timeout,
	}
}

// Initialize the connection pools
func InitConnectionPools() {
	connPoolInitOnce.Do(func() {
		authConnPool = NewConnectionPool(Config.GetAuthServiceAddr(), 5, 20, 10*time.Second)
		userConnPool = NewConnectionPool(Config.GetUserServiceAddr(), 5, 20, 10*time.Second)
		responseCache = lru.New(100) // Cache size of 100 items
	})
}

// Get returns a connection from the pool or creates a new one
func (p *ConnectionPool) Get() (*grpc.ClientConn, error) {
	select {
	case conn := <-p.connections:
		return conn, nil
	default:
		// Create a new connection with timeout
		ctx, cancel := context.WithTimeout(context.Background(), p.timeout)
		defer cancel()

		// In production, use TLS credentials instead of insecure
		conn, err := grpc.DialContext(ctx, p.serviceAddr,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithBlock())

		if err != nil {
			return nil, err
		}
		return conn, nil
	}
}

// Put returns a connection to the pool
func (p *ConnectionPool) Put(conn *grpc.ClientConn) {
	select {
	case p.connections <- conn:
		// Connection returned to pool
	default:
		// Pool is full, close the connection
		conn.Close()
	}
}

// Close closes all connections in the pool
func (p *ConnectionPool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	close(p.connections)
	for conn := range p.connections {
		conn.Close()
	}
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(maxTokens, tokensPerSec float64) RateLimiter {
	return RateLimiter{
		tokens:         maxTokens,
		maxTokens:      maxTokens,
		tokensPerSec:   tokensPerSec,
		lastRefillTime: time.Now(),
	}
}

// Allow checks if a request should be allowed
func (r *RateLimiter) Allow() bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(r.lastRefillTime).Seconds()
	r.lastRefillTime = now

	// Refill tokens based on elapsed time
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

// RegisterRequest represents the user registration payload
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
}

// LoginRequest represents the login payload
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// VerifyEmailRequest represents the request for verifying email
type VerifyEmailRequest struct {
	Email            string `json:"email" binding:"required,email"`
	VerificationCode string `json:"verification_code" binding:"required"`
}

// ResendCodeRequest represents the request for resending verification code
type ResendCodeRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// LogoutRequest represents the request for logging out
type LogoutRequest struct {
	AccessToken  string `json:"access_token" binding:"required"`
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// RefreshTokenRequest represents the request for refreshing token
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// RegisterResponse represents the response from user registration
type RegisterResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

// Initialize connection pools when package is loaded
func init() {
	InitConnectionPools()
	supabaseInitOnce.Do(func() {
		supabaseClient = supabase.NewClient(Config.Supabase.URL, Config.Supabase.AnonKey, nil)
	})
}

// RateLimitMiddleware limits the number of requests from a single IP
func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		rateLimiterMutex.RLock()
		limiter, exists := requestRateLimits[ip]
		rateLimiterMutex.RUnlock()

		if !exists {
			rateLimiterMutex.Lock()
			// Create a new rate limiter allowing 20 requests per minute
			limiter = NewRateLimiter(20, 0.33)
			requestRateLimits[ip] = limiter
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

// GetOAuthConfig godoc
// @Summary Get OAuth configuration
// @Description Get OAuth configuration details for the client
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "OAuth configuration"
// @Router /api/v1/auth/oauth-config [get]
func GetOAuthConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"google_client_id": Config.OAuth.GoogleClientID,
	})
}

func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

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

	// Get connection from pool
	conn, err := authConnPool.Get()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Failed to connect to auth service: " + err.Error(),
			Code:    "SERVICE_UNAVAILABLE",
		})
		return
	}
	defer authConnPool.Put(conn)

	client := authProto.NewAuthServiceClient(conn)

	// Set timeout for the request
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	req := &authProto.LoginRequest{
		Email:    request.Email,
		Password: request.Password,
	}

	resp, err := client.Login(ctx, req)
	if err != nil {
		// Properly handle gRPC status errors
		if st, ok := status.FromError(err); ok {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Success: false,
				Message: st.Message(),
				Code:    st.Code().String(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Success: false,
				Message: "Login failed: " + err.Error(),
				Code:    "INTERNAL_ERROR",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":       resp.Success,
		"message":       resp.Message,
		"access_token":  resp.AccessToken,
		"refresh_token": resp.RefreshToken,
		"user_id":       resp.UserId,
		"token_type":    resp.TokenType,
		"expires_in":    resp.ExpiresIn,
	})
}

// Register godoc
// @Summary Register a new user
// @Description Register a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param registration body RegisterRequest true "User registration data"
// @Success 200 {object} RegisterResponse "registration successful"
// @Failure 400 {object} ErrorResponse "invalid input"
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

	// Get connection from pool
	conn, err := authConnPool.Get()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Failed to connect to auth service: " + err.Error(),
			Code:    "SERVICE_UNAVAILABLE",
		})
		return
	}
	defer authConnPool.Put(conn)

	client := authProto.NewAuthServiceClient(conn)

	// Set timeout for the request
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	req := &authProto.RegisterRequest{
		Name:                  request.Name,
		Username:              request.Username,
		Email:                 request.Email,
		Password:              request.Password,
		ConfirmPassword:       request.ConfirmPassword,
		Gender:                request.Gender,
		DateOfBirth:           request.DateOfBirth,
		SecurityQuestion:      request.SecurityQuestion,
		SecurityAnswer:        request.SecurityAnswer,
		SubscribeToNewsletter: request.SubscribeToNewsletter,
		RecaptchaToken:        request.RecaptchaToken,
	}

	resp, err := client.Register(ctx, req)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Success: false,
				Message: st.Message(),
				Code:    st.Code().String(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Success: false,
				Message: "Registration failed: " + err.Error(),
				Code:    "INTERNAL_ERROR",
			})
		}
		return
	}

	c.JSON(http.StatusOK, RegisterResponse{
		Success: resp.Success,
		Message: resp.Message,
	})
}

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

	// Get connection from pool
	conn, err := authConnPool.Get()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Failed to connect to auth service: " + err.Error(),
			Code:    "SERVICE_UNAVAILABLE",
		})
		return
	}
	defer authConnPool.Put(conn)

	client := authProto.NewAuthServiceClient(conn)

	// Set timeout for the request
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	req := &authProto.RefreshTokenRequest{
		RefreshToken: request.RefreshToken,
	}

	resp, err := client.RefreshToken(ctx, req)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Success: false,
				Message: st.Message(),
				Code:    st.Code().String(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Success: false,
				Message: "Token refresh failed: " + err.Error(),
				Code:    "INTERNAL_ERROR",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":       resp.Success,
		"message":       resp.Message,
		"access_token":  resp.AccessToken,
		"refresh_token": resp.RefreshToken,
		"user_id":       resp.UserId,
		"token_type":    resp.TokenType,
		"expires_in":    resp.ExpiresIn,
	})
}

// GoogleAuth godoc
// @Summary Authenticate with Google
// @Description Use Google OAuth token to authenticate
// @Tags auth
// @Accept json
// @Produce json
// @Param token body map[string]interface{} true "Google token ID"
// @Success 200 {object} map[string]interface{} "tokens and user info"
// @Failure 400 {object} map[string]interface{} "bad request"
// @Failure 401 {object} map[string]interface{} "unauthorized"
// @Router /api/v1/auth/google [post]
func GoogleAuth(c *gin.Context) {
	var requestBody struct {
		TokenID string `json:"token_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Message: "Invalid request: " + err.Error(),
			Code:    "INVALID_REQUEST",
		})
		return
	}

	// Get connection from pool
	conn, err := authConnPool.Get()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Failed to connect to auth service: " + err.Error(),
			Code:    "SERVICE_UNAVAILABLE",
		})
		return
	}
	defer authConnPool.Put(conn)

	client := authProto.NewAuthServiceClient(conn)

	// Set timeout for the request
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	req := &authProto.GoogleLoginRequest{
		IdToken: requestBody.TokenID,
	}

	resp, err := client.GoogleLogin(ctx, req)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Success: false,
				Message: st.Message(),
				Code:    st.Code().String(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Success: false,
				Message: "Google authentication failed: " + err.Error(),
				Code:    "INTERNAL_ERROR",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":       resp.Success,
		"message":       resp.Message,
		"access_token":  resp.AccessToken,
		"refresh_token": resp.RefreshToken,
		"user_id":       resp.UserId,
		"token_type":    resp.TokenType,
		"expires_in":    resp.ExpiresIn,
	})
}

// VerifyEmail godoc
// @Summary Verify email address
// @Description Verify user's email address with verification code
// @Tags auth
// @Accept json
// @Produce json
// @Param verification body map[string]interface{} true "Email and verification code"
// @Success 200 {object} map[string]interface{} "tokens and user info"
// @Failure 400 {object} map[string]interface{} "bad request"
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

	// Get connection from pool
	conn, err := authConnPool.Get()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Failed to connect to auth service: " + err.Error(),
			Code:    "SERVICE_UNAVAILABLE",
		})
		return
	}
	defer authConnPool.Put(conn)

	client := authProto.NewAuthServiceClient(conn)

	// Set timeout for the request
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	req := &authProto.VerifyEmailRequest{
		Email:            request.Email,
		VerificationCode: request.VerificationCode,
	}

	resp, err := client.VerifyEmail(ctx, req)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Success: false,
				Message: st.Message(),
				Code:    st.Code().String(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Success: false,
				Message: "Email verification failed: " + err.Error(),
				Code:    "INTERNAL_ERROR",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":       resp.Success,
		"message":       resp.Message,
		"access_token":  resp.AccessToken,
		"refresh_token": resp.RefreshToken,
		"user_id":       resp.UserId,
		"token_type":    resp.TokenType,
		"expires_in":    resp.ExpiresIn,
	})
}

// ResendVerificationCode godoc
// @Summary Resend verification code
// @Description Resend verification code to user's email
// @Tags auth
// @Accept json
// @Produce json
// @Param email body map[string]interface{} true "User email"
// @Success 200 {object} map[string]interface{} "success message"
// @Failure 400 {object} map[string]interface{} "bad request"
// @Router /api/v1/auth/resend-code [post]
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

	// Get connection from pool
	conn, err := authConnPool.Get()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Failed to connect to auth service: " + err.Error(),
			Code:    "SERVICE_UNAVAILABLE",
		})
		return
	}
	defer authConnPool.Put(conn)

	client := authProto.NewAuthServiceClient(conn)

	// Set timeout for the request
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	req := &authProto.ResendVerificationCodeRequest{
		Email: request.Email,
	}

	resp, err := client.ResendVerificationCode(ctx, req)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Success: false,
				Message: st.Message(),
				Code:    st.Code().String(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Success: false,
				Message: "Failed to resend verification code: " + err.Error(),
				Code:    "INTERNAL_ERROR",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": resp.Success,
		"message": resp.Message,
	})
}

// GetUserProfile godoc
// @Summary Get user profile
// @Description Get the profile of the authenticated user
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} userProto.UserResponse "user profile"
// @Failure 401 {object} ErrorResponse "unauthorized"
// @Failure 500 {object} ErrorResponse "internal server error"
// @Router /api/v1/users/profile [get]
func GetUserProfile(c *gin.Context) {
	userIDAny, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Success: false,
			Message: "User ID not found in token",
			Code:    "UNAUTHORIZED",
		})
		return
	}

	userID, ok := userIDAny.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Invalid User ID format in token",
			Code:    "INTERNAL_ERROR",
		})
		return
	}

	// Get connection from user service pool
	conn, err := userConnPool.Get()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Failed to connect to user service: " + err.Error(),
			Code:    "SERVICE_UNAVAILABLE",
		})
		return
	}
	defer userConnPool.Put(conn)

	client := userProto.NewUserServiceClient(conn)

	// Set timeout for the request
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	req := &userProto.GetUserRequest{
		Id: userID,
	}

	resp, err := client.GetUser(ctx, req)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			// Map gRPC status code to HTTP status code if needed, otherwise use 500
			httpStatus := http.StatusInternalServerError
			if st.Code() == codes.NotFound {
				httpStatus = http.StatusNotFound
			}
			c.JSON(httpStatus, ErrorResponse{
				Success: false,
				Message: st.Message(),
				Code:    st.Code().String(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Success: false,
				Message: "Failed to get user profile: " + err.Error(),
				Code:    "INTERNAL_ERROR",
			})
		}
		return
	}

	c.JSON(http.StatusOK, resp)
}

// UpdateUserProfile godoc
// @Summary Update user profile
// @Description Update the profile of the authenticated user
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param profile body userProto.UpdateUserRequest true "User profile data to update"
// @Success 200 {object} userProto.UserResponse "updated profile"
// @Failure 400 {object} ErrorResponse "bad request"
// @Failure 401 {object} ErrorResponse "unauthorized"
// @Failure 500 {object} ErrorResponse "internal server error"
// @Router /api/v1/users/profile [put]
func UpdateUserProfile(c *gin.Context) {
	userIDAny, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Success: false,
			Message: "User ID not found in token",
			Code:    "UNAUTHORIZED",
		})
		return
	}

	userID, ok := userIDAny.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Invalid User ID format in token",
			Code:    "INTERNAL_ERROR",
		})
		return
	}

	var request userProto.UpdateUserRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Message: "Invalid request body: " + err.Error(),
			Code:    "INVALID_REQUEST",
		})
		return
	}

	// Ensure the user ID from the token matches the request (or set it)
	request.Id = userID

	// Get connection from user service pool
	conn, err := userConnPool.Get()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Failed to connect to user service: " + err.Error(),
			Code:    "SERVICE_UNAVAILABLE",
		})
		return
	}
	defer userConnPool.Put(conn)

	client := userProto.NewUserServiceClient(conn)

	// Set timeout for the request
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, err := client.UpdateUser(ctx, &request)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			httpStatus := http.StatusInternalServerError
			if st.Code() == codes.NotFound {
				httpStatus = http.StatusNotFound
			} else if st.Code() == codes.InvalidArgument {
				httpStatus = http.StatusBadRequest
			}
			c.JSON(httpStatus, ErrorResponse{
				Success: false,
				Message: st.Message(),
				Code:    st.Code().String(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Success: false,
				Message: "Failed to update user profile: " + err.Error(),
				Code:    "INTERNAL_ERROR",
			})
		}
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ListProducts godoc
// @Summary List products
// @Description Get a list of products
// @Tags products
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} map[string]interface{} "list of products"
// @Router /api/v1/products [get]
func ListProducts(c *gin.Context) {
	// This is just a stub - in a real implementation, this would call the product service via gRPC
	c.JSON(http.StatusOK, gin.H{
		"message": "list products endpoint",
	})
}

// GetProduct godoc
// @Summary Get product
// @Description Get a product by ID
// @Tags products
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "product details"
// @Failure 404 {object} map[string]interface{} "product not found"
// @Router /api/v1/products/{id} [get]
func GetProduct(c *gin.Context) {
	// This is just a stub - in a real implementation, this would call the product service via gRPC
	c.JSON(http.StatusOK, gin.H{
		"message": "get product endpoint",
	})
}

// CreateProduct godoc
// @Summary Create product
// @Description Create a new product
// @Tags products
// @Accept json
// @Produce json
// @Param product body map[string]interface{} true "Product data"
// @Security BearerAuth
// @Success 201 {object} map[string]interface{} "created product"
// @Failure 400 {object} map[string]interface{} "bad request"
// @Router /api/v1/products [post]
func CreateProduct(c *gin.Context) {
	// This is just a stub - in a real implementation, this would call the product service via gRPC
	c.JSON(http.StatusOK, gin.H{
		"message": "create product endpoint",
	})
}

// UpdateProduct godoc
// @Summary Update product
// @Description Update a product by ID
// @Tags products
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Param product body map[string]interface{} true "Updated product data"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "updated product"
// @Failure 400 {object} map[string]interface{} "bad request"
// @Failure 404 {object} map[string]interface{} "product not found"
// @Router /api/v1/products/{id} [put]
func UpdateProduct(c *gin.Context) {
	// This is just a stub - in a real implementation, this would call the product service via gRPC
	c.JSON(http.StatusOK, gin.H{
		"message": "update product endpoint",
	})
}

// DeleteProduct godoc
// @Summary Delete product
// @Description Delete a product by ID
// @Tags products
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Security BearerAuth
// @Success 204 "no content"
// @Failure 404 {object} map[string]interface{} "product not found"
// @Router /api/v1/products/{id} [delete]
func DeleteProduct(c *gin.Context) {
	// This is just a stub - in a real implementation, this would call the product service via gRPC
	c.JSON(http.StatusOK, gin.H{
		"message": "delete product endpoint",
	})
}

// CreatePayment godoc
// @Summary Create payment
// @Description Create a new payment
// @Tags payments
// @Accept json
// @Produce json
// @Param payment body map[string]interface{} true "Payment data"
// @Security BearerAuth
// @Success 201 {object} map[string]interface{} "created payment"
// @Failure 400 {object} map[string]interface{} "bad request"
// @Router /api/v1/payments [post]
func CreatePayment(c *gin.Context) {
	// This is just a stub - in a real implementation, this would call the payment service via gRPC
	c.JSON(http.StatusOK, gin.H{
		"message": "create payment endpoint",
	})
}

// GetPayment godoc
// @Summary Get payment
// @Description Get a payment by ID
// @Tags payments
// @Accept json
// @Produce json
// @Param id path string true "Payment ID"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "payment details"
// @Failure 404 {object} map[string]interface{} "payment not found"
// @Router /api/v1/payments/{id} [get]
func GetPayment(c *gin.Context) {
	// This is just a stub - in a real implementation, this would call the payment service via gRPC
	c.JSON(http.StatusOK, gin.H{
		"message": "get payment endpoint",
	})
}

// GetPaymentHistory godoc
// @Summary Get payment history
// @Description Get payment history for the authenticated user
// @Tags payments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} map[string]interface{} "payment history"
// @Router /api/v1/payments/history [get]
func GetPaymentHistory(c *gin.Context) {
	// This is just a stub - in a real implementation, this would call the payment service via gRPC
	c.JSON(http.StatusOK, gin.H{
		"message": "get payment history endpoint",
	})
}

// Helper function to upload a file to Supabase
func uploadToSupabase(fileHeader *multipart.FileHeader, bucketName string, destinationPath string) (string, error) {
	if fileHeader == nil {
		return "", nil // No file provided
	}

	file, err := fileHeader.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer file.Close()

	contentType := fileHeader.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream" // Default content type
	}

	// Remove context for now as UploadFile doesn't seem to use it
	// ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	// defer cancel()

	upsert := false // Define the boolean value
	fileOptions := supabase.FileOptions{
		ContentType: &contentType,
		Upsert:      &upsert,
	}

	_, err = supabaseClient.UploadFile(bucketName, destinationPath, file, fileOptions)

	if err != nil {
		return "", fmt.Errorf("failed to upload to supabase: %w", err)
	}

	// Construct the public URL
	publicURL := supabaseClient.GetPublicUrl(bucketName, destinationPath)

	// Note: GetPublicUrl might return the base URL + path, not necessarily a signed URL unless configured.
	// If you need temporary signed URLs, use client.CreateSignedUrl() instead after upload.
	// For simple public buckets, this structure is usually sufficient.
	return publicURL.SignedURL, nil // Assuming SignedURL gives the public accessible URL
}

// RegisterWithMedia godoc
// @Summary Register a new user with profile picture and banner
// @Description Register a new user account, including optional profile picture and banner uploads
// @Tags auth
// @Accept multipart/form-data
// @Produce json
// @Param name formData string true "User's full name"
// @Param username formData string true "Desired username"
// @Param email formData string true "User's email address"
// @Param password formData string true "Password (min 8 chars)"
// @Param confirm_password formData string true "Password confirmation"
// @Param gender formData string true "User's gender"
// @Param date_of_birth formData string true "Date of birth (e.g., MM-DD-YYYY)"
// @Param security_question formData string true "Security question"
// @Param security_answer formData string true "Security answer"
// @Param subscribe_to_newsletter formData boolean false "Subscribe to newsletter"
// @Param recaptcha_token formData string true "Google reCAPTCHA token"
// @Param profile_picture formData file false "Profile picture file"
// @Param banner_image formData file false "Banner image file"
// @Success 200 {object} RegisterResponse "registration successful"
// @Failure 400 {object} ErrorResponse "invalid input or file upload error"
// @Failure 500 {object} ErrorResponse "internal server error or service unavailable"
// @Router /api/v1/auth/register-with-media [post]
func RegisterWithMedia(c *gin.Context) {
	// Ensure services are initialized (especially Supabase client)
	// This might be redundant if init() works as expected, but safer.
	if supabaseClient == nil {
		InitServices()
	}
	if supabaseClient == nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Success: false, Message: "Supabase client not initialized", Code: "CONFIG_ERROR"})
		return
	}

	// Parse multipart form (adjust MaxMemory as needed)
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Success: false, Message: "Invalid form data: " + err.Error(), Code: "INVALID_REQUEST"})
		return
	}

	// Extract text fields
	values := form.Value
	name := values["name"][0]
	username := values["username"][0]
	email := values["email"][0]
	password := values["password"][0]
	confirmPassword := values["confirm_password"][0]
	gender := values["gender"][0]
	dateOfBirth := values["date_of_birth"][0]
	securityQuestion := values["security_question"][0]
	securityAnswer := values["security_answer"][0]
	subscribe := values["subscribe_to_newsletter"][0] == "true"
	recaptchaToken := values["recaptcha_token"][0]

	// Basic validation (can be expanded)
	if password != confirmPassword {
		c.JSON(http.StatusBadRequest, ErrorResponse{Success: false, Message: "Passwords do not match", Code: "VALIDATION_ERROR"})
		return
	}
	// Add more validation as needed...

	// Extract files
	profilePicFileHeader, _ := c.FormFile("profile_picture") // Ignore error if file not present
	bannerFileHeader, _ := c.FormFile("banner_image")        // Ignore error if file not present

	// Generate a unique ID for path generation (or use user ID after creation if preferred)
	uuidVal, _ := uuid.NewV4()
	userPathPrefix := uuidVal.String()

	// --- Upload files to Supabase --- Needs error handling improvement
	profilePicURL, err := uploadToSupabase(profilePicFileHeader, "profile-pictures", userPathPrefix+"/"+filepath.Base(profilePicFileHeader.Filename))
	if err != nil {
		log.Printf("Failed to upload profile picture: %v", err)
		// Decide if this error is fatal or if registration can continue without profile pic
		// c.JSON(http.StatusInternalServerError, ErrorResponse{Success: false, Message: "Failed to upload profile picture", Code: "UPLOAD_ERROR"})
		// return
	}
	bannerURL, err := uploadToSupabase(bannerFileHeader, "banner-images", userPathPrefix+"/"+filepath.Base(bannerFileHeader.Filename))
	if err != nil {
		log.Printf("Failed to upload banner image: %v", err)
		// Decide if this error is fatal or if registration can continue without banner
		// c.JSON(http.StatusInternalServerError, ErrorResponse{Success: false, Message: "Failed to upload banner image", Code: "UPLOAD_ERROR"})
		// return
	}

	// --- Call Auth Service ---
	conn, err := authConnPool.Get()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Success: false, Message: "Failed to connect to auth service: " + err.Error(), Code: "SERVICE_UNAVAILABLE"})
		return
	}
	defer authConnPool.Put(conn)

	client := authProto.NewAuthServiceClient(conn)
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second) // Increased timeout for potential upload + registration
	defer cancel()

	req := &authProto.RegisterRequest{
		Name:                  name,
		Username:              username,
		Email:                 email,
		Password:              password,
		ConfirmPassword:       confirmPassword,
		Gender:                gender,
		DateOfBirth:           dateOfBirth,
		SecurityQuestion:      securityQuestion,
		SecurityAnswer:        securityAnswer,
		SubscribeToNewsletter: subscribe,
		RecaptchaToken:        recaptchaToken,
		ProfilePictureUrl:     profilePicURL, // Pass Supabase URL
		BannerUrl:             bannerURL,     // Pass Supabase URL
	}

	resp, err := client.Register(ctx, req)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Success: false, Message: st.Message(), Code: st.Code().String()})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{Success: false, Message: "Registration failed: " + err.Error(), Code: "INTERNAL_ERROR"})
		}
		return
	}

	c.JSON(http.StatusOK, RegisterResponse{
		Success: resp.Success,
		Message: resp.Message,
	})
}

// Initialize connection pools and Supabase client
func InitServices() {
	connPoolInitOnce.Do(func() {
		authConnPool = NewConnectionPool(Config.GetAuthServiceAddr(), 5, 20, 10*time.Second)
		userConnPool = NewConnectionPool(Config.GetUserServiceAddr(), 5, 20, 10*time.Second)
		responseCache = lru.New(100)
	})
	supabaseInitOnce.Do(func() {
		supabaseClient = supabase.NewClient(Config.Supabase.URL, Config.Supabase.AnonKey, nil)
	})
}
