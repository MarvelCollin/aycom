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

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
		log.Printf("SendMessage: Community service client not initialized")
		return "", fmt.Errorf("community service client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Printf("SendMessage: Preparing gRPC call to send message. ChatID=%s, UserID=%s", chatID, userID)

	// Validate IDs before sending to gRPC
	_, err := uuid.Parse(chatID)
	if err != nil {
		log.Printf("SendMessage: Invalid chat ID format: %s (%v)", chatID, err)
		return "", fmt.Errorf("invalid chat ID format: %v", err)
	}

	_, err = uuid.Parse(userID)
	if err != nil {
		log.Printf("SendMessage: Invalid user ID format: %s (%v)", userID, err)
		return "", fmt.Errorf("invalid user ID format: %v", err)
	}

	log.Printf("SendMessage: Making gRPC call to send message. ChatID=%s, UserID=%s", chatID, userID)
	resp, err := c.grpcClient.SendMessage(ctx, &communityProto.SendMessageRequest{
		ChatId:   chatID,
		SenderId: userID,
		Content:  content,
	})

	if err != nil {
		log.Printf("SendMessage: Error from gRPC service: %v", err)

		// Add detailed error information
		st, ok := status.FromError(err)
		if ok {
			log.Printf("SendMessage: gRPC status code: %s, message: %s", st.Code(), st.Message())
			if st.Code() == codes.NotFound {
				return "", fmt.Errorf("chat not found: %v", err)
			}
			if st.Code() == codes.PermissionDenied {
				return "", fmt.Errorf("user not allowed to send message: %v", err)
			}
		}

		return "", fmt.Errorf("error sending message: %v", err)
	}

	if resp == nil {
		log.Printf("SendMessage: Received nil response from gRPC service")
		return "", fmt.Errorf("received nil response from service")
	}

	if resp.Message == nil {
		log.Printf("SendMessage: Received response with nil message field")
		return "", fmt.Errorf("received response with nil message field")
	}

	log.Printf("SendMessage: Successfully sent message with ID: %s", resp.Message.Id)
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
		chat := Chat{
			ID:          protoChat.Id,
			Name:        protoChat.Name,
			IsGroupChat: protoChat.IsGroup,
			CreatedBy:   protoChat.CreatedBy,
			CreatedAt:   protoChat.CreatedAt.AsTime(),
			UpdatedAt:   protoChat.UpdatedAt.AsTime(),
		}

		// Get participants for this chat
		participantResp, err := c.grpcClient.ListChatParticipants(ctx, &communityProto.ListChatParticipantsRequest{
			ChatId: protoChat.Id,
		})

		if err == nil && participantResp != nil && len(participantResp.Participants) > 0 {
			participantIDs := make([]string, len(participantResp.Participants))
			for j, participant := range participantResp.Participants {
				participantIDs[j] = participant.UserId
			}
			chat.Participants = participantIDs
		}

		// Get last message if available
		messagesResp, err := c.grpcClient.ListMessages(ctx, &communityProto.ListMessagesRequest{
			ChatId: protoChat.Id,
			Limit:  1,
			Offset: 0,
		})

		if err == nil && messagesResp != nil && len(messagesResp.Messages) > 0 {
			lastMsg := messagesResp.Messages[0]
			chat.LastMessage = &Message{
				ID:        lastMsg.Id,
				ChatID:    lastMsg.ChatId,
				SenderID:  lastMsg.SenderId,
				Content:   lastMsg.Content,
				Timestamp: lastMsg.SentAt.AsTime(),
				IsRead:    !lastMsg.Unsent,
				IsEdited:  false,
				IsDeleted: lastMsg.DeletedForAll,
			}
		}

		chats[i] = chat
	}

	return chats, nil
}

func (c *communityCommunicationClient) CreateChat(isGroup bool, name string, participantIDs []string, createdBy string) (*Chat, error) {
	if c.grpcClient == nil {
		return nil, fmt.Errorf("community service client not initialized")
	}

	// Validate inputs
	if len(participantIDs) == 0 {
		return nil, fmt.Errorf("at least one participant is required")
	}

	// For group chats, name is required
	if isGroup && (name == "" || len(name) == 0) {
		return nil, fmt.Errorf("name is required for group chats")
	}

	// Make sure the creator is included in participants
	creatorIncluded := false
	for _, id := range participantIDs {
		if id == createdBy {
			creatorIncluded = true
			break
		}
	}

	// If creator is not in the participants list, add them
	if !creatorIncluded {
		participantIDs = append(participantIDs, createdBy)
	}

	// Log the request details
	log.Printf("Creating chat: isGroup=%v, name=%s, participants=%v, createdBy=%s",
		isGroup, name, participantIDs, createdBy)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := c.grpcClient.CreateChat(ctx, &communityProto.CreateChatRequest{
		IsGroup:        isGroup,
		Name:           name,
		ParticipantIds: participantIDs,
		CreatedBy:      createdBy,
	})
	if err != nil {
		log.Printf("Error creating chat: %v", err)
		return nil, err
	}

	log.Printf("Chat created successfully: %s", resp.Chat.Id)

	chat := &Chat{
		ID:           resp.Chat.Id,
		Name:         resp.Chat.Name,
		IsGroupChat:  resp.Chat.IsGroup,
		CreatedBy:    resp.Chat.CreatedBy,
		CreatedAt:    resp.Chat.CreatedAt.AsTime(),
		UpdatedAt:    resp.Chat.UpdatedAt.AsTime(),
		Participants: participantIDs,
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
