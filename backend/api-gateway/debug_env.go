package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

func debug_env_main() {
	// Try to load .env
	loadEnvDebug()

	// Print environment variables
	fmt.Println("SUPABASE_URL:", os.Getenv("SUPABASE_URL"))
	fmt.Println("SUPABASE_ANON_KEY:", os.Getenv("SUPABASE_ANON_KEY"))
}

func loadEnvDebug() {
	// Get current working directory
	dir, err := os.Getwd()
	if err != nil {
		log.Printf("Error getting current directory: %v", err)
		return
	}
	fmt.Println("Current directory:", dir)

	// Try to load .env from current directory
	err = godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env from current directory:", err)

		// Try root directory
		rootEnvPath := filepath.Join(filepath.Dir(filepath.Dir(dir)), ".env")
		fmt.Println("Trying to load from:", rootEnvPath)

		err = godotenv.Load(rootEnvPath)
		if err != nil {
			fmt.Println("Error loading .env from root directory:", err)
		} else {
			fmt.Println("Successfully loaded .env from root directory")
		}
	} else {
		fmt.Println("Successfully loaded .env from current directory")
	}
}
 