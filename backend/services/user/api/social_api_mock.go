package handlers

import (
	"context"
	"log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"aycom/backend/services/user/model"
	"aycom/backend/services/user/service"
)

// MockSocialService provides temporary implementations for social functionalities
// until the proper proto-based implementations are available
type MockSocialService struct {
	svc service.UserService // The actual service implementation
}

// NewMockSocialService creates a new mock social service with the provided UserService
func NewMockSocialService(svc service.UserService) *MockSocialService {
	return &MockSocialService{svc: svc}
}

// FollowUser handles a user following another user
func (m *MockSocialService) FollowUser(ctx context.Context, followerId, followedId string) error {
	if followerId == "" || followedId == "" {
		return status.Error(codes.InvalidArgument, "Both follower and followed IDs are required")
	}

	req := &model.FollowUserRequest{
		FollowerID: followerId,
		FollowedID: followedId,
	}

	err := m.svc.FollowUser(ctx, req)
	if err != nil {
		log.Printf("Error following user: %v", err)
		return err
	}

	return nil
}

// UnfollowUser handles a user unfollowing another user
func (m *MockSocialService) UnfollowUser(ctx context.Context, followerId, followedId string) error {
	if followerId == "" || followedId == "" {
		return status.Error(codes.InvalidArgument, "Both follower and followed IDs are required")
	}

	req := &model.UnfollowUserRequest{
		FollowerID: followerId,
		FollowedID: followedId,
	}

	err := m.svc.UnfollowUser(ctx, req)
	if err != nil {
		log.Printf("Error unfollowing user: %v", err)
		return err
	}

	return nil
}

// GetFollowers returns a list of users who follow the specified user
func (m *MockSocialService) GetFollowers(ctx context.Context, userId string, page, limit int) ([]*model.User, int, error) {
	if userId == "" {
		return nil, 0, status.Error(codes.InvalidArgument, "User ID is required")
	}

	req := &model.GetFollowersRequest{
		UserID: userId,
		Page:   page,
		Limit:  limit,
	}

	followers, total, err := m.svc.GetFollowers(ctx, req)
	if err != nil {
		log.Printf("Error getting followers: %v", err)
		return nil, 0, err
	}

	return followers, total, nil
}

// GetFollowing returns a list of users the specified user is following
func (m *MockSocialService) GetFollowing(ctx context.Context, userId string, page, limit int) ([]*model.User, int, error) {
	if userId == "" {
		return nil, 0, status.Error(codes.InvalidArgument, "User ID is required")
	}

	req := &model.GetFollowingRequest{
		UserID: userId,
		Page:   page,
		Limit:  limit,
	}

	following, total, err := m.svc.GetFollowing(ctx, req)
	if err != nil {
		log.Printf("Error getting following: %v", err)
		return nil, 0, err
	}

	return following, total, nil
}

// SearchUsers searches for users based on a query
func (m *MockSocialService) SearchUsers(ctx context.Context, query, filter string, page, limit int) ([]*model.User, int, error) {
	if query == "" {
		return nil, 0, status.Error(codes.InvalidArgument, "Search query is required")
	}

	req := &model.SearchUsersRequest{
		Query:  query,
		Filter: filter,
		Page:   page,
		Limit:  limit,
	}

	users, total, err := m.svc.SearchUsers(ctx, req)
	if err != nil {
		log.Printf("Error searching users: %v", err)
		return nil, 0, err
	}

	return users, total, nil
}
