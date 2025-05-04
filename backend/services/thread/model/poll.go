package model

import (
	"time"

	"github.com/google/uuid"
)

type Poll struct {
	PollID     uuid.UUID  `gorm:"type:uuid;primaryKey;column:poll_id"`
	ThreadID   uuid.UUID  `gorm:"type:uuid;not null;unique;column:thread_id"`
	Question   string     `gorm:"type:text;not null"`
	ClosesAt   time.Time  `gorm:"type:timestamp with time zone;not null"`
	WhoCanVote string     `gorm:"type:varchar(20);not null"`
	CreatedAt  time.Time  `gorm:"autoCreateTime"`
	UpdatedAt  time.Time  `gorm:"autoUpdateTime"`
	DeletedAt  *time.Time `gorm:"index"`
}

type PollOption struct {
	OptionID  uuid.UUID  `gorm:"type:uuid;primaryKey;column:option_id"`
	PollID    uuid.UUID  `gorm:"type:uuid;not null;column:poll_id"`
	Text      string     `gorm:"type:varchar(100);not null"`
	CreatedAt time.Time  `gorm:"autoCreateTime"`
	DeletedAt *time.Time `gorm:"index"`
}

type PollVote struct {
	VoteID    uuid.UUID  `gorm:"type:uuid;primaryKey;column:vote_id"`
	PollID    uuid.UUID  `gorm:"type:uuid;not null;column:poll_id"`
	OptionID  uuid.UUID  `gorm:"type:uuid;not null;column:option_id"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null;column:user_id"`
	CreatedAt time.Time  `gorm:"autoCreateTime"`
	DeletedAt *time.Time `gorm:"index"`
}
