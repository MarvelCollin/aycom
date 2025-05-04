package model

import (
	"time"

	"github.com/google/uuid"
)

type Reply struct {
	ReplyID       uuid.UUID  `gorm:"type:uuid;primaryKey;column:reply_id"`
	ThreadID      uuid.UUID  `gorm:"type:uuid;not null;column:thread_id"`
	UserID        uuid.UUID  `gorm:"type:uuid;not null;column:user_id"`
	Content       string     `gorm:"type:text;not null"`
	IsPinned      bool       `gorm:"default:false"`
	ParentReplyID *uuid.UUID `gorm:"type:uuid"`
	CreatedAt     time.Time  `gorm:"autoCreateTime"`
	UpdatedAt     time.Time  `gorm:"autoUpdateTime"`
	DeletedAt     *time.Time `gorm:"index"`
}
