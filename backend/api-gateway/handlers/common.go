package handlers

import (
	"aycom/backend/proto/community"
	"aycom/backend/proto/user"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"aycom/backend/api-gateway/config"
)

var AppConfig *config.Config

var (
	threadConnPool     *ConnectionPool
	UserClient         user.UserServiceClient
	grpcClientInitOnce sync.Once
	CommunityClient    community.CommunityServiceClient
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

func NewConnectionPool(serviceAddr string, maxIdle, maxOpen int, timeout time.Duration) *ConnectionPool {
	return &ConnectionPool{
		connections: make(chan *grpc.ClientConn, maxIdle),
		serviceAddr: serviceAddr,
		maxIdle:     maxIdle,
		maxOpen:     maxOpen,
		timeout:     timeout,
	}
}

func (p *ConnectionPool) Get() (*grpc.ClientConn, error) {
	select {
	case conn := <-p.connections:
		return conn, nil
	default:
		opts := []grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		}

		conn, err := grpc.NewClient(p.serviceAddr, opts...)
		if err != nil {
			return nil, err
		}
		return conn, nil
	}
}

func (p *ConnectionPool) Put(conn *grpc.ClientConn) {
	select {
	case p.connections <- conn:

	default:

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

func InitGRPCServices() {

	grpcClientInitOnce.Do(func() {
		log.Println("Initializing gRPC clients...")

		userServiceAddr := AppConfig.Services.UserService
		log.Printf("Connecting to User service at %s", userServiceAddr)

		userConn, err := grpc.NewClient(userServiceAddr,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)

		if err != nil {
			log.Printf("Warning: Failed to connect to User service: %v", err)
		} else {
			UserClient = user.NewUserServiceClient(userConn)
			log.Printf("Connected to User service at %s", userServiceAddr)
		}

		if AppConfig.Services.ThreadService != "" {

			threadConnPool = NewConnectionPool(AppConfig.Services.ThreadService, 5, 20, 3*time.Second)
			log.Printf("Thread service connection pool initialized for %s", AppConfig.Services.ThreadService)
		}

		communityServiceAddr := AppConfig.Services.CommunityService
		log.Printf("Connecting to Community service at %s", communityServiceAddr)

		communityConn, err := grpc.NewClient(communityServiceAddr,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)

		if err != nil {
			log.Printf("Warning: Failed to connect to Community service: %v", err)
		} else {
			CommunityClient = community.NewCommunityServiceClient(communityConn)
			log.Printf("Connected to Community service at %s", communityServiceAddr)
		}

		log.Println("gRPC clients initialization complete.")
	})
}

func InitHandlers(cfg *config.Config) {

	AppConfig = cfg

	InitGRPCServices()
	InitUserServiceClient(cfg)
	InitThreadServiceClient(cfg)
	InitCommunityServiceClient(cfg)
	InitWebsocketServices()
}

func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func RateLimitMiddleware(c *gin.Context) {
	c.Next()
}

func ProxyServiceHealthCheck(serviceName, port string) gin.HandlerFunc {
	return func(c *gin.Context) {
		serviceHost := serviceName

		if AppConfig.Server.CORSOrigin == "*" || os.Getenv("ENVIRONMENT") == "development" {
			serviceHost = "localhost"
		}

		log.Printf("Proxying health check to %s at %s:%s", serviceName, serviceHost, port)

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

		log.Printf("Health check response from %s: %s (status: %d)", serviceName, string(body), resp.StatusCode)

		contentType := resp.Header.Get("Content-Type")
		if contentType != "" {
			c.Header("Content-Type", contentType)
		} else {
			c.Header("Content-Type", "application/json")
		}

		c.Status(resp.StatusCode)
		c.Writer.Write(body)
	}
}
