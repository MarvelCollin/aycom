package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Token manages user authentication tokens
type Token struct {
	TokenID      uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID       uuid.UUID `gorm:"type:uuid;not null;index"` // Maps to UserAuth.UserID
	RefreshToken string    `gorm:"type:varchar(255);not null;unique"`
	UserAgent    string    `gorm:"type:varchar(512)"` // Browser/client info
	IPAddress    string    `gorm:"type:varchar(50)"`  // IP address for security
	ExpiresAt    time.Time `gorm:"type:timestamp with time zone;not null"`
	IsRevoked    bool      `gorm:"type:boolean;not null;default:false"`
	CreatedAt    time.Time `gorm:"type:timestamp with time zone;not null;default:now()"`
	UpdatedAt    time.Time `gorm:"type:timestamp with time zone;not null;default:now()"`
}

// BeforeCreate will set default values before inserting a new record
func (t *Token) BeforeCreate(tx *gorm.DB) error {
	if t.TokenID == uuid.Nil {
		t.TokenID = uuid.New()
	}
	return nil
}
