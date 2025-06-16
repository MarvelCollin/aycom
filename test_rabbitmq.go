package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/streadway/amqp"
)

type Event struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Source      string                 `json:"source"`
	Time        time.Time              `json:"time"`
	Data        map[string]interface{} `json:"data"`
	ContentType string                 `json:"content_type"`
}

func main() {
	// Connect to RabbitMQ
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}
	defer ch.Close()

	fmt.Println("✅ Successfully connected to RabbitMQ!")

	// Test publishing an event
	testEvent := Event{
		ID:          "test-123",
		Type:        "test.event",
		Source:      "test_client",
		Time:        time.Now(),
		Data:        map[string]interface{}{"message": "Hello from test!"},
		ContentType: "application/json",
	}

	body, err := json.Marshal(testEvent)
	if err != nil {
		log.Fatalf("Failed to marshal event: %v", err)
	}

	err = ch.Publish(
		"events",     // exchange
		"test.event", // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})

	if err != nil {
		log.Fatalf("Failed to publish event: %v", err)
	}

	fmt.Println("✅ Successfully published test event!")
	fmt.Printf("Event: %+v\n", testEvent)

	// Check if queues exist
	queueNames := []string{"user_events", "thread_events", "order_events", "product_events"}

	fmt.Println("\n📋 Checking queues:")
	for _, queueName := range queueNames {
		queue, err := ch.QueueInspect(queueName)
		if err != nil {
			fmt.Printf("❌ Queue '%s': Error - %v\n", queueName, err)
		} else {
			fmt.Printf("✅ Queue '%s': %d messages, %d consumers\n", queueName, queue.Messages, queue.Consumers)
		}
	}

	fmt.Println("\n🎉 RabbitMQ test completed successfully!")
}
