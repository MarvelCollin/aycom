package service

import (
	"context"
	"errors"
	"log"
	"time"

	"aycom/backend/proto/thread"
	"aycom/backend/services/thread/model"
	"aycom/backend/services/thread/repository"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type ThreadService interface {
	CreateThread(ctx context.Context, req *thread.CreateThreadRequest) (*model.Thread, error)
	GetThreadByID(ctx context.Context, threadID string) (*model.Thread, error)
	GetThreadsByUserID(ctx context.Context, userID string, page, limit int) ([]*model.Thread, error)
	GetAllThreads(ctx context.Context, page, limit int) ([]*model.Thread, error)
	GetTotalThreadCount(ctx context.Context) (int64, error)
	UpdateThread(ctx context.Context, req *thread.UpdateThreadRequest) (*model.Thread, error)
	DeleteThread(ctx context.Context, threadID, userID string) error
	GetMediaByUserID(ctx context.Context, userID string, page, limit int) ([]*model.Media, error)

	PinThread(ctx context.Context, threadID, userID string) error
	UnpinThread(ctx context.Context, threadID, userID string) error
	PinReply(ctx context.Context, replyID, userID string) error
	UnpinReply(ctx context.Context, replyID, userID string) error
}

type threadService struct {
	threadRepo  repository.ThreadRepository
	mediaRepo   repository.MediaRepository
	hashtagRepo repository.HashtagRepository
	replyRepo   repository.ReplyRepository
}

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

func (s *threadService) CreateThread(ctx context.Context, req *thread.CreateThreadRequest) (*model.Thread, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "User ID is required")
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid user ID: %v", err)
	}

	var communityID *uuid.UUID
	if req.CommunityId != nil && *req.CommunityId != "" {
		commID, err := uuid.Parse(*req.CommunityId)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "Invalid community ID: %v", err)
		}
		communityID = &commID
	}

	var scheduledAt *time.Time
	if req.ScheduledAt != nil {
		t := req.ScheduledAt.AsTime()
		scheduledAt = &t
	}

	var whoCanReply string = "Everyone"
	if req.WhoCanReply != nil && *req.WhoCanReply != "" {
		whoCanReply = *req.WhoCanReply
	}

	isAdvertisement := false
	if req.IsAdvertisement != nil {
		isAdvertisement = *req.IsAdvertisement
	}

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

	if err := s.threadRepo.CreateThread(thread); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to create thread: %v", err)
	}

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

	if len(req.Hashtags) > 0 {
		for _, hashtagText := range req.Hashtags {
			hashtag, err := s.hashtagRepo.FindHashtagByText(hashtagText)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
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

func (s *threadService) GetThreadsByUserID(ctx context.Context, userID string, page, limit int) ([]*model.Thread, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	if userID == "" {
		log.Printf("Empty userID provided, falling back to GetAllThreads")
		return s.GetAllThreads(ctx, page, limit)
	}

	log.Printf("Getting threads for user ID: %s, page: %d, limit: %d", userID, page, limit)

	// Validate UUID
	_, err := uuid.Parse(userID)
	if err != nil {
		log.Printf("Invalid UUID format for user ID %s: %v", userID, err)
		return nil, status.Errorf(codes.InvalidArgument, "Invalid user ID format: %v", err)
	}

	// Database query with panic recovery
	var threads []*model.Thread
	func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("PANIC in GetThreadsByUserID: %v", r)
				err = status.Errorf(codes.Internal, "Internal server error occurred")
			}
		}()
		threads, err = s.threadRepo.FindThreadsByUserID(userID, page, limit)
	}()

	if err != nil {
		log.Printf("Error getting threads for user %s: %v", userID, err)
		return nil, status.Errorf(codes.Internal, "Failed to retrieve threads: %v", err)
	}

	log.Printf("Successfully retrieved %d threads for user %s", len(threads), userID)
	return threads, nil
}

func (s *threadService) GetAllThreads(ctx context.Context, page, limit int) ([]*model.Thread, error) {
	threads, err := s.threadRepo.FindAllThreads(page, limit)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to retrieve threads: %v", err)
	}

	return threads, nil
}

func (s *threadService) GetTotalThreadCount(ctx context.Context) (int64, error) {
	return s.threadRepo.CountAllThreads()
}

func (s *threadService) UpdateThread(ctx context.Context, req *thread.UpdateThreadRequest) (*model.Thread, error) {
	if req.ThreadId == "" {
		return nil, status.Error(codes.InvalidArgument, "Thread ID is required")
	}

	thread, err := s.threadRepo.FindThreadByID(req.ThreadId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "Thread with ID %s not found", req.ThreadId)
		}
		return nil, status.Errorf(codes.Internal, "Failed to retrieve thread: %v", err)
	}

	updated := false
	if req.Content != "" {
		thread.Content = req.Content
		updated = true
	}

	if req.IsPinned != nil {
		thread.IsPinned = *req.IsPinned
		updated = true
	}

	if len(req.AddHashtags) > 0 {
		for _, hashtagText := range req.AddHashtags {
			hashtag, err := s.hashtagRepo.FindHashtagByText(hashtagText)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
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

	if len(req.RemoveHashtags) > 0 {
		for _, hashtagText := range req.RemoveHashtags {
			hashtag, err := s.hashtagRepo.FindHashtagByText(hashtagText)
			if err != nil {
				continue
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

func (s *threadService) DeleteThread(ctx context.Context, threadID, userID string) error {
	if threadID == "" || userID == "" {
		return status.Error(codes.InvalidArgument, "Thread ID and User ID are required")
	}

	thread, err := s.threadRepo.FindThreadByID(threadID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return status.Errorf(codes.NotFound, "Thread with ID %s not found", threadID)
		}
		return status.Errorf(codes.Internal, "Failed to retrieve thread: %v", err)
	}

	if thread.UserID.String() != userID {
		return status.Error(codes.PermissionDenied, "User is not authorized to delete this thread")
	}

	if err := s.threadRepo.DeleteThread(threadID); err != nil {
		return status.Errorf(codes.Internal, "Failed to delete thread: %v", err)
	}

	return nil
}

type BookmarkReplyRequest struct {
	ReplyId string
	UserId  string
}

type RemoveReplyBookmarkRequest struct {
	ReplyId string
	UserId  string
}

func (s *threadService) PinThread(ctx context.Context, threadID, userID string) error {
	if threadID == "" {
		return status.Error(codes.InvalidArgument, "Thread ID is required")
	}

	if userID == "" {
		return status.Error(codes.InvalidArgument, "User ID is required")
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "Invalid user ID: %v", err)
	}

	_, err = uuid.Parse(threadID)
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "Invalid thread ID: %v", err)
	}

	thread, err := s.threadRepo.FindThreadByID(threadID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return status.Errorf(codes.NotFound, "Thread with ID %s not found", threadID)
		}
		return status.Errorf(codes.Internal, "Failed to get thread: %v", err)
	}

	if thread.UserID != userUUID {
		return status.Error(codes.PermissionDenied, "Only the owner of the thread can pin it")
	}

	thread.IsPinned = true
	thread.UpdatedAt = time.Now()

	if err := s.threadRepo.UpdateThread(thread); err != nil {
		return status.Errorf(codes.Internal, "Failed to pin thread: %v", err)
	}

	return nil
}

func (s *threadService) UnpinThread(ctx context.Context, threadID, userID string) error {
	if threadID == "" {
		return status.Error(codes.InvalidArgument, "Thread ID is required")
	}

	if userID == "" {
		return status.Error(codes.InvalidArgument, "User ID is required")
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "Invalid user ID: %v", err)
	}

	_, err = uuid.Parse(threadID)
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "Invalid thread ID: %v", err)
	}

	thread, err := s.threadRepo.FindThreadByID(threadID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return status.Errorf(codes.NotFound, "Thread with ID %s not found", threadID)
		}
		return status.Errorf(codes.Internal, "Failed to get thread: %v", err)
	}

	if thread.UserID != userUUID {
		return status.Error(codes.PermissionDenied, "Only the owner of the thread can unpin it")
	}

	thread.IsPinned = false
	thread.UpdatedAt = time.Now()

	if err := s.threadRepo.UpdateThread(thread); err != nil {
		return status.Errorf(codes.Internal, "Failed to unpin thread: %v", err)
	}

	return nil
}

func (s *threadService) PinReply(ctx context.Context, replyID, userID string) error {
	if replyID == "" {
		return status.Error(codes.InvalidArgument, "Reply ID is required")
	}

	if userID == "" {
		return status.Error(codes.InvalidArgument, "User ID is required")
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "Invalid user ID: %v", err)
	}

	_, err = uuid.Parse(replyID)
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "Invalid reply ID: %v", err)
	}

	reply, err := s.replyRepo.FindReplyByID(replyID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return status.Errorf(codes.NotFound, "Reply with ID %s not found", replyID)
		}
		return status.Errorf(codes.Internal, "Failed to get reply: %v", err)
	}

	if reply.UserID != userUUID {
		return status.Error(codes.PermissionDenied, "Only the owner of the reply can pin it")
	}

	reply.IsPinned = true
	reply.UpdatedAt = time.Now()

	if err := s.replyRepo.UpdateReply(reply); err != nil {
		return status.Errorf(codes.Internal, "Failed to pin reply: %v", err)
	}

	return nil
}

func (s *threadService) UnpinReply(ctx context.Context, replyID, userID string) error {
	if replyID == "" {
		return status.Error(codes.InvalidArgument, "Reply ID is required")
	}

	if userID == "" {
		return status.Error(codes.InvalidArgument, "User ID is required")
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "Invalid user ID: %v", err)
	}

	_, err = uuid.Parse(replyID)
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "Invalid reply ID: %v", err)
	}

	reply, err := s.replyRepo.FindReplyByID(replyID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return status.Errorf(codes.NotFound, "Reply with ID %s not found", replyID)
		}
		return status.Errorf(codes.Internal, "Failed to get reply: %v", err)
	}

	if reply.UserID != userUUID {
		return status.Error(codes.PermissionDenied, "Only the owner of the reply can unpin it")
	}

	reply.IsPinned = false
	reply.UpdatedAt = time.Now()

	if err := s.replyRepo.UpdateReply(reply); err != nil {
		return status.Errorf(codes.Internal, "Failed to unpin reply: %v", err)
	}

	return nil
}

func (s *threadService) GetMediaByUserID(ctx context.Context, userID string, page, limit int) ([]*model.Media, error) {
	if userID == "" {
		return nil, status.Error(codes.InvalidArgument, "User ID is required")
	}

	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	media, err := s.mediaRepo.FindMediaByUserID(userID, page, limit)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to retrieve user media: %v", err)
	}

	return media, nil
}
