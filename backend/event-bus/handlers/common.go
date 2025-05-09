package handlers

import (
	"os"
)

func getEnvFromHandler(key, fallback string) string {
	if val, exists := os.LookupEnv(key); exists {
		return val
	}
	return fallback
}
