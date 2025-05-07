package api

import (
	"context"
	"log"
	"time"

	"aycom/backend/proto/community"
	"aycom/backend/services/community/service"
)

// Message represents a gRPC message response
type Message struct {
	MessageId string `json:"message_id"`
	ChatId    string `json:"chat_id"`
	SenderId  string `json:"sender_id"`
	Content   string `json:"content"`
	Timestamp int64  `json:"timestamp"`
	IsRead    bool   `json:"is_read"`
	IsEdited  bool   `json:"is_edited,omitempty"`
	IsDeleted bool   `json:"is_deleted,omitempty"`
}

// CreateChatRequest represents a request to create a chat
type CreateChatRequest struct {
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	CreatorId      string   `json:"creator_id"`
	CommunityId    string   `json:"community_id"`
	IsGroupChat    bool     `json:"is_group_chat"`
	ParticipantIds []string `json:"participant_ids"`
}

// CreateChatResponse represents a response to a create chat request
type CreateChatResponse struct {
	ChatId string `json:"chat_id"`
	Chat   struct {
		Id           string    `json:"id"`
		Name         string    `json:"name"`
		Description  string    `json:"description"`
		IsGroupChat  bool      `json:"is_group_chat"`
		CreatorId    string    `json:"creator_id"`
		CreatedAt    time.Time `json:"created_at"`
		UpdatedAt    time.Time `json:"updated_at"`
		Participants []string  `json:"participants,omitempty"`
	} `json:"chat"`
}

// SendMessageRequest represents a request to send a message
type SendMessageRequest struct {
	ChatId  string `json:"chat_id"`
	UserId  string `json:"user_id"`
	Content string `json:"content"`
}

// SendMessageResponse represents a response to a send message request
type SendMessageResponse struct {
	MessageId string `json:"message_id"`
	Message   struct {
		Id        string `json:"id"`
		ChatId    string `json:"chat_id"`
		SenderId  string `json:"sender_id"`
		Content   string `json:"content"`
		Timestamp int64  `json:"timestamp"`
		IsRead    bool   `json:"is_read"`
		IsEdited  bool   `json:"is_edited"`
		IsDeleted bool   `json:"is_deleted"`
	} `json:"message"`
}

// MarkMessageAsReadRequest represents a request to mark a message as read
type MarkMessageAsReadRequest struct {
	ChatId    string `json:"chat_id"`
	MessageId string `json:"message_id"`
	UserId    string `json:"user_id"`
}

// MarkMessageAsReadResponse represents a response to a mark message as read request
type MarkMessageAsReadResponse struct {
	Success bool `json:"success"`
}

// ListMessagesRequest represents a request to list messages
type ListMessagesRequest struct {
	ChatId string `json:"chat_id"`
	Limit  int32  `json:"limit"`
	Offset int32  `json:"offset"`
}

// ListMessagesResponse represents a response to a list messages request
type ListMessagesResponse struct {
	Messages []*Message `json:"messages"`
}

type ChatHandler struct {
	chatService *service.ChatService
}

func NewChatHandler(chatService *service.ChatService) *ChatHandler {
	return &ChatHandler{
		chatService: chatService,
	}
}

// CreateChat creates a new chat
func (h *ChatHandler) CreateChat(ctx context.Context, req *CreateChatRequest) (*CreateChatResponse, error) {
	// Call service to create chat
	chat, err := h.chatService.CreateChat(
		req.Name,
		req.Description,
		req.CreatorId,
		req.IsGroupChat,
		req.ParticipantIds,
	)

	if err != nil {
		log.Printf("Error creating chat: %v", err)
		return nil, err
	}

	// Create the response
	response := &CreateChatResponse{
		ChatId: chat.Id,
	}

	// Populate chat details
	response.Chat.Id = chat.Id
	response.Chat.Name = chat.Name
	response.Chat.IsGroupChat = chat.IsGroup // Note: using IsGroup from proto
	response.Chat.CreatorId = chat.CreatedBy // Note: using CreatedBy from proto

	if chat.CreatedAt != nil {
		response.Chat.CreatedAt = chat.CreatedAt.AsTime()
	} else {
		response.Chat.CreatedAt = time.Now()
	}

	if chat.UpdatedAt != nil {
		response.Chat.UpdatedAt = chat.UpdatedAt.AsTime()
	} else {
		response.Chat.UpdatedAt = time.Now()
	}

	// Add participants from a separate call
	participants, err := h.chatService.ListParticipants(chat.Id, 100, 0)
	if err == nil {
		for _, p := range participants {
			response.Chat.Participants = append(response.Chat.Participants, p.UserId)
		}
	}

	return response, nil
}

// SendMessage sends a message in a chat
func (h *ChatHandler) SendMessage(ctx context.Context, req *SendMessageRequest) (*SendMessageResponse, error) {
	log.Printf("Handling SendMessage request. ChatID: %s, UserID: %s", req.ChatId, req.UserId)

	// Call service to send message
	msgId, err := h.chatService.SendMessage(req.ChatId, req.UserId, req.Content)
	if err != nil {
		log.Printf("Error sending message: %v", err)
		return nil, err
	}

	log.Printf("Message sent successfully. Message ID: %s", msgId)

	// Create the response
	response := &SendMessageResponse{
		MessageId: msgId,
	}

	// Initialize the message struct fields
	response.Message.Id = msgId
	response.Message.ChatId = req.ChatId
	response.Message.SenderId = req.UserId
	response.Message.Content = req.Content
	response.Message.Timestamp = time.Now().Unix()
	response.Message.IsRead = false
	response.Message.IsEdited = false
	response.Message.IsDeleted = false

	// Try to fetch the saved message for more accurate data
	messages, err := h.chatService.GetMessages(req.ChatId, 1, 0)
	if err == nil && len(messages) > 0 {
		// Find the message we just sent (should be the most recent one)
		var msg *community.Message
		for _, m := range messages {
			if m.Id == msgId {
				msg = m
				break
			}
		}

		// If we found the message, use its data
		if msg != nil {
			log.Printf("Found sent message in database, using its data")

			// Fill in full message details from the database
			response.Message.Id = msg.Id
			response.Message.ChatId = msg.ChatId
			response.Message.SenderId = msg.SenderId
			response.Message.Content = msg.Content

			if msg.SentAt != nil {
				response.Message.Timestamp = msg.SentAt.AsTime().Unix()
			}

			response.Message.IsRead = !msg.Unsent // Using the inverse of unsent
			response.Message.IsDeleted = msg.DeletedForAll
		}
	}

	return response, nil
}

// MarkMessageAsRead marks a message as read
func (h *ChatHandler) MarkMessageAsRead(ctx context.Context, req *MarkMessageAsReadRequest) (*MarkMessageAsReadResponse, error) {
	// Call service to mark message as read
	err := h.chatService.MarkMessageAsRead(req.ChatId, req.MessageId, req.UserId)
	if err != nil {
		log.Printf("Error marking message as read: %v", err)
		return &MarkMessageAsReadResponse{Success: false}, err
	}

	return &MarkMessageAsReadResponse{Success: true}, nil
}

// ListMessages lists messages in a chat
func (h *ChatHandler) ListMessages(ctx context.Context, req *ListMessagesRequest) (*ListMessagesResponse, error) {
	log.Printf("Handling ListMessages request. ChatID: %s, Limit: %d, Offset: %d",
		req.ChatId, req.Limit, req.Offset)

	// Set default values for limit and offset if not provided
	limit := int(req.Limit)
	if limit <= 0 {
		limit = 50 // Default limit
	}

	offset := int(req.Offset)
	if offset < 0 {
		offset = 0
	}

	// Call service to get messages
	messages, err := h.chatService.GetMessages(req.ChatId, limit, offset)
	if err != nil {
		log.Printf("Error getting messages: %v", err)
		return &ListMessagesResponse{Messages: []*Message{}}, err
	}

	log.Printf("Retrieved %d messages for chat %s", len(messages), req.ChatId)

	// Convert messages to response format
	responseMessages := make([]*Message, len(messages))
	for i, msg := range messages {
		timestamp := time.Now().Unix()
		if msg.SentAt != nil {
			timestamp = msg.SentAt.AsTime().Unix()
		}

		responseMessages[i] = &Message{
			MessageId: msg.Id,
			ChatId:    msg.ChatId,
			SenderId:  msg.SenderId,
			Content:   msg.Content,
			Timestamp: timestamp,
			IsRead:    !msg.Unsent, // Invert the unsent flag
			IsEdited:  false,       // Not tracking this yet
			IsDeleted: msg.DeletedForAll,
		}
	}

	return &ListMessagesResponse{
		Messages: responseMessages,
	}, nil
}

// SearchMessages searches for messages in a chat
func (h *ChatHandler) SearchMessages(ctx context.Context, req *ListMessagesRequest, query string) (*ListMessagesResponse, error) {
	// Set default values for limit and offset if not provided
	limit := int(req.Limit)
	if limit <= 0 {
		limit = 50 // Default limit
	}

	offset := int(req.Offset)
	if offset < 0 {
		offset = 0
	}

	// Call service to search messages
	messages, err := h.chatService.SearchMessages(req.ChatId, query, limit, offset)
	if err != nil {
		log.Printf("Error searching messages: %v", err)
		return &ListMessagesResponse{Messages: []*Message{}}, err
	}

	// Convert messages to response format
	responseMessages := make([]*Message, len(messages))
	for i, msg := range messages {
		timestamp := time.Now().Unix()
		if msg.SentAt != nil {
			timestamp = msg.SentAt.AsTime().Unix()
		}

		responseMessages[i] = &Message{
			MessageId: msg.Id,
			ChatId:    msg.ChatId,
			SenderId:  msg.SenderId,
			Content:   msg.Content,
			Timestamp: timestamp,
			IsRead:    !msg.Unsent, // Invert the unsent flag
			IsEdited:  false,       // Not tracking this yet
			IsDeleted: msg.DeletedForAll,
		}
	}

	return &ListMessagesResponse{
		Messages: responseMessages,
	}, nil
}
