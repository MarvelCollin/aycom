package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"aycom/backend/api-gateway/utils"
)

type NotificationClient struct {
	ID         string
	UserID     string
	Connection *websocket.Conn
	Send       chan []byte
}

func HandleNotificationsWebSocket(c *gin.Context) {
	log.Printf("WebSocket connection request received for notifications from IP: %s", c.ClientIP())

	
	origin := c.Request.Header.Get("Origin")
	if origin == "" {
		origin = "*"
	}

	c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")

	
	if c.Request.Method == "OPTIONS" {
		c.AbortWithStatus(http.StatusNoContent)
		return
	}

	
	userID := ""

	
	if contextUserID, exists := c.Get("userId"); exists {
		userID = contextUserID.(string)
		log.Printf("Got user ID from context: %s", userID)
	} else {
		
		token := c.Query("token")
		queryUserID := c.Query("user_id")

		log.Printf("Authenticating WebSocket from query params: token=%s..., user_id=%s",
			token[:min(len(token), 20)], queryUserID)

		if token != "" {
			
			jwtSecret := string(utils.GetJWTSecret())

			parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(jwtSecret), nil
			})

			if err != nil {
				log.Printf("JWT validation failed: %v", err)
				utils.SendErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid token")
				return
			}

			if !parsedToken.Valid {
				log.Printf("Invalid JWT token")
				utils.SendErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid token")
				return
			}

			
			if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok {
				if tokenUserID, ok := claims["user_id"].(string); ok {
					userID = tokenUserID
				} else if sub, ok := claims["sub"].(string); ok {
					userID = sub
				}

				
				if queryUserID != "" && userID != queryUserID {
					log.Printf("User ID mismatch: token=%s, query=%s", userID, queryUserID)
					utils.SendErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User ID mismatch")
					return
				}
			}
		} else if queryUserID != "" {
			
			log.Printf("WARNING: Using direct user ID without token validation (development only)")
			userID = queryUserID
		}

		if userID == "" {
			log.Printf("WebSocket connection rejected - no valid authentication")
			utils.SendErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
			return
		}
	}

	log.Printf("Handling WebSocket connection for user ID: %s", userID)
	log.Printf("Headers: %v", c.Request.Header)

	
	upgraderConfig := websocket.Upgrader{
		ReadBufferSize:  ReadBufferSize,
		WriteBufferSize: WriteBufferSize,
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			log.Printf("WebSocket connection attempt from origin: %s", origin)
			
			return true
		},
	}

	conn, err := upgraderConfig.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to set websocket upgrade: %v", err)
		return
	}

	log.Printf("WebSocket connection successfully established for user ID: %s", userID)

	client := &NotificationClient{
		ID:         uuid.New().String(),
		UserID:     userID,
		Connection: conn,
		Send:       make(chan []byte, AppConfig.WebSocket.SendBufferSize),
	}

	manager := GetWebSocketManager()
	wsClient := &Client{
		ID:         client.ID,
		UserID:     client.UserID,
		Connection: client.Connection,
		ChatID:     "",
		Send:       client.Send,
		Manager:    manager,
	}
	manager.register <- wsClient

	notificationManager.UserConnected(client.UserID)

	go client.notificationWritePump()
	go client.notificationReadPump(wsClient)

	
	testMessage := map[string]interface{}{
		"type":      "connection_established",
		"timestamp": time.Now().Format(time.RFC3339),
		"message":   "WebSocket connection established successfully",
	}

	if testMessageJSON, err := json.Marshal(testMessage); err == nil {
		select {
		case client.Send <- testMessageJSON:
			log.Printf("Sent test connection message to user %s", userID)
		default:
			log.Printf("Failed to send test message to user %s", userID)
		}
	}

	go sendUnreadNotifications(client)
}

func (c *NotificationClient) notificationReadPump(wsClient *Client) {
	defer func() {

		manager := GetWebSocketManager()
		manager.unregister <- wsClient

		notificationManager.UserDisconnected(c.UserID)

		c.Connection.Close()
	}()

	c.Connection.SetReadLimit(int64(AppConfig.WebSocket.MaxMessageSize))
	c.Connection.SetReadDeadline(time.Now().Add(AppConfig.WebSocket.ReadDeadlineTimeout))
	c.Connection.SetPongHandler(func(string) error {
		c.Connection.SetReadDeadline(time.Now().Add(AppConfig.WebSocket.ReadDeadlineTimeout))
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

		processNotificationAction(message, c.UserID)
	}
}

func (c *NotificationClient) notificationWritePump() {
	ticker := time.NewTicker(AppConfig.WebSocket.PingInterval)
	defer func() {
		ticker.Stop()
		c.Connection.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Connection.SetWriteDeadline(time.Now().Add(AppConfig.WebSocket.WriteDeadlineTimeout))
			if !ok {
				c.Connection.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Connection.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Connection.SetWriteDeadline(time.Now().Add(AppConfig.WebSocket.WriteDeadlineTimeout))
			if err := c.Connection.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func processNotificationAction(message []byte, userID string) {

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

		if err := notificationManager.MarkNotificationAsRead(userID, action.NotificationID); err != nil {
			log.Printf("Error marking notification as read: %v", err)
		}
	}
}

func sendUnreadNotifications(client *NotificationClient) {

	notifications := notificationManager.GetUserNotifications(client.UserID, 50, 0)

	var unreadNotifications []Notification
	for _, notification := range notifications {
		if !notification.Read {
			unreadNotifications = append(unreadNotifications, notification)
		}
	}

	if len(unreadNotifications) == 0 {
		return
	}

	notificationBundle := struct {
		Type          string         `json:"type"`
		Notifications []Notification `json:"notifications"`
	}{
		Type:          "notification_bundle",
		Notifications: unreadNotifications,
	}

	bundle, err := json.Marshal(notificationBundle)
	if err != nil {
		log.Printf("Error serializing notification bundle: %v", err)
		return
	}

	select {
	case client.Send <- bundle:

	default:
		log.Printf("Error: notification channel is full for user %s", client.UserID)
	}
}
