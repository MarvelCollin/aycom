package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Acad600-Tpa/WEB-MV-242/event-bus/config"
	"github.com/Acad600-Tpa/WEB-MV-242/event-bus/handlers"
	"github.com/Acad600-Tpa/WEB-MV-242/event-bus/publisher"
	"github.com/joho/godotenv"
	"github.com/streadway/amqp"
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

	// Initialize publisher
	pub := publisher.NewEventPublisher(ch)

	// Initialize event handlers
	userEventHandler := handlers.NewUserEventHandler(ch, pub)
	productEventHandler := handlers.NewProductEventHandler(ch, pub)
	orderEventHandler := handlers.NewOrderEventHandler(ch, pub)

	// Start the event handlers
	go userEventHandler.Start()
	go productEventHandler.Start()
	go orderEventHandler.Start()

	log.Println("Event bus started. Waiting for events...")

	// Wait for termination signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down event bus...")
}
