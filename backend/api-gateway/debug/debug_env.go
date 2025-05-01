package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

func main() {
	// Get current working directory
	dir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting current directory: %v", err)
	}
	fmt.Println("Current directory:", dir)

	// Try to load .env from current directory
	err = godotenv.Load()
	if err != nil {
		fmt.Printf("Error loading .env from current directory: %v\n", err)

		// Try project root directory (3 levels up from debug directory)
		rootDir := filepath.Dir(filepath.Dir(filepath.Dir(dir)))
		rootEnvPath := filepath.Join(rootDir, ".env")
		fmt.Println("Trying to load from:", rootEnvPath)

		err = godotenv.Load(rootEnvPath)
		if err != nil {
			fmt.Printf("Error loading .env from root directory: %v\n", err)
		} else {
			fmt.Println("Successfully loaded .env from root directory")
		}
	} else {
		fmt.Println("Successfully loaded .env from current directory")
	}

	// Print environment variables
	fmt.Println("SUPABASE_URL:", os.Getenv("SUPABASE_URL"))
	fmt.Println("SUPABASE_ANON_KEY:", os.Getenv("SUPABASE_ANON_KEY"))
}
