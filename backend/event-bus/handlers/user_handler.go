package handlers

import (
	"encoding/json"
	"log"
	"time"

	"aycom/backend/event-bus/publisher"

	"github.com/google/uuid"
	"github.com/streadway/amqp"
)

// UserEventHandler handles user-related events
type UserEventHandler struct {
	publisher publisher.EventPublisher
	conn      *amqp.Connection
	channel   *amqp.Channel
}

// NewUserEventHandler creates a new user event handler
func NewUserEventHandler(pub publisher.EventPublisher) *UserEventHandler {
	return &UserEventHandler{
		publisher: pub,
	}
}

// Start initializes the user event handler and starts consuming messages
func (h *UserEventHandler) Start() error {
	// Create a direct connection to RabbitMQ for consuming
	// This is separate from the publisher's connection
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

	// Declare a queue for user events
	q, err := h.channel.QueueDeclare(
		"user_events", // name
		true,          // durable
		false,         // delete when unused
		false,         // exclusive
		false,         // no-wait
		nil,           // arguments
	)
	if err != nil {
		h.cleanup()
		return err
	}

	// Bind the queue to the exchange with the appropriate routing key
	err = h.channel.QueueBind(
		q.Name,   // queue name
		"user.*", // routing key
		"events", // exchange
		false,    // no-wait
		nil,      // arguments
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

	log.Println("User event handler started successfully")
	return nil
}

// reconnect attempts to reestablish the RabbitMQ connection and restarts consumption
func (h *UserEventHandler) reconnect() error {
	h.cleanup()
	return h.Start()
}

// cleanup closes the channel and connection
func (h *UserEventHandler) cleanup() {
	if h.channel != nil {
		h.channel.Close()
	}
	if h.conn != nil {
		h.conn.Close()
	}
}

// processMessage handles an individual message
func (h *UserEventHandler) processMessage(msg amqp.Delivery) {
	log.Printf("Received user event with routing key: %s", msg.RoutingKey)

	// Parse event
	var event publisher.Event
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		log.Printf("Error unmarshaling event: %v", err)
		return
	}

	// Process based on routing key
	switch msg.RoutingKey {
	case "user.created":
		h.handleUserCreated(event)
	case "user.updated":
		h.handleUserUpdated(event)
	case "user.deleted":
		h.handleUserDeleted(event)
	default:
		log.Printf("Unhandled user event type: %s", msg.RoutingKey)
	}
}

// handleUserCreated processes user created events
func (h *UserEventHandler) handleUserCreated(event publisher.Event) {
	log.Printf("Processing user created event: %v", event)

	// Generate downstream event if needed
	if event.Source != "auth_service" {
		return
	}

	// Forward to relevant services
	notificationEvent := publisher.Event{
		ID:     uuid.New().String(),
		Type:   "notification.user_welcome",
		Source: "event_bus",
		Data:   event.Data,
	}

	if err := h.publisher.PublishEvent("notification.user_welcome", notificationEvent); err != nil {
		log.Printf("Error publishing notification event: %v", err)
	}
}

// handleUserUpdated processes user updated events
func (h *UserEventHandler) handleUserUpdated(event publisher.Event) {
	log.Printf("Processing user updated event: %v", event)

	// Process the event as needed
}

// handleUserDeleted processes user deleted events
func (h *UserEventHandler) handleUserDeleted(event publisher.Event) {
	log.Printf("Processing user deleted event: %v", event)

	// Process the event as needed
}
