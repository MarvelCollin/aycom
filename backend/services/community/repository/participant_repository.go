package repository

import (
	"aycom/backend/services/community/model"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ParticipantModel represents a chat participant in the database
type ParticipantModel struct {
	ChatID   uuid.UUID `gorm:"type:uuid;primaryKey;column:chat_id"`
	UserID   uuid.UUID `gorm:"type:uuid;primaryKey;column:user_id"`
	IsAdmin  bool      `gorm:"default:false;not null;column:is_admin"`
	JoinedAt time.Time `gorm:"default:now();not null;column:joined_at"`
}

// TableName sets the table name for the chat participant model
func (ParticipantModel) TableName() string {
	return "chat_participants"
}

// GormParticipantRepository implements model.ParticipantRepository
type GormParticipantRepository struct {
	db *gorm.DB
}

// NewParticipantRepository creates a new participant repository
func NewParticipantRepository(db *gorm.DB) model.ParticipantRepository {
	return &GormParticipantRepository{db: db}
}

// AddParticipant adds a participant to a chat
func (r *GormParticipantRepository) AddParticipant(participant *model.ParticipantDTO) error {
	chatUUID, err := uuid.Parse(participant.ChatID)
	if err != nil {
		return err
	}

	userUUID, err := uuid.Parse(participant.UserID)
	if err != nil {
		return err
	}

	dbParticipant := &ParticipantModel{
		ChatID:   chatUUID,
		UserID:   userUUID,
		IsAdmin:  participant.IsAdmin,
		JoinedAt: participant.JoinedAt,
	}
	return r.db.Create(dbParticipant).Error
}

// RemoveParticipant removes a participant from a chat
func (r *GormParticipantRepository) RemoveParticipant(chatID, userID string) error {
	chatUUID, err := uuid.Parse(chatID)
	if err != nil {
		return err
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return err
	}

	return r.db.Delete(&ParticipantModel{}, "chat_id = ? AND user_id = ?", chatUUID, userUUID).Error
}

// ListParticipantsByChatID lists all participants in a chat
func (r *GormParticipantRepository) ListParticipantsByChatID(chatID string, limit, offset int) ([]*model.ParticipantDTO, error) {
	chatUUID, err := uuid.Parse(chatID)
	if err != nil {
		return nil, err
	}

	var dbParticipants []ParticipantModel
	err = r.db.Where("chat_id = ?", chatUUID).
		Limit(limit).
		Offset(offset).
		Find(&dbParticipants).Error
	if err != nil {
		return nil, err
	}

	participants := make([]*model.ParticipantDTO, len(dbParticipants))
	for i, dbParticipant := range dbParticipants {
		participants[i] = &model.ParticipantDTO{
			ChatID:   dbParticipant.ChatID.String(),
			UserID:   dbParticipant.UserID.String(),
			JoinedAt: dbParticipant.JoinedAt,
			IsAdmin:  dbParticipant.IsAdmin,
		}
	}

	return participants, nil
}

// IsUserInChat checks if a user is in a chat
func (r *GormParticipantRepository) IsUserInChat(chatID, userID string) (bool, error) {
	chatUUID, err := uuid.Parse(chatID)
	if err != nil {
		return false, err
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return false, err
	}

	var count int64
	err = r.db.Model(&ParticipantModel{}).
		Where("chat_id = ? AND user_id = ?", chatUUID, userUUID).
		Count(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, nil
}
