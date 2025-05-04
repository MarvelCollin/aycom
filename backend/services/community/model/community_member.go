package model

import (
	"time"

	"github.com/google/uuid"
)

type CommunityMember struct {
	CommunityID uuid.UUID  `gorm:"type:uuid;primaryKey;column:community_id"`
	UserID      uuid.UUID  `gorm:"type:uuid;primaryKey;column:user_id"`
	Role        string     `gorm:"type:varchar(10);not null"`
	CreatedAt   time.Time  `gorm:"autoCreateTime"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime"`
	DeletedAt   *time.Time `gorm:"index"`
}
