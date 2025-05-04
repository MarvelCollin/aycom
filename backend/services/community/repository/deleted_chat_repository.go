package repository

import (
	"aycom/backend/services/community/model"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DeletedChatDBModel represents a deleted chat in the database
type DeletedChatDBModel struct {
	ChatID    uuid.UUID `gorm:"type:uuid;primaryKey;column:chat_id"`
	UserID    uuid.UUID `gorm:"type:uuid;primaryKey;column:user_id"`
	DeletedAt time.Time `gorm:"autoCreateTime"`
}

// TableName sets the table name for the deleted chat model
func (DeletedChatDBModel) TableName() string {
	return "deleted_chats"
}

// GormDeletedChatRepository implements model.DeletedChatRepository
type GormDeletedChatRepository struct {
	db *gorm.DB
}

// NewDeletedChatRepository creates a new deleted chat repository
func NewDeletedChatRepository(db *gorm.DB) model.DeletedChatRepository {
	return &GormDeletedChatRepository{db: db}
}

// MarkChatAsDeleted marks a chat as deleted for a user
func (r *GormDeletedChatRepository) MarkChatAsDeleted(chatID, userID string) error {
	chatUUID, err := uuid.Parse(chatID)
	if err != nil {
		return err
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return err
	}

	deletedChat := &DeletedChatDBModel{
		ChatID: chatUUID,
		UserID: userUUID,
	}

	return r.db.Create(deletedChat).Error
}

// IsDeletedForUser checks if a chat is deleted for a user
func (r *GormDeletedChatRepository) IsDeletedForUser(chatID, userID string) (bool, error) {
	chatUUID, err := uuid.Parse(chatID)
	if err != nil {
		return false, err
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return false, err
	}

	var count int64
	err = r.db.Model(&DeletedChatDBModel{}).
		Where("chat_id = ? AND user_id = ?", chatUUID, userUUID).
		Count(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, nil
}
