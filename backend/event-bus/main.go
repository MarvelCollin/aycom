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
	"github.com/streadway/amqp"
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

	// Connect to RabbitMQ
	conn, err := amqp.Dial(cfg.RabbitMQURL)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	// Create a channel
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}
	defer ch.Close()

	// Declare the exchange
	err = ch.ExchangeDeclare(
		"events", // name
		"topic",  // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare an exchange: %v", err)
	}

	// Initialize event publisher
	pub := publisher.NewEventPublisher(ch)

	// Initialize and start event handlers
	var wg sync.WaitGroup

	// Start user event handler
	userHandler := handlers.NewUserEventHandler(ch, pub)
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Println("Starting user event handler...")
		if err := userHandler.Start(); err != nil {
			log.Printf("Error in user event handler: %v", err)
		}
	}()

	// Start order event handler
	orderHandler := handlers.NewOrderEventHandler(ch, pub)
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Println("Starting order event handler...")
		if err := orderHandler.Start(); err != nil {
			log.Printf("Error in order event handler: %v", err)
		}
	}()

	// Start product event handler
	productHandler := handlers.NewProductEventHandler(ch, pub)
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
