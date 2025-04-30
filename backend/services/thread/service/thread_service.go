package service

import (
	"context"
	"errors"
	"time"

	"github.com/Acad600-Tpa/WEB-MV-242/backend/services/thread/model"
	"github.com/Acad600-Tpa/WEB-MV-242/backend/services/thread/proto"
	"github.com/Acad600-Tpa/WEB-MV-242/backend/services/thread/repository"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ThreadService defines the interface for thread operations
type ThreadService interface {
	CreateThread(ctx context.Context, req *proto.CreateThreadRequest) (*model.Thread, error)
	GetThreadByID(ctx context.Context, threadID string) (*model.Thread, error)
	GetThreadsByUserID(ctx context.Context, userID string, page, limit int) ([]*model.Thread, int64, error)
	UpdateThread(ctx context.Context, req *proto.UpdateThreadRequest) (*model.Thread, error)
	DeleteThread(ctx context.Context, threadID, userID string) error
	IncrementViewCount(ctx context.Context, threadID string) error
	GetThreadStats(ctx context.Context, threadID string) (replyCount, likeCount, repostCount int64, err error)
}

// threadService implements the ThreadService interface
type threadService struct {
	threadRepo  repository.ThreadRepository
	mediaRepo   repository.MediaRepository
	hashtagRepo repository.HashtagRepository
	mentionRepo repository.MentionRepository
}

// NewThreadService creates a new thread service
func NewThreadService(
	threadRepo repository.ThreadRepository,
	mediaRepo repository.MediaRepository,
	hashtagRepo repository.HashtagRepository,
	mentionRepo repository.MentionRepository,
) ThreadService {
	return &threadService{
		threadRepo:  threadRepo,
		mediaRepo:   mediaRepo,
		hashtagRepo: hashtagRepo,
		mentionRepo: mentionRepo,
	}
}

// CreateThread creates a new thread
func (s *threadService) CreateThread(ctx context.Context, req *proto.CreateThreadRequest) (*model.Thread, error) {
	// Validate required fields
	if req.UserId == "" || req.Content == "" {
		return nil, status.Error(codes.InvalidArgument, "User ID and content are required")
	}

	// Parse user ID
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid user ID: %v", err)
	}

	// Parse community ID if provided
	var communityID *uuid.UUID
	if req.CommunityId != "" {
		commID, err := uuid.Parse(req.CommunityId)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "Invalid community ID: %v", err)
		}
		communityID = &commID
	}

	// Parse scheduled time if provided
	var scheduledAt *time.Time
	if req.ScheduledAt != nil {
		t := req.ScheduledAt.AsTime()
		scheduledAt = &t
	}

	// Set default who_can_reply if not provided
	whoCanReply := req.WhoCanReply
	if whoCanReply == "" {
		whoCanReply = "Everyone"
	}

	// Create thread
	thread := &model.Thread{
		ThreadID:        uuid.New(),
		UserID:          userID,
		Content:         req.Content,
		WhoCanReply:     whoCanReply,
		ScheduledAt:     scheduledAt,
		CommunityID:     communityID,
		IsAdvertisement: req.IsAdvertisement,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Create thread in database
	if err := s.threadRepo.CreateThread(thread); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to create thread: %v", err)
	}

	// Process media attachments if any
	if len(req.Media) > 0 {
		for _, mediaInfo := range req.Media {
			media := &model.Media{
				MediaID:   uuid.New(),
				ThreadID:  &thread.ThreadID,
				Type:      mediaInfo.Type,
				URL:       mediaInfo.Url,
				CreatedAt: time.Now(),
			}
			if err := s.mediaRepo.CreateMedia(media); err != nil {
				return nil, status.Errorf(codes.Internal, "Failed to create media: %v", err)
			}
		}
	}

	// Process hashtags if any
	if len(req.Hashtags) > 0 {
		for _, hashtagText := range req.Hashtags {
			hashtag, err := s.hashtagRepo.FindOrCreateHashtagByText(hashtagText)
			if err != nil {
				return nil, status.Errorf(codes.Internal, "Failed to process hashtag: %v", err)
			}
			if err := s.hashtagRepo.AddHashtagToThread(thread.ThreadID.String(), hashtag.HashtagID.String()); err != nil {
				return nil, status.Errorf(codes.Internal, "Failed to link hashtag to thread: %v", err)
			}
		}
	}

	// Process mentions if any
	if len(req.MentionedUserIds) > 0 {
		for _, mentionedUserID := range req.MentionedUserIds {
			userUUID, err := uuid.Parse(mentionedUserID)
			if err != nil {
				return nil, status.Errorf(codes.InvalidArgument, "Invalid mentioned user ID: %v", err)
			}
			mention := &model.UserMention{
				MentionID:       uuid.New(),
				MentionedUserID: userUUID,
				ThreadID:        &thread.ThreadID,
				CreatedAt:       time.Now(),
			}
			if err := s.mentionRepo.CreateMention(mention); err != nil {
				return nil, status.Errorf(codes.Internal, "Failed to create mention: %v", err)
			}
		}
	}

	// TODO: Process poll if provided (implement in poll service)

	return thread, nil
}

// GetThreadByID retrieves a thread by its ID
func (s *threadService) GetThreadByID(ctx context.Context, threadID string) (*model.Thread, error) {
	if threadID == "" {
		return nil, status.Error(codes.InvalidArgument, "Thread ID is required")
	}

	thread, err := s.threadRepo.FindThreadByID(threadID)
	if err != nil {
		if errors.Is(err, repository.ErrThreadNotFound) {
			return nil, status.Errorf(codes.NotFound, "Thread with ID %s not found", threadID)
		}
		return nil, status.Errorf(codes.Internal, "Failed to retrieve thread: %v", err)
	}

	return thread, nil
}

// GetThreadsByUserID retrieves threads by user ID with pagination
func (s *threadService) GetThreadsByUserID(ctx context.Context, userID string, page, limit int) ([]*model.Thread, int64, error) {
	if userID == "" {
		return nil, 0, status.Error(codes.InvalidArgument, "User ID is required")
	}

	// Default pagination values if not provided
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	threads, totalCount, err := s.threadRepo.FindThreadsByUserID(userID, page, limit)
	if err != nil {
		return nil, 0, status.Errorf(codes.Internal, "Failed to retrieve threads: %v", err)
	}

	return threads, totalCount, nil
}

// UpdateThread updates a thread
func (s *threadService) UpdateThread(ctx context.Context, req *proto.UpdateThreadRequest) (*model.Thread, error) {
	if req.ThreadId == "" {
		return nil, status.Error(codes.InvalidArgument, "Thread ID is required")
	}

	// Get existing thread
	thread, err := s.threadRepo.FindThreadByID(req.ThreadId)
	if err != nil {
		if errors.Is(err, repository.ErrThreadNotFound) {
			return nil, status.Errorf(codes.NotFound, "Thread with ID %s not found", req.ThreadId)
		}
		return nil, status.Errorf(codes.Internal, "Failed to retrieve thread: %v", err)
	}

	// Update fields if provided
	updated := false
	if req.Content != "" {
		thread.Content = req.Content
		updated = true
	}

	// Update isPinned status
	thread.IsPinned = req.IsPinned

	// Add new categories if provided
	if len(req.AddCategoryNames) > 0 {
		// Logic to add categories would be here
		// Requires CategoryRepository implementation
		updated = true
	}

	// Remove categories if provided
	if len(req.RemoveCategoryNames) > 0 {
		// Logic to remove categories would be here
		// Requires CategoryRepository implementation
		updated = true
	}

	// Add new hashtags if provided
	if len(req.AddHashtags) > 0 {
		for _, hashtagText := range req.AddHashtags {
			hashtag, err := s.hashtagRepo.FindOrCreateHashtagByText(hashtagText)
			if err != nil {
				return nil, status.Errorf(codes.Internal, "Failed to process hashtag: %v", err)
			}
			if err := s.hashtagRepo.AddHashtagToThread(thread.ThreadID.String(), hashtag.HashtagID.String()); err != nil {
				return nil, status.Errorf(codes.Internal, "Failed to link hashtag to thread: %v", err)
			}
		}
		updated = true
	}

	// Remove hashtags if provided
	if len(req.RemoveHashtags) > 0 {
		for _, hashtagText := range req.RemoveHashtags {
			hashtag, err := s.hashtagRepo.FindHashtagByText(hashtagText)
			if err != nil {
				continue // Skip if hashtag doesn't exist
			}
			if err := s.hashtagRepo.RemoveHashtagFromThread(thread.ThreadID.String(), hashtag.HashtagID.String()); err != nil {
				return nil, status.Errorf(codes.Internal, "Failed to remove hashtag from thread: %v", err)
			}
		}
		updated = true
	}

	if updated {
		thread.UpdatedAt = time.Now()
		if err := s.threadRepo.UpdateThread(thread); err != nil {
			return nil, status.Errorf(codes.Internal, "Failed to update thread: %v", err)
		}
	}

	return thread, nil
}

// DeleteThread deletes a thread
func (s *threadService) DeleteThread(ctx context.Context, threadID, userID string) error {
	if threadID == "" || userID == "" {
		return status.Error(codes.InvalidArgument, "Thread ID and User ID are required")
	}

	// Get thread to check ownership
	thread, err := s.threadRepo.FindThreadByID(threadID)
	if err != nil {
		if errors.Is(err, repository.ErrThreadNotFound) {
			return status.Errorf(codes.NotFound, "Thread with ID %s not found", threadID)
		}
		return status.Errorf(codes.Internal, "Failed to retrieve thread: %v", err)
	}

	// Check if user is the owner
	if thread.UserID.String() != userID {
		return status.Error(codes.PermissionDenied, "User is not authorized to delete this thread")
	}

	// Delete thread from database
	if err := s.threadRepo.DeleteThread(threadID); err != nil {
		return status.Errorf(codes.Internal, "Failed to delete thread: %v", err)
	}

	return nil
}

// IncrementViewCount increments the view count for a thread
func (s *threadService) IncrementViewCount(ctx context.Context, threadID string) error {
	if threadID == "" {
		return status.Error(codes.InvalidArgument, "Thread ID is required")
	}

	if err := s.threadRepo.IncrementViewCount(threadID); err != nil {
		return status.Errorf(codes.Internal, "Failed to increment view count: %v", err)
	}

	return nil
}

// GetThreadStats retrieves thread statistics
func (s *threadService) GetThreadStats(ctx context.Context, threadID string) (replyCount, likeCount, repostCount int64, err error) {
	if threadID == "" {
		return 0, 0, 0, status.Error(codes.InvalidArgument, "Thread ID is required")
	}

	replyCount, likeCount, repostCount, err = s.threadRepo.GetThreadStats(threadID)
	if err != nil {
		return 0, 0, 0, status.Errorf(codes.Internal, "Failed to retrieve thread stats: %v", err)
	}

	return replyCount, likeCount, repostCount, nil
}

// Add ErrThreadNotFound to repository package to properly handle not found errors
var ErrThreadNotFound = errors.New("thread not found")
