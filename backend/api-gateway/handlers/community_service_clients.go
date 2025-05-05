package handlers

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"aycom/backend/api-gateway/config"
	communityProto "aycom/backend/proto/community"
	userProto "aycom/backend/proto/user"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// CommunityServiceClient provides methods to interact with the Community service
type CommunityServiceClient interface {
	SendMessage(chatID, userID, content string) (string, error)
	MarkMessageAsRead(chatID, messageID, userID string) error
	GetMessages(chatID string, limit, offset int) ([]Message, error)
	ValidateUser(userID string) (bool, error)
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

// GRPCCommunityServiceClient is an implementation of CommunityServiceClient
// that communicates with the Community service via gRPC
type GRPCCommunityServiceClient struct {
	client     communityProto.CommunityServiceClient
	userClient userProto.UserServiceClient
}

// Global instance of the community service client
var communityServiceClient CommunityServiceClient

// InitCommunityServiceClient initializes the community service client
func InitCommunityServiceClient(cfg *config.Config) {
	// Connect to services
	communityAddr := cfg.Services.CommunityService
	userAddr := cfg.Services.UserService

	log.Printf("Connecting to Community service at %s and User service at %s", communityAddr, userAddr)

	// Connect to Community service with retry mechanism
	var communityConn *grpc.ClientConn
	var communityErr error
	for i := 0; i < 5; i++ {
		communityConn, communityErr = grpc.Dial(
			communityAddr,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithBlock(),
			grpc.WithTimeout(5*time.Second),
		)
		if communityErr == nil {
			break
		}
		retryDelay := time.Duration(i+1) * time.Second
		log.Printf("Failed to connect to Community service (attempt %d/5): %v. Retrying in %v...",
			i+1, communityErr, retryDelay)
		time.Sleep(retryDelay)
	}
	if communityErr != nil {
		log.Fatalf("CRITICAL: Failed to connect to Community service after multiple attempts: %v", communityErr)
	}

	// Connect to User service with retry mechanism
	var userConn *grpc.ClientConn
	var userErr error
	for i := 0; i < 5; i++ {
		userConn, userErr = grpc.Dial(
			userAddr,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithBlock(),
			grpc.WithTimeout(5*time.Second),
		)
		if userErr == nil {
			break
		}
		retryDelay := time.Duration(i+1) * time.Second
		log.Printf("Failed to connect to User service (attempt %d/5): %v. Retrying in %v...",
			i+1, userErr, retryDelay)
		time.Sleep(retryDelay)
	}
	if userErr != nil {
		log.Fatalf("CRITICAL: Failed to connect to User service after multiple attempts: %v", userErr)
	}

	// Create clients with connections
	communityClient := communityProto.NewCommunityServiceClient(communityConn)
	userClient := userProto.NewUserServiceClient(userConn)

	log.Printf("Successfully connected to Community service at %s", communityAddr)
	log.Printf("Successfully connected to User service at %s", userAddr)

	communityServiceClient = &GRPCCommunityServiceClient{
		client:     communityClient,
		userClient: userClient,
	}

	log.Println("Community service client initialized successfully")
}

// SendMessage implements CommunityServiceClient
func (c *GRPCCommunityServiceClient) SendMessage(chatID, userID, content string) (string, error) {
	if c.client == nil {
		return "", fmt.Errorf("community service client not initialized")
	}

	// First validate that the user exists using the User service
	valid, err := c.ValidateUser(userID)
	if err != nil {
		return "", err
	}

	if !valid {
		return "", fmt.Errorf("user %s not found", userID)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Now send the message using the Community service
	resp, err := c.client.SendMessage(ctx, &communityProto.SendMessageRequest{
		ChatId:   chatID,
		SenderId: userID,
		Content:  content,
	})
	if err != nil {
		return "", err
	}

	return resp.Message.Id, nil
}

// MarkMessageAsRead implements CommunityServiceClient
func (c *GRPCCommunityServiceClient) MarkMessageAsRead(chatID, messageID, userID string) error {
	if c.client == nil {
		return fmt.Errorf("community service client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// For now just handle this as unsend - we can add proper mark as read later
	_, err := c.client.UnsendMessage(ctx, &communityProto.UnsendMessageRequest{
		MessageId: messageID,
	})
	return err
}

// GetMessages implements CommunityServiceClient
func (c *GRPCCommunityServiceClient) GetMessages(chatID string, limit, offset int) ([]Message, error) {
	if c.client == nil {
		return nil, fmt.Errorf("community service client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := c.client.ListMessages(ctx, &communityProto.ListMessagesRequest{
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

// ValidateUser implements CommunityServiceClient
func (c *GRPCCommunityServiceClient) ValidateUser(userID string) (bool, error) {
	if c.userClient == nil {
		return false, fmt.Errorf("user service client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Call User service to validate user
	_, err := c.userClient.GetUser(ctx, &userProto.GetUserRequest{
		UserId: userID,
	})

	if err != nil {
		// Check if the error is because the user doesn't exist
		if strings.Contains(err.Error(), "not found") {
			return false, nil
		}
		return false, err
	}

	return true, nil
}
