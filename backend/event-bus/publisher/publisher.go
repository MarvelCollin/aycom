package publisher

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/streadway/amqp"
)

// Event represents a message event in the system
type Event struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Source      string                 `json:"source"`
	Time        time.Time              `json:"time"`
	Data        map[string]interface{} `json:"data"`
	ContentType string                 `json:"content_type"`
}

// EventPublisher defines the interface for publishing events
type EventPublisher interface {
	PublishEvent(routingKey string, event Event) error
}

// eventPublisher implements the EventPublisher interface
type eventPublisher struct {
	channel *amqp.Channel
}

// NewEventPublisher creates a new event publisher
func NewEventPublisher(ch *amqp.Channel) EventPublisher {
	return &eventPublisher{
		channel: ch,
	}
}

// PublishEvent publishes an event to the message broker
func (p *eventPublisher) PublishEvent(routingKey string, event Event) error {
	// Set the time if not already set
	if event.Time.IsZero() {
		event.Time = time.Now()
	}

	// Marshal the event to JSON
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("error marshaling event: %w", err)
	}

	// Set default content type if not specified
	contentType := event.ContentType
	if contentType == "" {
		contentType = "application/json"
	}

	// Create the publishing message
	msg := amqp.Publishing{
		ContentType: contentType,
		Body:        body,
		Timestamp:   event.Time,
		MessageId:   event.ID,
		Type:        event.Type,
	}

	// Publish the message
	err = p.channel.Publish(
		"events",   // exchange
		routingKey, // routing key
		false,      // mandatory
		false,      // immediate
		msg,
	)

	if err != nil {
		return fmt.Errorf("error publishing event: %w", err)
	}

	log.Printf("Published event: %s with routing key: %s", event.Type, routingKey)
	return nil
}
