package model

import (
	"time"

	"github.com/google/uuid"
)

type DeletedChat struct {
	ChatID    uuid.UUID `gorm:"type:uuid;primaryKey;column:chat_id"`
	UserID    uuid.UUID `gorm:"type:uuid;primaryKey;column:user_id"`
	DeletedAt time.Time `gorm:"autoCreateTime"`
}
