package handlers

import (
	"encoding/json"
	"log"
	"time"

	"aycom/backend/event-bus/publisher"

	"github.com/streadway/amqp"
)

type ProductEventHandler struct {
	publisher publisher.EventPublisher
	conn      *amqp.Connection
	channel   *amqp.Channel
}

func NewProductEventHandler(pub publisher.EventPublisher) *ProductEventHandler {
	return &ProductEventHandler{
		publisher: pub,
	}
}

func (h *ProductEventHandler) Start() error {

	var err error
	retries := 0
	maxRetries := 5

	for retries < maxRetries {

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

	h.channel, err = h.conn.Channel()
	if err != nil {
		h.conn.Close()
		return err
	}

	err = h.channel.ExchangeDeclare(
		"events",
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		h.cleanup()
		return err
	}

	q, err := h.channel.QueueDeclare(
		"product_events",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		h.cleanup()
		return err
	}

	err = h.channel.QueueBind(
		q.Name,
		"product.*",
		"events",
		false,
		nil,
	)
	if err != nil {
		h.cleanup()
		return err
	}

	msgs, err := h.channel.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		h.cleanup()
		return err
	}

	closeChan := h.conn.NotifyClose(make(chan *amqp.Error, 1))

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

				time.Sleep(time.Second)
				if err := h.reconnect(); err != nil {
					log.Printf("Failed to reconnect: %v", err)
					return
				}
				return
			}
		}
	}()

	log.Println("Product event handler started successfully")
	return nil
}

func (h *ProductEventHandler) reconnect() error {
	h.cleanup()
	return h.Start()
}

func (h *ProductEventHandler) cleanup() {
	if h.channel != nil {
		h.channel.Close()
	}
	if h.conn != nil {
		h.conn.Close()
	}
}

func (h *ProductEventHandler) processMessage(msg amqp.Delivery) {
	log.Printf("Received product event with routing key: %s", msg.RoutingKey)

	var event publisher.Event
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		log.Printf("Error unmarshaling event: %v", err)
		return
	}

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

func (h *ProductEventHandler) handleProductCreated(event publisher.Event) {
	log.Printf("Processing product created event: %v", event)

}

func (h *ProductEventHandler) handleProductUpdated(event publisher.Event) {
	log.Printf("Processing product updated event: %v", event)

}

func (h *ProductEventHandler) handleProductDeleted(event publisher.Event) {
	log.Printf("Processing product deleted event: %v", event)

}

func (h *ProductEventHandler) handleProductStockChanged(event publisher.Event) {
	log.Printf("Processing product stock changed event: %v", event)

	if data, ok := event.Data["stock"].(float64); ok && data < 10 {

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
