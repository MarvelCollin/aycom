package handlers

import (
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/streadway/amqp"

	"aycom/backend/event-bus/publisher"

)

type UserEventHandler struct {
	publisher publisher.EventPublisher
	conn      *amqp.Connection
	channel   *amqp.Channel
}

func NewUserEventHandler(pub publisher.EventPublisher) *UserEventHandler {
	return &UserEventHandler{
		publisher: pub,
	}
}

func (h *UserEventHandler) Start() error {

	var err error
	retries := 0
	maxRetries := 5

	for retries < maxRetries {

		h.conn, err = amqp.Dial(getEnvFromHandler("RABBITMQ_URL", "amqp:
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
		"user_events",
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
		"user.*",
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

	log.Println("User event handler started successfully")
	return nil
}

func (h *UserEventHandler) reconnect() error {
	h.cleanup()
	return h.Start()
}

func (h *UserEventHandler) cleanup() {
	if h.channel != nil {
		h.channel.Close()
	}
	if h.conn != nil {
		h.conn.Close()
	}
}

func (h *UserEventHandler) processMessage(msg amqp.Delivery) {
	log.Printf("Received user event with routing key: %s", msg.RoutingKey)

	var event publisher.Event
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		log.Printf("Error unmarshaling event: %v", err)
		return
	}
	switch msg.RoutingKey {
	case "user.created":
		h.handleUserCreated(event)
	case "user.updated":
		h.handleUserUpdated(event)
	case "user.deleted":
		h.handleUserDeleted(event)
	case "user.followed":
		h.handleUserFollowed(event)
	case "user.unfollowed":
		h.handleUserUnfollowed(event)
	default:
		log.Printf("Unhandled user event type: %s", msg.RoutingKey)
	}
}

func (h *UserEventHandler) handleUserCreated(event publisher.Event) {
	log.Printf("Processing user created event: %v", event)

	if event.Source != "auth_service" {
		return
	}

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

func (h *UserEventHandler) handleUserUpdated(event publisher.Event) {
	log.Printf("Processing user updated event: %v", event)

}

func (h *UserEventHandler) handleUserDeleted(event publisher.Event) {
	log.Printf("Processing user deleted event: %v", event)
}

func (h *UserEventHandler) handleUserFollowed(event publisher.Event) {
	log.Printf("Processing user followed event: %v", event)

	followerID, ok := event.Data["follower_id"].(string)
	if !ok {
		log.Printf("Error: follower_id is missing or not a string")
		return
	}

	followedID, ok := event.Data["followed_id"].(string)
	if !ok {
		log.Printf("Error: followed_id is missing or not a string")
		return
	}

	log.Printf("User %s followed user %s", followerID, followedID)

	
	
	
	
}

func (h *UserEventHandler) handleUserUnfollowed(event publisher.Event) {
	log.Printf("Processing user unfollowed event: %v", event)

	followerID, ok := event.Data["follower_id"].(string)
	if !ok {
		log.Printf("Error: follower_id is missing or not a string")
		return
	}

	followedID, ok := event.Data["followed_id"].(string)
	if !ok {
		log.Printf("Error: followed_id is missing or not a string")
		return
	}

	log.Printf("User %s unfollowed user %s", followerID, followedID)

	
	
	
	
}
