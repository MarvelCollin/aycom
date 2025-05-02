package handlers

import (
	"encoding/json"
	"log"

	"aycom/backend/event-bus/publisher"

	"github.com/google/uuid"
	"github.com/streadway/amqp"
)

// UserEventHandler handles user-related events
type UserEventHandler struct {
	channel   *amqp.Channel
	publisher publisher.EventPublisher
}

// NewUserEventHandler creates a new user event handler
func NewUserEventHandler(ch *amqp.Channel, pub publisher.EventPublisher) *UserEventHandler {
	return &UserEventHandler{
		channel:   ch,
		publisher: pub,
	}
}

// Start initializes the user event handler and starts consuming messages
func (h *UserEventHandler) Start() error {
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
		return err
	}

	// Process incoming messages
	for msg := range msgs {
		log.Printf("Received user event with routing key: %s", msg.RoutingKey)

		// Parse event
		var event publisher.Event
		if err := json.Unmarshal(msg.Body, &event); err != nil {
			log.Printf("Error unmarshaling event: %v", err)
			continue
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

	return nil
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
