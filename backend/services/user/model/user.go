package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	UserID                 uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Username               string    `gorm:"type:varchar(50);not null;unique"`
	Name                   string    `gorm:"type:varchar(100);not null"`
	Email                  string    `gorm:"type:varchar(255);not null;unique"`
	ProfilePictureURL      string    `gorm:"type:varchar(512)"`
	BannerURL              string    `gorm:"type:varchar(512)"`
	Bio                    string    `gorm:"type:text"`
	Gender                 string    `gorm:"type:varchar(10)"`
	DateOfBirth            time.Time `gorm:"type:date;not null"`
	JoinedAt               time.Time `gorm:"type:timestamp with time zone;not null;default:now()"`
	IsBanned               bool      `gorm:"type:boolean;not null;default:false"`
	IsDeactivated          bool      `gorm:"type:boolean;not null;default:false"`
	IsPrivate              bool      `gorm:"type:boolean;not null;default:false"`
	IsPremium              bool      `gorm:"type:boolean;not null;default:false"`
	NewsletterSubscription bool      `gorm:"type:boolean;not null;default:false"`
	CreatedAt              time.Time `gorm:"type:timestamp with time zone;not null;default:now()"`
	UpdatedAt              time.Time `gorm:"type:timestamp with time zone;not null;default:now()"`
}

// BeforeCreate will set default values before inserting a new record
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.UserID == uuid.Nil {
		u.UserID = uuid.New()
	}
	return nil
}
