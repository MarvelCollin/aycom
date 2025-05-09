package model

type FollowUserRequest struct {
	FollowerID string
	FollowedID string
}

type UnfollowUserRequest struct {
	FollowerID string
	FollowedID string
}

type GetFollowersRequest struct {
	UserID string
	Page   int
	Limit  int
}

type GetFollowingRequest struct {
	UserID string
	Page   int
	Limit  int
}

type SearchUsersRequest struct {
	Query  string
	Filter string
	Page   int
	Limit  int
}
