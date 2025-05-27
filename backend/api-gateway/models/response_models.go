package models

import "time"

type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type Pagination struct {
	Total   int  `json:"total"`
	Page    int  `json:"page"`
	Limit   int  `json:"limit"`
	HasMore bool `json:"hasMore"`
}

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

type UserSearchResponse struct {
	Users      []User     `json:"users"`
	Pagination Pagination `json:"pagination"`
}

type UserRecommendationsResponse struct {
	Users []User `json:"users"`
}

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

type ThreadSearchResponse struct {
	Threads    []Thread   `json:"threads"`
	Pagination Pagination `json:"pagination"`
}

type BookmarksResponse struct {
	Bookmarks  []Thread   `json:"bookmarks"`
	Pagination Pagination `json:"pagination"`
}

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

type CommunitySearchResponse struct {
	Communities []Community `json:"communities"`
	Pagination  Pagination  `json:"pagination"`
}

type Media struct {
	ID        string `json:"id"`
	Type      string `json:"type"` 
	URL       string `json:"url"`
	Thumbnail string `json:"thumbnail,omitempty"`
}

type MediaUploadResponse struct {
	ID        string `json:"id"`
	Type      string `json:"type"` 
	URL       string `json:"url"`
	Thumbnail string `json:"thumbnail,omitempty"`
}

type MediaSearchResponse struct {
	Media      []Media    `json:"media"`
	Pagination Pagination `json:"pagination"`
}

type Notification struct {
	ID            string    `json:"id"`
	Type          string    `json:"type"` 
	UserID        string    `json:"user_id"`
	Username      string    `json:"username"`
	DisplayName   string    `json:"display_name"`
	Avatar        string    `json:"avatar"`
	Timestamp     time.Time `json:"timestamp"`
	ThreadID      string    `json:"thread_id,omitempty"`
	ThreadContent string    `json:"thread_content,omitempty"`
	IsRead        bool      `json:"is_read"`
}

type NotificationsResponse struct {
	Notifications []Notification `json:"notifications"`
	Pagination    Pagination     `json:"pagination"`
}

type NotificationPreferencesResponse struct {
	Preferences NotificationPreferences `json:"preferences"`
}

type NotificationPreferences struct {
	Likes          bool `json:"likes"`
	Comments       bool `json:"comments"`
	Follows        bool `json:"follows"`
	Mentions       bool `json:"mentions"`
	DirectMessages bool `json:"direct_messages"`
}