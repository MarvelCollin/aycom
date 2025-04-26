package main

import (
	"log"
	"os"

	"github.com/Acad600-Tpa/WEB-MV-242/api-gateway/config"
	"github.com/Acad600-Tpa/WEB-MV-242/api-gateway/routes"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Initialize configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Set up Gin
	router := gin.Default()

	// Register routes
	routes.RegisterRoutes(router, cfg)

	// Start server
	port := os.Getenv("API_GATEWAY_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("API Gateway starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start API Gateway: %v", err)
	}
}
