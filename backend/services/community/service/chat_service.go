package service

import (
	"errors"
	"log"
	"time"

	"aycom/backend/services/community/model"

	"github.com/google/uuid"
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
		ID:       p.ID,
		ChatID:   p.ChatID,
		UserID:   p.UserID,
		JoinedAt: p.JoinedAt,
	}
}

func fromModelParticipantDTO(p *model.ParticipantDTO) *Participant {
	return &Participant{
		ID:       p.ID,
		ChatID:   p.ChatID,
		UserID:   p.UserID,
		JoinedAt: p.JoinedAt,
	}
}

// CreateChat creates a new chat
func (s *ChatService) CreateChat(name, description string, creatorID, communityID uuid.UUID, isGroupChat bool, participantIDs []uuid.UUID) (*Chat, error) {
	chat := &Chat{
		ID:          uuid.New().String(),
		Name:        name,
		Description: description,
		CreatorID:   creatorID.String(),
		CommunityID: communityID.String(),
		IsGroupChat: isGroupChat,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.chatRepo.CreateChat(toModelChatDTO(chat)); err != nil {
		return nil, err
	}

	// Add creator as participant
	creatorParticipant := &Participant{
		ID:       uuid.New().String(),
		ChatID:   chat.ID,
		UserID:   creatorID.String(),
		JoinedAt: time.Now(),
	}
	if err := s.participantRepo.AddParticipant(toModelParticipantDTO(creatorParticipant)); err != nil {
		log.Printf("Error adding creator as participant: %v", err)
		// Don't return error, continue with adding other participants
	}

	// Add other participants
	for _, participantID := range participantIDs {
		if participantID == creatorID {
			continue // Creator already added
		}
		participant := &Participant{
			ID:       uuid.New().String(),
			ChatID:   chat.ID,
			UserID:   participantID.String(),
			JoinedAt: time.Now(),
		}
		if err := s.participantRepo.AddParticipant(toModelParticipantDTO(participant)); err != nil {
			log.Printf("Error adding participant %s: %v", participantID, err)
			// Don't return error, continue with adding other participants
		}
	}

	return chat, nil
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
func (s *ChatService) ListChats(userID string, limit, offset int) ([]*Chat, error) {
	chatDTOs, err := s.chatRepo.ListChatsByUserID(userID, limit, offset)
	if err != nil {
		return nil, err
	}

	chats := make([]*Chat, len(chatDTOs))
	for i, dto := range chatDTOs {
		chats[i] = fromModelChatDTO(dto)
	}
	return chats, nil
}

// AddParticipant adds a participant to a chat
func (s *ChatService) AddParticipant(chatID, userID, addedBy string) error {
	// Check if chat exists
	_, err := s.chatRepo.FindChatByID(chatID)
	if err != nil {
		return err
	}

	// Check if added by is in the chat
	isInChat, err := s.participantRepo.IsUserInChat(chatID, addedBy)
	if err != nil {
		return err
	}
	if !isInChat {
		return errors.New("only chat participants can add new participants")
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
		return err
	}

	chat := fromModelChatDTO(chatDTO)

	// Check if removed by is in the chat
	isInChat, err := s.participantRepo.IsUserInChat(chatID, removedBy)
	if err != nil {
		return err
	}
	if !isInChat {
		return errors.New("only chat participants can remove participants")
	}

	// Cannot remove creator unless it's a self-removal
	if userID == chat.CreatorID && userID != removedBy {
		return errors.New("cannot remove chat creator")
	}

	// Remove user from chat
	return s.participantRepo.RemoveParticipant(chatID, userID)
}

// ListParticipants lists participants in a chat
func (s *ChatService) ListParticipants(chatID string, limit, offset int) ([]*Participant, error) {
	participantDTOs, err := s.participantRepo.ListParticipantsByChatID(chatID, limit, offset)
	if err != nil {
		return nil, err
	}

	participants := make([]*Participant, len(participantDTOs))
	for i, dto := range participantDTOs {
		participants[i] = fromModelParticipantDTO(dto)
	}
	return participants, nil
}

// SendMessage sends a message to a chat
func (s *ChatService) SendMessage(chatID, userID, content string) (string, error) {
	// Check if chat exists
	_, err := s.chatRepo.FindChatByID(chatID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errors.New("chat not found")
		}
		return "", err
	}

	// Check if user is in the chat
	isInChat, err := s.participantRepo.IsUserInChat(chatID, userID)
	if err != nil {
		return "", err
	}
	if !isInChat {
		return "", errors.New("user is not a participant in this chat")
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
func (s *ChatService) GetMessages(chatID string, limit, offset int) ([]*Message, error) {
	// Check if chat exists
	_, err := s.chatRepo.FindChatByID(chatID)
	if err != nil {
		return nil, err
	}

	messageDTOs, err := s.messageRepo.FindMessagesByChatID(chatID, limit, offset)
	if err != nil {
		return nil, err
	}

	messages := make([]*Message, len(messageDTOs))
	for i, dto := range messageDTOs {
		messages[i] = fromModelMessageDTO(dto)
	}
	return messages, nil
}

// MarkMessageAsRead marks a message as read
func (s *ChatService) MarkMessageAsRead(chatID, messageID, userID string) error {
	// Check if chat exists
	_, err := s.chatRepo.FindChatByID(chatID)
	if err != nil {
		return err
	}

	// Check if user is in the chat
	isInChat, err := s.participantRepo.IsUserInChat(chatID, userID)
	if err != nil {
		return err
	}
	if !isInChat {
		return errors.New("user is not a participant in this chat")
	}

	// Mark message as read
	return s.messageRepo.MarkMessageAsRead(messageID, userID)
}

// DeleteMessage deletes a message
func (s *ChatService) DeleteMessage(chatID, messageID, userID string) error {
	// Get the message
	messageDTO, err := s.messageRepo.FindMessageByID(messageID)
	if err != nil {
		return err
	}

	message := fromModelMessageDTO(messageDTO)

	// Check if message belongs to the chat
	if message.ChatID != chatID {
		return errors.New("message does not belong to the specified chat")
	}

	// Only sender can delete
	if message.SenderID != userID {
		return errors.New("only the sender can delete a message")
	}

	return s.messageRepo.DeleteMessage(messageID)
}

// UnsendMessage unsends a message (marks as deleted but keeps in database)
func (s *ChatService) UnsendMessage(chatID, messageID, userID string) error {
	// Get the message
	messageDTO, err := s.messageRepo.FindMessageByID(messageID)
	if err != nil {
		return err
	}

	message := fromModelMessageDTO(messageDTO)

	// Check if message belongs to the chat
	if message.ChatID != chatID {
		return errors.New("message does not belong to the specified chat")
	}

	// Only sender can unsend
	if message.SenderID != userID {
		return errors.New("only the sender can unsend a message")
	}

	return s.messageRepo.UnsendMessage(messageID)
}

// SearchMessages searches for messages in a chat
func (s *ChatService) SearchMessages(chatID, query string, limit, offset int) ([]*Message, error) {
	// Check if chat exists
	_, err := s.chatRepo.FindChatByID(chatID)
	if err != nil {
		return nil, err
	}

	messageDTOs, err := s.messageRepo.SearchMessages(chatID, query, limit, offset)
	if err != nil {
		return nil, err
	}

	messages := make([]*Message, len(messageDTOs))
	for i, dto := range messageDTOs {
		messages[i] = fromModelMessageDTO(dto)
	}
	return messages, nil
}
