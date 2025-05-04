package api

import (
	"context"

	"aycom/backend/services/community/service"

	"github.com/google/uuid"
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
}

// MarkMessageAsReadRequest represents a request to mark a message as read
type MarkMessageAsReadRequest struct {
	ChatId    string `json:"chat_id"`
	MessageId string `json:"message_id"`
	UserId    string `json:"user_id"`
}

// MarkMessageAsReadResponse represents a response to a mark message as read request
type MarkMessageAsReadResponse struct {
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

// Return success response for now until we implement proper repositories
func (h *ChatHandler) CreateChat(ctx context.Context, req *CreateChatRequest) (*CreateChatResponse, error) {
	return &CreateChatResponse{
		ChatId: uuid.New().String(),
	}, nil
}

// Return success response for now
func (h *ChatHandler) SendMessage(ctx context.Context, req *SendMessageRequest) (*SendMessageResponse, error) {
	return &SendMessageResponse{
		MessageId: uuid.New().String(),
	}, nil
}

// Return success response for now
func (h *ChatHandler) MarkMessageAsRead(ctx context.Context, req *MarkMessageAsReadRequest) (*MarkMessageAsReadResponse, error) {
	return &MarkMessageAsReadResponse{}, nil
}

// Return empty list for now
func (h *ChatHandler) ListMessages(ctx context.Context, req *ListMessagesRequest) (*ListMessagesResponse, error) {
	return &ListMessagesResponse{
		Messages: []*Message{},
	}, nil
}
