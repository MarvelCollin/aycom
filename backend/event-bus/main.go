package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"aycom/backend/event-bus/config"
	"aycom/backend/event-bus/handlers"
	"aycom/backend/event-bus/publisher"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	fmt.Println("Event bus service starting...")

	// Create publisher with connection management
	pub, err := publisher.NewEventPublisher(cfg.RabbitMQURL)
	if err != nil {
		log.Printf("Warning: Initial connection to RabbitMQ failed: %v", err)
		log.Println("The system will continue to attempt reconnection in the background")
	}

	// Make sure to close the publisher when the application exits
	defer func() {
		if err := pub.Close(); err != nil {
			log.Printf("Error closing publisher: %v", err)
		}
	}()

	// Initialize and start event handlers
	var wg sync.WaitGroup

	// Start user event handler
	userHandler := handlers.NewUserEventHandler(pub)
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Println("Starting user event handler...")
		if err := userHandler.Start(); err != nil {
			log.Printf("Error in user event handler: %v", err)
		}
	}()

	// Start order event handler
	orderHandler := handlers.NewOrderEventHandler(pub)
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Println("Starting order event handler...")
		if err := orderHandler.Start(); err != nil {
			log.Printf("Error in order event handler: %v", err)
		}
	}()

	// Start product event handler
	productHandler := handlers.NewProductEventHandler(pub)
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Println("Starting product event handler...")
		if err := productHandler.Start(); err != nil {
			log.Printf("Error in product event handler: %v", err)
		}
	}()

	log.Println("Event bus service started successfully")

	// Set up graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down event bus...")

	// Give handlers time to finish ongoing processing
	time.Sleep(5 * time.Second)

	log.Println("Event bus stopped")
}
