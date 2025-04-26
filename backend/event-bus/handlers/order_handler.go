package handlers

import (
	"encoding/json"
	"log"

	"github.com/Acad600-Tpa/WEB-MV-242/event-bus/publisher"
	"github.com/streadway/amqp"
)

// OrderEventHandler handles order-related events
type OrderEventHandler struct {
	channel   *amqp.Channel
	publisher publisher.EventPublisher
}

// NewOrderEventHandler creates a new order event handler
func NewOrderEventHandler(ch *amqp.Channel, pub publisher.EventPublisher) *OrderEventHandler {
	return &OrderEventHandler{
		channel:   ch,
		publisher: pub,
	}
}

// Start initializes the order event handler and starts consuming messages
func (h *OrderEventHandler) Start() error {
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
		log.Printf("Received order event with routing key: %s", msg.RoutingKey)

		// Parse event
		var event publisher.Event
		if err := json.Unmarshal(msg.Body, &event); err != nil {
			log.Printf("Error unmarshaling event: %v", err)
			continue
		}

		// Process based on routing key
		switch msg.RoutingKey {
		case "order.created":
			h.handleOrderCreated(event)
		case "order.updated":
			h.handleOrderUpdated(event)
		case "order.canceled":
			h.handleOrderCanceled(event)
		case "order.completed":
			h.handleOrderCompleted(event)
		default:
			log.Printf("Unhandled order event type: %s", msg.RoutingKey)
		}
	}

	return nil
}

// handleOrderCreated processes order created events
func (h *OrderEventHandler) handleOrderCreated(event publisher.Event) {
	log.Printf("Processing order created event: %v", event)

	// Process the event as needed

	// Publish notification about new order
	notificationEvent := publisher.Event{
		Type:   "notification.new_order",
		Source: "event_bus",
		Data:   event.Data,
	}

	if err := h.publisher.PublishEvent("notification.new_order", notificationEvent); err != nil {
		log.Printf("Error publishing new order notification: %v", err)
	}

	// Trigger inventory update
	inventoryEvent := publisher.Event{
		Type:   "inventory.reserve",
		Source: "event_bus",
		Data:   event.Data,
	}

	if err := h.publisher.PublishEvent("inventory.reserve", inventoryEvent); err != nil {
		log.Printf("Error publishing inventory reserve event: %v", err)
	}
}

// handleOrderUpdated processes order updated events
func (h *OrderEventHandler) handleOrderUpdated(event publisher.Event) {
	log.Printf("Processing order updated event: %v", event)

	// Process the event as needed
}

// handleOrderCanceled processes order canceled events
func (h *OrderEventHandler) handleOrderCanceled(event publisher.Event) {
	log.Printf("Processing order canceled event: %v", event)

	// Process the event as needed

	// Trigger inventory release
	inventoryEvent := publisher.Event{
		Type:   "inventory.release",
		Source: "event_bus",
		Data:   event.Data,
	}

	if err := h.publisher.PublishEvent("inventory.release", inventoryEvent); err != nil {
		log.Printf("Error publishing inventory release event: %v", err)
	}
}

// handleOrderCompleted processes order completed events
func (h *OrderEventHandler) handleOrderCompleted(event publisher.Event) {
	log.Printf("Processing order completed event: %v", event)

	// Process the event as needed

	// Publish notification about completed order
	notificationEvent := publisher.Event{
		Type:   "notification.order_completed",
		Source: "event_bus",
		Data:   event.Data,
	}

	if err := h.publisher.PublishEvent("notification.order_completed", notificationEvent); err != nil {
		log.Printf("Error publishing order completed notification: %v", err)
	}
}
