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
	Categories  []Category `gorm:"many2many:community_categories;"`
	CreatedAt   time.Time  `gorm:"autoCreateTime"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime"`
	DeletedAt   *time.Time `gorm:"index"`
}

type Category struct {
	CategoryID  uuid.UUID   `gorm:"type:uuid;primaryKey;column:category_id;default:gen_random_uuid()"`
	Name        string      `gorm:"type:varchar(50);unique;not null"`
	Communities []Community `gorm:"many2many:community_categories;"`
	CreatedAt   time.Time   `gorm:"autoCreateTime"`
	UpdatedAt   time.Time   `gorm:"autoUpdateTime"`
	DeletedAt   *time.Time  `gorm:"index"`
}

type CommunityCategory struct {
	CommunityID uuid.UUID `gorm:"primaryKey;column:community_id"`
	CategoryID  uuid.UUID `gorm:"primaryKey;column:category_id"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
}
