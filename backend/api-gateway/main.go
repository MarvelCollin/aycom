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
	loadEnvFile()

	if err := godotenv.Load(); err != nil {
		rootEnvPath := filepath.Join("..", "..", "..", ".env")
		if err := godotenv.Load(rootEnvPath); err != nil {
			log.Printf("Warning: .env file not found or cannot be loaded: %v", err)
		} else {
			log.Printf("Loaded .env from root directory: %s", rootEnvPath)
		}
	}

	if os.Getenv("GIN_MODE") != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8083"
		log.Printf("Using default port: %s", port)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Printf("Warning: Failed to load configuration: %v. Using default configuration.", err)
		cfg = config.GetDefaultConfig()
	}

	handlers.InitHandlers(cfg)

	if err := utils.InitSupabase(); err != nil {
		log.Printf("Warning: Failed to initialize Supabase client: %v", err)
	} else {
		log.Println("Supabase client initialized successfully")
	}

	if err := utils.InitRedis(cfg); err != nil {
		log.Printf("Warning: Failed to initialize Redis client: %v", err)
	} else {
		log.Println("Redis client initialized successfully")
	}

	// Initialize RabbitMQ Event Publisher
	rabbitMQURL := os.Getenv("RABBITMQ_URL")
	if rabbitMQURL == "" {
		rabbitMQURL = "amqp://guest:guest@rabbitmq:5672/"
	}
	if err := utils.InitEventPublisher(rabbitMQURL); err != nil {
		log.Printf("Warning: Failed to initialize RabbitMQ Event Publisher: %v", err)
	} else {
		log.Println("RabbitMQ Event Publisher initialized successfully")
	}

	r := router.SetupRouter(cfg)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		fmt.Printf("API Gateway started on port: %s\n", port)
		fmt.Printf("Swagger UI available at: http://localhost:%s/swagger/index.html\n", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down API Gateway...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("API Gateway stopped")
}

func loadEnvFile() {
	err := godotenv.Load()
	if err == nil {
		log.Println("Loaded .env file from current directory")
		return
	}

	parentDir := ".."
	parentEnvPath := filepath.Join(parentDir, ".env")
	err = godotenv.Load(parentEnvPath)
	if err == nil {
		log.Printf("Loaded .env from %s", parentEnvPath)
		return
	}

	rootDir := filepath.Join(parentDir, "..")
	rootEnvPath := filepath.Join(rootDir, ".env")
	err = godotenv.Load(rootEnvPath)
	if err == nil {
		log.Printf("Loaded .env from %s", rootEnvPath)
		return
	}

	log.Println("Warning: .env file not found, using environment variables")
}
