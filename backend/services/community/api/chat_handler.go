package api

import (
	"aycom/backend/proto/community"
	"context"
	"log"
	"time"

	"aycom/backend/services/community/service"
)

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

type CreateChatRequest struct {
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	CreatorId      string   `json:"creator_id"`
	CommunityId    string   `json:"community_id"`
	IsGroupChat    bool     `json:"is_group_chat"`
	ParticipantIds []string `json:"participant_ids"`
}

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

type SendMessageRequest struct {
	ChatId  string `json:"chat_id"`
	UserId  string `json:"user_id"`
	Content string `json:"content"`
}

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

type MarkMessageAsReadRequest struct {
	ChatId    string `json:"chat_id"`
	MessageId string `json:"message_id"`
	UserId    string `json:"user_id"`
}

type MarkMessageAsReadResponse struct {
	Success bool `json:"success"`
}

type ListMessagesRequest struct {
	ChatId string `json:"chat_id"`
	Limit  int32  `json:"limit"`
	Offset int32  `json:"offset"`
}

type ListMessagesResponse struct {
	Messages []*Message `json:"messages"`
}

type ChatHandler struct {
	chatService service.ChatService
}

func NewChatHandler(chatService service.ChatService) *ChatHandler {
	return &ChatHandler{
		chatService: chatService,
	}
}

func (h *ChatHandler) CreateChat(ctx context.Context, req *CreateChatRequest) (*CreateChatResponse, error) {
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

	response := &CreateChatResponse{
		ChatId: chat.Id,
	}

	response.Chat.Id = chat.Id
	response.Chat.Name = chat.Name
	response.Chat.IsGroupChat = chat.IsGroup
	response.Chat.CreatorId = chat.CreatedBy

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

	participants, err := h.chatService.ListParticipants(chat.Id, 100, 0)
	if err == nil {
		for _, p := range participants {
			response.Chat.Participants = append(response.Chat.Participants, p.UserId)
		}
	}

	return response, nil
}

func (h *ChatHandler) SendMessage(ctx context.Context, req *SendMessageRequest) (*SendMessageResponse, error) {
	log.Printf("Handling SendMessage request. ChatID: %s, UserID: %s", req.ChatId, req.UserId)

	msgId, err := h.chatService.SendMessage(req.ChatId, req.UserId, req.Content)
	if err != nil {
		log.Printf("Error sending message: %v", err)
		return nil, err
	}

	log.Printf("Message sent successfully. Message ID: %s", msgId)

	response := &SendMessageResponse{
		MessageId: msgId,
	}

	response.Message.Id = msgId
	response.Message.ChatId = req.ChatId
	response.Message.SenderId = req.UserId
	response.Message.Content = req.Content
	response.Message.Timestamp = time.Now().Unix()
	response.Message.IsRead = false
	response.Message.IsEdited = false
	response.Message.IsDeleted = false

	messages, err := h.chatService.GetMessages(req.ChatId, 1, 0)
	if err == nil && len(messages) > 0 {
		var msg *community.Message
		for _, m := range messages {
			if m.Id == msgId {
				msg = m
				break
			}
		}

		if msg != nil {
			log.Printf("Found sent message in database, using its data")

			response.Message.Id = msg.Id
			response.Message.ChatId = msg.ChatId
			response.Message.SenderId = msg.SenderId
			response.Message.Content = msg.Content

			if msg.SentAt != nil {
				response.Message.Timestamp = msg.SentAt.AsTime().Unix()
			}

			response.Message.IsRead = !msg.Unsent
			response.Message.IsDeleted = msg.DeletedForAll
		}
	}

	return response, nil
}

func (h *ChatHandler) MarkMessageAsRead(ctx context.Context, req *MarkMessageAsReadRequest) (*MarkMessageAsReadResponse, error) {
	err := h.chatService.DeleteMessage(req.ChatId, req.MessageId, req.UserId)
	if err != nil {
		log.Printf("Error marking message as read: %v", err)
		return &MarkMessageAsReadResponse{Success: false}, err
	}

	return &MarkMessageAsReadResponse{Success: true}, nil
}

func (h *ChatHandler) ListMessages(ctx context.Context, req *ListMessagesRequest) (*ListMessagesResponse, error) {
	log.Printf("Handling ListMessages request. ChatID: %s, Limit: %d, Offset: %d",
		req.ChatId, req.Limit, req.Offset)

	limit := int(req.Limit)
	if limit <= 0 {
		limit = 50
	}

	offset := int(req.Offset)
	if offset < 0 {
		offset = 0
	}

	messages, err := h.chatService.GetMessages(req.ChatId, limit, offset)
	if err != nil {
		log.Printf("Error getting messages: %v", err)
		return &ListMessagesResponse{Messages: []*Message{}}, err
	}

	log.Printf("Retrieved %d messages for chat %s", len(messages), req.ChatId)

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
			IsRead:    !msg.Unsent,
			IsEdited:  false,
			IsDeleted: msg.DeletedForAll,
		}
	}

	return &ListMessagesResponse{
		Messages: responseMessages,
	}, nil
}

func (h *ChatHandler) SearchMessages(ctx context.Context, req *ListMessagesRequest, query string) (*ListMessagesResponse, error) {
	limit := int(req.Limit)
	if limit <= 0 {
		limit = 50
	}

	offset := int(req.Offset)
	if offset < 0 {
		offset = 0
	}

	messages, err := h.chatService.SearchMessages(req.ChatId, query, limit, offset)
	if err != nil {
		log.Printf("Error searching messages: %v", err)
		return &ListMessagesResponse{Messages: []*Message{}}, err
	}

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
			IsRead:    !msg.Unsent,
			IsEdited:  false,
			IsDeleted: msg.DeletedForAll,
		}
	}

	return &ListMessagesResponse{
		Messages: responseMessages,
	}, nil
}
