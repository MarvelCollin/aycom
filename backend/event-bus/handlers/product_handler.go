package handlers

import (
	"encoding/json"
	"log"

	"aycom/backend/event-bus/publisher"

	"github.com/streadway/amqp"
)

// ProductEventHandler handles product-related events
type ProductEventHandler struct {
	channel   *amqp.Channel
	publisher publisher.EventPublisher
}

// NewProductEventHandler creates a new product event handler
func NewProductEventHandler(ch *amqp.Channel, pub publisher.EventPublisher) *ProductEventHandler {
	return &ProductEventHandler{
		channel:   ch,
		publisher: pub,
	}
}

// Start initializes the product event handler and starts consuming messages
func (h *ProductEventHandler) Start() error {
	// Declare a queue for product events
	q, err := h.channel.QueueDeclare(
		"product_events", // name
		true,             // durable
		false,            // delete when unused
		false,            // exclusive
		false,            // no-wait
		nil,              // arguments
	)
	if err != nil {
		return err
	}

	// Bind the queue to the exchange with the appropriate routing key
	err = h.channel.QueueBind(
		q.Name,      // queue name
		"product.*", // routing key
		"events",    // exchange
		false,       // no-wait
		nil,         // arguments
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
		log.Printf("Received product event with routing key: %s", msg.RoutingKey)

		// Parse event
		var event publisher.Event
		if err := json.Unmarshal(msg.Body, &event); err != nil {
			log.Printf("Error unmarshaling event: %v", err)
			continue
		}

		// Process based on routing key
		switch msg.RoutingKey {
		case "product.created":
			h.handleProductCreated(event)
		case "product.updated":
			h.handleProductUpdated(event)
		case "product.deleted":
			h.handleProductDeleted(event)
		case "product.stock_changed":
			h.handleProductStockChanged(event)
		default:
			log.Printf("Unhandled product event type: %s", msg.RoutingKey)
		}
	}

	return nil
}

// handleProductCreated processes product created events
func (h *ProductEventHandler) handleProductCreated(event publisher.Event) {
	log.Printf("Processing product created event: %v", event)

	// Process the event as needed
}

// handleProductUpdated processes product updated events
func (h *ProductEventHandler) handleProductUpdated(event publisher.Event) {
	log.Printf("Processing product updated event: %v", event)

	// Process the event as needed
}

// handleProductDeleted processes product deleted events
func (h *ProductEventHandler) handleProductDeleted(event publisher.Event) {
	log.Printf("Processing product deleted event: %v", event)

	// Process the event as needed
}

// handleProductStockChanged processes product stock changed events
func (h *ProductEventHandler) handleProductStockChanged(event publisher.Event) {
	log.Printf("Processing product stock changed event: %v", event)

	// Process the event as needed

	// Check if we need to notify about low stock
	if data, ok := event.Data["stock"].(float64); ok && data < 10 {
		// Publish low stock notification event
		lowStockEvent := publisher.Event{
			Type:   "notification.low_stock",
			Source: "event_bus",
			Data: map[string]interface{}{
				"product_id":   event.Data["product_id"],
				"product_name": event.Data["product_name"],
				"stock":        data,
			},
		}

		if err := h.publisher.PublishEvent("notification.low_stock", lowStockEvent); err != nil {
			log.Printf("Error publishing low stock notification: %v", err)
		}
	}
}
