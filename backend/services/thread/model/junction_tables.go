package model

import (
	"time"

	"github.com/google/uuid"
)

type ThreadHashtag struct {
	ThreadID  uuid.UUID  `gorm:"type:uuid;primaryKey;column:thread_id"`
	HashtagID uuid.UUID  `gorm:"type:uuid;primaryKey;column:hashtag_id"`
	DeletedAt *time.Time `gorm:"index"`
}

// ThreadCategory moved to category.go
