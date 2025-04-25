package config

import (
	"log"
	"os"
	"strconv"
)

// Config represents the application configuration
type Config struct {
	Server   ServerConfig
	Services ServicesConfig
	Auth     AuthConfig
	Redis    RedisConfig
	RabbitMQ RabbitMQConfig
}

// ServerConfig contains server related configuration
type ServerConfig struct {
	Host string
	Port string
}

// ServicesConfig contains addresses of microservices
type ServicesConfig struct {
	Auth    string
	User    string
	Product string
}

// AuthConfig contains authentication related configuration
type AuthConfig struct {
	JWTSecret string
	TokenExp  int // token expiration in hours
}

// RedisConfig contains Redis connection configuration
type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

// RabbitMQConfig contains RabbitMQ connection configuration
type RabbitMQConfig struct {
	Host     string
	Port     string
	User     string
	Password string
}

// Load loads the configuration from environment variables
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
			Port: getEnv("SERVER_PORT", "8080"),
		},
		Services: ServicesConfig{
			Auth:    getEnv("AUTH_SERVICE", "auth:50051"),
			User:    getEnv("USER_SERVICE", "user:50052"),
			Product: getEnv("PRODUCT_SERVICE", "product:50053"),
		},
		Auth: AuthConfig{
			JWTSecret: getEnv("JWT_SECRET", "your-secret-key"),
			TokenExp:  getEnvAsInt("TOKEN_EXP", 24),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "redis"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		RabbitMQ: RabbitMQConfig{
			Host:     getEnv("RABBITMQ_HOST", "rabbitmq"),
			Port:     getEnv("RABBITMQ_PORT", "5672"),
			User:     getEnv("RABBITMQ_USER", "guest"),
			Password: getEnv("RABBITMQ_PASSWORD", "guest"),
		},
	}
}

// Helper function to get an environment variable with a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// Helper function to get an environment variable as an integer
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		log.Printf("Warning: Could not parse %s as int, using default: %d", key, defaultValue)
		return defaultValue
	}

	return value
}
