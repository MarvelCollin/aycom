package models

import "time"

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// SuccessResponse represents a success response
type SuccessResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// Pagination represents pagination information
type Pagination struct {
	Total   int  `json:"total"`
	Page    int  `json:"page"`
	Limit   int  `json:"limit"`
	HasMore bool `json:"hasMore"`
}

// User represents a user in the system
type User struct {
	ID                string    `json:"id"`
	Username          string    `json:"username"`
	Name              string    `json:"name"`
	ProfilePictureURL string    `json:"profile_picture_url"`
	Bio               string    `json:"bio"`
	IsVerified        bool      `json:"is_verified"`
	FollowerCount     int       `json:"follower_count"`
	IsFollowing       bool      `json:"is_following"`
	CreatedAt         time.Time `json:"created_at"`
}

// UserSearchResponse represents a response for user search
type UserSearchResponse struct {
	Users      []User     `json:"users"`
	Pagination Pagination `json:"pagination"`
}

// UserRecommendationsResponse represents a response for user recommendations
type UserRecommendationsResponse struct {
	Users []User `json:"users"`
}

// Thread represents a thread (post/tweet) in the system
type Thread struct {
	ID             string    `json:"id"`
	Content        string    `json:"content"`
	UserID         string    `json:"user_id"`
	Username       string    `json:"username"`
	DisplayName    string    `json:"display_name"`
	ProfilePicture string    `json:"profile_picture"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	LikeCount      int       `json:"like_count"`
	ReplyCount     int       `json:"reply_count"`
	RepostCount    int       `json:"repost_count"`
	IsLiked        bool      `json:"is_liked"`
	IsReposted     bool      `json:"is_reposted"`
	IsBookmarked   bool      `json:"is_bookmarked"`
	Media          []Media   `json:"media"`
	ParentID       string    `json:"parent_id,omitempty"`
}

// ThreadSearchResponse represents a response for thread search
type ThreadSearchResponse struct {
	Threads    []Thread   `json:"threads"`
	Pagination Pagination `json:"pagination"`
}

// BookmarksResponse represents a response for bookmarks
type BookmarksResponse struct {
	Bookmarks  []Thread   `json:"bookmarks"`
	Pagination Pagination `json:"pagination"`
}

// Community represents a community in the system
type Community struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Avatar      string    `json:"avatar"`
	Banner      string    `json:"banner"`
	MemberCount int       `json:"member_count"`
	IsJoined    bool      `json:"is_joined"`
	IsApproved  bool      `json:"is_approved"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CommunitySearchResponse represents a response for community search
type CommunitySearchResponse struct {
	Communities []Community `json:"communities"`
	Pagination  Pagination  `json:"pagination"`
}

// Media represents a media file (image, gif, video)
type Media struct {
	ID        string `json:"id"`
	Type      string `json:"type"` // image, gif, video
	URL       string `json:"url"`
	Thumbnail string `json:"thumbnail,omitempty"`
}

// MediaUploadResponse represents a response for media upload
type MediaUploadResponse struct {
	ID        string `json:"id"`
	Type      string `json:"type"` // image, gif, video
	URL       string `json:"url"`
	Thumbnail string `json:"thumbnail,omitempty"`
}

// MediaSearchResponse represents a response for media search
type MediaSearchResponse struct {
	Media      []Media    `json:"media"`
	Pagination Pagination `json:"pagination"`
}

// Notification represents a notification
type Notification struct {
	ID            string    `json:"id"`
	Type          string    `json:"type"` // like, repost, follow, mention
	UserID        string    `json:"user_id"`
	Username      string    `json:"username"`
	DisplayName   string    `json:"display_name"`
	Avatar        string    `json:"avatar"`
	Timestamp     time.Time `json:"timestamp"`
	ThreadID      string    `json:"thread_id,omitempty"`
	ThreadContent string    `json:"thread_content,omitempty"`
	IsRead        bool      `json:"is_read"`
}

// NotificationsResponse represents a response for notifications
type NotificationsResponse struct {
	Notifications []Notification `json:"notifications"`
	Pagination    Pagination     `json:"pagination"`
}

// NotificationPreferencesResponse contains user's notification preference settings
type NotificationPreferencesResponse struct {
	Preferences NotificationPreferences `json:"preferences"`
}

// NotificationPreferences contains various notification preference settings
type NotificationPreferences struct {
	Likes          bool `json:"likes"`
	Comments       bool `json:"comments"`
	Follows        bool `json:"follows"`
	Mentions       bool `json:"mentions"`
	DirectMessages bool `json:"direct_messages"`
}
