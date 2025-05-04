package model

import (
	"time"

	"github.com/google/uuid"
)

type DeletedChat struct {
	ChatID    uuid.UUID `gorm:"column:chat_id;primaryKey"`
	UserID    uuid.UUID `gorm:"column:user_id;primaryKey"`
	DeletedAt time.Time `gorm:"column:deleted_at"`
}

func (d *DeletedChat) TableName() string {
	return "deleted_chats"
}
