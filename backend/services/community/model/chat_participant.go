package model

import (
	"time"

	"github.com/google/uuid"
)

type ChatParticipant struct {
	ChatID   uuid.UUID `gorm:"type:uuid;primaryKey;column:chat_id"`
	UserID   uuid.UUID `gorm:"type:uuid;primaryKey;column:user_id"`
	JoinedAt time.Time `gorm:"column:joined_at;default:now();not null"`
	IsAdmin  bool      `gorm:"column:is_admin;default:false;not null"`
}

func (p *ChatParticipant) TableName() string {
	return "chat_participants"
}
