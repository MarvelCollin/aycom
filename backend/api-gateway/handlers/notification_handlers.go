package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"aycom/backend/api-gateway/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// NotificationType defines the types of notifications
type NotificationType string

// Notification types
const (
	NotificationTypeMessage   NotificationType = "message"
	NotificationTypeLike      NotificationType = "like"
	NotificationTypeFollow    NotificationType = "follow"
	NotificationTypeReply     NotificationType = "reply"
	NotificationTypeMention   NotificationType = "mention"
	NotificationTypeCommunity NotificationType = "community"
	NotificationTypeSystem    NotificationType = "system"
)

// Notification represents a user notification
type Notification struct {
	ID        string           `json:"id"`
	UserID    string           `json:"user_id"`
	Type      NotificationType `json:"type"`
	Content   string           `json:"content"`
	Data      json.RawMessage  `json:"data,omitempty"`
	Read      bool             `json:"read"`
	CreatedAt time.Time        `json:"created_at"`
}

// NotificationManager manages user notifications
type NotificationManager struct {
	userNotifications map[string][]Notification // userID -> notifications
	userConnections   map[string]bool           // userID -> hasActiveConnection
	mutex             sync.RWMutex
}

var (
	// Global notification manager
	notificationManager *NotificationManager
)

// Initialize the notification manager
func init() {
	notificationManager = NewNotificationManager()
}

// NewNotificationManager creates a new notification manager
func NewNotificationManager() *NotificationManager {
	return &NotificationManager{
		userNotifications: make(map[string][]Notification),
		userConnections:   make(map[string]bool),
	}
}

// AddNotification adds a notification for a user
func (nm *NotificationManager) AddNotification(userID string, notificationType NotificationType, content string, data interface{}) (string, error) {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	notificationID := uuid.New().String()
	notification := Notification{
		ID:        notificationID,
		UserID:    userID,
		Type:      notificationType,
		Content:   content,
		Data:      dataBytes,
		Read:      false,
		CreatedAt: time.Now(),
	}

	nm.mutex.Lock()
	defer nm.mutex.Unlock()

	if _, ok := nm.userNotifications[userID]; !ok {
		nm.userNotifications[userID] = []Notification{}
	}
	nm.userNotifications[userID] = append(nm.userNotifications[userID], notification)

	// Send notification if user is connected
	if nm.userConnections[userID] {
		// Use the WebSocket manager to send the notification
		broadcastNotificationToUser(userID, notification)
	}

	return notificationID, nil
}

// MarkNotificationAsRead marks a notification as read
func (nm *NotificationManager) MarkNotificationAsRead(userID, notificationID string) error {
	nm.mutex.Lock()
	defer nm.mutex.Unlock()

	notifications, ok := nm.userNotifications[userID]
	if !ok {
		return nil // No notifications for this user
	}

	for i := range notifications {
		if notifications[i].ID == notificationID {
			notifications[i].Read = true
			nm.userNotifications[userID] = notifications
			return nil
		}
	}

	return nil // Notification not found
}

// GetUserNotifications gets all notifications for a user
func (nm *NotificationManager) GetUserNotifications(userID string, limit, offset int) []Notification {
	nm.mutex.RLock()
	defer nm.mutex.RUnlock()

	notifications, ok := nm.userNotifications[userID]
	if !ok {
		return []Notification{}
	}

	if offset >= len(notifications) {
		return []Notification{}
	}

	end := offset + limit
	if end > len(notifications) {
		end = len(notifications)
	}

	return notifications[offset:end]
}

// UserConnected marks a user as connected
func (nm *NotificationManager) UserConnected(userID string) {
	nm.mutex.Lock()
	defer nm.mutex.Unlock()
	nm.userConnections[userID] = true
}

// UserDisconnected marks a user as disconnected
func (nm *NotificationManager) UserDisconnected(userID string) {
	nm.mutex.Lock()
	defer nm.mutex.Unlock()
	nm.userConnections[userID] = false
}

// SendNotification sends a notification to a user
func SendNotification(userID string, notificationType NotificationType, content string, data interface{}) (string, error) {
	return notificationManager.AddNotification(userID, notificationType, content, data)
}

// broadcastNotificationToUser sends a notification to a specific user
func broadcastNotificationToUser(userID string, notification Notification) {
	// Get WebSocket manager instance
	wsManager := GetWebSocketManager()

	// For now, we don't have a direct way to send to specific users
	// This is a simplified implementation - we'd need to modify WebSocketManager
	// to support direct user notifications
	log.Printf("Would send notification to user %s: %v", userID, notification)

	// Find any clients for this user in any chat rooms
	for _, client := range wsManager.clients {
		if client.UserID == userID {
			// Create a notification message
			notificationMessage := struct {
				Type         string      `json:"type"`
				Notification interface{} `json:"notification"`
			}{
				Type:         "notification",
				Notification: notification,
			}

			// Serialize the message
			messageData, err := json.Marshal(notificationMessage)
			if err != nil {
				log.Printf("Error serializing notification message: %v", err)
				continue
			}

			// Send the notification to the client
			select {
			case client.Send <- messageData:
				// Message sent successfully
				log.Printf("Sent notification to user %s", userID)
			default:
				// Failed to send message, channel might be full
				log.Printf("Failed to send notification to user %s", userID)
			}
			break // Send to only one client per user
		}
	}
}

func GetMentionNotifications(c *gin.Context) {
	// Get current user ID from JWT token
	_, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// TODO: Get mention notifications from notification service
	// Placeholder implementation
	c.JSON(http.StatusOK, gin.H{
		"notifications": []gin.H{},
		"pagination": gin.H{
			"total":   0,
			"page":    1,
			"limit":   10,
			"hasMore": false,
		},
	})
}

func MarkAllNotificationsAsRead(c *gin.Context) {
	// Get current user ID from JWT token
	_, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// TODO: Mark all notifications as read for this user
	// Placeholder implementation
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "All notifications marked as read",
	})
}

func DeleteNotification(c *gin.Context) {
	// Get notification ID from path
	notificationID := c.Param("id")
	if notificationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Notification ID is required"})
		return
	}

	// Get current user ID from JWT token
	_, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// TODO: Delete notification
	// Placeholder implementation
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Notification deleted",
	})
}

func UpdateNotificationSettings(c *gin.Context) {
	// Get current user ID from JWT token
	_, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// TODO: Implement actual settings update
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Notification settings updated successfully",
	})
}

func UpdateNotificationStatus(c *gin.Context) {
	// Get current user ID from JWT token
	_, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// TODO: Implement actual status update
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Notification status updated successfully",
	})
}

func GetNotificationPreferences(c *gin.Context) {
	// Get current user ID from JWT token
	_, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// TODO: Implement getting preferences from service
	prefs := models.NotificationPreferences{
		Likes:          true,
		Comments:       true,
		Follows:        true,
		Mentions:       true,
		DirectMessages: true,
	}

	c.JSON(http.StatusOK, models.NotificationPreferencesResponse{
		Preferences: prefs,
	})
}
