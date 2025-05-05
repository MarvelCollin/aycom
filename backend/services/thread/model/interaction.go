package model

import (
	"time"

	"github.com/google/uuid"
)

type Like struct {
	UserID    uuid.UUID  `gorm:"type:uuid;not null;column:user_id;primaryKey"`
	ThreadID  *uuid.UUID `gorm:"type:uuid;column:thread_id"`
	ReplyID   *uuid.UUID `gorm:"type:uuid;column:reply_id"`
	CreatedAt time.Time  `gorm:"autoCreateTime"`
	DeletedAt *time.Time `gorm:"index"`
}

type Repost struct {
	UserID     uuid.UUID  `gorm:"type:uuid;not null;column:user_id;primaryKey"`
	ThreadID   uuid.UUID  `gorm:"type:uuid;not null;column:thread_id;primaryKey"`
	RepostText *string    `gorm:"type:text"`
	CreatedAt  time.Time  `gorm:"autoCreateTime"`
	DeletedAt  *time.Time `gorm:"index"`
}

type Bookmark struct {
	UserID    uuid.UUID  `gorm:"type:uuid;not null;column:user_id;primaryKey"`
	ThreadID  uuid.UUID  `gorm:"type:uuid;not null;column:thread_id;primaryKey"`
	CreatedAt time.Time  `gorm:"autoCreateTime"`
	DeletedAt *time.Time `gorm:"index"`
}
