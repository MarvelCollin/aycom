package config

import "os"

type Config struct {
	RabbitMQURL string
}

func LoadConfig() (*Config, error) {
	cfg := &Config{
		RabbitMQURL: getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
