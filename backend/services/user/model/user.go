package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID                    uuid.UUID  `gorm:"type:uuid;primaryKey"`
	Username              string     `gorm:"unique;not null"`
	Email                 string     `gorm:"unique;not null"`
	Name                  string     `gorm:"not null"`
	PasswordHash          string     `gorm:"not null"`
	Gender                string     `gorm:"type:varchar(10)"`
	DateOfBirth           *time.Time `gorm:"type:date"`
	ProfilePictureURL     string     `gorm:"type:text"`
	BannerURL             string     `gorm:"type:text"`
	Bio                   string     `gorm:"type:text"`
	Location              string     `gorm:"type:text"`
	Website               string     `gorm:"type:text"`
	SecurityQuestion      string     `gorm:"type:text"`
	SecurityAnswer        string     `gorm:"type:text"`
	SubscribeToNewsletter bool       `gorm:"default:false"`
	IsVerified            bool       `gorm:"default:false"`
	FollowerCount         int        `gorm:"-"` // Virtual field, not stored in DB
	FollowingCount        int        `gorm:"-"` // Virtual field, not stored in DB
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

func (u *User) BeforeCreate() error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}
