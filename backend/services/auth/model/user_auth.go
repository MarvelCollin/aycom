package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserAuth contains authentication-specific user data
type UserAuth struct {
	AuthID       uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID       uuid.UUID  `gorm:"type:uuid;not null;unique"` // Maps to User.UserID in user service
	Email        string     `gorm:"type:varchar(255);not null;unique"`
	Username     string     `gorm:"type:varchar(50);not null;unique"` // Store username for login purposes
	PasswordHash string     `gorm:"type:varchar(255);not null"`
	PasswordSalt string     `gorm:"type:varchar(64);not null"`
	IsActivated  bool       `gorm:"type:boolean;not null;default:false"`
	IsAdmin      bool       `gorm:"type:boolean;not null;default:false"`
	LastLoginAt  *time.Time `gorm:"type:timestamp with time zone"`
	CreatedAt    time.Time  `gorm:"type:timestamp with time zone;not null;default:now()"`
	UpdatedAt    time.Time  `gorm:"type:timestamp with time zone;not null;default:now()"`
}

// BeforeCreate will set default values before inserting a new record
func (ua *UserAuth) BeforeCreate(tx *gorm.DB) error {
	if ua.AuthID == uuid.Nil {
		ua.AuthID = uuid.New()
	}
	return nil
}
