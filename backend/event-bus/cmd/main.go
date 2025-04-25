package main

import (
	"log"
	"os"
	"time"

	"github.com/streadway/amqp"
)

func main() {
	// Get RabbitMQ connection string from environment
	rabbitMQURL := os.Getenv("RABBITMQ_URL")
	if rabbitMQURL == "" {
		rabbitMQURL = "amqp://guest:guest@localhost:5672/"
	}

	// Keep trying to connect to RabbitMQ
	var conn *amqp.Connection
	var err error
	for {
		conn, err = amqp.Dial(rabbitMQURL)
		if err == nil {
			break
		}
		log.Printf("Failed to connect to RabbitMQ: %v. Retrying in 5 seconds...", err)
		time.Sleep(5 * time.Second)
	}
	defer conn.Close()

	log.Println("Connected to RabbitMQ")
	
	// Create a channel
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}
	defer ch.Close()

	// Declare exchange for event broadcasting
	err = ch.ExchangeDeclare(
		"events",  // name
		"fanout",  // type
		true,      // durable
		false,     // auto-deleted
		false,     // internal
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare an exchange: %v", err)
	}

	log.Println("Event bus service started. Waiting for events...")
	
	// Keep the application running
	forever := make(chan bool)
	<-forever
} 