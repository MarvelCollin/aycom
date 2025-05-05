package handlers

import (
	"encoding/json"
	"log"
	"time"

	"aycom/backend/event-bus/publisher"

	"github.com/streadway/amqp"
)

// OrderEventHandler handles order-related events
type OrderEventHandler struct {
	publisher publisher.EventPublisher
	conn      *amqp.Connection
	channel   *amqp.Channel
}

// NewOrderEventHandler creates a new order event handler
func NewOrderEventHandler(pub publisher.EventPublisher) *OrderEventHandler {
	return &OrderEventHandler{
		publisher: pub,
	}
}

// Start initializes the order event handler and starts consuming messages
func (h *OrderEventHandler) Start() error {
	// Create a direct connection to RabbitMQ for consuming
	var err error
	retries := 0
	maxRetries := 5

	for retries < maxRetries {
		// Connect to RabbitMQ for consuming
		h.conn, err = amqp.Dial(getEnvFromHandler("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"))
		if err == nil {
			break
		}
		retries++
		log.Printf("Failed to connect to RabbitMQ for consuming (attempt %d/%d): %v",
			retries, maxRetries, err)
		time.Sleep(time.Duration(retries) * time.Second)
	}

	if err != nil {
		return err
	}

	// Create a channel
	h.channel, err = h.conn.Channel()
	if err != nil {
		h.conn.Close()
		return err
	}

	// Declare the exchange (in case it doesn't exist yet)
	err = h.channel.ExchangeDeclare(
		"events", // name
		"topic",  // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	if err != nil {
		h.cleanup()
		return err
	}

	// Declare a queue for order events
	q, err := h.channel.QueueDeclare(
		"order_events", // name
		true,           // durable
		false,          // delete when unused
		false,          // exclusive
		false,          // no-wait
		nil,            // arguments
	)
	if err != nil {
		h.cleanup()
		return err
	}

	// Bind the queue to the exchange with the appropriate routing key
	err = h.channel.QueueBind(
		q.Name,    // queue name
		"order.*", // routing key
		"events",  // exchange
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		h.cleanup()
		return err
	}

	// Start consuming messages
	msgs, err := h.channel.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		h.cleanup()
		return err
	}

	// Set up a channel to handle connection closure
	closeChan := h.conn.NotifyClose(make(chan *amqp.Error, 1))

	// Create a goroutine to handle messages
	go func() {
		for {
			select {
			case msg, ok := <-msgs:
				if !ok {
					log.Println("Consumer channel closed, exiting consumer loop")
					return
				}
				h.processMessage(msg)
			case err := <-closeChan:
				log.Printf("RabbitMQ connection closed: %v, attempting to reconnect...", err)
				// Attempt to reconnect after a delay
				time.Sleep(time.Second)
				if err := h.reconnect(); err != nil {
					log.Printf("Failed to reconnect: %v", err)
					return
				}
				return // Exit this goroutine as reconnect will start a new one
			}
		}
	}()

	log.Println("Order event handler started successfully")
	return nil
}

// reconnect attempts to reestablish the RabbitMQ connection and restarts consumption
func (h *OrderEventHandler) reconnect() error {
	h.cleanup()
	return h.Start()
}

// cleanup closes the channel and connection
func (h *OrderEventHandler) cleanup() {
	if h.channel != nil {
		h.channel.Close()
	}
	if h.conn != nil {
		h.conn.Close()
	}
}

// processMessage handles an individual message
func (h *OrderEventHandler) processMessage(msg amqp.Delivery) {
	log.Printf("Received order event with routing key: %s", msg.RoutingKey)

	// Parse event
	var event publisher.Event
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		log.Printf("Error unmarshaling event: %v", err)
		return
	}

	// Process based on routing key
	switch msg.RoutingKey {
	case "order.created":
		h.handleOrderCreated(event)
	case "order.updated":
		h.handleOrderUpdated(event)
	case "order.completed":
		h.handleOrderCompleted(event)
	default:
		log.Printf("Unhandled order event type: %s", msg.RoutingKey)
	}
}

// handleOrderCreated processes order created events
func (h *OrderEventHandler) handleOrderCreated(event publisher.Event) {
	log.Printf("Processing order created event: %v", event)
	// Process the event as needed
}

// handleOrderUpdated processes order updated events
func (h *OrderEventHandler) handleOrderUpdated(event publisher.Event) {
	log.Printf("Processing order updated event: %v", event)
	// Process the event as needed
}

// handleOrderCompleted processes order completed events
func (h *OrderEventHandler) handleOrderCompleted(event publisher.Event) {
	log.Printf("Processing order completed event: %v", event)
	// Process the event as needed
}

// getEnvFromHandler gets an environment variable with fallback
// Using the function from user_handler.go
// func getEnvFromHandler(key, fallback string) string {
// 	if val, exists := os.LookupEnv(key); exists {
// 		return val
// 	}
// 	return fallback
// }
