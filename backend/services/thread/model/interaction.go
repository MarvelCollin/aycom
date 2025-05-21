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

	// Define unique index constraints using tags with clearer conditions
	// A user can like a specific thread once OR a specific reply once
	_ struct{} `gorm:"uniqueIndex:likes_user_thread_idx,where:thread_id IS NOT NULL AND reply_id IS NULL AND deleted_at IS NULL;uniqueIndex:likes_user_reply_idx,where:reply_id IS NOT NULL AND thread_id IS NULL AND deleted_at IS NULL"`
}

type Repost struct {
	UserID     uuid.UUID  `gorm:"type:uuid;not null;column:user_id;primaryKey"`
	ThreadID   uuid.UUID  `gorm:"type:uuid;not null;column:thread_id;primaryKey"`
	RepostText *string    `gorm:"type:text"`
	CreatedAt  time.Time  `gorm:"autoCreateTime"`
	DeletedAt  *time.Time `gorm:"index"`
}

type Bookmark struct {
	UserID    uuid.UUID  `gorm:"type:uuid;not null;column:user_id"`
	ThreadID  *uuid.UUID `gorm:"type:uuid;column:thread_id"`
	ReplyID   *uuid.UUID `gorm:"type:uuid;column:reply_id"`
	CreatedAt time.Time  `gorm:"autoCreateTime"`
	DeletedAt *time.Time `gorm:"index"`

	// Define constraints similar to the Like model for consistency
	_ struct{} `gorm:"check:((thread_id IS NOT NULL AND reply_id IS NULL) OR (thread_id IS NULL AND reply_id IS NOT NULL));uniqueIndex:bookmark_user_thread_idx,where:thread_id IS NOT NULL AND reply_id IS NULL AND deleted_at IS NULL;uniqueIndex:bookmark_user_reply_idx,where:reply_id IS NOT NULL AND thread_id IS NULL AND deleted_at IS NULL"`
}
