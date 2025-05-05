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

	// Check if user has any active connections
	wsManager.mutex.RLock()
	clientID, exists := wsManager.userToClient[userID]
	wsManager.mutex.RUnlock()

	if !exists {
		return
	}

	// Get the client connection
	wsManager.mutex.RLock()
	client, exists := wsManager.clients[clientID]
	wsManager.mutex.RUnlock()

	if !exists {
		return
	}

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
		return
	}

	// Send the notification to the client
	select {
	case client.Send <- messageData:
		// Message sent successfully
	default:
		// Failed to send message, client might be disconnected
		log.Printf("Failed to send notification to user %s", userID)
	}
}

// GetMentionNotifications handles fetching mention notifications for the current user
// @Summary Get mention notifications
// @Description Get notifications where the user is mentioned
// @Tags Notifications
// @Accept json
// @Produce json
// @Param page query int false "Page number for pagination"
// @Param limit query int false "Number of results per page"
// @Success 200 {object} models.NotificationsResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/notifications/mentions [get]
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

// MarkAllNotificationsAsRead marks all notifications for the current user as read
// @Summary Mark all notifications as read
// @Description Mark all notifications for the current user as read
// @Tags Notifications
// @Accept json
// @Produce json
// @Success 200 {object} models.SuccessResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/notifications/read-all [post]
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

// DeleteNotification deletes a notification
// @Summary Delete a notification
// @Description Delete a specific notification by ID
// @Tags Notifications
// @Accept json
// @Produce json
// @Param id path string true "Notification ID"
// @Success 200 {object} models.SuccessResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/notifications/{id} [delete]
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

// UpdateNotificationSettings updates the notification settings for the user
// @Summary Update notification settings
// @Description Update the notification settings for the current user
// @Tags Notifications
// @Accept json
// @Produce json
// @Success 200 {object} models.SuccessResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/notifications/settings [post]
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

// UpdateNotificationStatus updates the read status of a notification
// @Summary Update notification status
// @Description Update the read status of a specific notification
// @Tags Notifications
// @Accept json
// @Produce json
// @Param id path string true "Notification ID"
// @Success 200 {object} models.SuccessResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/notifications/{id}/status [post]
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

// GetNotificationPreferences gets the user's notification preferences
// @Summary Get notification preferences
// @Description Get the notification preferences for the current user
// @Tags Notifications
// @Accept json
// @Produce json
// @Success 200 {object} models.NotificationPreferencesResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/notifications/preferences [get]
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
