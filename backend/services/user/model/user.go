package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID                    uuid.UUID  `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	Username              string     `gorm:"unique;not null"`
	Email                 string     `gorm:"unique;not null"`
	Name                  string     `gorm:"not null"`
	Gender                string     `gorm:"type:varchar(10)"`
	DateOfBirth           *time.Time `gorm:"type:date"`
	ProfilePictureURL     string     `gorm:"type:text"`
	BannerURL             string     `gorm:"type:text"`
	SecurityQuestion      string     `gorm:"type:text"`
	SecurityAnswer        string     `gorm:"type:text"`
	SubscribeToNewsletter bool       `gorm:"default:false"`
	IsVerified            bool       `gorm:"default:false"`
	CreatedAt             time.Time
	UpdatedAt             time.Time
}
