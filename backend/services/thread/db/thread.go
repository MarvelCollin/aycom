package db

import (
	"time"

	"github.com/google/uuid"
)

type Thread struct {
	ThreadID        uuid.UUID  `gorm:"type:uuid;primaryKey;column:thread_id"`
	UserID          uuid.UUID  `gorm:"type:uuid;not null;column:user_id"`
	Content         string     `gorm:"type:text;not null"`
	IsPinned        bool       `gorm:"default:false"`
	WhoCanReply     string     `gorm:"type:varchar(20);not null"`
	ScheduledAt     *time.Time `gorm:"type:timestamp with time zone"`
	CommunityID     *uuid.UUID `gorm:"type:uuid"`
	IsAdvertisement bool       `gorm:"default:false"`
	CreatedAt       time.Time  `gorm:"autoCreateTime"`
	UpdatedAt       time.Time  `gorm:"autoUpdateTime"`
	DeletedAt       *time.Time `gorm:"index"`
	Media           []Media    `gorm:"foreignKey:ThreadID"`
}

type Reply struct {
	ReplyID       uuid.UUID  `gorm:"type:uuid;primaryKey;column:reply_id"`
	ThreadID      uuid.UUID  `gorm:"type:uuid;not null;column:thread_id"`
	UserID        uuid.UUID  `gorm:"type:uuid;not null;column:user_id"`
	Content       string     `gorm:"type:text;not null"`
	IsPinned      bool       `gorm:"default:false"`
	ParentReplyID *uuid.UUID `gorm:"type:uuid"`
	CreatedAt     time.Time  `gorm:"autoCreateTime"`
	UpdatedAt     time.Time  `gorm:"autoUpdateTime"`
	DeletedAt     *time.Time `gorm:"index"`
	Media         []Media    `gorm:"foreignKey:ReplyID"`
}

type Media struct {
	MediaID   uuid.UUID  `gorm:"type:uuid;primaryKey;column:media_id"`
	ThreadID  *uuid.UUID `gorm:"type:uuid;column:thread_id"`
	ReplyID   *uuid.UUID `gorm:"type:uuid;column:reply_id"`
	Type      string     `gorm:"type:varchar(10);not null"`
	URL       string     `gorm:"type:varchar(512);not null"`
	CreatedAt time.Time  `gorm:"autoCreateTime"`
	DeletedAt *time.Time `gorm:"index"`
}

type Hashtag struct {
	HashtagID uuid.UUID  `gorm:"type:uuid;primaryKey;column:hashtag_id"`
	Text      string     `gorm:"type:varchar(50);unique;not null"`
	CreatedAt time.Time  `gorm:"autoCreateTime"`
	DeletedAt *time.Time `gorm:"index"`
}

type UserMention struct {
	MentionID       uuid.UUID  `gorm:"type:uuid;primaryKey;column:mention_id"`
	MentionedUserID uuid.UUID  `gorm:"type:uuid;not null;column:mentioned_user_id"`
	ThreadID        *uuid.UUID `gorm:"type:uuid;column:thread_id"`
	ReplyID         *uuid.UUID `gorm:"type:uuid;column:reply_id"`
	CreatedAt       time.Time  `gorm:"autoCreateTime"`
	DeletedAt       *time.Time `gorm:"index"`
}

type Category struct {
	CategoryID uuid.UUID  `gorm:"type:uuid;primaryKey;column:category_id"`
	Name       string     `gorm:"type:varchar(50);not null"`
	Type       string     `gorm:"type:varchar(10);not null"`
	CreatedAt  time.Time  `gorm:"autoCreateTime"`
	UpdatedAt  time.Time  `gorm:"autoUpdateTime"`
	DeletedAt  *time.Time `gorm:"index"`
}

type Poll struct {
	PollID     uuid.UUID    `gorm:"type:uuid;primaryKey;column:poll_id"`
	ThreadID   uuid.UUID    `gorm:"type:uuid;not null;unique;column:thread_id"`
	Question   string       `gorm:"type:text;not null"`
	ClosesAt   time.Time    `gorm:"type:timestamp with time zone;not null"`
	WhoCanVote string       `gorm:"type:varchar(20);not null"`
	CreatedAt  time.Time    `gorm:"autoCreateTime"`
	UpdatedAt  time.Time    `gorm:"autoUpdateTime"`
	DeletedAt  *time.Time   `gorm:"index"`
	Options    []PollOption `gorm:"foreignKey:PollID"`
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

type Like struct {
	UserID    uuid.UUID  `gorm:"type:uuid;not null;column:user_id;primaryKey"`
	ThreadID  *uuid.UUID `gorm:"type:uuid;column:thread_id;primaryKey"`
	ReplyID   *uuid.UUID `gorm:"type:uuid;column:reply_id;primaryKey"`
	CreatedAt time.Time  `gorm:"autoCreateTime"`
	DeletedAt *time.Time `gorm:"index"`
}

type Repost struct {
	UserID     uuid.UUID  `gorm:"type:uuid;not null;column:user_id;primaryKey"`
	ThreadID   uuid.UUID  `gorm:"type:uuid;not null;column:thread_id;primaryKey"`
	RepostText *string    `gorm:"type:text"`
	CreatedAt  time.Time  `gorm:"autoCreateTime"`
	DeletedAt  *time.Time `gorm:"index"`
}

type Bookmark struct {
	UserID    uuid.UUID  `gorm:"type:uuid;not null;column:user_id;primaryKey"`
	ThreadID  uuid.UUID  `gorm:"type:uuid;not null;column:thread_id;primaryKey"`
	CreatedAt time.Time  `gorm:"autoCreateTime"`
	DeletedAt *time.Time `gorm:"index"`
}

type ThreadHashtag struct {
	ThreadID  uuid.UUID  `gorm:"type:uuid;primaryKey;column:thread_id"`
	HashtagID uuid.UUID  `gorm:"type:uuid;primaryKey;column:hashtag_id"`
	DeletedAt *time.Time `gorm:"index"`
}

type ThreadCategory struct {
	ThreadID   uuid.UUID  `gorm:"type:uuid;primaryKey;column:thread_id"`
	CategoryID uuid.UUID  `gorm:"type:uuid;primaryKey;column:category_id"`
	DeletedAt  *time.Time `gorm:"index"`
}
