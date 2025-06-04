package publisher

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
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

type EventPublisher interface {
	PublishEvent(routingKey string, event Event) error
	Close() error
}

type eventPublisher struct {
	connection    *amqp.Connection
	channel       *amqp.Channel
	url           string
	connected     bool
	mu            sync.Mutex
	reconnectChan chan struct{}
	closeChan     chan struct{}
}

func NewEventPublisher(url string) (EventPublisher, error) {
	publisher := &eventPublisher{
		url:           url,
		connected:     false,
		reconnectChan: make(chan struct{}, 1),
		closeChan:     make(chan struct{}, 1),
	}

	if err := publisher.connect(); err != nil {
		log.Printf("Failed to establish initial connection: %v, will retry in background", err)

		go publisher.reconnectLoop()

		publisher.reconnectChan <- struct{}{}
	}

	return publisher, nil
}

func (p *eventPublisher) connect() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.channel != nil {
		p.channel.Close()
		p.channel = nil
	}
	if p.connection != nil {
		p.connection.Close()
		p.connection = nil
	}

	var err error
	p.connection, err = amqp.Dial(p.url)
	if err != nil {
		p.connected = false
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	p.channel, err = p.connection.Channel()
	if err != nil {
		p.connection.Close()
		p.connection = nil
		p.connected = false
		return fmt.Errorf("failed to open a channel: %w", err)
	}

	err = p.channel.ExchangeDeclare(
		"events",
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		p.channel.Close()
		p.channel = nil
		p.connection.Close()
		p.connection = nil
		p.connected = false
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	closeChan := make(chan *amqp.Error, 1)
	p.channel.NotifyClose(closeChan)

	go func() {
		select {
		case <-closeChan:
			log.Println("RabbitMQ connection lost, reconnecting...")
			p.connected = false
			select {
			case p.reconnectChan <- struct{}{}:

			default:

			}
		case <-p.closeChan:
			return
		}
	}()

	p.connected = true
	log.Println("Successfully connected to RabbitMQ")
	return nil
}

func (p *eventPublisher) reconnectLoop() {
	backoff := 1 * time.Second
	maxBackoff := 30 * time.Second

	for {
		select {
		case <-p.reconnectChan:

			if err := p.connect(); err != nil {
				log.Printf("Failed to reconnect: %v, retrying in %v", err, backoff)

				time.AfterFunc(backoff, func() {
					select {
					case p.reconnectChan <- struct{}{}:

					default:

					}
				})

				backoff = backoff * 2
				if backoff > maxBackoff {
					backoff = maxBackoff
				}
			} else {

				backoff = 1 * time.Second
			}
		case <-p.closeChan:
			return
		}
	}
}

func (p *eventPublisher) PublishEvent(routingKey string, event Event) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.connected {
		return fmt.Errorf("not connected to RabbitMQ, event will be lost: %+v", event)
	}

	if event.Time.IsZero() {
		event.Time = time.Now()
	}

	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("error marshaling event: %w", err)
	}

	contentType := event.ContentType
	if contentType == "" {
		contentType = "application/json"
	}

	msg := amqp.Publishing{
		ContentType:  contentType,
		Body:         body,
		Timestamp:    event.Time,
		MessageId:    event.ID,
		Type:         event.Type,
		DeliveryMode: amqp.Persistent,
	}

	err = p.channel.Publish(
		"events",
		routingKey,
		true,
		false,
		msg,
	)

	if err != nil {

		p.connected = false
		select {
		case p.reconnectChan <- struct{}{}:

		default:

		}
		return fmt.Errorf("error publishing event: %w", err)
	}

	log.Printf("Published event: %s with routing key: %s", event.Type, routingKey)
	return nil
}

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
