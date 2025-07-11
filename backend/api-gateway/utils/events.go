package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/streadway/amqp"
)

type EventData map[string]interface{}

type Event struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	Source      string    `json:"source"`
	Time        time.Time `json:"time"`
	Data        EventData `json:"data"`
	ContentType string    `json:"content_type"`
}

type EventPublisher interface {
	PublishEvent(routingKey string, event Event) error
	Close() error
}

type rabbitMQEventPublisher struct {
	connection    *amqp.Connection
	channel       *amqp.Channel
	url           string
	connected     bool
	mu            sync.Mutex
	reconnectChan chan struct{}
	closeChan     chan struct{}
}

var (
	eventPublisher EventPublisher
	publisherOnce  sync.Once
)

func InitEventPublisher(rabbitMQURL string) error {
	var err error
	publisherOnce.Do(func() {
		eventPublisher, err = NewEventPublisher(rabbitMQURL)
	})
	return err
}

func GetEventPublisher() EventPublisher {
	return eventPublisher
}

func NewEventPublisher(url string) (EventPublisher, error) {
	publisher := &rabbitMQEventPublisher{
		url:           url,
		connected:     false,
		reconnectChan: make(chan struct{}, 1),
		closeChan:     make(chan struct{}, 1),
	}

	if err := publisher.connect(); err != nil {
		log.Printf("Failed to establish initial RabbitMQ connection: %v, will retry in background", err)
		go publisher.reconnectLoop()
		publisher.reconnectChan <- struct{}{}
	}

	return publisher, nil
}

func (p *rabbitMQEventPublisher) connect() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	var err error
	p.connection, err = amqp.Dial(p.url)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	p.channel, err = p.connection.Channel()
	if err != nil {
		p.connection.Close()
		return fmt.Errorf("failed to open RabbitMQ channel: %w", err)
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
		p.connection.Close()
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	p.connected = true
	log.Println("Successfully connected to RabbitMQ")
	return nil
}

func (p *rabbitMQEventPublisher) reconnectLoop() {
	for {
		select {
		case <-p.reconnectChan:
			for !p.connected {
				log.Println("RabbitMQ connection lost, reconnecting...")
				if err := p.connect(); err != nil {
					log.Printf("Failed to reconnect to RabbitMQ: %v", err)
					time.Sleep(5 * time.Second)
				}
			}
		case <-p.closeChan:
			return
		}
	}
}

func (p *rabbitMQEventPublisher) PublishEvent(routingKey string, event Event) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.connected {
		return fmt.Errorf("not connected to RabbitMQ, event will be lost: %+v", event)
	}

	
	if event.ID == "" {
		event.ID = uuid.New().String()
	}
	if event.Time.IsZero() {
		event.Time = time.Now()
	}
	if event.Source == "" {
		event.Source = "api_gateway"
	}
	if event.ContentType == "" {
		event.ContentType = "application/json"
	}

	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	err = p.channel.Publish(
		"events",
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)

	if err != nil {
		log.Printf("Failed to publish event: %v", err)
		p.connected = false
		go func() {
			p.reconnectChan <- struct{}{}
		}()
		return fmt.Errorf("failed to publish event: %w", err)
	}

	log.Printf("Successfully published event with routing key: %s", routingKey)
	return nil
}

func (p *rabbitMQEventPublisher) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.closeChan <- struct{}{}

	if p.channel != nil {
		p.channel.Close()
	}
	if p.connection != nil {
		p.connection.Close()
	}
	p.connected = false
	return nil
}


func PublishUserFollowedEvent(followerID, followedID string, additionalData EventData) error {
	if eventPublisher == nil {
		return fmt.Errorf("event publisher not initialized")
	}

	data := EventData{
		"follower_id": followerID,
		"followed_id": followedID,
		"action":      "follow",
	}

	
	for k, v := range additionalData {
		data[k] = v
	}

	event := Event{
		Type: "user.followed",
		Data: data,
	}

	return eventPublisher.PublishEvent("user.followed", event)
}

func PublishUserUnfollowedEvent(followerID, followedID string, additionalData EventData) error {
	if eventPublisher == nil {
		return fmt.Errorf("event publisher not initialized")
	}

	data := EventData{
		"follower_id": followerID,
		"followed_id": followedID,
		"action":      "unfollow",
	}

	
	for k, v := range additionalData {
		data[k] = v
	}

	event := Event{
		Type: "user.unfollowed",
		Data: data,
	}

	return eventPublisher.PublishEvent("user.unfollowed", event)
}

func PublishThreadLikedEvent(threadID, userID string, additionalData EventData) error {
	if eventPublisher == nil {
		return fmt.Errorf("event publisher not initialized")
	}

	data := EventData{
		"thread_id": threadID,
		"user_id":   userID,
		"action":    "like",
	}

	
	for k, v := range additionalData {
		data[k] = v
	}

	event := Event{
		Type: "thread.liked",
		Data: data,
	}

	return eventPublisher.PublishEvent("thread.liked", event)
}

func PublishThreadUnlikedEvent(threadID, userID string, additionalData EventData) error {
	if eventPublisher == nil {
		return fmt.Errorf("event publisher not initialized")
	}

	data := EventData{
		"thread_id": threadID,
		"user_id":   userID,
		"action":    "unlike",
	}

	
	for k, v := range additionalData {
		data[k] = v
	}

	event := Event{
		Type: "thread.unliked",
		Data: data,
	}

	return eventPublisher.PublishEvent("thread.unliked", event)
}

func PublishThreadBookmarkedEvent(threadID, userID string, additionalData EventData) error {
	if eventPublisher == nil {
		return fmt.Errorf("event publisher not initialized")
	}

	data := EventData{
		"thread_id": threadID,
		"user_id":   userID,
		"action":    "bookmark",
	}

	
	for k, v := range additionalData {
		data[k] = v
	}

	event := Event{
		Type: "thread.bookmarked",
		Data: data,
	}

	return eventPublisher.PublishEvent("thread.bookmarked", event)
}

func PublishThreadUnbookmarkedEvent(threadID, userID string, additionalData EventData) error {
	if eventPublisher == nil {
		return fmt.Errorf("event publisher not initialized")
	}

	data := EventData{
		"thread_id": threadID,
		"user_id":   userID,
		"action":    "unbookmark",
	}

	
	for k, v := range additionalData {
		data[k] = v
	}

	event := Event{
		Type: "thread.unbookmarked",
		Data: data,
	}

	return eventPublisher.PublishEvent("thread.unbookmarked", event)
}
