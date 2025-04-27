package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SecurityQuestion defines available security questions for account recovery
type SecurityQuestion struct {
	QuestionID uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Question   string    `gorm:"type:varchar(255);not null"`
	CreatedAt  time.Time `gorm:"type:timestamp with time zone;not null;default:now()"`
	UpdatedAt  time.Time `gorm:"type:timestamp with time zone;not null;default:now()"`
}

// BeforeCreate will set default values before inserting a new record
func (sq *SecurityQuestion) BeforeCreate(tx *gorm.DB) error {
	if sq.QuestionID == uuid.Nil {
		sq.QuestionID = uuid.New()
	}
	return nil
}
