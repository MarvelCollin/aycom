package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Config for WebSocket parameters
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

// SetCommunityServiceClient sets the client for use across handler functions
func SetCommunityServiceClient(client CommunityServiceClient) {
	communityServiceClient = client
}

// ChatMessage represents a message sent over WebSocket
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

// HandleCommunityChat handles WebSocket connections for community chat
func HandleCommunityChat(c *gin.Context) {
	// No authorization required anymore
	userID := "anonymous" // Default user ID for anonymous users

	chatID := c.Param("id")
	if chatID == "" {
		SendErrorResponse(c, http.StatusBadRequest, "invalid_request", "Chat ID is required")
		return
	}

	log.Printf("WebSocket connection request for chat %s from anonymous user", chatID)

	// Set up websocket connection
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins in development
		},
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to set websocket upgrade: %v", err)
		return
	}

	log.Printf("WebSocket connection established for chat %s, anonymous user", chatID)

	// Create client and register with WebSocket manager
	wsClient := &Client{
		ID:         uuid.New().String(),
		UserID:     userID,
		Connection: conn,
		ChatID:     chatID,
		Send:       make(chan []byte, WebSocketConfig.SendBufferSize),
		Manager:    GetWebSocketManager(),
	}

	// Register client with the WebSocket manager
	manager := GetWebSocketManager()
	manager.register <- wsClient

	// Start goroutines for reading and writing
	go communityChatReadPump(wsClient)
	go communityChatWritePump(wsClient)
}

// communityChatReadPump reads messages from the WebSocket connection
// for community chat messages
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

	for {
		_, message, err := c.Connection.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Error: %v", err)
			}
			break
		}

		// Process the message with our community chat handler
		processedMsg, err := ProcessIncomingMessage(message, c.UserID, c.ChatID)
		if err != nil {
			// If error, we'll just send the error response back to this client only
			c.Send <- processedMsg
		} else {
			// If no error, broadcast the processed message to all clients in the chat room
			c.Manager.broadcast <- BroadcastMessage{
				ChatID:  c.ChatID,
				Message: processedMsg,
				UserID:  c.UserID,
			}
		}
	}
}

// communityChatWritePump writes messages to the WebSocket connection
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

			// Add queued chat messages to the current websocket message
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

// ProcessIncomingMessage processes messages received from WebSocket clients
// and stores them in the database via the Community service gRPC client
func ProcessIncomingMessage(message []byte, userID, chatID string) ([]byte, error) {
	var chatMessage ChatMessage
	if err := json.Unmarshal(message, &chatMessage); err != nil {
		log.Printf("Error unmarshaling message: %v", err)
		return createErrorResponse("invalid_format", "Invalid message format"), err
	}

	// Default to the provided user ID and chat ID if not specified in the message
	if chatMessage.UserID == "" {
		chatMessage.UserID = userID
	}
	if chatMessage.ChatID == "" {
		chatMessage.ChatID = chatID
	}

	// Validate that the user ID in the message matches the authenticated user
	if chatMessage.UserID != userID {
		log.Printf("User ID mismatch: %s != %s", chatMessage.UserID, userID)
		return createErrorResponse("unauthorized", "User ID mismatch"), fmt.Errorf("user ID mismatch")
	}

	// Set timestamp if not provided
	if chatMessage.Timestamp.IsZero() {
		chatMessage.Timestamp = time.Now()
	}

	// Process message based on type
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
	default:
		log.Printf("Unknown message type: %s", chatMessage.Type)
		return createErrorResponse("invalid_type", "Unknown message type"), nil
	}
}

// processTextMessage handles text messages
func processTextMessage(message ChatMessage, originalID string) ([]byte, error) {
	log.Printf("Processing text message from user %s to chat %s: %s", message.UserID, message.ChatID, message.Content)

	// Save the message to the database
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

	// Update message ID from database
	message.MessageID = msgID

	// Create response map with all required fields
	responseMap := map[string]interface{}{
		"type":       message.Type,
		"content":    message.Content,
		"user_id":    message.UserID,
		"chat_id":    message.ChatID,
		"timestamp":  message.Timestamp.Unix(), // Convert to Unix timestamp
		"message_id": message.MessageID,
		"is_edited":  message.IsEdited,
		"is_deleted": message.IsDeleted,
		"is_read":    message.IsRead,
	}

	// Include the original client ID if it was a temporary ID (starts with temp-)
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

// processTypingIndicator handles typing indicators
func processTypingIndicator(message ChatMessage) ([]byte, error) {
	// Just relay the typing indicator, no need to save
	message.Content = "" // Ensure no content in typing indicators
	responseMsg, err := json.Marshal(message)
	if err != nil {
		return createErrorResponse("server_error", "Failed to process typing indicator"), err
	}
	return responseMsg, nil
}

// processReadReceipt handles read receipts
func processReadReceipt(message ChatMessage) ([]byte, error) {
	// Update the read status in the database
	client := GetCommunityServiceClient()
	err := client.MarkMessageAsRead(
		message.ChatID,
		message.UserID,
		message.MessageID,
	)

	if err != nil {
		return createErrorResponse("server_error", "Failed to mark message as read"), err
	}

	// Return the read receipt to be broadcast
	responseMsg, err := json.Marshal(message)
	if err != nil {
		return createErrorResponse("server_error", "Failed to process read receipt"), err
	}
	return responseMsg, nil
}

// processEditMessage handles message edits
func processEditMessage(message ChatMessage) ([]byte, error) {
	// Call the community service to update the message
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

// processDeleteMessage handles message deletion
func processDeleteMessage(message ChatMessage) ([]byte, error) {
	// Call the community service to delete the message
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
	message.Content = "" // Remove the content from deleted messages
	responseMsg, err := json.Marshal(message)
	if err != nil {
		return createErrorResponse("server_error", "Failed to process deletion"), err
	}
	return responseMsg, nil
}

// createErrorResponse creates a standardized error response
func createErrorResponse(code, message string) []byte {
	errorResponse := struct {
		Type    string `json:"type"`
		Error   string `json:"error"`
		Message string `json:"message"`
	}{
		Type:    "error",
		Error:   code,
		Message: message,
	}

	response, _ := json.Marshal(errorResponse)
	return response
}
