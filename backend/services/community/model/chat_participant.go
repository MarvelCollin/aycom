package model

import (
	"time"

	"github.com/google/uuid"
)

type ChatParticipant struct {
	ChatID    uuid.UUID  `gorm:"type:uuid;primaryKey;column:chat_id"`
	UserID    uuid.UUID  `gorm:"type:uuid;primaryKey;column:user_id"`
	JoinedAt  time.Time  `gorm:"autoCreateTime"`
	IsAdmin   bool       `gorm:"default:false;not null"`
	DeletedAt *time.Time `gorm:"index"`
}
