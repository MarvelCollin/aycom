package handlers

import (
	"context"
	"log"
	"sync"
	"time"

	"aycom/backend/api-gateway/config"

	"net/http"

	communityProto "aycom/backend/services/community/proto"
	userProto "aycom/backend/services/user/proto"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var Config *config.Config

var (
	authConnPool      *ConnectionPool
	threadConnPool    *ConnectionPool
	UserClient        userProto.UserServiceClient
	connPoolInitOnce  sync.Once
	requestRateLimits = make(map[string]*RateLimiter)
	rateLimiterMutex  sync.RWMutex
	// supabaseClient    *supabase.Client
	supabaseInitOnce   sync.Once
	grpcClientInitOnce sync.Once
	CommunityClient    communityProto.CommunityServiceClient
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

type AuthServiceResponse struct {
	Success      bool   `json:"success"`
	Message      string `json:"message"`
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	UserId       string `json:"user_id,omitempty"`
	TokenType    string `json:"token_type,omitempty"`
	ExpiresIn    int64  `json:"expires_in,omitempty"`
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

func InitServices() {
	grpcClientInitOnce.Do(func() {
		log.Println("Initializing gRPC clients...")

		userServiceAddr := Config.GetUserServiceAddr()
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
			UserClient = userProto.NewUserServiceClient(userConn)
			log.Printf("Connected to User service at %s", userServiceAddr)
		}

		authConnPool = NewConnectionPool(Config.GetAuthServiceAddr(), 5, 20, 10*time.Second)
		log.Printf("Auth service connection pool initialized for %s", Config.GetAuthServiceAddr())

		if Config.GetThreadServiceAddr() != "" {
			threadConnPool = NewConnectionPool(Config.GetThreadServiceAddr(), 5, 20, 10*time.Second)
			log.Printf("Thread service connection pool initialized for %s", Config.GetThreadServiceAddr())
		}

		communityServiceAddr := Config.GetCommunityServiceAddr()
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
			CommunityClient = communityProto.NewCommunityServiceClient(communityConn)
			log.Printf("Connected to Community service at %s", communityServiceAddr)
		}

		log.Println("gRPC clients initialization complete.")
	})

	supabaseInitOnce.Do(func() {
		if Config.Supabase.URL != "" && Config.Supabase.AnonKey != "" {
			// supabaseClient = supabase.NewClient(Config.Supabase.URL, Config.Supabase.AnonKey, nil)
			log.Println("Supabase client initialized")
		} else {
			log.Println("Warning: Supabase credentials not provided")
		}
	})
}

func uploadToSupabase() (string, error) {
	return "", nil
}

func InitHandlers(cfg *config.Config) {
	Config = cfg
	InitServices()
}

func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func RateLimitMiddleware(c *gin.Context) {
	c.Next()
}
