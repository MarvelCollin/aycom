package service

import (
	"aycom/backend/proto/community"
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"

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

type ChatService struct {
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
) *ChatService {
	return &ChatService{
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
		JoinedAt: p.JoinedAt,
		IsAdmin:  false,
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

	for _, participantID := range participantIDs {
		participant := &Participant{
			ID:       uuid.New().String(),
			ChatID:   chat.ID,
			UserID:   participantID,
			JoinedAt: time.Now(),
		}
		if err := s.participantRepo.AddParticipant(toModelParticipantDTO(participant)); err != nil {
			log.Printf("Error adding participant %s to chat %s: %v", participantID, chat.ID, err)

		}
	}

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

	return &community.Chat{
		Id:        chat.ID,
		Name:      chat.Name,
		IsGroup:   chat.IsGroupChat,
		CreatedBy: chat.CreatorID,
		CreatedAt: timestamppb.New(chat.CreatedAt),
		UpdatedAt: timestamppb.New(chat.UpdatedAt),
	}, nil
}

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

func (s *ChatService) AddParticipant(chatID, userID, addedBy string) error {

	_, err := s.chatRepo.FindChatByID(chatID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrChatNotFound
		}
		return err
	}

	isInChat, err := s.participantRepo.IsUserInChat(chatID, addedBy)
	if err != nil {
		return err
	}
	if !isInChat {
		return ErrNotChatParticipant
	}

	isInChat, err = s.participantRepo.IsUserInChat(chatID, userID)
	if err != nil {
		return err
	}
	if isInChat {
		return errors.New("user is already in the chat")
	}

	participant := &Participant{
		ID:       uuid.New().String(),
		ChatID:   chatID,
		UserID:   userID,
		JoinedAt: time.Now(),
	}
	return s.participantRepo.AddParticipant(toModelParticipantDTO(participant))
}

func (s *ChatService) RemoveParticipant(chatID, userID, removedBy string) error {

	chatDTO, err := s.chatRepo.FindChatByID(chatID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrChatNotFound
		}
		return err
	}

	chat := fromModelChatDTO(chatDTO)

	isInChat, err := s.participantRepo.IsUserInChat(chatID, removedBy)
	if err != nil {
		return err
	}
	if !isInChat {
		return ErrNotChatParticipant
	}

	if userID == chat.CreatorID && userID != removedBy {
		return ErrPermissionDenied
	}

	return s.participantRepo.RemoveParticipant(chatID, userID)
}

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
			IsAdmin:  false,
			JoinedAt: timestamppb.New(dto.JoinedAt),
		}
	}
	return participants, nil
}

func (s *ChatService) SendMessage(chatID, userID, content string) (string, error) {

	_, err := s.chatRepo.FindChatByID(chatID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", ErrChatNotFound
		}
		return "", err
	}

	isInChat, err := s.participantRepo.IsUserInChat(chatID, userID)
	if err != nil {
		return "", err
	}
	if !isInChat {
		return "", ErrNotChatParticipant
	}

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
		log.Printf("Error saving message to database: %v", err)
		return "", err
	}

	log.Printf("Message saved to database with ID: %s", messageID)
	return messageID, nil
}

func (s *ChatService) GetMessages(chatID string, limit, offset int) ([]*community.Message, error) {

	_, err := s.chatRepo.FindChatByID(chatID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrChatNotFound
		}
		return nil, err
	}

	messageDTOs, err := s.messageRepo.FindMessagesByChatID(chatID, limit, offset)
	if err != nil {
		log.Printf("Error retrieving messages from database: %v", err)
		return nil, err
	}

	log.Printf("Retrieved %d messages for chat %s", len(messageDTOs), chatID)

	messages := make([]*community.Message, len(messageDTOs))
	for i, dto := range messageDTOs {
		messages[i] = &community.Message{
			Id:               dto.ID,
			ChatId:           dto.ChatID,
			SenderId:         dto.SenderID,
			Content:          dto.Content,
			SentAt:           timestamppb.New(dto.Timestamp),
			Unsent:           !dto.IsRead,
			DeletedForAll:    dto.IsDeleted,
			DeletedForSender: false,
		}
	}
	return messages, nil
}

func (s *ChatService) MarkMessageAsRead(chatID, messageID, userID string) error {

	_, err := s.chatRepo.FindChatByID(chatID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrChatNotFound
		}
		return err
	}

	isInChat, err := s.participantRepo.IsUserInChat(chatID, userID)
	if err != nil {
		return err
	}
	if !isInChat {
		return ErrNotChatParticipant
	}

	return s.messageRepo.MarkMessageAsRead(messageID, userID)
}

func (s *ChatService) DeleteMessage(chatID, messageID, userID string) error {

	messageDTO, err := s.messageRepo.FindMessageByID(messageID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrMessageNotFound
		}
		return err
	}

	message := fromModelMessageDTO(messageDTO)

	if chatID == "" {
		chatID = message.ChatID
	}

	if message.ChatID != chatID {
		return ErrPermissionDenied
	}

	if message.SenderID != userID {
		return ErrPermissionDenied
	}

	return s.messageRepo.DeleteMessage(messageID)
}

func (s *ChatService) UnsendMessage(chatID, messageID, userID string) error {

	messageDTO, err := s.messageRepo.FindMessageByID(messageID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrMessageNotFound
		}
		return err
	}

	message := fromModelMessageDTO(messageDTO)

	if chatID == "" {
		chatID = message.ChatID
	}

	if message.ChatID != chatID {
		return ErrPermissionDenied
	}

	if message.SenderID != userID {
		return ErrPermissionDenied
	}

	return s.messageRepo.UnsendMessage(messageID)
}

func (s *ChatService) SearchMessages(chatID, query string, limit, offset int) ([]*community.Message, error) {

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
			Unsent:           !dto.IsRead,
			DeletedForAll:    dto.IsDeleted,
			DeletedForSender: false,
		}
	}
	return messages, nil
}

func (s *ChatService) EditMessage(chatID, userID, messageID, newContent string) error {

	messageDTO, err := s.messageRepo.FindMessageByID(messageID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrMessageNotFound
		}
		return err
	}

	if chatID == "" {
		chatID = messageDTO.ChatID
	}

	if messageDTO.ChatID != chatID {
		return ErrPermissionDenied
	}

	if messageDTO.SenderID != userID {
		return ErrPermissionDenied
	}

	messageDTO.Content = newContent
	messageDTO.IsEdited = true

	return s.messageRepo.UpdateMessage(messageDTO)
}

func (s *ChatService) GetMessageById(messageID string) (*community.Message, error) {

	messageDTO, err := s.messageRepo.FindMessageByID(messageID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrMessageNotFound
		}
		return nil, err
	}

	return &community.Message{
		Id:               messageDTO.ID,
		ChatId:           messageDTO.ChatID,
		SenderId:         messageDTO.SenderID,
		Content:          messageDTO.Content,
		SentAt:           timestamppb.New(messageDTO.Timestamp),
		Unsent:           !messageDTO.IsRead,
		DeletedForAll:    messageDTO.IsDeleted,
		DeletedForSender: false,
	}, nil
}
