package model

import (
	"time"

	"github.com/google/uuid"
)

type Media struct {
	MediaID   uuid.UUID  `gorm:"type:uuid;primaryKey;column:media_id"`
	ThreadID  *uuid.UUID `gorm:"type:uuid;column:thread_id"`
	ReplyID   *uuid.UUID `gorm:"type:uuid;column:reply_id"`
	Type      string     `gorm:"type:varchar(10);not null"`
	URL       string     `gorm:"type:varchar(512);not null"`
	CreatedAt time.Time  `gorm:"autoCreateTime"`
	DeletedAt *time.Time `gorm:"index"`
}
