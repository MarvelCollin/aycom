package service

import (
	"errors"
	"log"
	"time"

	"aycom/backend/proto/community"
	"aycom/backend/services/community/model"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

// Message represents a chat message
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

// Chat represents a chat room
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

// Participant represents a chat participant
type Participant struct {
	ID       string
	ChatID   string
	UserID   string
	JoinedAt time.Time
}

// DeletedChat represents a deleted chat
type DeletedChat struct {
	ChatID    string
	UserID    string
	DeletedAt time.Time
}

// Interfaces for repositories
type ChatRepository interface {
	CreateChat(chat *Chat) error
	FindChatByID(chatID string) (*Chat, error)
	ListChatsByUserID(userID string, limit, offset int) ([]*Chat, error)
	UpdateChat(chat *Chat) error
	DeleteChat(chatID string) error
}

// MessageRepository defines the methods for message operations
type MessageRepository interface {
	SaveMessage(message *Message) error
	FindMessageByID(messageID string) (*Message, error)
	FindMessagesByChatID(chatID string, limit, offset int) ([]*Message, error)
	MarkMessageAsRead(messageID, userID string) error
	DeleteMessage(messageID string) error
	UnsendMessage(messageID string) error
	SearchMessages(chatID, query string, limit, offset int) ([]*Message, error)
}

// ParticipantRepository defines the methods for participant operations
type ParticipantRepository interface {
	AddParticipant(participant *Participant) error
	RemoveParticipant(chatID, userID string) error
	ListParticipantsByChatID(chatID string, limit, offset int) ([]*Participant, error)
	IsUserInChat(chatID, userID string) (bool, error)
}

// DeletedChatRepository defines the methods for deleted chat operations
type DeletedChatRepository interface {
	MarkChatAsDeleted(chatID, userID string) error
	IsDeletedForUser(chatID, userID string) (bool, error)
}

// ChatService implements chat operations
type ChatService struct {
	chatRepo        model.ChatRepository
	messageRepo     model.MessageRepository
	participantRepo model.ParticipantRepository
	deletedChatRepo model.DeletedChatRepository
}

// NewChatService creates a new chat service
func NewChatService(
	chatRepo model.ChatRepository,
	participantRepo model.ParticipantRepository,
	messageRepo model.MessageRepository,
	deletedChatRepo model.DeletedChatRepository,
) *ChatService {
	return &ChatService{
		chatRepo:        chatRepo,
		participantRepo: participantRepo,
		messageRepo:     messageRepo,
		deletedChatRepo: deletedChatRepo,
	}
}

// Convert between model.ChatDTO and service Chat
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

// Convert between model.MessageDTO and service Message
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

// Convert between model.ParticipantDTO and service Participant
func toModelParticipantDTO(p *Participant) *model.ParticipantDTO {
	return &model.ParticipantDTO{
		ChatID:   p.ChatID,
		UserID:   p.UserID,
		JoinedAt: p.JoinedAt,
		IsAdmin:  false, // Default value, can be updated if needed
	}
}

func fromModelParticipantDTO(p *model.ParticipantDTO) *Participant {
	return &Participant{
		ID:       uuid.New().String(), // Generate a new ID since it's not in the DTO
		ChatID:   p.ChatID,
		UserID:   p.UserID,
		JoinedAt: p.JoinedAt,
	}
}

// CreateChat creates a new chat
func (s *ChatService) CreateChat(name string, description string, creatorID string, isGroupChat bool, participantIDs []string) (*community.Chat, error) {
	chat := &Chat{
		ID:          uuid.New().String(),
		Name:        name,
		Description: description,
		CreatorID:   creatorID,
		IsGroupChat: isGroupChat,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.chatRepo.CreateChat(toModelChatDTO(chat)); err != nil {
		return nil, err
	}

	// Add all participants to chat
	for _, participantID := range participantIDs {
		participant := &Participant{
			ID:       uuid.New().String(),
			ChatID:   chat.ID,
			UserID:   participantID,
			JoinedAt: time.Now(),
		}
		if err := s.participantRepo.AddParticipant(toModelParticipantDTO(participant)); err != nil {
			log.Printf("Error adding participant %s to chat %s: %v", participantID, chat.ID, err)
			// Continue adding other participants even if one fails
		}
	}

	// Add the creator as a participant if not already included
	creatorAlreadyAdded := false
	for _, participantID := range participantIDs {
		if participantID == creatorID {
			creatorAlreadyAdded = true
			break
		}
	}

	if !creatorAlreadyAdded {
		participant := &Participant{
			ID:       uuid.New().String(),
			ChatID:   chat.ID,
			UserID:   creatorID,
			JoinedAt: time.Now(),
		}
		if err := s.participantRepo.AddParticipant(toModelParticipantDTO(participant)); err != nil {
			log.Printf("Error adding creator %s to chat %s: %v", creatorID, chat.ID, err)
		}
	}

	// Return proto Chat
	return &community.Chat{
		Id:        chat.ID,
		Name:      chat.Name,
		IsGroup:   chat.IsGroupChat,
		CreatedBy: chat.CreatorID,
		CreatedAt: timestamppb.New(chat.CreatedAt),
		UpdatedAt: timestamppb.New(chat.UpdatedAt),
	}, nil
}

// GetChat gets a chat by ID
func (s *ChatService) GetChat(chatID string) (*Chat, error) {
	chatDTO, err := s.chatRepo.FindChatByID(chatID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrChatNotFound
		}
		return nil, err
	}
	return fromModelChatDTO(chatDTO), nil
}

// ListChats lists chats for a user
func (s *ChatService) ListChats(userID string, limit, offset int) ([]*community.Chat, error) {
	chatDTOs, err := s.chatRepo.ListChatsByUserID(userID, limit, offset)
	if err != nil {
		return nil, err
	}

	chats := make([]*community.Chat, len(chatDTOs))
	for i, dto := range chatDTOs {
		chats[i] = &community.Chat{
			Id:        dto.ID,
			Name:      dto.Name,
			IsGroup:   dto.IsGroupChat,
			CreatedBy: dto.CreatorID,
			CreatedAt: timestamppb.New(dto.CreatedAt),
			UpdatedAt: timestamppb.New(dto.UpdatedAt),
		}
	}

	return chats, nil
}

// AddParticipant adds a participant to a chat
func (s *ChatService) AddParticipant(chatID, userID, addedBy string) error {
	// Check if chat exists
	_, err := s.chatRepo.FindChatByID(chatID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrChatNotFound
		}
		return err
	}

	// Check if added by is in the chat
	isInChat, err := s.participantRepo.IsUserInChat(chatID, addedBy)
	if err != nil {
		return err
	}
	if !isInChat {
		return ErrNotChatParticipant
	}

	// Check if user is already in the chat
	isInChat, err = s.participantRepo.IsUserInChat(chatID, userID)
	if err != nil {
		return err
	}
	if isInChat {
		return errors.New("user is already in the chat")
	}

	// Add user to chat
	participant := &Participant{
		ID:       uuid.New().String(),
		ChatID:   chatID,
		UserID:   userID,
		JoinedAt: time.Now(),
	}
	return s.participantRepo.AddParticipant(toModelParticipantDTO(participant))
}

// RemoveParticipant removes a participant from a chat
func (s *ChatService) RemoveParticipant(chatID, userID, removedBy string) error {
	// Check if chat exists
	chatDTO, err := s.chatRepo.FindChatByID(chatID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrChatNotFound
		}
		return err
	}

	chat := fromModelChatDTO(chatDTO)

	// Check if removed by is in the chat
	isInChat, err := s.participantRepo.IsUserInChat(chatID, removedBy)
	if err != nil {
		return err
	}
	if !isInChat {
		return ErrNotChatParticipant
	}

	// Cannot remove creator unless it's a self-removal
	if userID == chat.CreatorID && userID != removedBy {
		return ErrPermissionDenied
	}

	// Remove user from chat
	return s.participantRepo.RemoveParticipant(chatID, userID)
}

// ListParticipants lists participants in a chat
func (s *ChatService) ListParticipants(chatID string, limit, offset int) ([]*community.ChatParticipant, error) {
	participantDTOs, err := s.participantRepo.ListParticipantsByChatID(chatID, limit, offset)
	if err != nil {
		return nil, err
	}

	participants := make([]*community.ChatParticipant, len(participantDTOs))
	for i, dto := range participantDTOs {
		participants[i] = &community.ChatParticipant{
			ChatId:   dto.ChatID,
			UserId:   dto.UserID,
			IsAdmin:  false, // Set default as false, would need additional logic to determine admin status
			JoinedAt: timestamppb.New(dto.JoinedAt),
		}
	}
	return participants, nil
}

// SendMessage sends a message to a chat
func (s *ChatService) SendMessage(chatID, userID, content string) (string, error) {
	// Check if chat exists
	_, err := s.chatRepo.FindChatByID(chatID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", ErrChatNotFound
		}
		return "", err
	}

	// Check if user is in the chat
	isInChat, err := s.participantRepo.IsUserInChat(chatID, userID)
	if err != nil {
		return "", err
	}
	if !isInChat {
		return "", ErrNotChatParticipant
	}

	// Create and save message
	messageID := uuid.New().String()
	message := &Message{
		ID:        messageID,
		ChatID:    chatID,
		SenderID:  userID,
		Content:   content,
		Timestamp: time.Now(),
		IsRead:    false,
		IsEdited:  false,
		IsDeleted: false,
	}

	if err := s.messageRepo.SaveMessage(toModelMessageDTO(message)); err != nil {
		return "", err
	}

	return messageID, nil
}

// GetMessages gets messages from a chat
func (s *ChatService) GetMessages(chatID string, limit, offset int) ([]*community.Message, error) {
	// Check if chat exists
	_, err := s.chatRepo.FindChatByID(chatID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrChatNotFound
		}
		return nil, err
	}

	messageDTOs, err := s.messageRepo.FindMessagesByChatID(chatID, limit, offset)
	if err != nil {
		return nil, err
	}

	messages := make([]*community.Message, len(messageDTOs))
	for i, dto := range messageDTOs {
		messages[i] = &community.Message{
			Id:               dto.ID,
			ChatId:           dto.ChatID,
			SenderId:         dto.SenderID,
			Content:          dto.Content,
			SentAt:           timestamppb.New(dto.Timestamp),
			Unsent:           !dto.IsRead, // Using IsRead as proxy for unsent status
			DeletedForAll:    dto.IsDeleted,
			DeletedForSender: false, // Not tracking per-user deletion yet
		}
	}
	return messages, nil
}

// MarkMessageAsRead marks a message as read
func (s *ChatService) MarkMessageAsRead(chatID, messageID, userID string) error {
	// Check if chat exists
	_, err := s.chatRepo.FindChatByID(chatID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrChatNotFound
		}
		return err
	}

	// Check if user is in the chat
	isInChat, err := s.participantRepo.IsUserInChat(chatID, userID)
	if err != nil {
		return err
	}
	if !isInChat {
		return ErrNotChatParticipant
	}

	// Mark message as read
	return s.messageRepo.MarkMessageAsRead(messageID, userID)
}

// DeleteMessage deletes a message
func (s *ChatService) DeleteMessage(chatID, messageID, userID string) error {
	// Get the message
	messageDTO, err := s.messageRepo.FindMessageByID(messageID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrMessageNotFound
		}
		return err
	}

	message := fromModelMessageDTO(messageDTO)

	// If chatID is not provided, use the one from the message
	if chatID == "" {
		chatID = message.ChatID
	}

	// Check if message belongs to the chat
	if message.ChatID != chatID {
		return ErrPermissionDenied
	}

	// Only sender can delete
	if message.SenderID != userID {
		return ErrPermissionDenied
	}

	return s.messageRepo.DeleteMessage(messageID)
}

// UnsendMessage unsends a message (marks as deleted but keeps in database)
func (s *ChatService) UnsendMessage(chatID, messageID, userID string) error {
	// Get the message
	messageDTO, err := s.messageRepo.FindMessageByID(messageID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrMessageNotFound
		}
		return err
	}

	message := fromModelMessageDTO(messageDTO)

	// If chatID is not provided, use the one from the message
	if chatID == "" {
		chatID = message.ChatID
	}

	// Check if message belongs to the chat
	if message.ChatID != chatID {
		return ErrPermissionDenied
	}

	// Only sender can unsend
	if message.SenderID != userID {
		return ErrPermissionDenied
	}

	return s.messageRepo.UnsendMessage(messageID)
}

// SearchMessages searches for messages in a chat
func (s *ChatService) SearchMessages(chatID, query string, limit, offset int) ([]*community.Message, error) {
	// Check if chat exists
	_, err := s.chatRepo.FindChatByID(chatID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrChatNotFound
		}
		return nil, err
	}

	messageDTOs, err := s.messageRepo.SearchMessages(chatID, query, limit, offset)
	if err != nil {
		return nil, err
	}

	messages := make([]*community.Message, len(messageDTOs))
	for i, dto := range messageDTOs {
		messages[i] = &community.Message{
			Id:               dto.ID,
			ChatId:           dto.ChatID,
			SenderId:         dto.SenderID,
			Content:          dto.Content,
			SentAt:           timestamppb.New(dto.Timestamp),
			Unsent:           !dto.IsRead, // Using IsRead as proxy for unsent status
			DeletedForAll:    dto.IsDeleted,
			DeletedForSender: false, // Not tracking per-user deletion yet
		}
	}
	return messages, nil
}
