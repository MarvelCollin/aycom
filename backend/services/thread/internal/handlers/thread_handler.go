package handlers

import (
	"context"

	"github.com/Acad600-Tpa/WEB-MV-242/backend/services/thread/model"
	"github.com/Acad600-Tpa/WEB-MV-242/backend/services/thread/proto"
	"github.com/Acad600-Tpa/WEB-MV-242/backend/services/thread/service"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ThreadHandler implements the gRPC service for threads
type ThreadHandler struct {
	proto.UnimplementedThreadServiceServer
	threadService      service.ThreadService
	replyService       service.ReplyService
	interactionService service.InteractionService
	pollService        service.PollService
}

// NewThreadHandler creates a new thread handler
func NewThreadHandler(
	threadService service.ThreadService,
	replyService service.ReplyService,
	interactionService service.InteractionService,
	pollService service.PollService,
) *ThreadHandler {
	return &ThreadHandler{
		threadService:      threadService,
		replyService:       replyService,
		interactionService: interactionService,
		pollService:        pollService,
	}
}

// CreateThread creates a new thread
func (h *ThreadHandler) CreateThread(ctx context.Context, req *proto.CreateThreadRequest) (*proto.ThreadResponse, error) {
	thread, err := h.threadService.CreateThread(ctx, req)
	if err != nil {
		return nil, err
	}

	// Convert thread to response
	response, err := h.convertThreadToResponse(ctx, thread)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to convert thread to response: %v", err)
	}

	return response, nil
}

// GetThreadById retrieves a thread by its ID
func (h *ThreadHandler) GetThreadById(ctx context.Context, req *proto.GetThreadRequest) (*proto.ThreadResponse, error) {
	thread, err := h.threadService.GetThreadByID(ctx, req.ThreadId)
	if err != nil {
		return nil, err
	}

	// Increment view count
	go func() {
		_ = h.threadService.IncrementViewCount(context.Background(), req.ThreadId)
	}()

	// Convert thread to response
	response, err := h.convertThreadToResponse(ctx, thread)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to convert thread to response: %v", err)
	}

	return response, nil
}

// GetThreadsByUser retrieves threads by a user with pagination
func (h *ThreadHandler) GetThreadsByUser(ctx context.Context, req *proto.GetThreadsByUserRequest) (*proto.ThreadsResponse, error) {
	// Get threads
	threads, totalCount, err := h.threadService.GetThreadsByUserID(ctx, req.UserId, int(req.Page), int(req.Limit))
	if err != nil {
		return nil, err
	}

	// Convert threads to response
	threadResponses := make([]*proto.ThreadResponse, 0, len(threads))
	for _, thread := range threads {
		response, err := h.convertThreadToResponse(ctx, thread)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Failed to convert thread to response: %v", err)
		}
		threadResponses = append(threadResponses, response)
	}

	return &proto.ThreadsResponse{
		Threads:    threadResponses,
		TotalCount: int32(totalCount),
		Page:       req.Page,
		Limit:      req.Limit,
	}, nil
}

// UpdateThread updates a thread
func (h *ThreadHandler) UpdateThread(ctx context.Context, req *proto.UpdateThreadRequest) (*proto.ThreadResponse, error) {
	thread, err := h.threadService.UpdateThread(ctx, req)
	if err != nil {
		return nil, err
	}

	// Convert thread to response
	response, err := h.convertThreadToResponse(ctx, thread)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to convert thread to response: %v", err)
	}

	return response, nil
}

// DeleteThread deletes a thread
func (h *ThreadHandler) DeleteThread(ctx context.Context, req *proto.DeleteThreadRequest) (*empty.Empty, error) {
	err := h.threadService.DeleteThread(ctx, req.ThreadId, req.UserId)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// CreateReply creates a new reply to a thread or another reply
func (h *ThreadHandler) CreateReply(ctx context.Context, req *proto.CreateReplyRequest) (*proto.ReplyResponse, error) {
	// Implementation will be added when ReplyService is created
	return nil, status.Error(codes.Unimplemented, "Method not implemented")
}

// GetRepliesByThread retrieves replies to a thread with pagination
func (h *ThreadHandler) GetRepliesByThread(ctx context.Context, req *proto.GetRepliesByThreadRequest) (*proto.RepliesResponse, error) {
	// Implementation will be added when ReplyService is created
	return nil, status.Error(codes.Unimplemented, "Method not implemented")
}

// UpdateReply updates a reply
func (h *ThreadHandler) UpdateReply(ctx context.Context, req *proto.UpdateReplyRequest) (*proto.ReplyResponse, error) {
	// Implementation will be added when ReplyService is created
	return nil, status.Error(codes.Unimplemented, "Method not implemented")
}

// DeleteReply deletes a reply
func (h *ThreadHandler) DeleteReply(ctx context.Context, req *proto.DeleteReplyRequest) (*empty.Empty, error) {
	// Implementation will be added when ReplyService is created
	return nil, status.Error(codes.Unimplemented, "Method not implemented")
}

// LikeThread adds a like to a thread
func (h *ThreadHandler) LikeThread(ctx context.Context, req *proto.LikeThreadRequest) (*empty.Empty, error) {
	// Implementation will be added when InteractionService is created
	return nil, status.Error(codes.Unimplemented, "Method not implemented")
}

// UnlikeThread removes a like from a thread
func (h *ThreadHandler) UnlikeThread(ctx context.Context, req *proto.UnlikeThreadRequest) (*empty.Empty, error) {
	// Implementation will be added when InteractionService is created
	return nil, status.Error(codes.Unimplemented, "Method not implemented")
}

// LikeReply adds a like to a reply
func (h *ThreadHandler) LikeReply(ctx context.Context, req *proto.LikeReplyRequest) (*empty.Empty, error) {
	// Implementation will be added when InteractionService is created
	return nil, status.Error(codes.Unimplemented, "Method not implemented")
}

// UnlikeReply removes a like from a reply
func (h *ThreadHandler) UnlikeReply(ctx context.Context, req *proto.UnlikeReplyRequest) (*empty.Empty, error) {
	// Implementation will be added when InteractionService is created
	return nil, status.Error(codes.Unimplemented, "Method not implemented")
}

// RepostThread reposts a thread
func (h *ThreadHandler) RepostThread(ctx context.Context, req *proto.RepostThreadRequest) (*empty.Empty, error) {
	// Implementation will be added when InteractionService is created
	return nil, status.Error(codes.Unimplemented, "Method not implemented")
}

// RemoveRepost removes a repost
func (h *ThreadHandler) RemoveRepost(ctx context.Context, req *proto.RemoveRepostRequest) (*empty.Empty, error) {
	// Implementation will be added when InteractionService is created
	return nil, status.Error(codes.Unimplemented, "Method not implemented")
}

// BookmarkThread bookmarks a thread
func (h *ThreadHandler) BookmarkThread(ctx context.Context, req *proto.BookmarkThreadRequest) (*empty.Empty, error) {
	// Implementation will be added when InteractionService is created
	return nil, status.Error(codes.Unimplemented, "Method not implemented")
}

// RemoveBookmark removes a bookmark
func (h *ThreadHandler) RemoveBookmark(ctx context.Context, req *proto.RemoveBookmarkRequest) (*empty.Empty, error) {
	// Implementation will be added when InteractionService is created
	return nil, status.Error(codes.Unimplemented, "Method not implemented")
}

// CreatePoll creates a poll for a thread
func (h *ThreadHandler) CreatePoll(ctx context.Context, req *proto.CreatePollRequest) (*proto.PollResponse, error) {
	// Implementation will be added when PollService is created
	return nil, status.Error(codes.Unimplemented, "Method not implemented")
}

// VotePoll adds a vote to a poll option
func (h *ThreadHandler) VotePoll(ctx context.Context, req *proto.VotePollRequest) (*empty.Empty, error) {
	// Implementation will be added when PollService is created
	return nil, status.Error(codes.Unimplemented, "Method not implemented")
}

// GetPollResults gets the results of a poll
func (h *ThreadHandler) GetPollResults(ctx context.Context, req *proto.GetPollResultsRequest) (*proto.PollResultsResponse, error) {
	// Implementation will be added when PollService is created
	return nil, status.Error(codes.Unimplemented, "Method not implemented")
}

// Helper function to convert a Thread model to a ThreadResponse proto
func (h *ThreadHandler) convertThreadToResponse(ctx context.Context, thread *model.Thread) (*proto.ThreadResponse, error) {
	// Create a basic response with available data
	response := &proto.ThreadResponse{
		ThreadId:    thread.ThreadID.String(),
		UserId:      thread.UserID.String(),
		Content:     thread.Content,
		IsPinned:    thread.IsPinned,
		WhoCanReply: thread.WhoCanReply,
		CreatedAt:   timestamppb.New(thread.CreatedAt),
		UpdatedAt:   timestamppb.New(thread.UpdatedAt),
	}

	if thread.ScheduledAt != nil {
		response.ScheduledAt = timestamppb.New(*thread.ScheduledAt)
	}

	if thread.CommunityID != nil {
		response.CommunityId = thread.CommunityID.String()
	}

	response.IsAdvertisement = thread.IsAdvertisement

	// Get thread stats
	replyCount, likeCount, repostCount, err := h.threadService.GetThreadStats(ctx, thread.ThreadID.String())
	if err == nil {
		response.ReplyCount = int32(replyCount)
		response.LikeCount = int32(likeCount)
		response.RepostCount = int32(repostCount)
	}

	// Add media if available
	if len(thread.Media) > 0 {
		response.Media = make([]*proto.MediaResponse, 0, len(thread.Media))
		for _, media := range thread.Media {
			mediaResp := &proto.MediaResponse{
				MediaId: media.MediaID.String(),
				Type:    media.Type,
				Url:     media.URL,
			}
			response.Media = append(response.Media, mediaResp)
		}
	}

	return response, nil
}

// Helper function to get thread stats (counts of replies, likes, reposts)
func (h *ThreadHandler) getThreadStats(ctx context.Context, threadID string) (replyCount, likeCount, repostCount int64, err error) {
	// This would be implemented to get statistics such as like counts, repost counts, etc.
	// For now, we return placeholders
	return 0, 0, 0, nil
}
