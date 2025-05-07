package model

import (
	"time"

	"github.com/google/uuid"
)

type DeletedChat struct {
	ChatID    uuid.UUID `gorm:"type:uuid;primaryKey;column:chat_id"`
	UserID    uuid.UUID `gorm:"type:uuid;primaryKey;column:user_id"`
	DeletedAt time.Time `gorm:"column:deleted_at;default:now();not null"`
}

func (d *DeletedChat) TableName() string {
	return "deleted_chats"
}
