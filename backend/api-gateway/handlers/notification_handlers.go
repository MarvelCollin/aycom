package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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

	seedSampleNotifications()
}

func seedSampleNotifications() {

	sampleUserID := "user-123"

	likeData := map[string]interface{}{
		"username":       "alice_wonder",
		"display_name":   "Alice Wonder",
		"avatar":         "https://i.pravatar.cc/150?img=1",
		"thread_id":      "thread-456",
		"thread_content": "Just posted my first thread on AYCOM! 🎉",
	}
	notificationManager.AddNotification(sampleUserID, NotificationTypeLike, "liked your post", likeData)

	bookmarkData := map[string]interface{}{
		"username":       "bob_builder",
		"display_name":   "Bob Builder",
		"avatar":         "https://i.pravatar.cc/150?img=2",
		"thread_id":      "thread-789",
		"thread_content": "Amazing tutorial on React hooks!",
	}
	notificationManager.AddNotification(sampleUserID, "bookmark", "bookmarked your post", bookmarkData)

	replyData := map[string]interface{}{
		"username":       "charlie_code",
		"display_name":   "Charlie Code",
		"avatar":         "https://i.pravatar.cc/150?img=3",
		"thread_id":      "thread-101",
		"thread_content": "Great question! Here's my take on this...",
	}
	notificationManager.AddNotification(sampleUserID, NotificationTypeReply, "replied to your post", replyData)

	mentionData := map[string]interface{}{
		"username":       "diana_dev",
		"display_name":   "Diana Developer",
		"avatar":         "https://i.pravatar.cc/150?img=4",
		"thread_id":      "thread-202",
		"thread_content": "Hey @you, what do you think about this new framework?",
	}
	notificationManager.AddNotification(sampleUserID, NotificationTypeMention, "mentioned you in a post", mentionData)

	followData := map[string]interface{}{
		"username":     "evan_explorer",
		"display_name": "Evan Explorer",
		"avatar":       "https://i.pravatar.cc/150?img=5",
	}
	notificationManager.AddNotification(sampleUserID, NotificationTypeFollow, "started following you", followData)
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
	userID, exists := c.Get("userId")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	notifications := notificationManager.GetUserNotifications(userID.(string), 50, 0)

	var mentions []gin.H
	for _, notification := range notifications {
		if notification.Type == NotificationTypeMention {
			mentions = append(mentions, gin.H{
				"id":         notification.ID,
				"type":       notification.Type,
				"user_id":    notification.UserID,
				"content":    notification.Content,
				"data":       notification.Data,
				"is_read":    notification.Read,
				"timestamp":  notification.CreatedAt.Format(time.RFC3339),
				"created_at": notification.CreatedAt.Format(time.RFC3339),
			})
		}
	}

	utils.SendSuccessResponse(c, http.StatusOK, gin.H{
		"mentions": mentions,
		"pagination": gin.H{
			"total_count":  len(mentions),
			"current_page": 1,
			"per_page":     50,
			"has_more":     false,
		},
	})
}

func GetUserNotifications(c *gin.Context) {
	userID, exists := c.Get("userId")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	notifications := notificationManager.GetUserNotifications(userID.(string), 50, 0)

	var notificationList []gin.H
	for _, notification := range notifications {
		notificationData := gin.H{
			"id":         notification.ID,
			"type":       notification.Type,
			"user_id":    notification.UserID,
			"content":    notification.Content,
			"is_read":    notification.Read,
			"timestamp":  notification.CreatedAt.Format(time.RFC3339),
			"created_at": notification.CreatedAt.Format(time.RFC3339),
		}

		var data map[string]interface{}
		if err := json.Unmarshal(notification.Data, &data); err == nil {
			if username, ok := data["username"].(string); ok {
				notificationData["username"] = username
			}
			if displayName, ok := data["display_name"].(string); ok {
				notificationData["display_name"] = displayName
			}
			if avatar, ok := data["avatar"].(string); ok {
				notificationData["avatar"] = avatar
			}
			if threadID, ok := data["thread_id"].(string); ok {
				notificationData["thread_id"] = threadID
			}
			if threadContent, ok := data["thread_content"].(string); ok {
				notificationData["thread_content"] = threadContent
			}
		}

		notificationList = append(notificationList, notificationData)
	}

	utils.SendSuccessResponse(c, http.StatusOK, gin.H{
		"notifications": notificationList,
		"pagination": gin.H{
			"total_count":  len(notificationList),
			"current_page": 1,
			"per_page":     50,
			"has_more":     false,
		},
	})
}

func MarkNotificationAsRead(c *gin.Context) {
	notificationID := c.Param("id")
	if notificationID == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "BAD_REQUEST", "Notification ID is required")
		return
	}

	userID, exists := c.Get("userId")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	err := notificationManager.MarkNotificationAsRead(userID.(string), notificationID)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to mark notification as read")
		return
	}

	utils.SendSuccessResponse(c, http.StatusOK, gin.H{
		"message": "Notification marked as read successfully",
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

func GetUserInteractionNotifications(c *gin.Context) {
	userID, exists := c.Get("userId")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	userIDStr := userID.(string)
	log.Printf("Fetching interaction notifications for user: %s", userIDStr)

	likes, err := fetchUserLikes(userIDStr)
	if err != nil {
		log.Printf("Error fetching user likes: %v", err)
		likes = []gin.H{} 
	}

	bookmarks, err := fetchUserBookmarks(userIDStr)
	if err != nil {
		log.Printf("Error fetching user bookmarks: %v", err)
		bookmarks = []gin.H{} 
	}

	replies, err := fetchUserReplies(userIDStr)
	if err != nil {
		log.Printf("Error fetching user replies: %v", err)
		replies = []gin.H{} 
	}

	notifications := notificationManager.GetUserNotifications(userIDStr, 100, 0)
	var follows []gin.H
	for _, notification := range notifications {
		if notification.Type == NotificationTypeFollow {
			notificationData := gin.H{
				"id":         notification.ID,
				"type":       notification.Type,
				"user_id":    notification.UserID,
				"content":    notification.Content,
				"is_read":    notification.Read,
				"timestamp":  notification.CreatedAt.Format(time.RFC3339),
				"created_at": notification.CreatedAt.Format(time.RFC3339),
			}

			var data map[string]interface{}
			if err := json.Unmarshal(notification.Data, &data); err == nil {
				if username, ok := data["username"].(string); ok {
					notificationData["username"] = username
				}
				if displayName, ok := data["display_name"].(string); ok {
					notificationData["display_name"] = displayName
				}
				if avatar, ok := data["avatar"].(string); ok {
					notificationData["avatar"] = avatar
				}
			}
			follows = append(follows, notificationData)
		}
	}

	utils.SendSuccessResponse(c, http.StatusOK, gin.H{
		"interactions": gin.H{
			"likes":     likes,
			"bookmarks": bookmarks,
			"replies":   replies,
			"follows":   follows,
		},
		"total_count": len(likes) + len(bookmarks) + len(replies) + len(follows),
	})
}

func fetchUserLikes(userID string) ([]gin.H, error) {
	log.Printf("Fetching likes for user: %s", userID)

	resp, err := http.Get(fmt.Sprintf("http://localhost:8083/api/v1/threads/user/%s/likes", userID))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user likes: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch user likes: status %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	var response struct {
		Data struct {
			Threads []struct {
				ThreadID  string    `json:"thread_id"`
				Content   string    `json:"content"`
				UserID    string    `json:"user_id"`
				Username  string    `json:"username"`
				Name      string    `json:"name"`
				Avatar    string    `json:"profile_picture_url"`
				CreatedAt time.Time `json:"created_at"`
				LikedAt   time.Time `json:"liked_at"`
			} `json:"threads"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	var likes []gin.H
	for _, thread := range response.Data.Threads {
		likes = append(likes, gin.H{
			"id":             uuid.New().String(), 
			"type":           "like",
			"user_id":        thread.UserID,
			"username":       thread.Username,
			"display_name":   thread.Name,
			"avatar":         thread.Avatar,
			"thread_id":      thread.ThreadID,
			"thread_content": thread.Content,
			"is_read":        false, 
			"timestamp":      thread.LikedAt.Format(time.RFC3339),
			"created_at":     thread.LikedAt.Format(time.RFC3339),
		})
	}

	return likes, nil
}

func fetchUserBookmarks(userID string) ([]gin.H, error) {
	log.Printf("Fetching bookmarks for user: %s", userID)

	resp, err := http.Get("http://localhost:8083/api/v1/bookmarks")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user bookmarks: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch user bookmarks: status %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	var response struct {
		Data struct {
			Bookmarks []struct {
				ThreadID     string    `json:"thread_id"`
				Content      string    `json:"content"`
				UserID       string    `json:"user_id"`
				Username     string    `json:"username"`
				Name         string    `json:"name"`
				Avatar       string    `json:"profile_picture_url"`
				CreatedAt    time.Time `json:"created_at"`
				BookmarkedAt time.Time `json:"bookmarked_at"`
			} `json:"bookmarks"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	var bookmarks []gin.H
	for _, bookmark := range response.Data.Bookmarks {
		bookmarks = append(bookmarks, gin.H{
			"id":             uuid.New().String(), 
			"type":           "bookmark",
			"user_id":        bookmark.UserID,
			"username":       bookmark.Username,
			"display_name":   bookmark.Name,
			"avatar":         bookmark.Avatar,
			"thread_id":      bookmark.ThreadID,
			"thread_content": bookmark.Content,
			"is_read":        false, 
			"timestamp":      bookmark.BookmarkedAt.Format(time.RFC3339),
			"created_at":     bookmark.BookmarkedAt.Format(time.RFC3339),
		})
	}

	return bookmarks, nil
}

func fetchUserReplies(userID string) ([]gin.H, error) {
	log.Printf("Fetching replies for user: %s", userID)

	resp, err := http.Get(fmt.Sprintf("http://localhost:8083/api/v1/threads/user/%s/replies", userID))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user replies: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch user replies: status %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	var response struct {
		Data struct {
			Replies []struct {
				ReplyID       string    `json:"reply_id"`
				ThreadID      string    `json:"thread_id"`
				Content       string    `json:"content"`
				UserID        string    `json:"user_id"`
				Username      string    `json:"username"`
				Name          string    `json:"name"`
				Avatar        string    `json:"profile_picture_url"`
				CreatedAt     time.Time `json:"created_at"`
				ThreadContent string    `json:"thread_content"`
			} `json:"replies"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	var replies []gin.H
	for _, reply := range response.Data.Replies {
		replies = append(replies, gin.H{
			"id":             uuid.New().String(), 
			"type":           "reply",
			"user_id":        reply.UserID,
			"username":       reply.Username,
			"display_name":   reply.Name,
			"avatar":         reply.Avatar,
			"thread_id":      reply.ThreadID,
			"thread_content": reply.ThreadContent,
			"reply_content":  reply.Content,
			"is_read":        false, 
			"timestamp":      reply.CreatedAt.Format(time.RFC3339),
			"created_at":     reply.CreatedAt.Format(time.RFC3339),
		})
	}

	return replies, nil
}