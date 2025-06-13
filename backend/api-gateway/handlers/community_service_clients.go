package handlers

import (
	communityProto "aycom/backend/proto/community"
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"aycom/backend/api-gateway/config"
)

type CommunityServiceClient interface {
	ValidateUser(userID string) (bool, error)
	SendMessage(chatID, userID, content string) (string, error)
	MarkMessageAsRead(chatID, userID, messageID string) error
	GetMessages(chatID string, limit, offset int) ([]Message, error)
	EditMessage(chatID, userID, messageID, newContent string) error
	DeleteMessage(chatID, userID, messageID string) error
	GetChats(userID string, limit, offset int) ([]Chat, error)
	CreateChat(isGroup bool, name string, participantIDs []string, createdBy string) (*Chat, error)
	IsUserChatParticipant(chatID, userID string) (bool, error)
	GetChatParticipants(chatID string) ([]string, error)
}

type Message struct {
	ID        string    `json:"id"`
	ChatID    string    `json:"chat_id"`
	SenderID  string    `json:"sender_id"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	IsRead    bool      `json:"is_read"`
	IsEdited  bool      `json:"is_edited,omitempty"`
	IsDeleted bool      `json:"is_deleted,omitempty"`
}

type Chat struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	IsGroupChat  bool      `json:"is_group_chat"`
	CreatedBy    string    `json:"created_by"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Participants []string  `json:"participants,omitempty"`
	LastMessage  *Message  `json:"last_message,omitempty"`
}

type communityCommunicationClient struct {
	grpcClient communityProto.CommunityServiceClient
}

var communityServiceClient CommunityServiceClient

func GetCommunityServiceClient() CommunityServiceClient {
	if communityServiceClient == nil {
		log.Println("Warning: Community service client not initialized, using fallback")
		return &communityCommunicationClient{}
	}
	return communityServiceClient
}

func InitCommunityServiceClient(cfg *config.Config) {
	log.Println("Initializing Community service client...")

	if CommunityClient != nil {
		client := &communityCommunicationClient{
			grpcClient: CommunityClient,
		}
		SetCommunityServiceClient(client)
		log.Println("Community service client initialized successfully")
	} else {
		log.Println("Warning: Community gRPC client not available, using fallback implementation")

		communityServiceAddr := cfg.Services.CommunityService

		conn, err := grpc.Dial(communityServiceAddr,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			log.Printf("Failed to connect to Community service: %v", err)
			return
		}
		client := &communityCommunicationClient{
			grpcClient: communityProto.NewCommunityServiceClient(conn),
		}
		SetCommunityServiceClient(client)
	}
}

func (c *communityCommunicationClient) ValidateUser(userID string) (bool, error) {
	if c.grpcClient == nil {
		return false, fmt.Errorf("community service client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := c.grpcClient.ListChats(ctx, &communityProto.ListChatsRequest{
		UserId: userID,
	})

	if err != nil {
		log.Printf("Error validating user %s: %v", userID, err)
		return false, err
	}

	return true, nil
}

func (c *communityCommunicationClient) SendMessage(chatID, userID, content string) (string, error) {
	if c.grpcClient == nil {
		return "", fmt.Errorf("community service client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Printf("Sending message to chat %s from user %s: %s", chatID, userID, content)

	resp, err := c.grpcClient.SendMessage(ctx, &communityProto.SendMessageRequest{
		ChatId:   chatID,
		SenderId: userID,
		Content:  content,
	})
	if err != nil {
		log.Printf("Error sending message to gRPC service: %v", err)
		return "", err
	}

	log.Printf("Successfully sent message, got ID: %s", resp.Message.Id)
	return resp.Message.Id, nil
}

func (c *communityCommunicationClient) MarkMessageAsRead(chatID, userID, messageID string) error {
	if c.grpcClient == nil {
		return fmt.Errorf("community service client not initialized")
	}

	log.Printf("Marking message as read: chat %s, user %s, message %s", chatID, userID, messageID)

	return nil
}

func (c *communityCommunicationClient) GetMessages(chatID string, limit, offset int) ([]Message, error) {
	if c.grpcClient == nil {
		return nil, fmt.Errorf("community service client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Printf("Fetching messages for chat %s (limit: %d, offset: %d)", chatID, limit, offset)

	resp, err := c.grpcClient.ListMessages(ctx, &communityProto.ListMessagesRequest{
		ChatId: chatID,
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		log.Printf("Error fetching messages from gRPC service: %v", err)
		return nil, err
	}

	log.Printf("Retrieved %d messages from service for chat %s", len(resp.Messages), chatID)

	messages := make([]Message, len(resp.Messages))
	for i, msg := range resp.Messages {
		sentTime := time.Now()
		if msg.SentAt != nil {
			sentTime = msg.SentAt.AsTime()
		}

		messages[i] = Message{
			ID:        msg.Id,
			ChatID:    msg.ChatId,
			SenderID:  msg.SenderId,
			Content:   msg.Content,
			Timestamp: sentTime,
			IsRead:    !msg.Unsent,
			IsEdited:  false,
			IsDeleted: msg.DeletedForAll || msg.DeletedForSender,
		}
	}

	return messages, nil
}

func (c *communityCommunicationClient) EditMessage(chatID, userID, messageID, newContent string) error {
	if c.grpcClient == nil {
		return fmt.Errorf("community service client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := c.grpcClient.UnsendMessage(ctx, &communityProto.UnsendMessageRequest{
		MessageId: messageID,
	})
	if err != nil {
		return fmt.Errorf("failed to unsend message: %w", err)
	}

	_, err = c.grpcClient.SendMessage(ctx, &communityProto.SendMessageRequest{
		ChatId:   chatID,
		SenderId: userID,
		Content:  newContent,
	})
	if err != nil {
		return fmt.Errorf("failed to send edited message: %w", err)
	}

	return nil
}

func (c *communityCommunicationClient) DeleteMessage(chatID, userID, messageID string) error {
	if c.grpcClient == nil {
		return fmt.Errorf("community service client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := c.grpcClient.DeleteMessage(ctx, &communityProto.DeleteMessageRequest{
		MessageId: messageID,
	})
	if err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}

	return nil
}

func (c *communityCommunicationClient) GetChats(userID string, limit, offset int) ([]Chat, error) {
	if c.grpcClient == nil {
		return nil, fmt.Errorf("community service client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	log.Printf("Fetching chats for user %s (limit: %d, offset: %d)", userID, limit, offset)

	resp, err := c.grpcClient.ListChats(ctx, &communityProto.ListChatsRequest{
		UserId: userID,
	})
	if err != nil {
		return nil, err
	}

	chats := make([]Chat, len(resp.Chats))
	for i, protoChat := range resp.Chats {
		chats[i] = Chat{
			ID:          protoChat.Id,
			Name:        protoChat.Name,
			IsGroupChat: protoChat.IsGroup,
			CreatedBy:   protoChat.CreatedBy,
			CreatedAt:   protoChat.CreatedAt.AsTime(),
			UpdatedAt:   protoChat.UpdatedAt.AsTime(),
		}
	}

	return chats, nil
}

func (c *communityCommunicationClient) CreateChat(isGroup bool, name string, participantIDs []string, createdBy string) (*Chat, error) {
	if c.grpcClient == nil {
		return nil, fmt.Errorf("community service client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := c.grpcClient.CreateChat(ctx, &communityProto.CreateChatRequest{
		IsGroup:        isGroup,
		Name:           name,
		ParticipantIds: participantIDs,
		CreatedBy:      createdBy,
	})
	if err != nil {
		return nil, err
	}

	chat := &Chat{
		ID:           resp.Chat.Id,
		Name:         resp.Chat.Name,
		IsGroupChat:  resp.Chat.IsGroup,
		CreatedBy:    resp.Chat.CreatedBy,
		CreatedAt:    resp.Chat.CreatedAt.AsTime(),
		UpdatedAt:    resp.Chat.UpdatedAt.AsTime(),
		Participants: participantIDs, // Include the participants that were provided
	}

	return chat, nil
}

func (c *communityCommunicationClient) IsUserChatParticipant(chatID, userID string) (bool, error) {
	if c.grpcClient == nil {
		return false, fmt.Errorf("community service client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := c.grpcClient.ListChatParticipants(ctx, &communityProto.ListChatParticipantsRequest{
		ChatId: chatID,
	})
	if err != nil {
		return false, err
	}

	for _, participant := range resp.Participants {
		if participant.UserId == userID {
			return true, nil
		}
	}

	return false, nil
}

func (c *communityCommunicationClient) GetChatParticipants(chatID string) ([]string, error) {
	if c.grpcClient == nil {
		return nil, fmt.Errorf("community service client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := c.grpcClient.ListChatParticipants(ctx, &communityProto.ListChatParticipantsRequest{
		ChatId: chatID,
	})
	if err != nil {
		return nil, err
	}

	participantIDs := make([]string, len(resp.Participants))
	for i, p := range resp.Participants {
		participantIDs[i] = p.UserId
	}

	return participantIDs, nil
}
