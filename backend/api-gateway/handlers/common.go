package handlers

import (
	"context"
	"log"
	"sync"
	"time"

	"aycom/backend/api-gateway/config"

	// supabase "github.com/supabase-community/storage-go"
	"net/http"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Config stores the application configuration
var Config *config.Config

// Connection pools and global variables
var (
	authConnPool      *ConnectionPool
	threadConnPool    *ConnectionPool
	connPoolInitOnce  sync.Once
	requestRateLimits = make(map[string]*RateLimiter)
	rateLimiterMutex  sync.RWMutex
	// supabaseClient    *supabase.Client
	supabaseInitOnce sync.Once
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
	Success      bool   `json:"success"`
	Message      string `json:"message"`
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	UserId       string `json:"user_id,omitempty"`
	TokenType    string `json:"token_type,omitempty"`
	ExpiresIn    int64  `json:"expires_in,omitempty"`
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

// InitServices initializes connection pools and services
func InitServices() {
	// Initialize connection pools
	connPoolInitOnce.Do(func() {
		authConnPool = NewConnectionPool(Config.GetAuthServiceAddr(), 5, 20, 10*time.Second)
		if Config.GetThreadServiceAddr() != "" {
			threadConnPool = NewConnectionPool(Config.GetThreadServiceAddr(), 5, 20, 10*time.Second)
		}
	})

	// Initialize Supabase client
	supabaseInitOnce.Do(func() {
		if Config.Supabase.URL != "" && Config.Supabase.AnonKey != "" {
			// supabaseClient = supabase.NewClient(Config.Supabase.URL, Config.Supabase.AnonKey, nil)
			log.Println("Supabase client initialized")
		} else {
			log.Println("Warning: Supabase credentials not provided")
		}
	})
}

// uploadToSupabase uploads a file to Supabase storage
func uploadToSupabase() (string, error) {
	return "", nil
}

// InitHandlers initializes the handlers package with configuration
func InitHandlers(cfg *config.Config) {
	Config = cfg
	InitServices()
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
