package model

import (
	"time"

	"github.com/google/uuid"
)

type Category struct {
	CategoryID uuid.UUID  `gorm:"type:uuid;primaryKey;column:category_id"`
	Name       string     `gorm:"type:varchar(50);not null"`
	Type       string     `gorm:"type:varchar(10);not null"`
	CreatedAt  time.Time  `gorm:"autoCreateTime"`
	UpdatedAt  time.Time  `gorm:"autoUpdateTime"`
	DeletedAt  *time.Time `gorm:"index"`
}

type ThreadCategory struct {
	ThreadID   uuid.UUID  `gorm:"type:uuid;primaryKey;column:thread_id"`
	CategoryID uuid.UUID  `gorm:"type:uuid;primaryKey;column:category_id"`
	DeletedAt  *time.Time `gorm:"index"`
}
