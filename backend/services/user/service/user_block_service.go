package service

import (
	"context"

	"aycom/backend/services/user/repository"
)

// UserBlockService handles the block and report functionality
type UserBlockService struct {
	blockService *BlockService
}

// NewUserBlockService creates a new UserBlockService
func NewUserBlockService(
	blockRepo *repository.BlockRepository,
	reportRepo *repository.ReportRepository,
	userRepo repository.UserRepository,
) *UserBlockService {
	return &UserBlockService{
		blockService: NewBlockService(blockRepo, reportRepo, userRepo),
	}
}

// BlockUser creates a block relationship
func (s *UserBlockService) BlockUser(ctx context.Context, blockerID, blockedID string) error {
	return s.blockService.BlockUser(ctx, blockerID, blockedID)
}

// UnblockUser removes a block relationship
func (s *UserBlockService) UnblockUser(ctx context.Context, blockerID, blockedID string) error {
	return s.blockService.UnblockUser(ctx, blockerID, blockedID)
}

// IsUserBlocked checks if a user is blocked
func (s *UserBlockService) IsUserBlocked(ctx context.Context, userID, blockedByID string) (bool, error) {
	return s.blockService.IsUserBlocked(ctx, userID, blockedByID)
}

// ReportUser reports a user
func (s *UserBlockService) ReportUser(ctx context.Context, reporterID, reportedID, reason string) error {
	return s.blockService.ReportUser(ctx, reporterID, reportedID, reason)
}

// GetBlockedUsers returns a list of blocked users
func (s *UserBlockService) GetBlockedUsers(ctx context.Context, userID string, page, limit int) ([]map[string]interface{}, int64, error) {
	return s.blockService.GetBlockedUsers(ctx, userID, page, limit)
}
