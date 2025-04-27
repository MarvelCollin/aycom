package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// OAuthConnection stores third-party authentication provider details
type OAuthConnection struct {
	ConnectionID  uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID        uuid.UUID `gorm:"type:uuid;not null;index"`  // Maps to UserAuth.UserID
	Provider      string    `gorm:"type:varchar(50);not null"` // "google", "github", etc.
	ProviderID    string    `gorm:"type:varchar(255);not null"`
	ProviderEmail string    `gorm:"type:varchar(255)"`
	AccessToken   string    `gorm:"type:text"`
	RefreshToken  string    `gorm:"type:text"`
	ExpiresAt     time.Time `gorm:"type:timestamp with time zone"`
	CreatedAt     time.Time `gorm:"type:timestamp with time zone;not null;default:now()"`
	UpdatedAt     time.Time `gorm:"type:timestamp with time zone;not null;default:now()"`
}

// BeforeCreate will set default values before inserting a new record
func (oc *OAuthConnection) BeforeCreate(tx *gorm.DB) error {
	if oc.ConnectionID == uuid.Nil {
		oc.ConnectionID = uuid.New()
	}
	return nil
}
