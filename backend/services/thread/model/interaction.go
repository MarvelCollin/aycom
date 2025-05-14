package model

import (
	"time"

	"github.com/google/uuid"
)

type Like struct {
	UserID    uuid.UUID  `gorm:"type:uuid;not null;column:user_id"`
	ThreadID  *uuid.UUID `gorm:"type:uuid;column:thread_id"`
	ReplyID   *uuid.UUID `gorm:"type:uuid;column:reply_id"`
	CreatedAt time.Time  `gorm:"autoCreateTime"`
	DeletedAt *time.Time `gorm:"index"`

	// Define unique index constraints using tags, matching what we set up in the database:
	// 1. A user can like a thread once (when ThreadID is not null and ReplyID is null)
	// 2. A user can like a reply once (when ReplyID is not null and ThreadID is null)
	_ struct{} `gorm:"uniqueIndex:likes_user_thread_idx,where:thread_id IS NOT NULL AND reply_id IS NULL;uniqueIndex:likes_user_reply_idx,where:reply_id IS NOT NULL AND thread_id IS NULL"`
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
	ReplyID   *uuid.UUID `gorm:"type:uuid;column:reply_id"`
	CreatedAt time.Time  `gorm:"autoCreateTime"`
	DeletedAt *time.Time `gorm:"index"`
}
