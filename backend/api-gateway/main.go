package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"aycom/backend/api-gateway/config"
	"aycom/backend/api-gateway/handlers"
	"aycom/backend/api-gateway/router"
)

func main() {
	// Load .env file from project root
	loadEnvFile()

	// Get port from environment with fallback
	port := os.Getenv("API_GATEWAY_PORT")
	if port == "" {
		port = "8081"
	}

	// Set Gin to release mode in production
	if os.Getenv("GIN_MODE") != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Make the config available to handlers
	handlers.Config = cfg

	// Initialize services
	handlers.InitServices()

	// Set up router with all routes
	r := router.SetupRouter()

	// Create the server
	server := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// Start the server in a goroutine
	go func() {
		fmt.Printf("API Gateway started on port: %s\n", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Set up graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down API Gateway...")

	// Create a deadline context for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("API Gateway stopped")
}

// loadEnvFile loads the .env file from the project root or current directory
func loadEnvFile() {
	// Try loading .env from the absolute project root path first (with explicit file path)
	rootEnvPath := "C:\\BINUS\\TPA\\Web\\AYCOM\\.env"
	err := godotenv.Load(rootEnvPath)
	if err == nil {
		log.Printf("Loaded .env from %s", rootEnvPath)
		return
	}

	// If that fails, try to load .env from current directory
	err = godotenv.Load()
	if err == nil {
		log.Println("Loaded .env from current directory")
		return
	}

	// Try project root directory (2 levels up from api-gateway directory)
	dir, _ := os.Getwd()
	rootDir := filepath.Dir(filepath.Dir(dir))
	rootEnvPath = filepath.Join(rootDir, ".env")

	err = godotenv.Load(rootEnvPath)
	if err == nil {
		log.Printf("Loaded .env from %s", rootEnvPath)
		return
	}

	// If all attempts fail, log a warning
	log.Println("Warning: .env file not found, using environment variables")
}
