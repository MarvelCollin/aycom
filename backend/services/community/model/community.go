package model

import (
	"time"

	"github.com/google/uuid"
)

type Community struct {
	CommunityID uuid.UUID  `gorm:"type:uuid;primaryKey;column:community_id"`
	Name        string     `gorm:"type:varchar(100);unique;not null"`
	Description string     `gorm:"type:text;not null"`
	LogoURL     string     `gorm:"type:varchar(512);not null"`
	BannerURL   string     `gorm:"type:varchar(512);not null"`
	CreatorID   uuid.UUID  `gorm:"type:uuid;not null"`
	IsApproved  bool       `gorm:"default:false;not null"`
	CreatedAt   time.Time  `gorm:"autoCreateTime"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime"`
	DeletedAt   *time.Time `gorm:"index"`
}
