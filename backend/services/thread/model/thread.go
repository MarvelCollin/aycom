package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Thread struct {
	ThreadID        uuid.UUID      `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"thread_id"`
	UserID          uuid.UUID      `gorm:"type:uuid;not null" json:"user_id"`
	Content         string         `gorm:"type:text;not null" json:"content"`
	IsPinned        bool           `gorm:"default:false" json:"is_pinned"`
	WhoCanReply     string         `gorm:"type:varchar(20);not null;check:who_can_reply IN ('Everyone', 'Accounts You Follow', 'Verified Accounts')" json:"who_can_reply"`
	ScheduledAt     *time.Time     `gorm:"type:timestamp with time zone" json:"scheduled_at"`
	CommunityID     *uuid.UUID     `gorm:"type:uuid" json:"community_id"`
	IsAdvertisement bool           `gorm:"default:false" json:"is_advertisement"`
	CreatedAt       time.Time      `gorm:"type:timestamp with time zone;not null;default:now()" json:"created_at"`
	UpdatedAt       time.Time      `gorm:"type:timestamp with time zone;not null;default:now()" json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`

	Media            []Media          `gorm:"foreignKey:ThreadID" json:"media,omitempty"`
	Likes            []Like           `gorm:"foreignKey:ThreadID" json:"likes,omitempty"`
	Reposts          []Repost         `gorm:"foreignKey:ThreadID" json:"reposts,omitempty"`
	Bookmarks        []Bookmark       `gorm:"foreignKey:ThreadID" json:"bookmarks,omitempty"`
	Replies          []Reply          `gorm:"foreignKey:ThreadID" json:"replies,omitempty"`
	ThreadHashtags   []ThreadHashtag  `gorm:"foreignKey:ThreadID" json:"thread_hashtags,omitempty"`
	ThreadCategories []ThreadCategory `gorm:"foreignKey:ThreadID" json:"thread_categories,omitempty"`
	UserMentions     []UserMention    `gorm:"foreignKey:ThreadID" json:"user_mentions,omitempty"`
	Poll             *Poll            `gorm:"foreignKey:ThreadID" json:"poll,omitempty"`
}

// TableName sets the table name for Thread model
func (Thread) TableName() string {
	return "threads"
}

type Reply struct {
	ReplyID       uuid.UUID      `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"reply_id"`
	ThreadID      uuid.UUID      `gorm:"type:uuid;not null" json:"thread_id"`
	UserID        uuid.UUID      `gorm:"type:uuid;not null" json:"user_id"`
	Content       string         `gorm:"type:text;not null" json:"content"`
	IsPinned      bool           `gorm:"default:false" json:"is_pinned"`
	ParentReplyID *uuid.UUID     `gorm:"type:uuid" json:"parent_reply_id"`
	CreatedAt     time.Time      `gorm:"type:timestamp with time zone;not null;default:now()" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"type:timestamp with time zone;not null;default:now()" json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Thread       *Thread       `gorm:"foreignKey:ThreadID" json:"-"`
	ParentReply  *Reply        `gorm:"foreignKey:ParentReplyID" json:"parent_reply,omitempty"`
	ChildReplies []Reply       `gorm:"foreignKey:ParentReplyID" json:"child_replies,omitempty"`
	Media        []Media       `gorm:"foreignKey:ReplyID" json:"media,omitempty"`
	UserMentions []UserMention `gorm:"foreignKey:ReplyID" json:"user_mentions,omitempty"`
	Likes        []Like        `gorm:"foreignKey:ReplyID" json:"likes,omitempty"`
}

// TableName sets the table name for Reply model
func (Reply) TableName() string {
	return "replies"
}

// Media represents an image, GIF, or video attached to a thread or reply
type Media struct {
	MediaID   uuid.UUID      `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"media_id"`
	ThreadID  *uuid.UUID     `gorm:"type:uuid" json:"thread_id"`
	ReplyID   *uuid.UUID     `gorm:"type:uuid" json:"reply_id"`
	Type      string         `gorm:"type:varchar(10);not null;check:type IN ('Image', 'GIF', 'Video')" json:"type"`
	URL       string         `gorm:"type:varchar(512);not null" json:"url"`
	CreatedAt time.Time      `gorm:"type:timestamp with time zone;not null;default:now()" json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationship references without constraint names
	Thread *Thread `gorm:"foreignKey:ThreadID" json:"-"`
	Reply  *Reply  `gorm:"foreignKey:ReplyID" json:"-"`
}

// Like represents a user's like on a thread or reply
type Like struct {
	UserID    uuid.UUID      `gorm:"primaryKey;type:uuid" json:"user_id"`
	ThreadID  *uuid.UUID     `gorm:"primaryKey;type:uuid" json:"thread_id"`
	ReplyID   *uuid.UUID     `gorm:"primaryKey;type:uuid" json:"reply_id"`
	CreatedAt time.Time      `gorm:"type:timestamp with time zone;not null;default:now()" json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Thread *Thread `gorm:"foreignKey:ThreadID" json:"-"`
	Reply  *Reply  `gorm:"foreignKey:ReplyID" json:"-"`
}

// Repost represents a user's repost of a thread
type Repost struct {
	UserID     uuid.UUID      `gorm:"primaryKey;type:uuid" json:"user_id"`
	ThreadID   uuid.UUID      `gorm:"primaryKey;type:uuid" json:"thread_id"`
	RepostText *string        `gorm:"type:text" json:"repost_text"`
	CreatedAt  time.Time      `gorm:"type:timestamp with time zone;not null;default:now()" json:"created_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Thread *Thread `gorm:"foreignKey:ThreadID" json:"-"`
}

// Bookmark represents a user's bookmark of a thread
type Bookmark struct {
	UserID    uuid.UUID      `gorm:"primaryKey;type:uuid" json:"user_id"`
	ThreadID  uuid.UUID      `gorm:"primaryKey;type:uuid" json:"thread_id"`
	CreatedAt time.Time      `gorm:"type:timestamp with time zone;not null;default:now()" json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Thread *Thread `gorm:"foreignKey:ThreadID" json:"-"`
}

// Hashtag represents a hashtag that can be used in threads
type Hashtag struct {
	HashtagID uuid.UUID      `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"hashtag_id"`
	Text      string         `gorm:"type:varchar(50);not null;unique" json:"text"`
	CreatedAt time.Time      `gorm:"type:timestamp with time zone;not null;default:now()" json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	ThreadHashtags []ThreadHashtag `gorm:"foreignKey:HashtagID" json:"-"`
}

// ThreadHashtag represents the many-to-many relationship between threads and hashtags
type ThreadHashtag struct {
	ThreadID  uuid.UUID      `gorm:"primaryKey;type:uuid" json:"thread_id"`
	HashtagID uuid.UUID      `gorm:"primaryKey;type:uuid" json:"hashtag_id"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Thread  *Thread  `gorm:"foreignKey:ThreadID" json:"-"`
	Hashtag *Hashtag `gorm:"foreignKey:HashtagID" json:"-"`
}

// UserMention represents a user mention in a thread or reply
type UserMention struct {
	MentionID       uuid.UUID      `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"mention_id"`
	MentionedUserID uuid.UUID      `gorm:"type:uuid;not null" json:"mentioned_user_id"`
	ThreadID        *uuid.UUID     `gorm:"type:uuid" json:"thread_id"`
	ReplyID         *uuid.UUID     `gorm:"type:uuid" json:"reply_id"`
	CreatedAt       time.Time      `gorm:"type:timestamp with time zone;not null;default:now()" json:"created_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Thread *Thread `gorm:"foreignKey:ThreadID" json:"-"`
	Reply  *Reply  `gorm:"foreignKey:ReplyID" json:"-"`
}

// Category represents a category for threads or communities
type Category struct {
	CategoryID uuid.UUID      `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"category_id"`
	Name       string         `gorm:"type:varchar(50);not null" json:"name"`
	Type       string         `gorm:"type:varchar(10);not null;check:type IN ('Thread', 'Community')" json:"type"`
	CreatedAt  time.Time      `gorm:"type:timestamp with time zone;not null;default:now()" json:"created_at"`
	UpdatedAt  time.Time      `gorm:"type:timestamp with time zone;not null;default:now()" json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	ThreadCategories []ThreadCategory `gorm:"foreignKey:CategoryID" json:"-"`
}

// ThreadCategory represents the many-to-many relationship between threads and categories
type ThreadCategory struct {
	ThreadID   uuid.UUID      `gorm:"primaryKey;type:uuid" json:"thread_id"`
	CategoryID uuid.UUID      `gorm:"primaryKey;type:uuid" json:"category_id"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Thread   *Thread   `gorm:"foreignKey:ThreadID" json:"-"`
	Category *Category `gorm:"foreignKey:CategoryID" json:"-"`
}

// Poll represents a poll attached to a thread
type Poll struct {
	PollID     uuid.UUID      `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"poll_id"`
	ThreadID   uuid.UUID      `gorm:"type:uuid;not null;unique" json:"thread_id"`
	Question   string         `gorm:"type:text;not null" json:"question"`
	ClosesAt   time.Time      `gorm:"type:timestamp with time zone;not null" json:"closes_at"`
	WhoCanVote string         `gorm:"type:varchar(20);not null;check:who_can_vote IN ('Everyone', 'Accounts You Follow', 'Verified Accounts')" json:"who_can_vote"`
	CreatedAt  time.Time      `gorm:"type:timestamp with time zone;not null;default:now()" json:"created_at"`
	UpdatedAt  time.Time      `gorm:"type:timestamp with time zone;not null;default:now()" json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Thread  *Thread      `gorm:"foreignKey:ThreadID" json:"-"`
	Options []PollOption `gorm:"foreignKey:PollID" json:"options,omitempty"`
	Votes   []PollVote   `gorm:"foreignKey:PollID" json:"votes,omitempty"`
}

// PollOption represents an option in a poll
type PollOption struct {
	OptionID  uuid.UUID      `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"option_id"`
	PollID    uuid.UUID      `gorm:"type:uuid;not null" json:"poll_id"`
	Text      string         `gorm:"type:varchar(100);not null" json:"text"`
	CreatedAt time.Time      `gorm:"type:timestamp with time zone;not null;default:now()" json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Poll  *Poll      `gorm:"foreignKey:PollID" json:"-"`
	Votes []PollVote `gorm:"foreignKey:OptionID" json:"votes,omitempty"`
}

// PollVote represents a user's vote on a poll option
type PollVote struct {
	VoteID    uuid.UUID      `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"vote_id"`
	PollID    uuid.UUID      `gorm:"type:uuid;not null" json:"poll_id"`
	OptionID  uuid.UUID      `gorm:"type:uuid;not null" json:"option_id"`
	UserID    uuid.UUID      `gorm:"type:uuid;not null" json:"user_id"`
	CreatedAt time.Time      `gorm:"type:timestamp with time zone;not null;default:now()" json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Poll   *Poll       `gorm:"foreignKey:PollID" json:"-"`
	Option *PollOption `gorm:"foreignKey:OptionID" json:"-"`
}
