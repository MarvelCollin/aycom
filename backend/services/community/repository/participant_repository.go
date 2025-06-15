package repository

import (
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"aycom/backend/services/community/model"
)

type ParticipantModel struct {
	ChatID   uuid.UUID `gorm:"type:uuid;primaryKey;column:chat_id"`
	UserID   uuid.UUID `gorm:"type:uuid;primaryKey;column:user_id"`
	IsAdmin  bool      `gorm:"default:false;not null;column:is_admin"`
	JoinedAt time.Time `gorm:"default:now();not null;column:joined_at"`
}

func (ParticipantModel) TableName() string {
	return "chat_participants"
}

type GormParticipantRepository struct {
	db *gorm.DB
}

func NewParticipantRepository(db *gorm.DB) model.ParticipantRepository {
	return &GormParticipantRepository{db: db}
}

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

func (r *GormParticipantRepository) IsUserInChat(chatID, userID string) (bool, error) {
	// Validate input parameters
	if chatID == "" {
		return false, fmt.Errorf("chat ID cannot be empty")
	}
	if userID == "" {
		return false, fmt.Errorf("user ID cannot be empty")
	}

	// Log the received IDs for debugging
	log.Printf("IsUserInChat check with chatID=%s, userID=%s", chatID, userID)

	// Validate UUID format for chat ID
	chatUUID, err := uuid.Parse(chatID)
	if err != nil {
		log.Printf("Error parsing chat ID: %v", err)
		return false, fmt.Errorf("invalid chat ID format: %v", err)
	}

	// Validate UUID format for user ID
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		log.Printf("Error parsing user ID: %v", err)
		return false, fmt.Errorf("invalid user ID format: %v", err)
	}

	// Query the database
	var count int64
	err = r.db.Model(&ParticipantModel{}).
		Where("chat_id = ? AND user_id = ?", chatUUID, userUUID).
		Count(&count).Error
	if err != nil {
		log.Printf("Database error in IsUserInChat: %v", err)
		return false, fmt.Errorf("database error: %v", err)
	}

	log.Printf("IsUserInChat result: user %s in chat %s = %v", userID, chatID, count > 0)
	return count > 0, nil
}
