package model

import (
	"github.com/google/uuid"
)

type Contact struct {
	ID     uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	UserID uuid.UUID `gorm:"type:uuid;not null;index"`
	Type   string    `gorm:"type:varchar(20);not null"`
	Value  string    `gorm:"type:text;not null"`
}
