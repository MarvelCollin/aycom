package model

import (
	"time"

	"github.com/google/uuid"
)

type Follow struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey"`
	FollowerID uuid.UUID `gorm:"type:uuid;not null;index:idx_follower_followed"`
	FollowedID uuid.UUID `gorm:"type:uuid;not null;index:idx_follower_followed"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (f *Follow) BeforeCreate() error {
	if f.ID == uuid.Nil {
		f.ID = uuid.New()
	}
	return nil
}
