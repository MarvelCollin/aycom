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

type ReplyService interface {
	CreateReply(ctx context.Context, req *thread.CreateReplyRequest) (*model.Reply, error)
	GetReplyByID(ctx context.Context, replyID string) (*model.Reply, error)
	GetRepliesByThreadID(ctx context.Context, threadID string, page, limit int) ([]*model.Reply, error)
	UpdateReply(ctx context.Context, req *thread.UpdateReplyRequest) (*model.Reply, error)
	DeleteReply(ctx context.Context, replyID, userID string) error
	CountRepliesByParentID(ctx context.Context, parentID string) (int64, error)
	FindRepliesByParentID(ctx context.Context, parentReplyID string, page, limit int) ([]*model.Reply, error)
	GetRepliesByUserID(ctx context.Context, userID string, page, limit int) ([]*model.Reply, error)
}

// Add UserRelationService interface
type UserRelationService interface {
	CheckUserFollows(ctx context.Context, followerID, followedID string) (bool, error)
	CheckUserVerified(ctx context.Context, userID string) (bool, error)
}

type replyService struct {
	replyRepo       repository.ReplyRepository
	threadRepo      repository.ThreadRepository
	mediaRepo       repository.MediaRepository
	userRelationSvc UserRelationService // Add this field
}

func NewReplyService(
	replyRepo repository.ReplyRepository,
	threadRepo repository.ThreadRepository,
	mediaRepo repository.MediaRepository,
	userRelationSvc UserRelationService, // Add this parameter
) ReplyService {
	return &replyService{
		replyRepo:       replyRepo,
		threadRepo:      threadRepo,
		mediaRepo:       mediaRepo,
		userRelationSvc: userRelationSvc,
	}
}

func (s *replyService) CreateReply(ctx context.Context, req *thread.CreateReplyRequest) (*model.Reply, error) {
	if req.ThreadId == "" || req.UserId == "" || req.Content == "" {
		return nil, status.Error(codes.InvalidArgument, "Thread ID, User ID, and content are required")
	}

	// Get the thread to check reply permissions
	targetThread, err := s.threadRepo.FindThreadByID(req.ThreadId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "Thread with ID %s not found", req.ThreadId)
		}
		return nil, status.Errorf(codes.Internal, "Failed to retrieve thread: %v", err)
	}

	// Check who can reply restriction
	if targetThread.WhoCanReply != "Everyone" {
		threadOwnerID := targetThread.UserID.String()

		// If the user is the thread owner, they can always reply
		if req.UserId != threadOwnerID {
			// Only check restrictions if userRelationSvc is available
			if s.userRelationSvc != nil {
				switch targetThread.WhoCanReply {
				case "Accounts You Follow":
					// Check if the thread owner follows the user trying to reply
					follows, err := s.userRelationSvc.CheckUserFollows(ctx, threadOwnerID, req.UserId)
					if err != nil {
						return nil, status.Errorf(codes.Internal, "Failed to check follow relationship: %v", err)
					}
					if !follows {
						return nil, status.Errorf(codes.PermissionDenied, "Only accounts that the thread owner follows can reply to this thread")
					}
				case "Verified Accounts":
					// Check if user is verified
					verified, err := s.userRelationSvc.CheckUserVerified(ctx, req.UserId)
					if err != nil {
						return nil, status.Errorf(codes.Internal, "Failed to check verification status: %v", err)
					}
					if !verified {
						return nil, status.Errorf(codes.PermissionDenied, "Only verified accounts can reply to this thread")
					}
				default:
					// Unknown restriction, default to allowing the reply
				}
			} else {
				// Log warning but allow the reply if userRelationSvc is not available
				log.Printf("WARNING: Cannot check reply permissions for thread %s - userRelationSvc is nil", req.ThreadId)
			}
		}
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid user ID: %v", err)
	}

	threadID, err := uuid.Parse(req.ThreadId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid thread ID: %v", err)
	}

	var parentReplyID *uuid.UUID
	if req.ParentReplyId != nil && *req.ParentReplyId != "" {

		_, err := s.replyRepo.FindReplyByID(*req.ParentReplyId)
		if err != nil {
			return nil, status.Errorf(codes.NotFound, "Parent reply with ID %s not found", *req.ParentReplyId)
		}

		parentID, err := uuid.Parse(*req.ParentReplyId)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "Invalid parent reply ID: %v", err)
		}
		parentReplyID = &parentID
	}

	reply := &model.Reply{
		ReplyID:       uuid.New(),
		ThreadID:      threadID,
		UserID:        userID,
		Content:       req.Content,
		ParentReplyID: parentReplyID,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.replyRepo.CreateReply(reply); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to create reply: %v", err)
	}

	if len(req.Media) > 0 {
		for _, mediaInfo := range req.Media {
			media := &model.Media{
				MediaID:   uuid.New(),
				ReplyID:   &reply.ReplyID,
				Type:      mediaInfo.Type,
				URL:       mediaInfo.Url,
				CreatedAt: time.Now(),
			}

			if err := s.mediaRepo.CreateMedia(media); err != nil {
				return nil, status.Errorf(codes.Internal, "Failed to create media: %v", err)
			}
		}
	}

	return reply, nil
}

func (s *replyService) GetReplyByID(ctx context.Context, replyID string) (*model.Reply, error) {
	if replyID == "" {
		return nil, status.Error(codes.InvalidArgument, "Reply ID is required")
	}

	reply, err := s.replyRepo.FindReplyByID(replyID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "Reply with ID %s not found", replyID)
		}
		return nil, status.Errorf(codes.Internal, "Failed to retrieve reply: %v", err)
	}

	return reply, nil
}

func (s *replyService) GetRepliesByThreadID(ctx context.Context, threadID string, page, limit int) ([]*model.Reply, error) {
	if threadID == "" {
		return nil, status.Error(codes.InvalidArgument, "Thread ID is required")
	}

	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	replies, err := s.replyRepo.FindRepliesByThreadID(threadID, page, limit)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to retrieve replies: %v", err)
	}

	return replies, nil
}

func (s *replyService) UpdateReply(ctx context.Context, req *thread.UpdateReplyRequest) (*model.Reply, error) {
	if req.ReplyId == "" || req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "Reply ID and User ID are required")
	}

	reply, err := s.replyRepo.FindReplyByID(req.ReplyId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "Reply with ID %s not found", req.ReplyId)
		}
		return nil, status.Errorf(codes.Internal, "Failed to retrieve reply: %v", err)
	}

	if reply.UserID.String() != req.UserId {
		return nil, status.Error(codes.PermissionDenied, "User is not authorized to update this reply")
	}

	updated := false
	if req.Content != "" {
		reply.Content = req.Content
		updated = true
	}

	if req.IsPinned != nil {
		reply.IsPinned = *req.IsPinned
		updated = true
	}

	if updated {
		reply.UpdatedAt = time.Now()
		if err := s.replyRepo.UpdateReply(reply); err != nil {
			return nil, status.Errorf(codes.Internal, "Failed to update reply: %v", err)
		}
	}

	return reply, nil
}

func (s *replyService) DeleteReply(ctx context.Context, replyID, userID string) error {
	if replyID == "" || userID == "" {
		return status.Error(codes.InvalidArgument, "Reply ID and User ID are required")
	}

	reply, err := s.replyRepo.FindReplyByID(replyID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return status.Errorf(codes.NotFound, "Reply with ID %s not found", replyID)
		}
		return status.Errorf(codes.Internal, "Failed to retrieve reply: %v", err)
	}

	if reply.UserID.String() != userID {
		return status.Error(codes.PermissionDenied, "User is not authorized to delete this reply")
	}

	if err := s.replyRepo.DeleteReply(replyID); err != nil {
		return status.Errorf(codes.Internal, "Failed to delete reply: %v", err)
	}

	return nil
}

func (s *replyService) CountRepliesByParentID(ctx context.Context, parentID string) (int64, error) {
	if parentID == "" {
		return 0, status.Error(codes.InvalidArgument, "Parent ID is required")
	}

	count, err := s.replyRepo.CountRepliesByParentID(parentID)
	if err != nil {
		return 0, status.Errorf(codes.Internal, "Failed to count replies: %v", err)
	}

	return count, nil
}

func (s *replyService) FindRepliesByParentID(ctx context.Context, parentReplyID string, page, limit int) ([]*model.Reply, error) {
	if parentReplyID == "" {
		return nil, status.Error(codes.InvalidArgument, "Parent reply ID is required")
	}

	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	replies, err := s.replyRepo.FindRepliesByParentID(parentReplyID, page, limit)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to retrieve replies: %v", err)
	}

	return replies, nil
}

func (s *replyService) GetRepliesByUserID(ctx context.Context, userID string, page, limit int) ([]*model.Reply, error) {
	if userID == "" {
		return nil, status.Error(codes.InvalidArgument, "User ID is required")
	}

	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	replies, err := s.replyRepo.FindRepliesByUserID(userID, page, limit)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to retrieve replies: %v", err)
	}

	return replies, nil
}
