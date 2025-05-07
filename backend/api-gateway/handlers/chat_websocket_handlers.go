package handlers

import (
	"encoding/json"
	"log"
	"net/http"
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
	// Validate user access to chat
	userID, exists := c.Get("userId")
	if !exists {
		SendErrorResponse(c, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	chatID := c.Param("id")
	if chatID == "" {
		SendErrorResponse(c, http.StatusBadRequest, "invalid_request", "Chat ID is required")
		return
	}

	// Validate user exists
	client := GetCommunityServiceClient()
	isValid, err := client.ValidateUser(userID.(string))
	if err != nil {
		log.Printf("Error validating user: %v", err)
		SendErrorResponse(c, http.StatusInternalServerError, "server_error", "Failed to validate user")
		return
	}

	if !isValid {
		SendErrorResponse(c, http.StatusNotFound, "not_found", "User not found")
		return
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to set websocket upgrade: %v", err)
		return
	}

	// Create client and register with WebSocket manager
	wsClient := &Client{
		ID:         uuid.New().String(),
		UserID:     userID.(string),
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
	// Parse the message from JSON
	var chatMessage ChatMessage
	if err := json.Unmarshal(message, &chatMessage); err != nil {
		log.Printf("Error unmarshaling message: %v", err)
		return createErrorResponse("invalid_message_format", "Could not parse message"), err
	}

	// Set message metadata
	chatMessage.UserID = userID
	chatMessage.ChatID = chatID
	chatMessage.Timestamp = time.Now()

	// Generate a unique message ID if not provided
	if chatMessage.MessageID == "" {
		chatMessage.MessageID = uuid.New().String()
	}

	// Process based on message type
	switch chatMessage.Type {
	case "text":
		return processTextMessage(chatMessage)
	case "typing":
		return processTypingIndicator(chatMessage)
	case "read":
		return processReadReceipt(chatMessage)
	case "edit":
		return processEditMessage(chatMessage)
	case "delete":
		return processDeleteMessage(chatMessage)
	default:
		return createErrorResponse("unknown_message_type", "Message type not supported"), nil
	}
}

// processTextMessage handles text messages
func processTextMessage(message ChatMessage) ([]byte, error) {
	// Save the message to the database
	client := GetCommunityServiceClient()
	msgID, err := client.SendMessage(
		message.ChatID,
		message.UserID,
		message.Content,
	)

	if err != nil {
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

	// Update message ID from database
	message.MessageID = msgID

	// Return the message to be broadcast
	responseMsg, err := json.Marshal(message)
	if err != nil {
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
