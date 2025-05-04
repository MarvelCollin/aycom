package model

import (
	"time"

	"github.com/google/uuid"
)

type Hashtag struct {
	HashtagID uuid.UUID  `gorm:"type:uuid;primaryKey;column:hashtag_id"`
	Text      string     `gorm:"type:varchar(50);unique;not null"`
	CreatedAt time.Time  `gorm:"autoCreateTime"`
	DeletedAt *time.Time `gorm:"index"`
}
