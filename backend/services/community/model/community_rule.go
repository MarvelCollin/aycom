package model

import (
	"time"

	"github.com/google/uuid"
)

type CommunityRule struct {
	RuleID      uuid.UUID  `gorm:"type:uuid;primaryKey;column:rule_id"`
	CommunityID uuid.UUID  `gorm:"type:uuid;not null"`
	RuleText    string     `gorm:"type:text;not null"`
	CreatedAt   time.Time  `gorm:"autoCreateTime"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime"`
	DeletedAt   *time.Time `gorm:"index"`
}
