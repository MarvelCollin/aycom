package service

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"

	"aycom/backend/services/user/repository"
)

// BlockService handles user block functionality
type BlockService struct {
	blockRepo  *repository.BlockRepository
	reportRepo *repository.ReportRepository
	userRepo   repository.UserRepository
}

// NewBlockService creates a new BlockService
func NewBlockService(blockRepo *repository.BlockRepository, reportRepo *repository.ReportRepository, userRepo repository.UserRepository) *BlockService {
	return &BlockService{
		blockRepo:  blockRepo,
		reportRepo: reportRepo,
		userRepo:   userRepo,
	}
}

// BlockUser creates a block relationship between users
func (s *BlockService) BlockUser(ctx context.Context, blockerID, blockedID string) error {
	// Validate UUIDs
	blockerUUID, err := uuid.Parse(blockerID)
	if err != nil {
		return fmt.Errorf("invalid blocker ID: %w", err)
	}

	blockedUUID, err := uuid.Parse(blockedID)
	if err != nil {
		return fmt.Errorf("invalid blocked ID: %w", err)
	}
	// Check if users exist
	_, err = s.userRepo.FindUserByID(blockerUUID.String())
	if err != nil {
		return fmt.Errorf("blocker user not found: %w", err)
	}

	_, err = s.userRepo.FindUserByID(blockedUUID.String())
	if err != nil {
		return fmt.Errorf("blocked user not found: %w", err)
	}

	// Create block relationship
	log.Printf("Creating block: %s blocks %s", blockerUUID, blockedUUID)
	return s.blockRepo.BlockUser(blockerUUID, blockedUUID)
}

// UnblockUser removes a block relationship between users
func (s *BlockService) UnblockUser(ctx context.Context, blockerID, blockedID string) error {
	// Validate UUIDs
	blockerUUID, err := uuid.Parse(blockerID)
	if err != nil {
		return fmt.Errorf("invalid blocker ID: %w", err)
	}

	blockedUUID, err := uuid.Parse(blockedID)
	if err != nil {
		return fmt.Errorf("invalid blocked ID: %w", err)
	}

	// Remove block relationship
	log.Printf("Removing block: %s unblocks %s", blockerUUID, blockedUUID)
	return s.blockRepo.UnblockUser(blockerUUID, blockedUUID)
}

// IsUserBlocked checks if a user is blocked by another user
func (s *BlockService) IsUserBlocked(ctx context.Context, userID, blockedByID string) (bool, error) {
	// Validate UUIDs
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return false, fmt.Errorf("invalid user ID: %w", err)
	}

	blockedByUUID, err := uuid.Parse(blockedByID)
	if err != nil {
		return false, fmt.Errorf("invalid blocked_by ID: %w", err)
	}

	return s.blockRepo.IsUserBlocked(userUUID, blockedByUUID)
}

// ReportUser creates a new user report
func (s *BlockService) ReportUser(ctx context.Context, reporterID, reportedID string, reason string) error {
	// Validate UUIDs
	reporterUUID, err := uuid.Parse(reporterID)
	if err != nil {
		return fmt.Errorf("invalid reporter ID: %w", err)
	}

	reportedUUID, err := uuid.Parse(reportedID)
	if err != nil {
		return fmt.Errorf("invalid reported ID: %w", err)
	}
	// Check if users exist
	_, err = s.userRepo.FindUserByID(reporterUUID.String())
	if err != nil {
		return fmt.Errorf("reporter user not found: %w", err)
	}

	_, err = s.userRepo.FindUserByID(reportedUUID.String())
	if err != nil {
		return fmt.Errorf("reported user not found: %w", err)
	}

	// Create report
	log.Printf("Creating report: %s reports %s: %s", reporterUUID, reportedUUID, reason)
	return s.reportRepo.ReportUser(reporterUUID, reportedUUID, reason)
}

// GetBlockedUsers returns users blocked by the specified user
func (s *BlockService) GetBlockedUsers(ctx context.Context, userID string, page, limit int) ([]map[string]interface{}, int64, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, 0, fmt.Errorf("invalid user ID: %w", err)
	}

	blockedUsers, total, err := s.blockRepo.GetBlockedUsers(userUUID, page, limit)
	if err != nil {
		return nil, 0, err
	}

	// Convert to generic map format
	result := make([]map[string]interface{}, len(blockedUsers))
	for i, user := range blockedUsers {
		result[i] = map[string]interface{}{
			"id":                  user.ID.String(),
			"username":            user.Username,
			"display_name":        user.Name,
			"profile_picture_url": user.ProfilePictureURL,
			"is_verified":         user.IsVerified,
		}
	}

	return result, total, nil
}
