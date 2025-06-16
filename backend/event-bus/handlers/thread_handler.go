package handlers

import (
	"encoding/json"
	"log"
	"time"

	"aycom/backend/event-bus/publisher"

	"github.com/streadway/amqp"
)

type ThreadEventHandler struct {
	publisher publisher.EventPublisher
	conn      *amqp.Connection
	channel   *amqp.Channel
}

func NewThreadEventHandler(pub publisher.EventPublisher) *ThreadEventHandler {
	return &ThreadEventHandler{
		publisher: pub,
	}
}

func (h *ThreadEventHandler) Start() error {
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
		"thread_events",
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
		"thread.*",
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
					log.Println("Thread consumer channel closed, exiting consumer loop")
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

	log.Println("Thread event handler started successfully")
	return nil
}

func (h *ThreadEventHandler) reconnect() error {
	h.cleanup()
	return h.Start()
}

func (h *ThreadEventHandler) cleanup() {
	if h.channel != nil {
		h.channel.Close()
	}
	if h.conn != nil {
		h.conn.Close()
	}
}

func (h *ThreadEventHandler) processMessage(msg amqp.Delivery) {
	log.Printf("Received thread event with routing key: %s", msg.RoutingKey)

	var event publisher.Event
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		log.Printf("Error unmarshaling event: %v", err)
		return
	}

	switch msg.RoutingKey {
	case "thread.liked":
		h.handleThreadLiked(event)
	case "thread.unliked":
		h.handleThreadUnliked(event)
	case "thread.bookmarked":
		h.handleThreadBookmarked(event)
	case "thread.unbookmarked":
		h.handleThreadUnbookmarked(event)
	case "thread.created":
		h.handleThreadCreated(event)
	case "thread.updated":
		h.handleThreadUpdated(event)
	case "thread.deleted":
		h.handleThreadDeleted(event)
	default:
		log.Printf("Unhandled thread event type: %s", msg.RoutingKey)
	}
}

func (h *ThreadEventHandler) handleThreadLiked(event publisher.Event) {
	log.Printf("Processing thread liked event: %v", event)

	threadID, ok := event.Data["thread_id"].(string)
	if !ok {
		log.Printf("Error: thread_id is missing or not a string")
		return
	}

	userID, ok := event.Data["user_id"].(string)
	if !ok {
		log.Printf("Error: user_id is missing or not a string")
		return
	}

	log.Printf("Thread %s was liked by user %s", threadID, userID)

	// Here you can add logic to:
	// - Update analytics/metrics
	// - Send notifications to thread author
	// - Update recommendation algorithms
	// - Invalidate cache for thread stats
}

func (h *ThreadEventHandler) handleThreadUnliked(event publisher.Event) {
	log.Printf("Processing thread unliked event: %v", event)

	threadID, ok := event.Data["thread_id"].(string)
	if !ok {
		log.Printf("Error: thread_id is missing or not a string")
		return
	}

	userID, ok := event.Data["user_id"].(string)
	if !ok {
		log.Printf("Error: user_id is missing or not a string")
		return
	}

	log.Printf("Thread %s was unliked by user %s", threadID, userID)

	// Here you can add logic to:
	// - Update analytics/metrics
	// - Update recommendation algorithms
	// - Invalidate cache for thread stats
}

func (h *ThreadEventHandler) handleThreadBookmarked(event publisher.Event) {
	log.Printf("Processing thread bookmarked event: %v", event)

	threadID, ok := event.Data["thread_id"].(string)
	if !ok {
		log.Printf("Error: thread_id is missing or not a string")
		return
	}

	userID, ok := event.Data["user_id"].(string)
	if !ok {
		log.Printf("Error: user_id is missing or not a string")
		return
	}

	log.Printf("Thread %s was bookmarked by user %s", threadID, userID)

	// Here you can add logic to:
	// - Update user preferences for recommendations
	// - Send notifications to thread author
	// - Update analytics
	// - Invalidate cache for user bookmarks
}

func (h *ThreadEventHandler) handleThreadUnbookmarked(event publisher.Event) {
	log.Printf("Processing thread unbookmarked event: %v", event)

	threadID, ok := event.Data["thread_id"].(string)
	if !ok {
		log.Printf("Error: thread_id is missing or not a string")
		return
	}

	userID, ok := event.Data["user_id"].(string)
	if !ok {
		log.Printf("Error: user_id is missing or not a string")
		return
	}

	log.Printf("Thread %s was unbookmarked by user %s", threadID, userID)

	// Here you can add logic to:
	// - Update user preferences
	// - Update analytics
	// - Invalidate cache for user bookmarks
}

func (h *ThreadEventHandler) handleThreadCreated(event publisher.Event) {
	log.Printf("Processing thread created event: %v", event)

	// Logic for when a new thread is created
	// - Send notifications to followers
	// - Update community statistics
	// - Trigger content moderation
}

func (h *ThreadEventHandler) handleThreadUpdated(event publisher.Event) {
	log.Printf("Processing thread updated event: %v", event)

	// Logic for when a thread is updated
	// - Invalidate caches
	// - Re-trigger content moderation if needed
}

func (h *ThreadEventHandler) handleThreadDeleted(event publisher.Event) {
	log.Printf("Processing thread deleted event: %v", event)

	// Logic for when a thread is deleted
	// - Clean up related data
	// - Update statistics
	// - Invalidate caches
}
