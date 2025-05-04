package repository

import (
	"aycom/backend/services/community/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ChatParticipantRepository interface {
	Add(participant *model.ChatParticipant) error
	Remove(chatID, userID uuid.UUID) error
	FindByChat(chatID uuid.UUID) ([]*model.ChatParticipant, error)
	FindByUser(userID uuid.UUID) ([]*model.ChatParticipant, error)
	Update(participant *model.ChatParticipant) error
}

type GormChatParticipantRepository struct {
	db *gorm.DB
}

func NewChatParticipantRepository(db *gorm.DB) ChatParticipantRepository {
	return &GormChatParticipantRepository{db: db}
}

func (r *GormChatParticipantRepository) Add(participant *model.ChatParticipant) error {
	return r.db.Create(participant).Error
}

func (r *GormChatParticipantRepository) Remove(chatID, userID uuid.UUID) error {
	return r.db.Delete(&model.ChatParticipant{}, "chat_id = ? AND user_id = ?", chatID, userID).Error
}

func (r *GormChatParticipantRepository) FindByChat(chatID uuid.UUID) ([]*model.ChatParticipant, error) {
	var participants []*model.ChatParticipant
	err := r.db.Where("chat_id = ?", chatID).Find(&participants).Error
	return participants, err
}

func (r *GormChatParticipantRepository) FindByUser(userID uuid.UUID) ([]*model.ChatParticipant, error) {
	var participants []*model.ChatParticipant
	err := r.db.Where("user_id = ?", userID).Find(&participants).Error
	return participants, err
}

func (r *GormChatParticipantRepository) Update(participant *model.ChatParticipant) error {
	return r.db.Save(participant).Error
}
