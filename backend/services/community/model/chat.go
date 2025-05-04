package model

import (
	"time"

	"github.com/google/uuid"
)

type Chat struct {
	ChatID    uuid.UUID  `gorm:"type:uuid;primaryKey;column:chat_id"`
	IsGroup   bool       `gorm:"default:false;not null"`
	Name      string     `gorm:"type:varchar(100)"`
	CreatedBy uuid.UUID  `gorm:"type:uuid;not null"`
	CreatedAt time.Time  `gorm:"autoCreateTime"`
	UpdatedAt time.Time  `gorm:"autoUpdateTime"`
	DeletedAt *time.Time `gorm:"index"`
}
