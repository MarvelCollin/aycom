package repository

import (
	"aycom/backend/services/community/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ChatRepository interface {
	Create(chat *model.Chat) error
	FindByID(id uuid.UUID) (*model.Chat, error)
	FindByUser(userID uuid.UUID) ([]*model.Chat, error)
	Update(chat *model.Chat) error
	Delete(id uuid.UUID) error
}

type GormChatRepository struct {
	db *gorm.DB
}

func NewChatRepository(db *gorm.DB) ChatRepository {
	return &GormChatRepository{db: db}
}

func (r *GormChatRepository) Create(chat *model.Chat) error {
	return r.db.Create(chat).Error
}

func (r *GormChatRepository) FindByID(id uuid.UUID) (*model.Chat, error) {
	var chat model.Chat
	err := r.db.First(&chat, "chat_id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &chat, nil
}

func (r *GormChatRepository) FindByUser(userID uuid.UUID) ([]*model.Chat, error) {
	var chats []*model.Chat
	err := r.db.Joins("JOIN chat_participants ON chats.chat_id = chat_participants.chat_id").Where("chat_participants.user_id = ?", userID).Find(&chats).Error
	return chats, err
}

func (r *GormChatRepository) Update(chat *model.Chat) error {
	return r.db.Save(chat).Error
}

func (r *GormChatRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&model.Chat{}, "chat_id = ?", id).Error
}
