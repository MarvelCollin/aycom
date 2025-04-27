package config

import (
	"fmt"
	"os"
)

// ServiceConfig contains the configuration for remote microservices
type ServiceConfig struct {
	AuthServiceHost string
	AuthServicePort string
	UserServiceHost string
	UserServicePort string
	// Add more services as needed
}

// Config represents the API Gateway configuration
type Config struct {
	Services  ServiceConfig
	JWTSecret string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	cfg := &Config{
		Services: ServiceConfig{
			AuthServiceHost: getEnv("AUTH_SERVICE_HOST", "auth_service"),
			AuthServicePort: getEnv("AUTH_SERVICE_PORT", "9090"),
			UserServiceHost: getEnv("USER_SERVICE_HOST", "user_service"),
			UserServicePort: getEnv("USER_SERVICE_PORT", "9091"),
		},
		JWTSecret: getEnv("JWT_SECRET", "default-secret-key"),
	}

	return cfg, nil
}

// GetAuthServiceAddr returns the full address for the auth service
func (c *Config) GetAuthServiceAddr() string {
	return fmt.Sprintf("%s:%s", c.Services.AuthServiceHost, c.Services.AuthServicePort)
}

// GetUserServiceAddr returns the full address for the user service
func (c *Config) GetUserServiceAddr() string {
	return fmt.Sprintf("%s:%s", c.Services.UserServiceHost, c.Services.UserServicePort)
}

// Helper to get environment variables with fallbacks
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
