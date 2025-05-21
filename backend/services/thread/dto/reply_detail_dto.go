package dto

import (
	"time"

	"aycom/backend/services/thread/model"
)

// UserDTO provides consistent user information
type UserDTO struct {
	ID                string  `json:"id"`
	Username          string  `json:"username"`
	Name              string  `json:"name"`
	ProfilePictureURL *string `json:"profile_picture_url"`
}

// ReplyDetailDTO provides a consistent structure for reply data
type ReplyDetailDTO struct {
	ID            string    `json:"id"`
	ThreadID      string    `json:"thread_id"`
	Content       string    `json:"content"`
	UserID        string    `json:"user_id"`
	User          UserDTO   `json:"user"`
	ParentReplyID *string   `json:"parent_reply_id,omitempty"`
	IsPinned      bool      `json:"is_pinned"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	// Counts
	LikesCount     int64 `json:"likes_count"`
	RepliesCount   int64 `json:"replies_count"`
	BookmarksCount int64 `json:"bookmarks_count"`

	// State flags for the requesting user
	IsLiked      bool `json:"is_liked"`
	IsBookmarked bool `json:"is_bookmarked"`

	// Media attachments
	Media []MediaDTO `json:"media,omitempty"`

	// For UI convenience, include these fields as well
	AuthorUsername string `json:"author_username"` // Same as User.Username
	AuthorName     string `json:"author_name"`     // Same as User.Name
	AuthorAvatar   string `json:"author_avatar"`   // Same as User.ProfilePictureURL
}

// MediaDTO represents media attached to a reply
type MediaDTO struct {
	ID           string `json:"id"`
	URL          string `json:"url"`
	Type         string `json:"type"`
	ThumbnailURL string `json:"thumbnail_url,omitempty"`
}

// FromReply converts a model.Reply to a ReplyDetailDTO
func FromReply(reply *model.Reply, user UserDTO, likesCount, repliesCount, bookmarksCount int64, isLiked, isBookmarked bool, media []MediaDTO) *ReplyDetailDTO {
	var parentReplyID *string
	if reply.ParentReplyID != nil {
		parentReplyIDStr := reply.ParentReplyID.String()
		parentReplyID = &parentReplyIDStr
	}

	// Default avatar if not provided
	avatarURL := ""
	if user.ProfilePictureURL != nil {
		avatarURL = *user.ProfilePictureURL
	}

	return &ReplyDetailDTO{
		ID:            reply.ReplyID.String(),
		ThreadID:      reply.ThreadID.String(),
		Content:       reply.Content,
		UserID:        reply.UserID.String(),
		User:          user,
		ParentReplyID: parentReplyID,
		IsPinned:      reply.IsPinned,
		CreatedAt:     reply.CreatedAt,
		UpdatedAt:     reply.UpdatedAt,

		LikesCount:     likesCount,
		RepliesCount:   repliesCount,
		BookmarksCount: bookmarksCount,

		IsLiked:      isLiked,
		IsBookmarked: isBookmarked,

		Media: media,

		// For UI convenience
		AuthorUsername: user.Username,
		AuthorName:     user.Name,
		AuthorAvatar:   avatarURL,
	}
}
