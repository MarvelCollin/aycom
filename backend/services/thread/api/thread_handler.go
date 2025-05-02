package handlers

import (
	"context"

	"aycom/backend/services/thread/db"
	"aycom/backend/services/thread/proto"
	"aycom/backend/services/thread/service"

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
func (h *ThreadHandler) DeleteThread(ctx context.Context, req *proto.DeleteThreadRequest) (*emptypb.Empty, error) {
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
func (h *ThreadHandler) DeleteReply(ctx context.Context, req *proto.DeleteReplyRequest) (*emptypb.Empty, error) {
	// Implementation will be added when ReplyService is created
	return nil, status.Error(codes.Unimplemented, "Method not implemented")
}

// LikeThread adds a like to a thread
func (h *ThreadHandler) LikeThread(ctx context.Context, req *proto.LikeThreadRequest) (*emptypb.Empty, error) {
	// Implementation will be added when InteractionService is created
	return nil, status.Error(codes.Unimplemented, "Method not implemented")
}

// UnlikeThread removes a like from a thread
func (h *ThreadHandler) UnlikeThread(ctx context.Context, req *proto.UnlikeThreadRequest) (*emptypb.Empty, error) {
	// Implementation will be added when InteractionService is created
	return nil, status.Error(codes.Unimplemented, "Method not implemented")
}

// LikeReply adds a like to a reply
func (h *ThreadHandler) LikeReply(ctx context.Context, req *proto.LikeReplyRequest) (*emptypb.Empty, error) {
	// Implementation will be added when InteractionService is created
	return nil, status.Error(codes.Unimplemented, "Method not implemented")
}

// UnlikeReply removes a like from a reply
func (h *ThreadHandler) UnlikeReply(ctx context.Context, req *proto.UnlikeReplyRequest) (*emptypb.Empty, error) {
	// Implementation will be added when InteractionService is created
	return nil, status.Error(codes.Unimplemented, "Method not implemented")
}

// RepostThread reposts a thread
func (h *ThreadHandler) RepostThread(ctx context.Context, req *proto.RepostThreadRequest) (*emptypb.Empty, error) {
	// Implementation will be added when InteractionService is created
	return nil, status.Error(codes.Unimplemented, "Method not implemented")
}

// RemoveRepost removes a repost
func (h *ThreadHandler) RemoveRepost(ctx context.Context, req *proto.RemoveRepostRequest) (*emptypb.Empty, error) {
	// Implementation will be added when InteractionService is created
	return nil, status.Error(codes.Unimplemented, "Method not implemented")
}

// BookmarkThread bookmarks a thread
func (h *ThreadHandler) BookmarkThread(ctx context.Context, req *proto.BookmarkThreadRequest) (*emptypb.Empty, error) {
	// Implementation will be added when InteractionService is created
	return nil, status.Error(codes.Unimplemented, "Method not implemented")
}

// RemoveBookmark removes a bookmark
func (h *ThreadHandler) RemoveBookmark(ctx context.Context, req *proto.RemoveBookmarkRequest) (*emptypb.Empty, error) {
	// Implementation will be added when InteractionService is created
	return nil, status.Error(codes.Unimplemented, "Method not implemented")
}

// CreatePoll creates a poll for a thread
func (h *ThreadHandler) CreatePoll(ctx context.Context, req *proto.CreatePollRequest) (*proto.PollResponse, error) {
	// Implementation will be added when PollService is created
	return nil, status.Error(codes.Unimplemented, "Method not implemented")
}

// VotePoll adds a vote to a poll option
func (h *ThreadHandler) VotePoll(ctx context.Context, req *proto.VotePollRequest) (*emptypb.Empty, error) {
	// Implementation will be added when PollService is created
	return nil, status.Error(codes.Unimplemented, "Method not implemented")
}

// GetPollResults gets the results of a poll
func (h *ThreadHandler) GetPollResults(ctx context.Context, req *proto.GetPollResultsRequest) (*proto.PollResultsResponse, error) {
	// Implementation will be added when PollService is created
	return nil, status.Error(codes.Unimplemented, "Method not implemented")
}

// Helper function to convert a Thread model to a ThreadResponse proto
func (h *ThreadHandler) convertThreadToResponse(ctx context.Context, thread *db.Thread) (*proto.ThreadResponse, error) {
	// Create a basic response with available data
	response := &proto.ThreadResponse{
		Id:        thread.ThreadID.String(),
		ThreadId:  thread.ThreadID.String(),
		UserId:    thread.UserID.String(),
		Content:   thread.Content,
		CreatedAt: timestamppb.New(thread.CreatedAt),
		UpdatedAt: timestamppb.New(thread.UpdatedAt),
	}

	// Set optional fields properly
	isPinned := thread.IsPinned
	response.IsPinned = &isPinned

	whoCanReply := thread.WhoCanReply
	response.WhoCanReply = &whoCanReply

	if thread.ScheduledAt != nil {
		response.ScheduledAt = timestamppb.New(*thread.ScheduledAt)
	}

	if thread.CommunityID != nil {
		communityID := thread.CommunityID.String()
		response.CommunityId = &communityID
	}

	isAd := thread.IsAdvertisement
	response.IsAdvertisement = &isAd

	// Get thread stats
	replyCount, likeCount, repostCount, err := h.threadService.GetThreadStats(ctx, thread.ThreadID.String())
	if err == nil {
		// Convert int32 to int64
		response.ReplyCount = int64(replyCount)
		response.LikeCount = int64(likeCount)
		response.RepostCount = int64(repostCount)
	}

	// Add media if available
	if len(thread.Media) > 0 {
		response.Media = make([]*proto.MediaResponse, 0, len(thread.Media))
		for _, media := range thread.Media {
			mediaResp := &proto.MediaResponse{
				Id:   media.MediaID.String(),
				Type: media.Type,
				Url:  media.URL,
			}
			response.Media = append(response.Media, mediaResp)
		}
	}

	return response, nil
}
