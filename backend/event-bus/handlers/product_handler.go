package handlers

import (
	"encoding/json"
	"log"
	"time"

	"aycom/backend/event-bus/publisher"

	"github.com/streadway/amqp"
)

// ProductEventHandler handles product-related events
type ProductEventHandler struct {
	publisher publisher.EventPublisher
	conn      *amqp.Connection
	channel   *amqp.Channel
}

// NewProductEventHandler creates a new product event handler
func NewProductEventHandler(pub publisher.EventPublisher) *ProductEventHandler {
	return &ProductEventHandler{
		publisher: pub,
	}
}

// Start initializes the product event handler and starts consuming messages
func (h *ProductEventHandler) Start() error {
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
		h.cleanup()
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

	log.Println("Product event handler started successfully")
	return nil
}

// reconnect attempts to reestablish the RabbitMQ connection and restarts consumption
func (h *ProductEventHandler) reconnect() error {
	h.cleanup()
	return h.Start()
}

// cleanup closes the channel and connection
func (h *ProductEventHandler) cleanup() {
	if h.channel != nil {
		h.channel.Close()
	}
	if h.conn != nil {
		h.conn.Close()
	}
}

// processMessage handles an individual message
func (h *ProductEventHandler) processMessage(msg amqp.Delivery) {
	log.Printf("Received product event with routing key: %s", msg.RoutingKey)

	// Parse event
	var event publisher.Event
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		log.Printf("Error unmarshaling event: %v", err)
		return
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
