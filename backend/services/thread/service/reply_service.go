package service

import (
	"context"
	"errors"
	"time"

	"github.com/Acad600-Tpa/WEB-MV-242/backend/services/thread/model"
	"github.com/Acad600-Tpa/WEB-MV-242/backend/services/thread/proto/thread-service/proto"
	"github.com/Acad600-Tpa/WEB-MV-242/backend/services/thread/repository"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ReplyService defines the interface for reply operations
type ReplyService interface {
	CreateReply(ctx context.Context, req *proto.CreateReplyRequest) (*model.Reply, error)
	GetReplyByID(ctx context.Context, replyID string) (*model.Reply, error)
	GetRepliesByThreadID(ctx context.Context, threadID string, parentReplyID *string, page, limit int) ([]*model.Reply, int64, error)
	UpdateReply(ctx context.Context, req *proto.UpdateReplyRequest) (*model.Reply, error)
	DeleteReply(ctx context.Context, replyID, userID string) error
}

// replyService implements the ReplyService interface
type replyService struct {
	replyRepo   repository.ReplyRepository
	threadRepo  repository.ThreadRepository
	mediaRepo   repository.MediaRepository
	mentionRepo repository.MentionRepository
}

// NewReplyService creates a new reply service
func NewReplyService(
	replyRepo repository.ReplyRepository,
	threadRepo repository.ThreadRepository,
	mediaRepo repository.MediaRepository,
	mentionRepo repository.MentionRepository,
) ReplyService {
	return &replyService{
		replyRepo:   replyRepo,
		threadRepo:  threadRepo,
		mediaRepo:   mediaRepo,
		mentionRepo: mentionRepo,
	}
}

// CreateReply creates a new reply to a thread or another reply
func (s *replyService) CreateReply(ctx context.Context, req *proto.CreateReplyRequest) (*model.Reply, error) {
	// Validate required fields
	if req.ThreadId == "" || req.UserId == "" || req.Content == "" {
		return nil, status.Error(codes.InvalidArgument, "Thread ID, User ID, and content are required")
	}

	// Verify thread exists
	_, err := s.threadRepo.FindThreadByID(req.ThreadId)
	if err != nil {
		if errors.Is(err, repository.ErrThreadNotFound) {
			return nil, status.Errorf(codes.NotFound, "Thread with ID %s not found", req.ThreadId)
		}
		return nil, status.Errorf(codes.Internal, "Failed to retrieve thread: %v", err)
	}

	// Parse required IDs
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid user ID: %v", err)
	}

	threadID, err := uuid.Parse(req.ThreadId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid thread ID: %v", err)
	}

	// Parse parent reply ID if provided
	var parentReplyID *uuid.UUID
	if req.ParentReplyId != "" {
		// Verify parent reply exists
		_, err := s.replyRepo.FindReplyByID(req.ParentReplyId)
		if err != nil {
			return nil, status.Errorf(codes.NotFound, "Parent reply with ID %s not found", req.ParentReplyId)
		}

		parentID, err := uuid.Parse(req.ParentReplyId)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "Invalid parent reply ID: %v", err)
		}
		parentReplyID = &parentID
	}

	// Create reply
	reply := &model.Reply{
		ReplyID:       uuid.New(),
		ThreadID:      threadID,
		UserID:        userID,
		Content:       req.Content,
		ParentReplyID: parentReplyID,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Create reply in database
	if err := s.replyRepo.CreateReply(reply); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to create reply: %v", err)
	}

	// Process media attachments if any
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
				ReplyID:         &reply.ReplyID,
				CreatedAt:       time.Now(),
			}
			if err := s.mentionRepo.CreateMention(mention); err != nil {
				return nil, status.Errorf(codes.Internal, "Failed to create mention: %v", err)
			}
		}
	}

	return reply, nil
}

// GetReplyByID retrieves a reply by its ID
func (s *replyService) GetReplyByID(ctx context.Context, replyID string) (*model.Reply, error) {
	if replyID == "" {
		return nil, status.Error(codes.InvalidArgument, "Reply ID is required")
	}

	reply, err := s.replyRepo.FindReplyByID(replyID)
	if err != nil {
		if errors.Is(err, repository.ErrReplyNotFound) {
			return nil, status.Errorf(codes.NotFound, "Reply with ID %s not found", replyID)
		}
		return nil, status.Errorf(codes.Internal, "Failed to retrieve reply: %v", err)
	}

	return reply, nil
}

// GetRepliesByThreadID retrieves replies to a thread with pagination
func (s *replyService) GetRepliesByThreadID(ctx context.Context, threadID string, parentReplyID *string, page, limit int) ([]*model.Reply, int64, error) {
	if threadID == "" {
		return nil, 0, status.Error(codes.InvalidArgument, "Thread ID is required")
	}

	// Default pagination values if not provided
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	replies, totalCount, err := s.replyRepo.FindRepliesByThreadID(threadID, parentReplyID, page, limit)
	if err != nil {
		return nil, 0, status.Errorf(codes.Internal, "Failed to retrieve replies: %v", err)
	}

	return replies, totalCount, nil
}

// UpdateReply updates a reply
func (s *replyService) UpdateReply(ctx context.Context, req *proto.UpdateReplyRequest) (*model.Reply, error) {
	if req.ReplyId == "" || req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "Reply ID and User ID are required")
	}

	// Get existing reply
	reply, err := s.replyRepo.FindReplyByID(req.ReplyId)
	if err != nil {
		if errors.Is(err, repository.ErrReplyNotFound) {
			return nil, status.Errorf(codes.NotFound, "Reply with ID %s not found", req.ReplyId)
		}
		return nil, status.Errorf(codes.Internal, "Failed to retrieve reply: %v", err)
	}

	// Check if user is the owner
	if reply.UserID.String() != req.UserId {
		return nil, status.Error(codes.PermissionDenied, "User is not authorized to update this reply")
	}

	// Update fields if provided
	updated := false
	if req.Content != "" {
		reply.Content = req.Content
		updated = true
	}

	// Update isPinned status
	reply.IsPinned = req.IsPinned
	updated = true

	if updated {
		reply.UpdatedAt = time.Now()
		if err := s.replyRepo.UpdateReply(reply); err != nil {
			return nil, status.Errorf(codes.Internal, "Failed to update reply: %v", err)
		}
	}

	return reply, nil
}

// DeleteReply deletes a reply
func (s *replyService) DeleteReply(ctx context.Context, replyID, userID string) error {
	if replyID == "" || userID == "" {
		return status.Error(codes.InvalidArgument, "Reply ID and User ID are required")
	}

	// Get reply to check ownership
	reply, err := s.replyRepo.FindReplyByID(replyID)
	if err != nil {
		if errors.Is(err, repository.ErrReplyNotFound) {
			return status.Errorf(codes.NotFound, "Reply with ID %s not found", replyID)
		}
		return status.Errorf(codes.Internal, "Failed to retrieve reply: %v", err)
	}

	// Check if user is the owner
	if reply.UserID.String() != userID {
		return status.Error(codes.PermissionDenied, "User is not authorized to delete this reply")
	}

	// Delete reply from database
	if err := s.replyRepo.DeleteReply(replyID); err != nil {
		return status.Errorf(codes.Internal, "Failed to delete reply: %v", err)
	}

	return nil
}

// Add ErrReplyNotFound to repository package to properly handle not found errors
var ErrReplyNotFound = errors.New("reply not found")
