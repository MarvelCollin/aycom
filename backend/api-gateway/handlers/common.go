package handlers

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"aycom/backend/api-gateway/config"
	"aycom/backend/proto/community"
	"aycom/backend/proto/user"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Config stores the application configuration
var Config *config.Config

// Connection pools and global variables
var (
	authConnPool       *ConnectionPool
	threadConnPool     *ConnectionPool
	UserClient         user.UserServiceClient
	connPoolInitOnce   sync.Once
	requestRateLimits  = make(map[string]*RateLimiter)
	rateLimiterMutex   sync.RWMutex
	supabaseInitOnce   sync.Once
	grpcClientInitOnce sync.Once
	CommunityClient    community.CommunityServiceClient
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

// RateLimiter implements a token bucket rate limiter
type RateLimiter struct {
	tokens         float64
	maxTokens      float64
	tokensPerSec   float64
	lastRefillTime time.Time
	mu             sync.Mutex
}

// ErrorResponse is a standard error response
type ErrorResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

// AuthServiceResponse is the standard response from auth service
type AuthServiceResponse struct {
	Success      bool        `json:"success"`
	Message      string      `json:"message"`
	AccessToken  string      `json:"access_token,omitempty"`
	RefreshToken string      `json:"refresh_token,omitempty"`
	UserId       string      `json:"user_id,omitempty"`
	TokenType    string      `json:"token_type,omitempty"`
	ExpiresIn    int64       `json:"expires_in,omitempty"`
	User         interface{} `json:"user,omitempty"`
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

// Get retrieves a connection from the pool or creates a new one
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

// Put returns a connection to the pool
func (p *ConnectionPool) Put(conn *grpc.ClientConn) {
	select {
	case p.connections <- conn:
		// Connection successfully returned to pool
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
func NewRateLimiter(maxTokens, tokensPerSec float64) *RateLimiter {
	return &RateLimiter{
		tokens:         maxTokens,
		maxTokens:      maxTokens,
		tokensPerSec:   tokensPerSec,
		lastRefillTime: time.Now(),
	}
}

// Allow checks if a request is allowed by the rate limiter
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

// InitGRPCServices initializes connection pools and gRPC clients
func InitGRPCServices() {
	// Use grpcClientInitOnce to initialize all clients together
	grpcClientInitOnce.Do(func() {
		log.Println("Initializing gRPC clients...")

		// Initialize User Service Client
		userServiceAddr := Config.Services.UserService
		log.Printf("Connecting to User service at %s", userServiceAddr)
		userConn, err := grpc.Dial(
			userServiceAddr,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithBlock(),
			grpc.WithTimeout(5*time.Second),
		)
		if err != nil {
			log.Fatalf("Failed to connect to User service: %v", err)
		} else {
			UserClient = user.NewUserServiceClient(userConn)
			log.Printf("Connected to User service at %s", userServiceAddr)
		}

		// Initialize Thread Service Client (using pool for example)
		if Config.Services.ThreadService != "" {
			threadConnPool = NewConnectionPool(Config.Services.ThreadService, 5, 20, 10*time.Second)
			log.Printf("Thread service connection pool initialized for %s", Config.Services.ThreadService)
		}

		// Initialize Community Service Client
		communityServiceAddr := Config.Services.CommunityService
		log.Printf("Connecting to Community service at %s", communityServiceAddr)
		communityConn, err := grpc.Dial(
			communityServiceAddr,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithBlock(),
			grpc.WithTimeout(5*time.Second),
		)
		if err != nil {
			log.Fatalf("Failed to connect to Community service: %v", err)
		} else {
			CommunityClient = community.NewCommunityServiceClient(communityConn)
			log.Printf("Connected to Community service at %s", communityServiceAddr)
		}

		log.Println("gRPC clients initialization complete.")
	})
}

// InitHandlers initializes the handlers package with configuration
func InitHandlers(cfg *config.Config) {
	// Set the global configuration first
	Config = cfg

	// Initialize services in the correct order
	InitGRPCServices()
	InitCommunityServiceClient(cfg)
	InitWebsocketServices() // This must be after setting Config
}

// @Summary Health check
// @Description Returns the health status of the API
// @Tags Health
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /health [get]
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// RateLimitMiddleware is a stub middleware for rate limiting
func RateLimitMiddleware(c *gin.Context) {
	// Allow all requests for now
	c.Next()
}

// SendErrorResponse sends a standardized error response
func SendErrorResponse(c *gin.Context, status int, code, message string) {
	c.JSON(status, ErrorResponse{
		Success: false,
		Message: message,
		Code:    code,
	})
}

// SendSuccessResponse sends a standardized success response
func SendSuccessResponse(c *gin.Context, status int, data interface{}) {
	c.JSON(status, gin.H{
		"success": true,
		"data":    data,
	})
}

// GetJWTSecret returns the JWT secret from environment or a default value
func GetJWTSecret() []byte {
	jwtSecret := []byte(os.Getenv("JWT_SECRET"))
	if len(jwtSecret) == 0 {
		jwtSecret = []byte("wompwomp123") // Fallback to default in .env
		log.Println("Warning: JWT_SECRET environment variable not set, using default.")
	}
	return jwtSecret
}

// ProxyServiceHealthCheck creates a handler that proxies health check requests to a service
func ProxyServiceHealthCheck(serviceName, port string) gin.HandlerFunc {
	return func(c *gin.Context) {
		serviceHost := serviceName

		// For local development, services are on localhost
		if Config.Server.CORSOrigin == "*" || os.Getenv("ENVIRONMENT") == "development" {
			serviceHost = "localhost"
		}

		// Log the request details
		log.Printf("Proxying health check to %s at %s:%s", serviceName, serviceHost, port)

		// Attempt to contact the service health endpoint
		client := &http.Client{
			Timeout: 5 * time.Second,
		}

		url := fmt.Sprintf("http://%s:%s/health", serviceHost, port)
		log.Printf("Sending health check request to: %s", url)

		resp, err := client.Get(url)
		if err != nil {
			log.Printf("Health check error for %s: %v", serviceName, err)
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":  "error",
				"message": fmt.Sprintf("Failed to connect to %s: %v", serviceName, err),
				"service": serviceName,
			})
			return
		}
		defer resp.Body.Close()

		// Read response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Failed to read response from %s: %v", serviceName, err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": fmt.Sprintf("Failed to read response from %s: %v", serviceName, err),
				"service": serviceName,
			})
			return
		}

		// Log successful response
		log.Printf("Health check response from %s: %s (status: %d)", serviceName, string(body), resp.StatusCode)

		// Set content type from original response
		contentType := resp.Header.Get("Content-Type")
		if contentType != "" {
			c.Header("Content-Type", contentType)
		} else {
			c.Header("Content-Type", "application/json")
		}

		// Return the same status code and body
		c.Status(resp.StatusCode)
		c.Writer.Write(body)
	}
}
