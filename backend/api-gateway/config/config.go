package config

import (
	"fmt"
	"os"
)

type ServiceConfig struct {
	AuthServiceHost string
	AuthServicePort string
	UserServiceHost string
	UserServicePort string
}

type OAuthConfig struct {
	GoogleClientID     string
	GoogleClientSecret string
}

type SupabaseConfig struct {
	URL     string
	AnonKey string
}

type Config struct {
	Services  ServiceConfig
	OAuth     OAuthConfig
	Supabase  SupabaseConfig
	JWTSecret string
}

func LoadConfig() (*Config, error) {
	cfg := &Config{
		Services: ServiceConfig{
			AuthServiceHost: getEnv("AUTH_SERVICE_HOST", "auth_service"),
			AuthServicePort: getEnv("AUTH_SERVICE_PORT", "9090"),
			UserServiceHost: getEnv("USER_SERVICE_HOST", "user_service"),
			UserServicePort: getEnv("USER_SERVICE_PORT", "9091"),
		},
		OAuth: OAuthConfig{
			GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
			GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
		},
		Supabase: SupabaseConfig{
			URL:     getEnv("SUPABASE_URL", ""),
			AnonKey: getEnv("SUPABASE_ANON_KEY", ""),
		},
		JWTSecret: getEnv("JWT_SECRET", "default-secret-key"),
	}

	if cfg.Supabase.URL == "" || cfg.Supabase.AnonKey == "" {
		return nil, fmt.Errorf("SUPABASE_URL and SUPABASE_ANON_KEY environment variables must be set")
	}

	return cfg, nil
}

func (c *Config) GetAuthServiceAddr() string {
	return fmt.Sprintf("%s:%s", c.Services.AuthServiceHost, c.Services.AuthServicePort)
}

func (c *Config) GetUserServiceAddr() string {
	return fmt.Sprintf("%s:%s", c.Services.UserServiceHost, c.Services.UserServicePort)
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
