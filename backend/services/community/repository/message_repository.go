package repository

import (
	"aycom/backend/services/community/model"
	"log"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MessageDBModel represents a message in the database
type MessageDBModel struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey;column:message_id"`
	ChatID    uuid.UUID  `gorm:"type:uuid;not null;column:chat_id"`
	SenderID  uuid.UUID  `gorm:"type:uuid;not null;column:sender_id"`
	Content   string     `gorm:"type:text;not null"`
	SentAt    time.Time  `gorm:"not null;autoCreateTime"`
	IsRead    bool       `gorm:"default:false"`
	IsEdited  bool       `gorm:"default:false"`
	IsDeleted bool       `gorm:"default:false"`
	DeletedAt *time.Time `gorm:"index"`
}

// TableName sets the table name for the message model
func (MessageDBModel) TableName() string {
	return "messages"
}

// GormMessageRepository implements model.MessageRepository
type GormMessageRepository struct {
	db *gorm.DB
}

// NewMessageRepository creates a new message repository
func NewMessageRepository(db *gorm.DB) model.MessageRepository {
	return &GormMessageRepository{db: db}
}

// SaveMessage saves a message to the database
func (r *GormMessageRepository) SaveMessage(message *model.MessageDTO) error {
	log.Printf("Saving message to database: ID=%s, ChatID=%s, SenderID=%s",
		message.ID, message.ChatID, message.SenderID)

	msgID, err := uuid.Parse(message.ID)
	if err != nil {
		log.Printf("Error parsing message ID: %v", err)
		return err
	}

	chatID, err := uuid.Parse(message.ChatID)
	if err != nil {
		log.Printf("Error parsing chat ID: %v", err)
		return err
	}

	senderID, err := uuid.Parse(message.SenderID)
	if err != nil {
		log.Printf("Error parsing sender ID: %v", err)
		return err
	}

	dbMessage := &MessageDBModel{
		ID:        msgID,
		ChatID:    chatID,
		SenderID:  senderID,
		Content:   message.Content,
		SentAt:    message.Timestamp,
		IsRead:    message.IsRead,
		IsEdited:  message.IsEdited,
		IsDeleted: message.IsDeleted,
	}

	err = r.db.Create(dbMessage).Error
	if err != nil {
		log.Printf("Error saving message to database: %v", err)
		return err
	}

	log.Printf("Message saved successfully to database with ID: %s", message.ID)
	return nil
}

// FindMessageByID finds a message by ID
func (r *GormMessageRepository) FindMessageByID(messageID string) (*model.MessageDTO, error) {
	msgID, err := uuid.Parse(messageID)
	if err != nil {
		return nil, err
	}

	var dbMessage MessageDBModel
	err = r.db.First(&dbMessage, "message_id = ?", msgID).Error
	if err != nil {
		return nil, err
	}

	return &model.MessageDTO{
		ID:        dbMessage.ID.String(),
		ChatID:    dbMessage.ChatID.String(),
		SenderID:  dbMessage.SenderID.String(),
		Content:   dbMessage.Content,
		Timestamp: dbMessage.SentAt,
		IsRead:    dbMessage.IsRead,
		IsEdited:  dbMessage.IsEdited,
		IsDeleted: dbMessage.IsDeleted,
	}, nil
}

// FindMessagesByChatID finds messages by chat ID
func (r *GormMessageRepository) FindMessagesByChatID(chatID string, limit, offset int) ([]*model.MessageDTO, error) {
	log.Printf("Finding messages for chat ID: %s (limit: %d, offset: %d)", chatID, limit, offset)

	chatUUID, err := uuid.Parse(chatID)
	if err != nil {
		log.Printf("Error parsing chat ID: %v", err)
		return nil, err
	}

	var dbMessages []MessageDBModel
	query := r.db.Where("chat_id = ?", chatUUID)

	// Add ordering by sent_at in descending order (latest messages first)
	query = query.Order("sent_at DESC")

	// Apply limit and offset
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	// Execute the query
	err = query.Find(&dbMessages).Error
	if err != nil {
		log.Printf("Database error retrieving messages: %v", err)
		return nil, err
	}

	log.Printf("Found %d messages for chat ID: %s", len(dbMessages), chatID)

	messages := make([]*model.MessageDTO, len(dbMessages))
	for i, dbMessage := range dbMessages {
		messages[i] = &model.MessageDTO{
			ID:        dbMessage.ID.String(),
			ChatID:    dbMessage.ChatID.String(),
			SenderID:  dbMessage.SenderID.String(),
			Content:   dbMessage.Content,
			Timestamp: dbMessage.SentAt,
			IsRead:    dbMessage.IsRead,
			IsEdited:  dbMessage.IsEdited,
			IsDeleted: dbMessage.IsDeleted,
		}
	}

	return messages, nil
}

// MarkMessageAsRead marks a message as read
func (r *GormMessageRepository) MarkMessageAsRead(messageID, userID string) error {
	msgID, err := uuid.Parse(messageID)
	if err != nil {
		return err
	}

	return r.db.Model(&MessageDBModel{}).
		Where("message_id = ?", msgID).
		Update("is_read", true).
		Error
}

// DeleteMessage deletes a message from the database
func (r *GormMessageRepository) DeleteMessage(messageID string) error {
	msgID, err := uuid.Parse(messageID)
	if err != nil {
		return err
	}

	return r.db.Delete(&MessageDBModel{}, "message_id = ?", msgID).Error
}

// UnsendMessage marks a message as deleted but doesn't remove it
func (r *GormMessageRepository) UnsendMessage(messageID string) error {
	msgID, err := uuid.Parse(messageID)
	if err != nil {
		return err
	}

	return r.db.Model(&MessageDBModel{}).
		Where("message_id = ?", msgID).
		Update("is_deleted", true).
		Error
}

// SearchMessages searches for messages by content
func (r *GormMessageRepository) SearchMessages(chatID, query string, limit, offset int) ([]*model.MessageDTO, error) {
	chatUUID, err := uuid.Parse(chatID)
	if err != nil {
		return nil, err
	}

	var dbMessages []MessageDBModel
	err = r.db.Where("chat_id = ? AND content LIKE ?", chatUUID, "%"+query+"%").
		Order("sent_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&dbMessages).Error
	if err != nil {
		return nil, err
	}

	messages := make([]*model.MessageDTO, len(dbMessages))
	for i, dbMessage := range dbMessages {
		messages[i] = &model.MessageDTO{
			ID:        dbMessage.ID.String(),
			ChatID:    dbMessage.ChatID.String(),
			SenderID:  dbMessage.SenderID.String(),
			Content:   dbMessage.Content,
			Timestamp: dbMessage.SentAt,
			IsRead:    dbMessage.IsRead,
			IsEdited:  dbMessage.IsEdited,
			IsDeleted: dbMessage.IsDeleted,
		}
	}

	return messages, nil
}

// UpdateMessage updates a message in the database
func (r *GormMessageRepository) UpdateMessage(message *model.MessageDTO) error {
	msgID, err := uuid.Parse(message.ID)
	if err != nil {
		return err
	}

	// Prepare message data to update
	updateData := map[string]interface{}{
		"content":    message.Content,
		"is_edited":  message.IsEdited,
		"is_read":    message.IsRead,
		"is_deleted": message.IsDeleted,
	}

	// Update the message
	return r.db.Model(&MessageDBModel{}).
		Where("message_id = ?", msgID).
		Updates(updateData).
		Error
}
