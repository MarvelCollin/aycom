package model

import (
	"time"

	"github.com/google/uuid"
)

type UserMention struct {
	MentionID       uuid.UUID  `gorm:"type:uuid;primaryKey;column:mention_id"`
	MentionedUserID uuid.UUID  `gorm:"type:uuid;not null;column:mentioned_user_id"`
	ThreadID        *uuid.UUID `gorm:"type:uuid;column:thread_id"`
	ReplyID         *uuid.UUID `gorm:"type:uuid;column:reply_id"`
	CreatedAt       time.Time  `gorm:"autoCreateTime"`
	DeletedAt       *time.Time `gorm:"index"`
}
