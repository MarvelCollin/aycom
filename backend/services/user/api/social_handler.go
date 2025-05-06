package handlers

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	userProto "aycom/backend/proto/user"
	"aycom/backend/services/user/model"
)

// FollowUserRequest is a temporary struct until proto is updated
type FollowUserRequest struct {
	FollowerId string
	FollowedId string
}

// FollowUserResponse is a temporary struct until proto is updated
type FollowUserResponse struct {
	Success bool
	Message string
}

// UnfollowUserRequest is a temporary struct until proto is updated
type UnfollowUserRequest struct {
	FollowerId string
	FollowedId string
}

// UnfollowUserResponse is a temporary struct until proto is updated
type UnfollowUserResponse struct {
	Success bool
	Message string
}

// GetFollowersRequest is a temporary struct until proto is updated
type GetFollowersRequest struct {
	UserId string
	Page   int32
	Limit  int32
}

// GetFollowersResponse is a temporary struct until proto is updated
type GetFollowersResponse struct {
	Followers  []*userProto.User
	TotalCount int32
	Page       int32
	Limit      int32
}

// GetFollowingRequest is a temporary struct until proto is updated
type GetFollowingRequest struct {
	UserId string
	Page   int32
	Limit  int32
}

// GetFollowingResponse is a temporary struct until proto is updated
type GetFollowingResponse struct {
	Following  []*userProto.User
	TotalCount int32
	Page       int32
	Limit      int32
}

// SearchUsersRequest is a temporary struct until proto is updated
type SearchUsersRequest struct {
	Query  string
	Filter string
	Page   int32
	Limit  int32
}

// SearchUsersResponse is a temporary struct until proto is updated
type SearchUsersResponse struct {
	Users      []*userProto.User
	TotalCount int32
}

// SocialHandlerImpl provides temporary implementations for social functions
type SocialHandlerImpl struct {
	*UserHandler // Embed the UserHandler for access to service
}

// FollowUser handles the FollowUser gRPC request
func (h *SocialHandlerImpl) FollowUser(ctx context.Context, req *FollowUserRequest) (*FollowUserResponse, error) {
	if req.FollowerId == "" || req.FollowedId == "" {
		return nil, status.Error(codes.InvalidArgument, "Both follower and followed IDs are required")
	}

	// Convert proto request to model request
	modelReq := &model.FollowUserRequest{
		FollowerID: req.FollowerId,
		FollowedID: req.FollowedId,
	}

	err := h.svc.FollowUser(ctx, modelReq)
	if err != nil {
		return nil, err
	}

	return &FollowUserResponse{
		Success: true,
		Message: "User followed successfully",
	}, nil
}

// UnfollowUser handles the UnfollowUser gRPC request
func (h *SocialHandlerImpl) UnfollowUser(ctx context.Context, req *UnfollowUserRequest) (*UnfollowUserResponse, error) {
	if req.FollowerId == "" || req.FollowedId == "" {
		return nil, status.Error(codes.InvalidArgument, "Both follower and followed IDs are required")
	}

	// Convert proto request to model request
	modelReq := &model.UnfollowUserRequest{
		FollowerID: req.FollowerId,
		FollowedID: req.FollowedId,
	}

	err := h.svc.UnfollowUser(ctx, modelReq)
	if err != nil {
		return nil, err
	}

	return &UnfollowUserResponse{
		Success: true,
		Message: "User unfollowed successfully",
	}, nil
}

// GetFollowers handles the GetFollowers gRPC request
func (h *SocialHandlerImpl) GetFollowers(ctx context.Context, req *GetFollowersRequest) (*GetFollowersResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "User ID is required")
	}

	// Convert proto request to model request
	modelReq := &model.GetFollowersRequest{
		UserID: req.UserId,
		Page:   int(req.Page),
		Limit:  int(req.Limit),
	}

	followers, total, err := h.svc.GetFollowers(ctx, modelReq)
	if err != nil {
		return nil, err
	}

	// Convert model users to proto users
	protoFollowers := make([]*userProto.User, len(followers))
	for i, follower := range followers {
		protoFollowers[i] = mapUserModelToProto(follower)
	}

	return &GetFollowersResponse{
		Followers:  protoFollowers,
		TotalCount: int32(total),
		Page:       req.Page,
		Limit:      req.Limit,
	}, nil
}

// GetFollowing handles the GetFollowing gRPC request
func (h *SocialHandlerImpl) GetFollowing(ctx context.Context, req *GetFollowingRequest) (*GetFollowingResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "User ID is required")
	}

	// Convert proto request to model request
	modelReq := &model.GetFollowingRequest{
		UserID: req.UserId,
		Page:   int(req.Page),
		Limit:  int(req.Limit),
	}

	following, total, err := h.svc.GetFollowing(ctx, modelReq)
	if err != nil {
		return nil, err
	}

	// Convert model users to proto users
	protoFollowing := make([]*userProto.User, len(following))
	for i, followed := range following {
		protoFollowing[i] = mapUserModelToProto(followed)
	}

	return &GetFollowingResponse{
		Following:  protoFollowing,
		TotalCount: int32(total),
		Page:       req.Page,
		Limit:      req.Limit,
	}, nil
}

// SearchUsers handles the SearchUsers gRPC request
func (h *SocialHandlerImpl) SearchUsers(ctx context.Context, req *SearchUsersRequest) (*SearchUsersResponse, error) {
	if req.Query == "" {
		return nil, status.Error(codes.InvalidArgument, "Search query is required")
	}

	// Convert proto request to model request
	modelReq := &model.SearchUsersRequest{
		Query:  req.Query,
		Filter: req.Filter,
		Page:   int(req.Page),
		Limit:  int(req.Limit),
	}

	users, total, err := h.svc.SearchUsers(ctx, modelReq)
	if err != nil {
		return nil, err
	}

	// Convert model users to proto users
	protoUsers := make([]*userProto.User, len(users))
	for i, user := range users {
		protoUsers[i] = mapUserModelToProto(user)
	}

	return &SearchUsersResponse{
		Users:      protoUsers,
		TotalCount: int32(total),
	}, nil
}
