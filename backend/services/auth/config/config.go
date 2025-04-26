package config

import (
	"fmt"
	"os"
)

// DatabaseConfig contains database connection settings
type DatabaseConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	DBName   string
	SSLMode  string
}

// Config represents the auth service configuration
type Config struct {
	Port         string
	Database     DatabaseConfig
	JWTSecret    string
	AccessTTL    int // Time-to-live for access tokens in minutes
	RefreshTTL   int // Time-to-live for refresh tokens in days
	RedisAddress string
	RabbitMQURL  string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	cfg := &Config{
		Port: getEnv("AUTH_SERVICE_PORT", "9090"),
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			Username: getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			DBName:   getEnv("DB_NAME", "auth_service"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		JWTSecret:    getEnv("JWT_SECRET", "your-secret-key"),
		AccessTTL:    getEnvAsInt("ACCESS_TOKEN_TTL", 15), // 15 minutes default
		RefreshTTL:   getEnvAsInt("REFRESH_TOKEN_TTL", 7), // 7 days default
		RedisAddress: getEnv("REDIS_ADDRESS", "localhost:6379"),
		RabbitMQURL:  getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
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

// Helper to get environment variables as integers with fallbacks
func getEnvAsInt(key string, fallback int) int {
	value := getEnv(key, "")
	if value == "" {
		return fallback
	}

	intValue, err := parseInt(value)
	if err != nil {
		return fallback
	}

	return intValue
}

// Helper to parse string to int
func parseInt(value string) (int, error) {
	var intValue int
	_, err := fmt.Sscanf(value, "%d", &intValue)
	return intValue, err
}
