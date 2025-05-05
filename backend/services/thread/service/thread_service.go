package service

import (
	"context"
	"errors"
	"time"

	"aycom/backend/proto/thread"
	"aycom/backend/services/thread/model"
	"aycom/backend/services/thread/repository"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

// ThreadService defines the interface for thread operations
type ThreadService interface {
	CreateThread(ctx context.Context, req *thread.CreateThreadRequest) (*model.Thread, error)
	GetThreadByID(ctx context.Context, threadID string) (*model.Thread, error)
	GetThreadsByUserID(ctx context.Context, userID string, page, limit int) ([]*model.Thread, error)
	GetAllThreads(ctx context.Context, page, limit int) ([]*model.Thread, error)
	UpdateThread(ctx context.Context, req *thread.UpdateThreadRequest) (*model.Thread, error)
	DeleteThread(ctx context.Context, threadID, userID string) error
}

// threadService implements the ThreadService interface
type threadService struct {
	threadRepo  repository.ThreadRepository
	mediaRepo   repository.MediaRepository
	hashtagRepo repository.HashtagRepository
	replyRepo   repository.ReplyRepository
}

// NewThreadService creates a new thread service
func NewThreadService(
	threadRepo repository.ThreadRepository,
	mediaRepo repository.MediaRepository,
	hashtagRepo repository.HashtagRepository,
	replyRepo repository.ReplyRepository,
) ThreadService {
	return &threadService{
		threadRepo:  threadRepo,
		mediaRepo:   mediaRepo,
		hashtagRepo: hashtagRepo,
		replyRepo:   replyRepo,
	}
}

// CreateThread creates a new thread
func (s *threadService) CreateThread(ctx context.Context, req *thread.CreateThreadRequest) (*model.Thread, error) {
	// Validate required fields
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "User ID is required")
	}

	// Parse user ID
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid user ID: %v", err)
	}

	// Parse community ID if provided
	var communityID *uuid.UUID
	if req.CommunityId != nil && *req.CommunityId != "" {
		commID, err := uuid.Parse(*req.CommunityId)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "Invalid community ID: %v", err)
		}
		communityID = &commID
	}

	// Set scheduled post time if provided
	var scheduledAt *time.Time
	if req.ScheduledAt != nil {
		t := req.ScheduledAt.AsTime()
		scheduledAt = &t
	}

	// Set who can reply if provided
	var whoCanReply string = "Everyone" // Default value
	if req.WhoCanReply != nil && *req.WhoCanReply != "" {
		whoCanReply = *req.WhoCanReply
	}

	// Set isAdvertisement if provided
	isAdvertisement := false // Default value
	if req.IsAdvertisement != nil {
		isAdvertisement = *req.IsAdvertisement
	}

	// Create thread
	threadID := uuid.New()
	thread := &model.Thread{
		ThreadID:        threadID,
		UserID:          userID,
		Content:         req.Content,
		CommunityID:     communityID,
		ScheduledAt:     scheduledAt,
		WhoCanReply:     whoCanReply,
		IsAdvertisement: isAdvertisement,
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
			// First look for existing hashtag
			hashtag, err := s.hashtagRepo.FindHashtagByText(hashtagText)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					// Create new hashtag
					hashtag = &model.Hashtag{
						HashtagID: uuid.New(),
						Text:      hashtagText,
						CreatedAt: time.Now(),
					}
					if err := s.hashtagRepo.CreateHashtag(hashtag); err != nil {
						return nil, status.Errorf(codes.Internal, "Failed to create hashtag: %v", err)
					}
				} else {
					return nil, status.Errorf(codes.Internal, "Failed to process hashtag: %v", err)
				}
			}

			if err := s.hashtagRepo.AddHashtagToThread(thread.ThreadID.String(), hashtag.HashtagID.String()); err != nil {
				return nil, status.Errorf(codes.Internal, "Failed to link hashtag to thread: %v", err)
			}
		}
	}

	return thread, nil
}

// GetThreadByID retrieves a thread by its ID
func (s *threadService) GetThreadByID(ctx context.Context, threadID string) (*model.Thread, error) {
	if threadID == "" {
		return nil, status.Error(codes.InvalidArgument, "Thread ID is required")
	}

	thread, err := s.threadRepo.FindThreadByID(threadID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "Thread with ID %s not found", threadID)
		}
		return nil, status.Errorf(codes.Internal, "Failed to retrieve thread: %v", err)
	}

	return thread, nil
}

// GetThreadsByUserID retrieves threads by user ID with pagination
func (s *threadService) GetThreadsByUserID(ctx context.Context, userID string, page, limit int) ([]*model.Thread, error) {
	// Default pagination values if not provided
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	// If userID is empty, get all threads instead
	if userID == "" {
		return s.GetAllThreads(ctx, page, limit)
	}

	threads, err := s.threadRepo.FindThreadsByUserID(userID, page, limit)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to retrieve threads: %v", err)
	}

	return threads, nil
}

// GetAllThreads retrieves all threads with pagination
func (s *threadService) GetAllThreads(ctx context.Context, page, limit int) ([]*model.Thread, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	threads, err := s.threadRepo.FindAllThreads(page, limit)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to retrieve threads: %v", err)
	}

	return threads, nil
}

// UpdateThread updates a thread
func (s *threadService) UpdateThread(ctx context.Context, req *thread.UpdateThreadRequest) (*model.Thread, error) {
	if req.ThreadId == "" {
		return nil, status.Error(codes.InvalidArgument, "Thread ID is required")
	}

	// Get existing thread
	thread, err := s.threadRepo.FindThreadByID(req.ThreadId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
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

	// Update isPinned status if provided
	if req.IsPinned != nil {
		thread.IsPinned = *req.IsPinned
		updated = true
	}

	// Add new hashtags if provided
	if len(req.AddHashtags) > 0 {
		for _, hashtagText := range req.AddHashtags {
			// First look for existing hashtag
			hashtag, err := s.hashtagRepo.FindHashtagByText(hashtagText)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					// Create new hashtag
					hashtag = &model.Hashtag{
						HashtagID: uuid.New(),
						Text:      hashtagText,
						CreatedAt: time.Now(),
					}
					if err := s.hashtagRepo.CreateHashtag(hashtag); err != nil {
						return nil, status.Errorf(codes.Internal, "Failed to create hashtag: %v", err)
					}
				} else {
					return nil, status.Errorf(codes.Internal, "Failed to process hashtag: %v", err)
				}
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
		if errors.Is(err, gorm.ErrRecordNotFound) {
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

// Request/response types needed for the service implementation
type BookmarkReplyRequest struct {
	ReplyId string
	UserId  string
}

type RemoveReplyBookmarkRequest struct {
	ReplyId string
	UserId  string
}
