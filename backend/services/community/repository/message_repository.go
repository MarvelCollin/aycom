package repository

import (
	"aycom/backend/services/community/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MessageRepository interface {
	Create(message *model.Message) error
	FindByID(id uuid.UUID) (*model.Message, error)
	FindByChat(chatID uuid.UUID, offset, limit int) ([]*model.Message, error)
	FindByUser(userID uuid.UUID, offset, limit int) ([]*model.Message, error)
	SearchInChat(chatID uuid.UUID, query string, offset, limit int) ([]*model.Message, error)
	Update(message *model.Message) error
	Delete(id uuid.UUID) error
}

type GormMessageRepository struct {
	db *gorm.DB
}

func NewMessageRepository(db *gorm.DB) MessageRepository {
	return &GormMessageRepository{db: db}
}

func (r *GormMessageRepository) Create(message *model.Message) error {
	return r.db.Create(message).Error
}

func (r *GormMessageRepository) FindByID(id uuid.UUID) (*model.Message, error) {
	var message model.Message
	err := r.db.First(&message, "message_id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &message, nil
}

func (r *GormMessageRepository) FindByChat(chatID uuid.UUID, offset, limit int) ([]*model.Message, error) {
	var messages []*model.Message
	err := r.db.Where("chat_id = ?", chatID).Order("sent_at ASC").Offset(offset).Limit(limit).Find(&messages).Error
	return messages, err
}

func (r *GormMessageRepository) FindByUser(userID uuid.UUID, offset, limit int) ([]*model.Message, error) {
	var messages []*model.Message
	err := r.db.Where("sender_id = ?", userID).Order("sent_at DESC").Offset(offset).Limit(limit).Find(&messages).Error
	return messages, err
}

func (r *GormMessageRepository) SearchInChat(chatID uuid.UUID, query string, offset, limit int) ([]*model.Message, error) {
	var messages []*model.Message
	err := r.db.Where("chat_id = ? AND content ILIKE ?", chatID, "%"+query+"%").Order("sent_at ASC").Offset(offset).Limit(limit).Find(&messages).Error
	return messages, err
}

func (r *GormMessageRepository) Update(message *model.Message) error {
	return r.db.Save(message).Error
}

func (r *GormMessageRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&model.Message{}, "message_id = ?", id).Error
}
