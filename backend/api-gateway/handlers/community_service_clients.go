package handlers

import (
	"context"
	"fmt"
	"log"
	"time"

	"aycom/backend/api-gateway/config"
	communityProto "aycom/backend/proto/community"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// CommunityServiceClient defines the methods used from the Community service
type CommunityServiceClient interface {
	ValidateUser(userID string) (bool, error)
	SendMessage(chatID, userID, content string) (string, error)
	MarkMessageAsRead(chatID, userID, messageID string) error
	GetMessages(chatID string, limit, offset int) ([]Message, error)
}

// Message represents a chat message
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

// Default implementation using gRPC client
type communityCommunicationClient struct {
	grpcClient communityProto.CommunityServiceClient
}

// Global instance of the community service client
var communityServiceClient CommunityServiceClient

// GetCommunityServiceClient returns the current community service client
func GetCommunityServiceClient() CommunityServiceClient {
	if communityServiceClient == nil {
		log.Println("Warning: Community service client not initialized, using fallback")
		// Create and return a fallback client
		return &communityCommunicationClient{}
	}
	return communityServiceClient
}

// InitCommunityServiceClient initializes the Community service client
func InitCommunityServiceClient(cfg *config.Config) {
	log.Println("Initializing Community service client...")

	// Use the existing CommunityClient gRPC client that was initialized in InitGRPCServices
	if CommunityClient != nil {
		client := &communityCommunicationClient{
			grpcClient: CommunityClient,
		}
		SetCommunityServiceClient(client)
		log.Println("Community service client initialized successfully")
	} else {
		log.Println("Warning: Community gRPC client not available, using fallback implementation")
		// Create a fallback implementation or retry connection
		communityServiceAddr := cfg.Services.CommunityService

		conn, err := grpc.NewClient(communityServiceAddr,
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

// ValidateUser checks if a user ID is valid with the Community service
func (c *communityCommunicationClient) ValidateUser(userID string) (bool, error) {
	// This is a stub implementation; in a real app, we'd make a gRPC call to the Community service
	log.Printf("Validating user %s with Community service", userID)
	// For development purposes, consider all users valid
	return true, nil
}

// SendMessage sends a message to a chat through the Community service
func (c *communityCommunicationClient) SendMessage(chatID, userID, content string) (string, error) {
	// This is a stub implementation; in a real app, we'd make a gRPC call to the Community service
	log.Printf("Sending message from user %s to chat %s: %s", userID, chatID, content)
	// Return a dummy message ID
	return "msg_" + userID + "_" + chatID, nil
}

// MarkMessageAsRead marks a message as read through the Community service
func (c *communityCommunicationClient) MarkMessageAsRead(chatID, userID, messageID string) error {
	// This is a stub implementation; in a real app, we'd make a gRPC call to the Community service
	log.Printf("Marking message %s as read by user %s in chat %s", messageID, userID, chatID)
	return nil
}

// GetMessages implements CommunityServiceClient
func (c *communityCommunicationClient) GetMessages(chatID string, limit, offset int) ([]Message, error) {
	if c.grpcClient == nil {
		return nil, fmt.Errorf("community service client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := c.grpcClient.ListMessages(ctx, &communityProto.ListMessagesRequest{
		ChatId: chatID,
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, err
	}

	messages := make([]Message, len(resp.Messages))
	for i, msg := range resp.Messages {
		messages[i] = Message{
			ID:        msg.Id,
			ChatID:    msg.ChatId,
			SenderID:  msg.SenderId,
			Content:   msg.Content,
			Timestamp: msg.SentAt.AsTime(),
			IsRead:    !msg.Unsent, // Use unsent as a proxy for read status
			IsEdited:  false,       // No edit tracking in proto
			IsDeleted: msg.DeletedForAll || msg.DeletedForSender,
		}
	}

	return messages, nil
}
