package repository

import (
	"aycom/backend/services/community/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DeletedChatRepository interface {
	Add(deleted *model.DeletedChat) error
	Remove(chatID, userID uuid.UUID) error
	FindByUser(userID uuid.UUID) ([]*model.DeletedChat, error)
	FindByChat(chatID uuid.UUID) ([]*model.DeletedChat, error)
}

type GormDeletedChatRepository struct {
	db *gorm.DB
}

func NewDeletedChatRepository(db *gorm.DB) DeletedChatRepository {
	return &GormDeletedChatRepository{db: db}
}

func (r *GormDeletedChatRepository) Add(deleted *model.DeletedChat) error {
	return r.db.Create(deleted).Error
}

func (r *GormDeletedChatRepository) Remove(chatID, userID uuid.UUID) error {
	return r.db.Delete(&model.DeletedChat{}, "chat_id = ? AND user_id = ?", chatID, userID).Error
}

func (r *GormDeletedChatRepository) FindByUser(userID uuid.UUID) ([]*model.DeletedChat, error) {
	var deleted []*model.DeletedChat
	err := r.db.Where("user_id = ?", userID).Find(&deleted).Error
	return deleted, err
}

func (r *GormDeletedChatRepository) FindByChat(chatID uuid.UUID) ([]*model.DeletedChat, error) {
	var deleted []*model.DeletedChat
	err := r.db.Where("chat_id = ?", chatID).Find(&deleted).Error
	return deleted, err
} 