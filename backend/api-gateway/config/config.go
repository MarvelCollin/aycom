package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the API Gateway
type Config struct {
	Server    ServerConfig
	Auth      AuthConfig
	Services  ServicesConfig
	RateLimit RateLimitConfig
	OAuth     OAuthConfig
	WebSocket WebSocketConfig
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	CORSOrigin   string
}

// AuthConfig holds authentication-related configuration
type AuthConfig struct {
	JWTSecret            string
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
	CookieDomain         string
	CookieSecure         bool
}

// OAuthConfig holds OAuth provider configuration
type OAuthConfig struct {
	GoogleClientID string
}

// ServicesConfig holds the addresses of all microservices
type ServicesConfig struct {
	UserService      string
	ThreadService    string
	CommunityService string
	AIService        string
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Limit    int64
	Burst    int
	Duration time.Duration
}

// WebSocketConfig holds WebSocket-related configuration
type WebSocketConfig struct {
	ReadBufferSize       int
	WriteBufferSize      int
	SendBufferSize       int
	ReadDeadlineTimeout  time.Duration
	WriteDeadlineTimeout time.Duration
	PingInterval         time.Duration
	MaxMessageSize       int
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	// Get service host and port variables
	userServiceHost := getEnv("USER_SERVICE_HOST", "localhost")
	userServicePort := getEnv("USER_SERVICE_PORT", "50052")
	threadServiceHost := getEnv("THREAD_SERVICE_HOST", "localhost")
	threadServicePort := getEnv("THREAD_SERVICE_PORT", "9092")
	communityServiceHost := getEnv("COMMUNITY_SERVICE_HOST", "localhost")
	communityServicePort := getEnv("COMMUNITY_SERVICE_PORT", "9093")
	aiServiceHost := getEnv("AI_SERVICE_HOST", "localhost")
	aiServicePort := getEnv("AI_SERVICE_PORT", "5000")

	// Log the service addresses for debugging
	fmt.Printf("Loading configuration with the following service addresses:\n")
	fmt.Printf("- User Service: %s:%s\n", userServiceHost, userServicePort)
	fmt.Printf("- Thread Service: %s:%s\n", threadServiceHost, threadServicePort)
	fmt.Printf("- Community Service: %s:%s\n", communityServiceHost, communityServicePort)
	fmt.Printf("- AI Service: %s:%s\n", aiServiceHost, aiServicePort)

	cfg := &Config{
		Server: ServerConfig{
			Port:         getEnv("API_GATEWAY_PORT", "8083"),
			ReadTimeout:  getDurationEnv("READ_TIMEOUT", 10*time.Second),
			WriteTimeout: getDurationEnv("WRITE_TIMEOUT", 10*time.Second),
			CORSOrigin:   getEnv("CORS_ORIGIN", "http://localhost:3000"),
		},
		Auth: AuthConfig{
			JWTSecret:            getEnv("JWT_SECRET", "your-secret-key"),
			AccessTokenDuration:  getDurationEnv("ACCESS_TOKEN_DURATION", 15*time.Minute),
			RefreshTokenDuration: getDurationEnv("REFRESH_TOKEN_DURATION", 7*24*time.Hour),
			CookieDomain:         getEnv("COOKIE_DOMAIN", "localhost"),
			CookieSecure:         getBoolEnv("COOKIE_SECURE", false),
		},
		Services: ServicesConfig{
			UserService:      fmt.Sprintf("%s:%s", userServiceHost, userServicePort),
			ThreadService:    fmt.Sprintf("%s:%s", threadServiceHost, threadServicePort),
			CommunityService: fmt.Sprintf("%s:%s", communityServiceHost, communityServicePort),
			AIService:        fmt.Sprintf("%s:%s", aiServiceHost, aiServicePort),
		},
		RateLimit: RateLimitConfig{
			Limit:    getInt64Env("RATE_LIMIT", 10),
			Burst:    getIntEnv("RATE_BURST", 20),
			Duration: getDurationEnv("RATE_DURATION", time.Minute),
		},
		OAuth: OAuthConfig{
			GoogleClientID: getEnv("GOOGLE_CLIENT_ID", ""),
		},
		WebSocket: WebSocketConfig{
			ReadBufferSize:       getIntEnv("WS_READ_BUFFER_SIZE", 1024),
			WriteBufferSize:      getIntEnv("WS_WRITE_BUFFER_SIZE", 1024),
			SendBufferSize:       getIntEnv("WS_SEND_BUFFER_SIZE", 256),
			ReadDeadlineTimeout:  getDurationEnv("WS_READ_DEADLINE", 60*time.Second),
			WriteDeadlineTimeout: getDurationEnv("WS_WRITE_DEADLINE", 10*time.Second),
			PingInterval:         getDurationEnv("WS_PING_INTERVAL", 54*time.Second),
			MaxMessageSize:       getIntEnv("WS_MAX_MESSAGE_SIZE", 4096),
		},
	}

	return cfg, nil
}

// Helper functions for loading environment variables with defaults
func getEnv(key, defaultVal string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultVal
}

func getIntEnv(key string, defaultVal int) int {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultVal
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		fmt.Printf("Warning: Invalid int value for %s: %s, using default: %d\n", key, valueStr, defaultVal)
		return defaultVal
	}
	return value
}

func getInt64Env(key string, defaultVal int64) int64 {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultVal
	}

	value, err := strconv.ParseInt(valueStr, 10, 64)
	if err != nil {
		fmt.Printf("Warning: Invalid int64 value for %s: %s, using default: %d\n", key, valueStr, defaultVal)
		return defaultVal
	}
	return value
}

func getBoolEnv(key string, defaultVal bool) bool {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultVal
	}

	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		fmt.Printf("Warning: Invalid bool value for %s: %s, using default: %v\n", key, valueStr, defaultVal)
		return defaultVal
	}
	return value
}

func getDurationEnv(key string, defaultVal time.Duration) time.Duration {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultVal
	}

	value, err := time.ParseDuration(valueStr)
	if err != nil {
		fmt.Printf("Warning: Invalid duration value for %s: %s, using default: %v\n", key, valueStr, defaultVal)
		return defaultVal
	}
	return value
}

// GetAuthServiceAddr returns the address of the auth service
func (c *Config) GetAuthServiceAddr() string {
	return c.Services.UserService
}

// GetDefaultConfig returns a default configuration
func GetDefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port:         getEnv("API_GATEWAY_PORT", "8083"),
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			CORSOrigin:   "http://localhost:3000",
		},
		Auth: AuthConfig{
			JWTSecret:            getEnv("JWT_SECRET", "default-secret-key-for-development"),
			AccessTokenDuration:  15 * time.Minute,
			RefreshTokenDuration: 7 * 24 * time.Hour,
			CookieDomain:         "localhost",
			CookieSecure:         false,
		},
		Services: ServicesConfig{
			UserService:      fmt.Sprintf("%s:%s", getEnv("USER_SERVICE_HOST", "localhost"), getEnv("USER_SERVICE_PORT", "50052")),
			ThreadService:    fmt.Sprintf("%s:%s", getEnv("THREAD_SERVICE_HOST", "localhost"), getEnv("THREAD_SERVICE_PORT", "9092")),
			CommunityService: fmt.Sprintf("%s:%s", getEnv("COMMUNITY_SERVICE_HOST", "localhost"), getEnv("COMMUNITY_SERVICE_PORT", "9093")),
			AIService:        fmt.Sprintf("%s:%s", getEnv("AI_SERVICE_HOST", "localhost"), getEnv("AI_SERVICE_PORT", "5000")),
		},
		RateLimit: RateLimitConfig{
			Limit:    100,
			Burst:    20,
			Duration: time.Minute,
		},
		OAuth: OAuthConfig{
			GoogleClientID: getEnv("GOOGLE_CLIENT_ID", ""),
		},
		WebSocket: WebSocketConfig{
			ReadBufferSize:       1024,
			WriteBufferSize:      1024,
			SendBufferSize:       256,
			ReadDeadlineTimeout:  60 * time.Second,
			WriteDeadlineTimeout: 10 * time.Second,
			PingInterval:         54 * time.Second,
			MaxMessageSize:       4096,
		},
	}
}
