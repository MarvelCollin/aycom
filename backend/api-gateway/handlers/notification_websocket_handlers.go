package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// NotificationClient represents a WebSocket client specifically for notifications
type NotificationClient struct {
	ID         string
	UserID     string
	Connection *websocket.Conn
	Send       chan []byte
}

// HandleNotificationsWebSocket handles WebSocket connections for notifications
func HandleNotificationsWebSocket(c *gin.Context) {
	userID, exists := c.Get("userId")
	if !exists {
		SendErrorResponse(c, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to set websocket upgrade: %v", err)
		return
	}

	// Create a notification-specific client
	client := &NotificationClient{
		ID:         uuid.New().String(),
		UserID:     userID.(string),
		Connection: conn,
		Send:       make(chan []byte, Config.WebSocket.SendBufferSize),
	}

	// Register with WebSocket manager
	manager := GetWebSocketManager()
	wsClient := &Client{
		ID:         client.ID,
		UserID:     client.UserID,
		Connection: client.Connection,
		ChatID:     "", // No specific chat for notifications
		Send:       client.Send,
		Manager:    manager,
	}
	manager.register <- wsClient

	// Notify the notification manager that this user is connected
	notificationManager.UserConnected(client.UserID)

	// Start goroutines for reading and writing
	go client.notificationWritePump()
	go client.notificationReadPump(wsClient)

	// Send any unread notifications to the client
	go sendUnreadNotifications(client)
}

// notificationReadPump reads messages from the notification WebSocket connection
func (c *NotificationClient) notificationReadPump(wsClient *Client) {
	defer func() {
		// Unregister with WebSocket manager
		manager := GetWebSocketManager()
		manager.unregister <- wsClient

		// Notify the notification manager that this user is disconnected
		notificationManager.UserDisconnected(c.UserID)

		c.Connection.Close()
	}()

	c.Connection.SetReadLimit(int64(Config.WebSocket.MaxMessageSize))
	c.Connection.SetReadDeadline(time.Now().Add(Config.WebSocket.ReadDeadlineTimeout))
	c.Connection.SetPongHandler(func(string) error {
		c.Connection.SetReadDeadline(time.Now().Add(Config.WebSocket.ReadDeadlineTimeout))
		return nil
	})

	for {
		_, message, err := c.Connection.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Error: %v", err)
			}
			break
		}

		// Process client actions for notifications (e.g., mark as read)
		processNotificationAction(message, c.UserID)
	}
}

// notificationWritePump writes messages to the notification WebSocket connection
func (c *NotificationClient) notificationWritePump() {
	ticker := time.NewTicker(Config.WebSocket.PingInterval)
	defer func() {
		ticker.Stop()
		c.Connection.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Connection.SetWriteDeadline(time.Now().Add(Config.WebSocket.WriteDeadlineTimeout))
			if !ok {
				c.Connection.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Connection.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued notification messages to the current websocket message
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Connection.SetWriteDeadline(time.Now().Add(Config.WebSocket.WriteDeadlineTimeout))
			if err := c.Connection.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// processNotificationAction processes actions from clients (e.g., marking notifications as read)
func processNotificationAction(message []byte, userID string) {
	// Parse the action
	var action struct {
		Type           string `json:"type"`
		NotificationID string `json:"notification_id"`
	}

	if err := json.Unmarshal(message, &action); err != nil {
		log.Printf("Error parsing notification action: %v", err)
		return
	}

	switch action.Type {
	case "mark_read":
		// Mark notification as read
		if err := notificationManager.MarkNotificationAsRead(userID, action.NotificationID); err != nil {
			log.Printf("Error marking notification as read: %v", err)
		}
	}
}

// sendUnreadNotifications sends all unread notifications to a newly connected client
func sendUnreadNotifications(client *NotificationClient) {
	// Get unread notifications for the user
	notifications := notificationManager.GetUserNotifications(client.UserID, 50, 0)

	// Only send unread notifications
	var unreadNotifications []Notification
	for _, notification := range notifications {
		if !notification.Read {
			unreadNotifications = append(unreadNotifications, notification)
		}
	}

	if len(unreadNotifications) == 0 {
		return
	}

	// Create a notification bundle
	notificationBundle := struct {
		Type          string         `json:"type"`
		Notifications []Notification `json:"notifications"`
	}{
		Type:          "notification_bundle",
		Notifications: unreadNotifications,
	}

	// Serialize the bundle
	bundle, err := json.Marshal(notificationBundle)
	if err != nil {
		log.Printf("Error serializing notification bundle: %v", err)
		return
	}

	// Send the bundle to the client
	select {
	case client.Send <- bundle:
		// Bundle sent successfully
	default:
		log.Printf("Error: notification channel is full for user %s", client.UserID)
	}
}

// GetUserNotifications handles the API request to get a user's notifications
func GetUserNotifications(c *gin.Context) {
	userID, exists := c.Get("userId")
	if !exists {
		SendErrorResponse(c, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 10
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// Get the notifications for the user
	notifications := notificationManager.GetUserNotifications(userID.(string), limit, offset)

	SendSuccessResponse(c, http.StatusOK, gin.H{
		"notifications": notifications,
		"limit":         limit,
		"offset":        offset,
		"total":         len(notifications), // This is the returned count, not total in DB
	})
}

// MarkNotificationAsRead marks a notification as read
func MarkNotificationAsRead(c *gin.Context) {
	userID, exists := c.Get("userId")
	if !exists {
		SendErrorResponse(c, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	notificationID := c.Param("id")
	if notificationID == "" {
		SendErrorResponse(c, http.StatusBadRequest, "invalid_request", "Notification ID is required")
		return
	}

	// Mark the notification as read
	err := notificationManager.MarkNotificationAsRead(userID.(string), notificationID)
	if err != nil {
		SendErrorResponse(c, http.StatusInternalServerError, "server_error", "Failed to mark notification as read")
		return
	}

	SendSuccessResponse(c, http.StatusOK, gin.H{"success": true})
}
