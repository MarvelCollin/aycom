package service

import (
	"context"

	"aycom/backend/services/user/repository"
)

type UserBlockService struct {
	blockService *BlockService
}

func NewUserBlockService(
	blockRepo *repository.BlockRepository,
	reportRepo *repository.ReportRepository,
	userRepo repository.UserRepository,
) *UserBlockService {
	return &UserBlockService{
		blockService: NewBlockService(blockRepo, reportRepo, userRepo),
	}
}

func (s *UserBlockService) BlockUser(ctx context.Context, blockerID, blockedID string) error {
	return s.blockService.BlockUser(ctx, blockerID, blockedID)
}

func (s *UserBlockService) UnblockUser(ctx context.Context, blockerID, blockedID string) error {
	return s.blockService.UnblockUser(ctx, blockerID, blockedID)
}

func (s *UserBlockService) IsUserBlocked(ctx context.Context, userID, blockedByID string) (bool, error) {
	return s.blockService.IsUserBlocked(ctx, userID, blockedByID)
}

func (s *UserBlockService) ReportUser(ctx context.Context, reporterID, reportedID, reason string) error {
	return s.blockService.ReportUser(ctx, reporterID, reportedID, reason)
}

func (s *UserBlockService) GetBlockedUsers(ctx context.Context, userID string, page, limit int) ([]map[string]interface{}, int64, error) {
	return s.blockService.GetBlockedUsers(ctx, userID, page, limit)
}
