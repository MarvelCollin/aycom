package model

import (
	"time"

	"github.com/google/uuid"
)

type UserProfile struct {
	ID                uuid.UUID  `gorm:"type:uuid;primaryKey"`
	UserID            uuid.UUID  `gorm:"type:uuid;not null;index"`
	Bio               string     `gorm:"type:text"`
	Location          string     `gorm:"type:text"`
	Website           string     `gorm:"type:text"`
	Birthdate         *time.Time `gorm:"type:date"`
	ProfilePictureURL string     `gorm:"type:text"`
	BannerURL         string     `gorm:"type:text"`
	CreatedAt         time.Time
	UpdatedAt         time.Time
}
