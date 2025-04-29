package config

import (
	"fmt"
	"log"
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

// LoadDatabaseConfig loads database configuration from environment variables
func LoadDatabaseConfig() DatabaseConfig {
	cfg := DatabaseConfig{
		Host:     os.Getenv("DATABASE_HOST"),
		Port:     os.Getenv("DATABASE_PORT"),
		Username: os.Getenv("DATABASE_USER"),
		Password: os.Getenv("DATABASE_PASSWORD"),
		DBName:   os.Getenv("DATABASE_NAME"),
		SSLMode:  os.Getenv("DATABASE_SSL_MODE"),
	}

	// Set defaults if environment variables are not set
	if cfg.Host == "" {
		cfg.Host = "auth_db" // Default service name in Docker Compose
	}
	if cfg.Port == "" {
		cfg.Port = "5432"
	}
	if cfg.Username == "" {
		cfg.Username = "kolin" // Default user from docker-compose
	}
	if cfg.Password == "" {
		cfg.Password = "kolin" // Default password from docker-compose
	}
	if cfg.DBName == "" {
		cfg.DBName = "auth_db" // Default db name from docker-compose
	}
	if cfg.SSLMode == "" {
		cfg.SSLMode = "disable" // Default to disable for local development
	}

	log.Printf("Loaded DB Config: Host=%s, Port=%s, User=%s, DBName=%s, SSLMode=%s",
		cfg.Host, cfg.Port, cfg.Username, cfg.DBName, cfg.SSLMode)

	return cfg
}
