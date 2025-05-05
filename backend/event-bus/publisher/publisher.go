package publisher

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
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
	Close() error
}

// eventPublisher implements the EventPublisher interface
type eventPublisher struct {
	connection    *amqp.Connection
	channel       *amqp.Channel
	url           string
	connected     bool
	mu            sync.Mutex
	reconnectChan chan struct{}
	closeChan     chan struct{}
}

// NewEventPublisher creates a new event publisher with connection management
func NewEventPublisher(url string) (EventPublisher, error) {
	publisher := &eventPublisher{
		url:           url,
		connected:     false,
		reconnectChan: make(chan struct{}, 1),
		closeChan:     make(chan struct{}, 1),
	}

	// Establish initial connection
	if err := publisher.connect(); err != nil {
		log.Printf("Failed to establish initial connection: %v, will retry in background", err)
		// Start reconnection goroutine
		go publisher.reconnectLoop()
		// Signal for immediate reconnection attempt
		publisher.reconnectChan <- struct{}{}
	}

	return publisher, nil
}

// connect establishes a connection to RabbitMQ and sets up a channel
func (p *eventPublisher) connect() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Close existing connections if any
	if p.channel != nil {
		p.channel.Close()
		p.channel = nil
	}
	if p.connection != nil {
		p.connection.Close()
		p.connection = nil
	}

	// Connect to RabbitMQ
	var err error
	p.connection, err = amqp.Dial(p.url)
	if err != nil {
		p.connected = false
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	// Create a channel
	p.channel, err = p.connection.Channel()
	if err != nil {
		p.connection.Close()
		p.connection = nil
		p.connected = false
		return fmt.Errorf("failed to open a channel: %w", err)
	}

	// Declare the exchange
	err = p.channel.ExchangeDeclare(
		"events", // name
		"topic",  // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	if err != nil {
		p.channel.Close()
		p.channel = nil
		p.connection.Close()
		p.connection = nil
		p.connected = false
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	// Set up notification of disconnection
	closeChan := make(chan *amqp.Error, 1)
	p.channel.NotifyClose(closeChan)

	// Start a goroutine to handle connection closure
	go func() {
		select {
		case <-closeChan:
			log.Println("RabbitMQ connection lost, reconnecting...")
			p.connected = false
			select {
			case p.reconnectChan <- struct{}{}:
				// Signal sent
			default:
				// Channel already has a pending signal
			}
		case <-p.closeChan:
			return // Exit the goroutine when publisher is closed
		}
	}()

	p.connected = true
	log.Println("Successfully connected to RabbitMQ")
	return nil
}

// reconnectLoop continuously attempts to reconnect to RabbitMQ
func (p *eventPublisher) reconnectLoop() {
	backoff := 1 * time.Second
	maxBackoff := 30 * time.Second

	for {
		select {
		case <-p.reconnectChan:
			// Attempt to reconnect
			if err := p.connect(); err != nil {
				log.Printf("Failed to reconnect: %v, retrying in %v", err, backoff)

				// Set up a timer for next reconnect attempt with exponential backoff
				time.AfterFunc(backoff, func() {
					select {
					case p.reconnectChan <- struct{}{}:
						// Signal sent
					default:
						// Channel already has a pending signal
					}
				})

				// Increase backoff time, but don't exceed max
				backoff = backoff * 2
				if backoff > maxBackoff {
					backoff = maxBackoff
				}
			} else {
				// Reset backoff time on successful connection
				backoff = 1 * time.Second
			}
		case <-p.closeChan:
			return // Exit the loop when publisher is closed
		}
	}
}

// PublishEvent publishes an event to the message broker
func (p *eventPublisher) PublishEvent(routingKey string, event Event) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// If not connected, attempt to reconnect
	if !p.connected {
		return fmt.Errorf("not connected to RabbitMQ, event will be lost: %+v", event)
	}

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
		ContentType:  contentType,
		Body:         body,
		Timestamp:    event.Time,
		MessageId:    event.ID,
		Type:         event.Type,
		DeliveryMode: amqp.Persistent, // Make messages persistent
	}

	// Publish the message
	err = p.channel.Publish(
		"events",   // exchange
		routingKey, // routing key
		true,       // mandatory (message must be routed to a queue)
		false,      // immediate
		msg,
	)

	if err != nil {
		// Mark as disconnected so next publish will trigger reconnect
		p.connected = false
		select {
		case p.reconnectChan <- struct{}{}:
			// Signal for reconnection
		default:
			// Channel already has a signal
		}
		return fmt.Errorf("error publishing event: %w", err)
	}

	log.Printf("Published event: %s with routing key: %s", event.Type, routingKey)
	return nil
}

// Close cleans up resources
func (p *eventPublisher) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	close(p.closeChan)

	if p.channel != nil {
		p.channel.Close()
		p.channel = nil
	}

	if p.connection != nil {
		p.connection.Close()
		p.connection = nil
	}

	p.connected = false
	return nil
}
