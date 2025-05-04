package model

import (
	"time"

	"github.com/google/uuid"
)

type CommunityJoinRequest struct {
	RequestID   uuid.UUID  `gorm:"type:uuid;primaryKey;column:request_id"`
	CommunityID uuid.UUID  `gorm:"type:uuid;not null"`
	UserID      uuid.UUID  `gorm:"type:uuid;not null"`
	Status      string     `gorm:"type:varchar(10);not null"`
	CreatedAt   time.Time  `gorm:"autoCreateTime"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime"`
	DeletedAt   *time.Time `gorm:"index"`
}
