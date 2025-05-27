package repository

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"aycom/backend/services/community/model"
)

type DeletedChatDBModel struct {
	ChatID    uuid.UUID `gorm:"type:uuid;primaryKey;column:chat_id"`
	UserID    uuid.UUID `gorm:"type:uuid;primaryKey;column:user_id"`
	DeletedAt time.Time `gorm:"autoCreateTime"`
}

func (DeletedChatDBModel) TableName() string {
	return "deleted_chats"
}

type GormDeletedChatRepository struct {
	db *gorm.DB
}

func NewDeletedChatRepository(db *gorm.DB) model.DeletedChatRepository {
	return &GormDeletedChatRepository{db: db}
}

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
