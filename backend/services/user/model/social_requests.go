package model

// FollowUserRequest contains data for follow request
type FollowUserRequest struct {
	FollowerID string
	FollowedID string
}

// UnfollowUserRequest contains data for unfollow request
type UnfollowUserRequest struct {
	FollowerID string
	FollowedID string
}

// GetFollowersRequest contains data for fetching followers
type GetFollowersRequest struct {
	UserID string
	Page   int
	Limit  int
}

// GetFollowingRequest contains data for fetching users being followed
type GetFollowingRequest struct {
	UserID string
	Page   int
	Limit  int
}

// SearchUsersRequest contains data for searching users
type SearchUsersRequest struct {
	Query  string
	Filter string
	Page   int
	Limit  int
}
