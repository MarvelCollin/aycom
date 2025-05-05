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

	"github.com/google/uuid"
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

// MockCommunityServiceClient is a mock implementation for testing
type MockCommunityServiceClient struct {
	messages map[string][]Message // chatID -> messages
	users    map[string]bool      // userID -> exists
}

// Global instance of the community service client
var communityServiceClient CommunityServiceClient

// InitCommunityServiceClient initializes the community service client
func InitCommunityServiceClient(cfg *config.Config) {
	// Try to connect to the real services first
	communityAddr := cfg.Services.CommunityService
	userAddr := cfg.Services.UserService

	log.Printf("Attempting to connect to Community service at %s and User service at %s", communityAddr, userAddr)

	// Connect to Community service
	communityConn, err := grpc.Dial(communityAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithTimeout(5*time.Second),
	)
	if err != nil {
		log.Printf("CRITICAL: Failed to connect to Community service: %v", err)
		// Continue to try User service anyway
	}

	// Connect to User service for user validation
	userConn, err := grpc.Dial(userAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithTimeout(5*time.Second),
	)
	if err != nil {
		log.Printf("CRITICAL: Failed to connect to User service: %v", err)
		// Continue anyway
	}

	// Create clients with whatever connections we have
	var communityClient communityProto.CommunityServiceClient
	var userClient userProto.UserServiceClient

	if communityConn != nil {
		communityClient = communityProto.NewCommunityServiceClient(communityConn)
		log.Printf("Successfully connected to Community service at %s", communityAddr)
	}

	if userConn != nil {
		userClient = userProto.NewUserServiceClient(userConn)
		log.Printf("Successfully connected to User service at %s", userAddr)
	}

	communityServiceClient = &GRPCCommunityServiceClient{
		client:     communityClient,
		userClient: userClient,
	}
}

// NewMockCommunityServiceClient creates a new mock community service client
func NewMockCommunityServiceClient() *MockCommunityServiceClient {
	return &MockCommunityServiceClient{
		messages: make(map[string][]Message),
		users:    map[string]bool{"mock-user-1": true, "mock-user-2": true},
	}
}

// SendMessage implements CommunityServiceClient
func (m *MockCommunityServiceClient) SendMessage(chatID, userID, content string) (string, error) {
	// Validate user exists
	if !m.users[userID] {
		return "", fmt.Errorf("user %s not found", userID)
	}

	messageID := uuid.New().String()
	message := Message{
		ID:        messageID,
		ChatID:    chatID,
		SenderID:  userID,
		Content:   content,
		Timestamp: time.Now(),
		IsRead:    false,
		IsEdited:  false,
		IsDeleted: false,
	}

	if _, ok := m.messages[chatID]; !ok {
		m.messages[chatID] = []Message{}
	}
	m.messages[chatID] = append(m.messages[chatID], message)

	log.Printf("[MOCK] Message sent - ID: %s, Chat: %s, User: %s, Content: %s",
		messageID, chatID, userID, content)

	return messageID, nil
}

// MarkMessageAsRead implements CommunityServiceClient
func (m *MockCommunityServiceClient) MarkMessageAsRead(chatID, messageID, userID string) error {
	log.Printf("[MOCK] Message marked as read - ID: %s, Chat: %s, User: %s",
		messageID, chatID, userID)
	return nil
}

// GetMessages implements CommunityServiceClient
func (m *MockCommunityServiceClient) GetMessages(chatID string, limit, offset int) ([]Message, error) {
	messages, ok := m.messages[chatID]
	if !ok {
		return []Message{}, nil
	}

	if offset >= len(messages) {
		return []Message{}, nil
	}

	end := offset + limit
	if end > len(messages) {
		end = len(messages)
	}

	return messages[offset:end], nil
}

// ValidateUser implements CommunityServiceClient
func (m *MockCommunityServiceClient) ValidateUser(userID string) (bool, error) {
	exists := m.users[userID]
	log.Printf("[MOCK] User validation - ID: %s, Exists: %v", userID, exists)
	return exists, nil
}

// SendMessage implements CommunityServiceClient for the gRPC version
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

// MarkMessageAsRead implements CommunityServiceClient for the gRPC version
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

// GetMessages implements CommunityServiceClient for the gRPC version
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

// ValidateUser implements CommunityServiceClient for the gRPC version
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
