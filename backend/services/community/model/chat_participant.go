package model

import (
	"time"

	"github.com/google/uuid"
)

type ChatParticipant struct {
	ID        uuid.UUID `gorm:"primaryKey;column:id"`
	ChatID    uuid.UUID `gorm:"column:chat_id"`
	UserID    uuid.UUID `gorm:"column:user_id"`
	JoinedAt  time.Time `gorm:"column:joined_at"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (p *ChatParticipant) TableName() string {
	return "chat_participants"
}
