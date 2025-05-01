package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Acad600-Tpa/WEB-MV-242/backend/api-gateway/config"
	"github.com/Acad600-Tpa/WEB-MV-242/backend/api-gateway/docs"
	// Import swagger docs
	_ "github.com/Acad600-Tpa/WEB-MV-242/backend/api-gateway/docs"
	"github.com/Acad600-Tpa/WEB-MV-242/backend/api-gateway/handlers"
	"github.com/Acad600-Tpa/WEB-MV-242/backend/api-gateway/router"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title AYCOM API Gateway
// @version 1.0
// @description API Gateway for AYCOM microservices
// @host localhost:8080
// @BasePath /api/v1

func main() {
	// Initialize Swagger docs
	docs.SwaggerInfo.Title = "AYCOM API Gateway"
	docs.SwaggerInfo.Description = "API Gateway for AYCOM microservices"
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Host = "localhost:8080"
	docs.SwaggerInfo.BasePath = "/api/v1"
	docs.SwaggerInfo.Schemes = []string{"http", "https"}

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

	// Add Swagger documentation endpoint
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Get port from environment
	port := os.Getenv("API_GATEWAY_PORT")
	if port == "" {
		port = "8080"
	}

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
