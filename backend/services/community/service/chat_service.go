package service

import (
	"aycom/backend/proto/community"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	"aycom/backend/services/community/model"
)

type Message struct {
	ID        string
	ChatID    string
	SenderID  string
	Content   string
	Timestamp time.Time
	IsRead    bool
	IsEdited  bool
	IsDeleted bool
}

type Chat struct {
	ID          string
	Name        string
	Description string
	CreatorID   string
	CommunityID string
	IsGroupChat bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Participant struct {
	ID       string
	ChatID   string
	UserID   string
	JoinedAt time.Time
}

type DeletedChat struct {
	ChatID    string
	UserID    string
	DeletedAt time.Time
}

type ChatRepository interface {
	CreateChat(chat *Chat) error
	FindChatByID(chatID string) (*Chat, error)
	ListChatsByUserID(userID string, limit, offset int) ([]*Chat, error)
	UpdateChat(chat *Chat) error
	DeleteChat(chatID string) error
}

type MessageRepository interface {
	SaveMessage(message *Message) error
	FindMessageByID(messageID string) (*Message, error)
	FindMessagesByChatID(chatID string, limit, offset int) ([]*Message, error)
	MarkMessageAsRead(messageID, userID string) error
	DeleteMessage(messageID string) error
	UnsendMessage(messageID string) error
	SearchMessages(chatID, query string, limit, offset int) ([]*Message, error)
	UpdateMessage(message *Message) error
}

type ParticipantRepository interface {
	AddParticipant(participant *Participant) error
	RemoveParticipant(chatID, userID string) error
	ListParticipantsByChatID(chatID string, limit, offset int) ([]*Participant, error)
	IsUserInChat(chatID, userID string) (bool, error)
}

type DeletedChatRepository interface {
	MarkChatAsDeleted(chatID, userID string) error
	IsDeletedForUser(chatID, userID string) (bool, error)
}

type ChatService interface {
	ListChats(userID string, limit, offset int) ([]*community.Chat, error)
	CreateChat(name string, description string, creatorID string, isGroupChat bool, participantIDs []string) (*community.Chat, error)
	AddParticipant(chatID, userID, addedBy string) error
	RemoveParticipant(chatID, userID, removedBy string) error
	ListParticipants(chatID string, limit, offset int) ([]*community.ChatParticipant, error)
	SendMessage(chatID, userID, content string) (string, error)
	GetMessages(chatID string, limit, offset int) ([]*community.Message, error)
	DeleteMessage(chatID, messageID, userID string) error
	UnsendMessage(chatID, messageID, userID string) error
	SearchMessages(chatID, query string, limit, offset int) ([]*community.Message, error)
}

type chatService struct {
	chatRepo        model.ChatRepository
	messageRepo     model.MessageRepository
	participantRepo model.ParticipantRepository
	deletedChatRepo model.DeletedChatRepository
}

func NewChatService(
	chatRepo model.ChatRepository,
	participantRepo model.ParticipantRepository,
	messageRepo model.MessageRepository,
	deletedChatRepo model.DeletedChatRepository,
) ChatService {
	return &chatService{
		chatRepo:        chatRepo,
		participantRepo: participantRepo,
		messageRepo:     messageRepo,
		deletedChatRepo: deletedChatRepo,
	}
}

func toModelChatDTO(c *Chat) *model.ChatDTO {
	return &model.ChatDTO{
		ID:          c.ID,
		Name:        c.Name,
		Description: c.Description,
		CreatorID:   c.CreatorID,
		CommunityID: c.CommunityID,
		IsGroupChat: c.IsGroupChat,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}
}

func fromModelChatDTO(c *model.ChatDTO) *Chat {
	return &Chat{
		ID:          c.ID,
		Name:        c.Name,
		Description: c.Description,
		CreatorID:   c.CreatorID,
		CommunityID: c.CommunityID,
		IsGroupChat: c.IsGroupChat,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}
}

func toModelMessageDTO(m *Message) *model.MessageDTO {
	return &model.MessageDTO{
		ID:        m.ID,
		ChatID:    m.ChatID,
		SenderID:  m.SenderID,
		Content:   m.Content,
		Timestamp: m.Timestamp,
		IsRead:    m.IsRead,
		IsEdited:  m.IsEdited,
		IsDeleted: m.IsDeleted,
	}
}

func fromModelMessageDTO(m *model.MessageDTO) *Message {
	return &Message{
		ID:        m.ID,
		ChatID:    m.ChatID,
		SenderID:  m.SenderID,
		Content:   m.Content,
		Timestamp: m.Timestamp,
		IsRead:    m.IsRead,
		IsEdited:  m.IsEdited,
		IsDeleted: m.IsDeleted,
	}
}

func toModelParticipantDTO(p *Participant) *model.ParticipantDTO {
	return &model.ParticipantDTO{
		ChatID:   p.ChatID,
		UserID:   p.UserID,
		IsAdmin:  false,
		JoinedAt: p.JoinedAt,
	}
}

func fromModelParticipantDTO(p *model.ParticipantDTO) *Participant {
	return &Participant{
		ID:       uuid.New().String(),
		ChatID:   p.ChatID,
		UserID:   p.UserID,
		JoinedAt: p.JoinedAt,
	}
}

func (s *chatService) CreateChat(name string, description string, creatorID string, isGroupChat bool, participantIDs []string) (*community.Chat, error) {
	log.Printf("Creating chat: name=%s, creator=%s, isGroup=%v, participants=%v", name, creatorID, isGroupChat, participantIDs)

	// Generate a new UUID for the chat
	chatID := uuid.New().String()

	// Create the chat
	now := time.Now()
	chat := &model.ChatDTO{
		ID:          chatID,
		Name:        name,
		IsGroupChat: isGroupChat,
		CreatorID:   creatorID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.chatRepo.CreateChat(toModelChatDTO(fromModelChatDTO(chat))); err != nil {
		log.Printf("Error creating chat: %v", err)
		return nil, fmt.Errorf("failed to create chat: %v", err)
	}

	// Add participants
	for _, userID := range participantIDs {
		participant := &model.ParticipantDTO{
			ChatID:   chatID,
			UserID:   userID,
			IsAdmin:  userID == creatorID, // Creator is admin
			JoinedAt: now,
		}

		if err := s.participantRepo.AddParticipant(participant); err != nil {
			log.Printf("Error adding participant %s to chat %s: %v", userID, chatID, err)
			// We don't return error here to avoid leaving chat with no participants
			// In production, this should use transactions
		}
	}

	// Return the created chat
	return &community.Chat{
		Id:        chatID,
		Name:      name,
		IsGroup:   isGroupChat,
		CreatedBy: creatorID,
		CreatedAt: timestamppb.New(now),
		UpdatedAt: timestamppb.New(now),
	}, nil
}

func (s *chatService) ListChats(userID string, limit, offset int) ([]*community.Chat, error) {
	chats, err := s.chatRepo.ListChatsByUserID(userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list chats: %v", err)
	}

	protoChats := make([]*community.Chat, 0, len(chats))
	for _, chat := range chats {
		// Skip deleted chats
		isDeleted, err := s.deletedChatRepo.IsDeletedForUser(chat.ID, userID)
		if err != nil {
			log.Printf("Error checking if chat %s is deleted for user %s: %v", chat.ID, userID, err)
			continue
		}

		if isDeleted {
			continue
		}

		protoChats = append(protoChats, &community.Chat{
			Id:        chat.ID,
			Name:      chat.Name,
			IsGroup:   chat.IsGroupChat,
			CreatedBy: chat.CreatorID,
			CreatedAt: timestamppb.New(chat.CreatedAt),
			UpdatedAt: timestamppb.New(chat.UpdatedAt),
		})
	}

	return protoChats, nil
}

func (s *chatService) ListParticipants(chatID string, limit, offset int) ([]*community.ChatParticipant, error) {
	participants, err := s.participantRepo.ListParticipantsByChatID(chatID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list participants: %v", err)
	}

	protoParticipants := make([]*community.ChatParticipant, len(participants))
	for i, participant := range participants {
		protoParticipants[i] = &community.ChatParticipant{
			ChatId:   participant.ChatID,
			UserId:   participant.UserID,
			IsAdmin:  participant.IsAdmin,
			JoinedAt: timestamppb.New(participant.JoinedAt),
		}
	}

	return protoParticipants, nil
}

func (s *chatService) AddParticipant(chatID, userID, addedBy string) error {
	// Check if the chat exists
	_, err := s.chatRepo.FindChatByID(chatID)
	if err != nil {
		return fmt.Errorf("chat not found: %v", err)
	}

	// Check if the adder is an admin
	isAdmin, err := s.isUserChatAdmin(chatID, addedBy)
	if err != nil {
		return fmt.Errorf("failed to check admin status: %v", err)
	}

	if !isAdmin {
		return fmt.Errorf("only admins can add participants")
	}

	// Add the participant
	participant := &model.ParticipantDTO{
		ChatID:   chatID,
		UserID:   userID,
		IsAdmin:  false,
		JoinedAt: time.Now(),
	}

	if err := s.participantRepo.AddParticipant(participant); err != nil {
		return fmt.Errorf("failed to add participant: %v", err)
	}

	return nil
}

func (s *chatService) RemoveParticipant(chatID, userID, removedBy string) error {
	// Check if the chat exists
	_, err := s.chatRepo.FindChatByID(chatID)
	if err != nil {
		return fmt.Errorf("chat not found: %v", err)
	}

	// Check if the remover is an admin
	isAdmin, err := s.isUserChatAdmin(chatID, removedBy)
	if err != nil {
		return fmt.Errorf("failed to check admin status: %v", err)
	}

	if !isAdmin && removedBy != userID {
		return fmt.Errorf("only admins can remove other participants")
	}

	// Remove the participant
	if err := s.participantRepo.RemoveParticipant(chatID, userID); err != nil {
		return fmt.Errorf("failed to remove participant: %v", err)
	}

	return nil
}

func (s *chatService) SendMessage(chatID, userID, content string) (string, error) {
	// Check if the user is a participant
	isParticipant, err := s.participantRepo.IsUserInChat(chatID, userID)
	if err != nil {
		return "", fmt.Errorf("failed to check if user is in chat: %v", err)
	}

	if !isParticipant {
		return "", fmt.Errorf("user is not a participant in this chat")
	}

	// Create the message
	messageID := uuid.New().String()
	message := &model.MessageDTO{
		ID:        messageID,
		ChatID:    chatID,
		SenderID:  userID,
		Content:   content,
		Timestamp: time.Now(),
		IsRead:    false,
		IsEdited:  false,
		IsDeleted: false,
	}

	if err := s.messageRepo.SaveMessage(message); err != nil {
		return "", fmt.Errorf("failed to save message: %v", err)
	}

	return messageID, nil
}

func (s *chatService) GetMessages(chatID string, limit, offset int) ([]*community.Message, error) {
	messages, err := s.messageRepo.FindMessagesByChatID(chatID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %v", err)
	}

	protoMessages := make([]*community.Message, len(messages))
	for i, message := range messages {
		protoMessages[i] = &community.Message{
			Id:       message.ID,
			ChatId:   message.ChatID,
			SenderId: message.SenderID,
			Content:  message.Content,
			SentAt:   timestamppb.New(message.Timestamp),
			Unsent:   message.IsDeleted,
		}
	}

	return protoMessages, nil
}

func (s *chatService) DeleteMessage(chatID, messageID, userID string) error {
	// Get the message
	message, err := s.messageRepo.FindMessageByID(messageID)
	if err != nil {
		return fmt.Errorf("failed to find message: %v", err)
	}

	// Check if the message belongs to the chat
	if message.ChatID != chatID {
		return fmt.Errorf("message does not belong to this chat")
	}

	// Check if the user is the sender or an admin
	if message.SenderID != userID {
		isAdmin, err := s.isUserChatAdmin(chatID, userID)
		if err != nil {
			return fmt.Errorf("failed to check admin status: %v", err)
		}

		if !isAdmin {
			return fmt.Errorf("only the sender or an admin can delete a message")
		}
	}

	// Delete the message
	if err := s.messageRepo.DeleteMessage(messageID); err != nil {
		return fmt.Errorf("failed to delete message: %v", err)
	}

	return nil
}

func (s *chatService) UnsendMessage(chatID, messageID, userID string) error {
	// Get the message
	message, err := s.messageRepo.FindMessageByID(messageID)
	if err != nil {
		return fmt.Errorf("failed to find message: %v", err)
	}

	// Check if the message belongs to the chat
	if message.ChatID != chatID {
		return fmt.Errorf("message does not belong to this chat")
	}

	// Check if the user is the sender
	if message.SenderID != userID {
		return fmt.Errorf("only the sender can unsend a message")
	}

	// Unsend the message
	if err := s.messageRepo.UnsendMessage(messageID); err != nil {
		return fmt.Errorf("failed to unsend message: %v", err)
	}

	return nil
}

func (s *chatService) SearchMessages(chatID, query string, limit, offset int) ([]*community.Message, error) {
	messages, err := s.messageRepo.SearchMessages(chatID, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search messages: %v", err)
	}

	protoMessages := make([]*community.Message, len(messages))
	for i, message := range messages {
		protoMessages[i] = &community.Message{
			Id:       message.ID,
			ChatId:   message.ChatID,
			SenderId: message.SenderID,
			Content:  message.Content,
			SentAt:   timestamppb.New(message.Timestamp),
			Unsent:   message.IsDeleted,
		}
	}

	return protoMessages, nil
}

func (s *chatService) isUserChatAdmin(chatID, userID string) (bool, error) {
	participants, err := s.participantRepo.ListParticipantsByChatID(chatID, 100, 0)
	if err != nil {
		return false, fmt.Errorf("failed to list participants: %v", err)
	}

	for _, participant := range participants {
		if participant.UserID == userID && participant.IsAdmin {
			return true, nil
		}
	}

	return false, nil
}
