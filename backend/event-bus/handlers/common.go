package handlers

import (
	"os"
)

// getEnvFromHandler gets an environment variable with fallback
func getEnvFromHandler(key, fallback string) string {
	if val, exists := os.LookupEnv(key); exists {
		return val
	}
	return fallback
}
