package model

import (
	"time"

	"github.com/google/uuid"
)

type Thread struct {
	ThreadID        uuid.UUID  `gorm:"type:uuid;primaryKey;column:thread_id"`
	UserID          uuid.UUID  `gorm:"type:uuid;not null;column:user_id"`
	Content         string     `gorm:"type:text;not null"`
	IsPinned        bool       `gorm:"default:false"`
	WhoCanReply     string     `gorm:"type:varchar(20);not null"`
	ScheduledAt     *time.Time `gorm:"type:timestamp with time zone"`
	CommunityID     *uuid.UUID `gorm:"type:uuid"`
	IsAdvertisement bool       `gorm:"default:false"`
	CreatedAt       time.Time  `gorm:"autoCreateTime"`
	UpdatedAt       time.Time  `gorm:"autoUpdateTime"`
	DeletedAt       *time.Time `gorm:"index"`
}
