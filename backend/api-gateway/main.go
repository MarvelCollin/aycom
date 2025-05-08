// @title AYCOM API Gateway
// @version 1.0
// @description This is the API Gateway for the AYCOM platform.
// @host localhost:8083
// @BasePath /api/v1
// @schemes http https
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
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
	"aycom/backend/api-gateway/utils"
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
		log.Printf("Warning: Failed to load configuration: %v. Using default configuration.", err)
		// Use default configuration instead of failing
		cfg = config.GetDefaultConfig()
	}

	// Initialize handlers with configuration
	handlers.InitHandlers(cfg)

	// Initialize Supabase client
	if err := utils.InitSupabase(); err != nil {
		log.Printf("Warning: Failed to initialize Supabase client: %v", err)
	} else {
		log.Println("Supabase client initialized successfully")
	}

	// Set up router with all routes
	r := router.SetupRouter(cfg)

	// Create the server
	server := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// Start the server in a goroutine
	go func() {
		fmt.Printf("API Gateway started on port: %s\n", port)
		fmt.Printf("Swagger UI available at: http://localhost:%s/swagger/index.html\n", port)
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
	// Try loading .env from current directory first
	err := godotenv.Load()
	if err == nil {
		log.Println("Loaded .env file from current directory")
		return
	}

	// Try project root directory (parent of api-gateway directory)
	parentDir := ".."
	parentEnvPath := filepath.Join(parentDir, ".env")
	err = godotenv.Load(parentEnvPath)
	if err == nil {
		log.Printf("Loaded .env from %s", parentEnvPath)
		return
	}

	// Try 2 levels up (from backend/api-gateway to root)
	rootDir := filepath.Join(parentDir, "..")
	rootEnvPath := filepath.Join(rootDir, ".env")
	err = godotenv.Load(rootEnvPath)
	if err == nil {
		log.Printf("Loaded .env from %s", rootEnvPath)
		return
	}

	// Log error and continue with environment variables
	log.Println("Warning: .env file not found, using environment variables")
}
