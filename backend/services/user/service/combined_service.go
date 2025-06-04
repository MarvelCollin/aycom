package service

import (
	"context"
	"aycom/backend/proto/user"

	"aycom/backend/services/user/model"

)

type CombinedService struct {
	userService  UserService
	blockService *UserBlockService
}

func NewCombinedService(userService UserService, blockService *UserBlockService) *CombinedService {
	return &CombinedService{
		userService:  userService,
		blockService: blockService,
	}
}

func (s *CombinedService) CreateUserProfile(ctx context.Context, req *user.CreateUserRequest) (*model.User, error) {
	return s.userService.CreateUserProfile(ctx, req)
}

func (s *CombinedService) GetUserByID(ctx context.Context, id string) (*model.User, error) {
	return s.userService.GetUserByID(ctx, id)
}

func (s *CombinedService) GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	return s.userService.GetUserByUsername(ctx, username)
}

func (s *CombinedService) UpdateUserProfile(ctx context.Context, req *user.UpdateUserRequest) (*model.User, error) {
	return s.userService.UpdateUserProfile(ctx, req)
}

func (s *CombinedService) UpdateUserVerificationStatus(ctx context.Context, req *user.UpdateUserVerificationStatusRequest) error {
	return s.userService.UpdateUserVerificationStatus(ctx, req)
}

func (s *CombinedService) DeleteUser(ctx context.Context, id string) error {
	return s.userService.DeleteUser(ctx, id)
}

func (s *CombinedService) LoginUser(ctx context.Context, req *user.LoginUserRequest) (*model.User, error) {
	return s.userService.LoginUser(ctx, req)
}

func (s *CombinedService) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	return s.userService.GetUserByEmail(ctx, email)
}

func (s *CombinedService) FollowUser(ctx context.Context, req *model.FollowUserRequest) error {
	return s.userService.FollowUser(ctx, req)
}

func (s *CombinedService) UnfollowUser(ctx context.Context, req *model.UnfollowUserRequest) error {
	return s.userService.UnfollowUser(ctx, req)
}

func (s *CombinedService) GetFollowers(ctx context.Context, req *model.GetFollowersRequest) ([]*model.User, int, error) {
	return s.userService.GetFollowers(ctx, req)
}

func (s *CombinedService) GetFollowing(ctx context.Context, req *model.GetFollowingRequest) ([]*model.User, int, error) {
	return s.userService.GetFollowing(ctx, req)
}

func (s *CombinedService) IsFollowing(ctx context.Context, followerID, followedID string) (bool, error) {
	return s.userService.IsFollowing(ctx, followerID, followedID)
}

func (s *CombinedService) SearchUsers(ctx context.Context, req *model.SearchUsersRequest) ([]*model.User, int, error) {
	return s.userService.SearchUsers(ctx, req)
}

func (s *CombinedService) GetRecommendedUsers(ctx context.Context, limit int) ([]*model.User, error) {
	return s.userService.GetRecommendedUsers(ctx, limit)
}

func (s *CombinedService) GetAllUsers(ctx context.Context, page, limit int, sortBy string, ascending bool) ([]*model.User, int, error) {
	return s.userService.GetAllUsers(ctx, page, limit, sortBy, ascending)
}

func (s *CombinedService) BlockUser(ctx context.Context, blockerID, blockedID string) error {
	return s.blockService.BlockUser(ctx, blockerID, blockedID)
}

func (s *CombinedService) UnblockUser(ctx context.Context, unblockerID, unblockedID string) error {
	return s.blockService.UnblockUser(ctx, unblockerID, unblockedID)
}

func (s *CombinedService) IsUserBlocked(ctx context.Context, userID, blockedByID string) (bool, error) {
	return s.blockService.IsUserBlocked(ctx, userID, blockedByID)
}

func (s *CombinedService) ReportUser(ctx context.Context, reporterID, reportedID string, reason string) error {
	return s.blockService.ReportUser(ctx, reporterID, reportedID, reason)
}

func (s *CombinedService) GetBlockedUsers(ctx context.Context, userID string, page, limit int) ([]map[string]interface{}, int64, error) {
	return s.blockService.GetBlockedUsers(ctx, userID, page, limit)
}