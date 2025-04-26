package config

import (
	"os"
)

// ServiceConfig contains the configuration for remote microservices
type ServiceConfig struct {
	UserServiceAddr    string
	ProductServiceAddr string
	PaymentServiceAddr string
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
			UserServiceAddr:    getEnv("USER_SERVICE_ADDR", "localhost:50051"),
			ProductServiceAddr: getEnv("PRODUCT_SERVICE_ADDR", "localhost:50052"),
			PaymentServiceAddr: getEnv("PAYMENT_SERVICE_ADDR", "localhost:50053"),
		},
		JWTSecret: getEnv("JWT_SECRET", "default-secret-key"),
	}

	return cfg, nil
}

// Helper to get environment variables with fallbacks
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
