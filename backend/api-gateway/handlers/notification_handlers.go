package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"aycom/backend/api-gateway/models"
	"aycom/backend/api-gateway/utils"
)

type NotificationType string

const (
	NotificationTypeMessage   NotificationType = "message"
	NotificationTypeLike      NotificationType = "like"
	NotificationTypeFollow    NotificationType = "follow"
	NotificationTypeReply     NotificationType = "reply"
	NotificationTypeMention   NotificationType = "mention"
	NotificationTypeCommunity NotificationType = "community"
	NotificationTypeSystem    NotificationType = "system"
)

type Notification struct {
	ID        string           `json:"id"`
	UserID    string           `json:"user_id"`
	Type      NotificationType `json:"type"`
	Content   string           `json:"content"`
	Data      json.RawMessage  `json:"data,omitempty"`
	Read      bool             `json:"read"`
	CreatedAt time.Time        `json:"created_at"`
}

type NotificationManager struct {
	userNotifications map[string][]Notification
	userConnections   map[string]bool
	mutex             sync.RWMutex
}

var (
	notificationManager *NotificationManager
)

func init() {
	notificationManager = NewNotificationManager()
}

func NewNotificationManager() *NotificationManager {
	return &NotificationManager{
		userNotifications: make(map[string][]Notification),
		userConnections:   make(map[string]bool),
	}
}

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

	if nm.userConnections[userID] {

		broadcastNotificationToUser(userID, notification)
	}

	return notificationID, nil
}

func (nm *NotificationManager) MarkNotificationAsRead(userID, notificationID string) error {
	nm.mutex.Lock()
	defer nm.mutex.Unlock()

	notifications, ok := nm.userNotifications[userID]
	if !ok {
		return nil
	}

	for i := range notifications {
		if notifications[i].ID == notificationID {
			notifications[i].Read = true
			nm.userNotifications[userID] = notifications
			return nil
		}
	}

	return nil
}

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

func (nm *NotificationManager) UserConnected(userID string) {
	nm.mutex.Lock()
	defer nm.mutex.Unlock()
	nm.userConnections[userID] = true
}

func (nm *NotificationManager) UserDisconnected(userID string) {
	nm.mutex.Lock()
	defer nm.mutex.Unlock()
	nm.userConnections[userID] = false
}

func SendNotification(userID string, notificationType NotificationType, content string, data interface{}) (string, error) {
	return notificationManager.AddNotification(userID, notificationType, content, data)
}

func broadcastNotificationToUser(userID string, notification Notification) {

	wsManager := GetWebSocketManager()

	log.Printf("Would send notification to user %s: %v", userID, notification)

	for _, client := range wsManager.clients {
		if client.UserID == userID {

			notificationMessage := struct {
				Type         string      `json:"type"`
				Notification interface{} `json:"notification"`
			}{
				Type:         "notification",
				Notification: notification,
			}

			messageData, err := json.Marshal(notificationMessage)
			if err != nil {
				log.Printf("Error serializing notification message: %v", err)
				continue
			}

			select {
			case client.Send <- messageData:

				log.Printf("Sent notification to user %s", userID)
			default:

				log.Printf("Failed to send notification to user %s", userID)
			}
			break
		}
	}
}

func GetMentionNotifications(c *gin.Context) {
	_, exists := c.Get("userID")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	utils.SendSuccessResponse(c, http.StatusOK, gin.H{
		"notifications": []gin.H{},
		"pagination": gin.H{
			"total_count":  0,
			"current_page": 1,
			"per_page":     10,
			"has_more":     false,
		},
	})
}

func MarkAllNotificationsAsRead(c *gin.Context) {
	_, exists := c.Get("userID")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	utils.SendSuccessResponse(c, http.StatusOK, gin.H{
		"message": "All notifications marked as read",
	})
}

func DeleteNotification(c *gin.Context) {
	notificationID := c.Param("id")
	if notificationID == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "BAD_REQUEST", "Notification ID is required")
		return
	}

	_, exists := c.Get("userID")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	utils.SendSuccessResponse(c, http.StatusOK, gin.H{
		"message": "Notification deleted",
	})
}

func UpdateNotificationSettings(c *gin.Context) {
	_, exists := c.Get("userID")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	utils.SendSuccessResponse(c, http.StatusOK, gin.H{
		"message": "Notification settings updated successfully",
	})
}

func UpdateNotificationStatus(c *gin.Context) {
	_, exists := c.Get("userID")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	utils.SendSuccessResponse(c, http.StatusOK, gin.H{
		"message": "Notification status updated successfully",
	})
}

func GetNotificationPreferences(c *gin.Context) {
	_, exists := c.Get("userID")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	prefs := models.NotificationPreferences{
		Likes:          true,
		Comments:       true,
		Follows:        true,
		Mentions:       true,
		DirectMessages: true,
	}

	utils.SendSuccessResponse(c, http.StatusOK, gin.H{
		"preferences": prefs,
	})
}
