package handlers

import (
	"aycom/backend/api-gateway/utils"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var WebSocketConfig = struct {
	MaxMessageSize       int
	SendBufferSize       int
	ReadDeadlineTimeout  time.Duration
	WriteDeadlineTimeout time.Duration
	PingInterval         time.Duration
}{
	MaxMessageSize:       4096,
	SendBufferSize:       256,
	ReadDeadlineTimeout:  60 * time.Second,
	WriteDeadlineTimeout: 10 * time.Second,
	PingInterval:         30 * time.Second,
}

func SetCommunityServiceClient(client CommunityServiceClient) {
	communityServiceClient = client
}

type ChatMessage struct {
	Type      string    `json:"type"`
	Content   string    `json:"content"`
	UserID    string    `json:"user_id"`
	ChatID    string    `json:"chat_id"`
	Timestamp time.Time `json:"timestamp"`
	MessageID string    `json:"message_id,omitempty"`
	IsEdited  bool      `json:"is_edited,omitempty"`
	IsDeleted bool      `json:"is_deleted,omitempty"`
	IsRead    bool      `json:"is_read,omitempty"`
}

func HandleCommunityChat(c *gin.Context) {

	userID := "anonymous"

	directUserID := c.Query("user_id")
	if directUserID != "" {
		userID = directUserID
		log.Printf("Using directly provided user_id: %s", userID)
	} else {

		token := c.GetHeader("Authorization")
		if token == "" {

			token = c.Query("token")
			if token != "" {
				log.Printf("DEBUG: Got token from query parameter: %s... (length: %d)", token[:min(len(token), 20)]+"...", len(token))

				token = "Bearer " + token
				log.Printf("DEBUG: Added Bearer prefix: %s... (length: %d)", token[:min(len(token), 27)]+"...", len(token))
			} else {
				log.Printf("DEBUG: No token found in query parameter")
			}
		}

		if token != "" && strings.HasPrefix(token, "Bearer ") {

			tokenString := token[7:]
			log.Printf("DEBUG: Parsing token string: %s... (length: %d)", tokenString[:min(len(tokenString), 20)]+"...", len(tokenString))

			jwtSecret := string(utils.GetJWTSecret())

			parsedToken, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {

				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					log.Printf("DEBUG: Unexpected signing method: %v", token.Header["alg"])
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(jwtSecret), nil
			})

			if err != nil {
				log.Printf("DEBUG: JWT parse error: %v", err)
			} else {
				log.Printf("DEBUG: Token parsed successfully, valid: %v", parsedToken.Valid)
			}

			if err == nil && parsedToken.Valid {

				if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok {
					log.Printf("DEBUG: Token claims: %+v", claims)

					var uid string

					if userIDFromToken, ok := claims["user_id"].(string); ok {
						uid = userIDFromToken
					} else if sub, ok := claims["sub"].(string); ok {
						uid = sub
					}

					if uid != "" {
						userID = uid
						log.Printf("Authenticated WebSocket connection for user %s", userID)
					} else {
						log.Printf("DEBUG: user_id/sub claim not found or not a string")
					}
				} else {
					log.Printf("DEBUG: Failed to get claims from token")
				}
			} else {
				log.Printf("Invalid token: %v", err)
			}
		} else {
			log.Printf("DEBUG: No valid Bearer token found, using anonymous user")
		}
	}

	chatID := c.Param("id")
	if chatID == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "BAD_REQUEST", "Chat ID is required")
		return
	}

	log.Printf("WebSocket connection request for chat %s from user %s", chatID, userID)

	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to set websocket upgrade: %v", err)
		return
	}

	log.Printf("WebSocket connection established for chat %s, user %s", chatID, userID)

	wsClient := &Client{
		ID:         uuid.New().String(),
		UserID:     userID,
		Connection: conn,
		ChatID:     chatID,
		Send:       make(chan []byte, WebSocketConfig.SendBufferSize),
		Manager:    GetWebSocketManager(),
	}

	manager := GetWebSocketManager()
	manager.register <- wsClient

	go communityChatReadPump(wsClient)
	go communityChatWritePump(wsClient)
}

func communityChatReadPump(c *Client) {
	defer func() {
		c.Manager.unregister <- c
		c.Connection.Close()
	}()

	c.Connection.SetReadLimit(int64(WebSocketConfig.MaxMessageSize))
	c.Connection.SetReadDeadline(time.Now().Add(WebSocketConfig.ReadDeadlineTimeout))
	c.Connection.SetPongHandler(func(string) error {
		c.Connection.SetReadDeadline(time.Now().Add(WebSocketConfig.ReadDeadlineTimeout))
		return nil
	})

	log.Printf("WebSocket read pump started for client %s, user %s, chat %s", c.ID, c.UserID, c.ChatID)

	initialMessage := map[string]interface{}{
		"type":      "connection_established",
		"chat_id":   c.ChatID,
		"user_id":   c.UserID,
		"timestamp": time.Now(),
	}

	initialJSON, err := json.Marshal(initialMessage)
	if err == nil {
		c.Send <- initialJSON
		log.Printf("Sent connection confirmation to client %s", c.ID)
	}

	for {
		_, message, err := c.Connection.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Error: %v", err)
			}
			break
		}

		processedMsg, err := ProcessIncomingMessage(message, c.UserID, c.ChatID)
		if err != nil {

			c.Send <- processedMsg
		} else {

			c.Manager.broadcast <- BroadcastMessage{
				ChatID:  c.ChatID,
				Message: processedMsg,
				UserID:  c.UserID,
			}
		}
	}
}

func communityChatWritePump(c *Client) {
	ticker := time.NewTicker(WebSocketConfig.PingInterval)
	defer func() {
		ticker.Stop()
		c.Connection.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Connection.SetWriteDeadline(time.Now().Add(WebSocketConfig.WriteDeadlineTimeout))
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
			c.Connection.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Connection.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func ProcessIncomingMessage(message []byte, userID, chatID string) ([]byte, error) {
	log.Printf("Processing raw message: %s", string(message))

	var chatMessage ChatMessage
	if err := json.Unmarshal(message, &chatMessage); err != nil {
		log.Printf("Error unmarshaling message: %v", err)
		return createErrorResponse("invalid_format", "Invalid message format"), err
	}

	log.Printf("Received message type: %s from user %s in chat %s", chatMessage.Type, userID, chatID)

	if chatMessage.UserID == "" {
		chatMessage.UserID = userID
		log.Printf("Set missing UserID to %s", userID)
	}
	if chatMessage.ChatID == "" {
		chatMessage.ChatID = chatID
		log.Printf("Set missing ChatID to %s", chatID)
	}

	if chatMessage.UserID != userID {
		log.Printf("User ID mismatch: %s != %s", chatMessage.UserID, userID)
		return createErrorResponse("unauthorized", "User ID mismatch"), fmt.Errorf("user ID mismatch")
	}

	if chatMessage.Timestamp.IsZero() {
		chatMessage.Timestamp = time.Now()
	}

	var originalID string
	if chatMessage.MessageID != "" {
		originalID = chatMessage.MessageID
	}

	switch chatMessage.Type {
	case "text":
		return processTextMessage(chatMessage, originalID)
	case "typing":
		return processTypingIndicator(chatMessage)
	case "read":
		return processReadReceipt(chatMessage)
	case "edit":
		return processEditMessage(chatMessage)
	case "delete":
		return processDeleteMessage(chatMessage)
	case "connection_check":

		chatMessage.Timestamp = time.Now()
		responseMsg, err := json.Marshal(map[string]interface{}{
			"type":      "connection_ack",
			"user_id":   chatMessage.UserID,
			"chat_id":   chatMessage.ChatID,
			"timestamp": chatMessage.Timestamp.Unix(),
			"message":   "Connection established",
		})
		if err != nil {
			return createErrorResponse("server_error", "Failed to create connection acknowledgment"), err
		}
		return responseMsg, nil
	default:
		log.Printf("Unknown message type: %s", chatMessage.Type)
		return createErrorResponse("invalid_type", "Unknown message type"), nil
	}
}

func processTextMessage(message ChatMessage, originalID string) ([]byte, error) {
	log.Printf("Processing text message from user %s to chat %s: %s", message.UserID, message.ChatID, message.Content)

	client := GetCommunityServiceClient()
	msgID, err := client.SendMessage(
		message.ChatID,
		message.UserID,
		message.Content,
	)

	if err != nil {
		log.Printf("Error saving message to database: %v", err)
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.NotFound:
				return createErrorResponse("not_found", "Chat or user not found"), err
			case codes.PermissionDenied:
				return createErrorResponse("permission_denied", "Not allowed to send messages in this chat"), err
			default:
				return createErrorResponse("server_error", "Failed to save message"), err
			}
		}
		return createErrorResponse("server_error", "Failed to save message"), err
	}

	log.Printf("Message saved with ID: %s (original ID: %s)", msgID, originalID)

	message.MessageID = msgID

	responseMap := map[string]interface{}{
		"type":       message.Type,
		"content":    message.Content,
		"user_id":    message.UserID,
		"chat_id":    message.ChatID,
		"timestamp":  message.Timestamp.Unix(),
		"message_id": message.MessageID,
		"is_edited":  message.IsEdited,
		"is_deleted": message.IsDeleted,
		"is_read":    message.IsRead,
	}

	if originalID != "" && strings.HasPrefix(originalID, "temp-") {
		responseMap["original_id"] = originalID
	}

	responseMsg, err := json.Marshal(responseMap)
	if err != nil {
		log.Printf("Error serializing response: %v", err)
		return createErrorResponse("server_error", "Failed to process message"), err
	}

	return responseMsg, nil
}

func processTypingIndicator(message ChatMessage) ([]byte, error) {

	message.Content = ""
	responseMsg, err := json.Marshal(message)
	if err != nil {
		return createErrorResponse("server_error", "Failed to process typing indicator"), err
	}
	return responseMsg, nil
}

func processReadReceipt(message ChatMessage) ([]byte, error) {

	client := GetCommunityServiceClient()
	err := client.MarkMessageAsRead(
		message.ChatID,
		message.UserID,
		message.MessageID,
	)

	if err != nil {
		return createErrorResponse("server_error", "Failed to mark message as read"), err
	}

	responseMsg, err := json.Marshal(message)
	if err != nil {
		return createErrorResponse("server_error", "Failed to process read receipt"), err
	}
	return responseMsg, nil
}

func processEditMessage(message ChatMessage) ([]byte, error) {

	client := GetCommunityServiceClient()
	err := client.EditMessage(
		message.ChatID,
		message.UserID,
		message.MessageID,
		message.Content,
	)

	if err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.NotFound:
				return createErrorResponse("not_found", "Message not found"), err
			case codes.PermissionDenied:
				return createErrorResponse("permission_denied", "Not allowed to edit this message"), err
			default:
				return createErrorResponse("server_error", "Failed to edit message"), err
			}
		}
		return createErrorResponse("server_error", "Failed to edit message"), err
	}

	message.IsEdited = true
	responseMsg, err := json.Marshal(message)
	if err != nil {
		return createErrorResponse("server_error", "Failed to process edit"), err
	}
	return responseMsg, nil
}

func processDeleteMessage(message ChatMessage) ([]byte, error) {

	client := GetCommunityServiceClient()
	err := client.DeleteMessage(
		message.ChatID,
		message.UserID,
		message.MessageID,
	)

	if err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.NotFound:
				return createErrorResponse("not_found", "Message not found"), err
			case codes.PermissionDenied:
				return createErrorResponse("permission_denied", "Not allowed to delete this message"), err
			default:
				return createErrorResponse("server_error", "Failed to delete message"), err
			}
		}
		return createErrorResponse("server_error", "Failed to delete message"), err
	}

	message.IsDeleted = true
	message.Content = ""
	responseMsg, err := json.Marshal(message)
	if err != nil {
		return createErrorResponse("server_error", "Failed to process deletion"), err
	}
	return responseMsg, nil
}

func createErrorResponse(code, message string) []byte {
	response := gin.H{
		"success": false,
		"error": gin.H{
			"code":    code,
			"message": message,
		},
	}

	data, err := json.Marshal(response)
	if err != nil {
		log.Printf("Error marshaling error response: %v", err)
		return []byte(`{"success":false,"error":{"code":"INTERNAL_ERROR","message":"Failed to generate error response"}}`)
	}

	return data
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
