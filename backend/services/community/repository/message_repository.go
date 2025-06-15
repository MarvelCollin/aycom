package repository

import (
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"aycom/backend/services/community/model"
)

type MessageDBModel struct {
	ID               uuid.UUID  `gorm:"type:uuid;primaryKey;column:message_id"`
	ChatID           uuid.UUID  `gorm:"type:uuid;not null;column:chat_id"`
	SenderID         uuid.UUID  `gorm:"type:uuid;not null;column:sender_id"`
	Content          string     `gorm:"type:text;column:content"`
	MediaURL         string     `gorm:"type:varchar(512);column:media_url"`
	MediaType        string     `gorm:"type:varchar(10);column:media_type"`
	SentAt           time.Time  `gorm:"column:sent_at;not null"`
	Unsent           bool       `gorm:"default:false;not null;column:unsent"`
	UnsentAt         *time.Time `gorm:"column:unsent_at"`
	DeletedForSender bool       `gorm:"default:false;not null;column:deleted_for_sender"`
	DeletedForAll    bool       `gorm:"default:false;not null;column:deleted_for_all"`
	ReplyToMessageID *uuid.UUID `gorm:"type:uuid;column:reply_to_message_id"`
	IsRead           bool       `gorm:"default:false;column:is_read"`
	IsEdited         bool       `gorm:"default:false;column:is_edited"`
	IsDeleted        bool       `gorm:"default:false;column:is_deleted"`
	CreatedAt        time.Time  `gorm:"column:created_at"`
	UpdatedAt        time.Time  `gorm:"column:updated_at"`
	DeletedAt        *time.Time `gorm:"index;column:deleted_at"`
}

func (MessageDBModel) TableName() string {
	return "messages"
}

type GormMessageRepository struct {
	db *gorm.DB
}

func NewMessageRepository(db *gorm.DB) model.MessageRepository {
	return &GormMessageRepository{db: db}
}

func (r *GormMessageRepository) SaveMessage(message *model.MessageDTO) error {
	log.Printf("Saving message to database: ID=%s, ChatID=%s, SenderID=%s",
		message.ID, message.ChatID, message.SenderID)

	// Validate inputs
	if message.ID == "" {
		log.Printf("Error: Message ID is empty")
		return fmt.Errorf("message ID cannot be empty")
	}
	if message.ChatID == "" {
		log.Printf("Error: Chat ID is empty")
		return fmt.Errorf("chat ID cannot be empty")
	}
	if message.SenderID == "" {
		log.Printf("Error: Sender ID is empty")
		return fmt.Errorf("sender ID cannot be empty")
	}

	// Parse UUIDs with error handling
	msgID, err := uuid.Parse(message.ID)
	if err != nil {
		log.Printf("Error parsing message ID: %v", err)
		return fmt.Errorf("invalid message ID format: %v", err)
	}

	chatID, err := uuid.Parse(message.ChatID)
	if err != nil {
		log.Printf("Error parsing chat ID: %v", err)
		return fmt.Errorf("invalid chat ID format: %v", err)
	}

	senderID, err := uuid.Parse(message.SenderID)
	if err != nil {
		log.Printf("Error parsing sender ID: %v", err)
		return fmt.Errorf("invalid sender ID format: %v", err)
	}

	// Verify chat exists
	var chatCount int64
	err = r.db.Table("chats").Where("chat_id = ?", chatID).Count(&chatCount).Error
	if err != nil {
		log.Printf("Error checking if chat exists: %v", err)
		return fmt.Errorf("failed to verify chat: %v", err)
	}
	if chatCount == 0 {
		log.Printf("Error: Chat with ID %s does not exist", message.ChatID)
		return fmt.Errorf("chat with ID %s does not exist", message.ChatID)
	}

	// Verify sender is a participant
	var participantCount int64
	err = r.db.Table("chat_participants").
		Where("chat_id = ? AND user_id = ?", chatID, senderID).
		Count(&participantCount).Error
	if err != nil {
		log.Printf("Error checking if sender is a participant: %v", err)
		return fmt.Errorf("failed to verify sender participation: %v", err)
	}
	if participantCount == 0 {
		log.Printf("Error: User %s is not a participant in chat %s", message.SenderID, message.ChatID)
		return fmt.Errorf("user %s is not a participant in chat %s", message.SenderID, message.ChatID)
	}

	// Create the message
	dbMessage := &MessageDBModel{
		ID:               msgID,
		ChatID:           chatID,
		SenderID:         senderID,
		Content:          message.Content,
		SentAt:           message.Timestamp,
		Unsent:           false,
		DeletedForSender: false,
		DeletedForAll:    false,
		IsRead:           message.IsRead,
		IsEdited:         message.IsEdited,
		IsDeleted:        message.IsDeleted,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Use a transaction to ensure data integrity
	tx := r.db.Begin()
	if tx.Error != nil {
		log.Printf("Error starting transaction: %v", tx.Error)
		return fmt.Errorf("failed to start transaction: %v", tx.Error)
	}

	err = tx.Create(dbMessage).Error
	if err != nil {
		tx.Rollback()
		log.Printf("Error saving message to database: %v", err)
		return fmt.Errorf("failed to save message: %v", err)
	}

	// Update chat's updated_at timestamp
	err = tx.Table("chats").Where("chat_id = ?", chatID).
		Update("updated_at", time.Now()).Error
	if err != nil {
		tx.Rollback()
		log.Printf("Error updating chat timestamp: %v", err)
		return fmt.Errorf("failed to update chat timestamp: %v", err)
	}

	if err := tx.Commit().Error; err != nil {
		log.Printf("Error committing transaction: %v", err)
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	log.Printf("Message saved successfully to database with ID: %s", message.ID)
	return nil
}

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
		ID:               dbMessage.ID.String(),
		ChatID:           dbMessage.ChatID.String(),
		SenderID:         dbMessage.SenderID.String(),
		Content:          dbMessage.Content,
		Timestamp:        dbMessage.SentAt,
		IsRead:           dbMessage.IsRead,
		IsEdited:         dbMessage.IsEdited,
		IsDeleted:        dbMessage.IsDeleted,
		Unsent:           dbMessage.Unsent,
		DeletedForSender: dbMessage.DeletedForSender,
		DeletedForAll:    dbMessage.DeletedForAll,
	}, nil
}

func (r *GormMessageRepository) FindMessagesByChatID(chatID string, limit, offset int) ([]*model.MessageDTO, error) {
	log.Printf("Finding messages for chat ID: %s (limit: %d, offset: %d)", chatID, limit, offset)

	chatUUID, err := uuid.Parse(chatID)
	if err != nil {
		log.Printf("Error parsing chat ID: %v", err)
		return nil, err
	}

	var dbMessages []MessageDBModel
	query := r.db.Where("chat_id = ?", chatUUID)

	query = query.Order("sent_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	err = query.Find(&dbMessages).Error
	if err != nil {
		log.Printf("Database error retrieving messages: %v", err)
		return nil, err
	}

	log.Printf("Found %d messages for chat ID: %s", len(dbMessages), chatID)

	messages := make([]*model.MessageDTO, len(dbMessages))
	for i, dbMessage := range dbMessages {
		messages[i] = &model.MessageDTO{
			ID:               dbMessage.ID.String(),
			ChatID:           dbMessage.ChatID.String(),
			SenderID:         dbMessage.SenderID.String(),
			Content:          dbMessage.Content,
			MediaURL:         dbMessage.MediaURL,
			MediaType:        dbMessage.MediaType,
			Timestamp:        dbMessage.SentAt,
			Unsent:           dbMessage.Unsent,
			UnsentAt:         dbMessage.UnsentAt,
			DeletedForSender: dbMessage.DeletedForSender,
			DeletedForAll:    dbMessage.DeletedForAll,
			IsRead:           dbMessage.IsRead,
			IsEdited:         dbMessage.IsEdited,
			IsDeleted:        dbMessage.IsDeleted,
			CreatedAt:        dbMessage.CreatedAt,
			UpdatedAt:        dbMessage.UpdatedAt,
		}
		if dbMessage.ReplyToMessageID != nil {
			messages[i].ReplyToMessageID = dbMessage.ReplyToMessageID.String()
		}
	}

	return messages, nil
}

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

func (r *GormMessageRepository) DeleteMessage(messageID string) error {
	msgID, err := uuid.Parse(messageID)
	if err != nil {
		return err
	}

	return r.db.Delete(&MessageDBModel{}, "message_id = ?", msgID).Error
}

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

func (r *GormMessageRepository) UpdateMessage(message *model.MessageDTO) error {
	msgID, err := uuid.Parse(message.ID)
	if err != nil {
		return err
	}

	updateData := map[string]interface{}{
		"content":    message.Content,
		"is_edited":  message.IsEdited,
		"is_read":    message.IsRead,
		"is_deleted": message.IsDeleted,
	}

	return r.db.Model(&MessageDBModel{}).
		Where("message_id = ?", msgID).
		Updates(updateData).
		Error
}
