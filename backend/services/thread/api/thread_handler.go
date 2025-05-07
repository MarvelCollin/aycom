package handlers

import (
	"aycom/backend/proto/thread"
	"context"
	"log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"aycom/backend/services/thread/model"
	"aycom/backend/services/thread/repository"
	"aycom/backend/services/thread/service"

)

// ThreadHandler handles gRPC requests for the Thread service
type ThreadHandler struct {
	thread.UnimplementedThreadServiceServer
	threadService      service.ThreadService
	replyService       service.ReplyService
	interactionService service.InteractionService
	pollService        service.PollService
	interactionRepo    repository.InteractionRepository
	userClient         service.UserClient
	hashtagRepo        repository.HashtagRepository
}

// NewThreadHandler creates a new ThreadHandler instance
func NewThreadHandler(
	threadService service.ThreadService,
	replyService service.ReplyService,
	interactionService service.InteractionService,
	pollService service.PollService,
	interactionRepo repository.InteractionRepository,
	userClient service.UserClient,
	hashtagRepo repository.HashtagRepository,
) *ThreadHandler {
	return &ThreadHandler{
		threadService:      threadService,
		replyService:       replyService,
		interactionService: interactionService,
		pollService:        pollService,
		interactionRepo:    interactionRepo,
		userClient:         userClient,
		hashtagRepo:        hashtagRepo,
	}
}

// CreateThread creates a new thread
func (h *ThreadHandler) CreateThread(ctx context.Context, req *thread.CreateThreadRequest) (*thread.ThreadResponse, error) {
	t, err := h.threadService.CreateThread(ctx, req)
	if err != nil {
		return nil, err
	}

	// Convert thread to response
	response, err := h.convertThreadToResponse(ctx, t)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to convert thread to response: %v", err)
	}

	return response, nil
}

// GetThreadById retrieves a thread by its ID
func (h *ThreadHandler) GetThreadById(ctx context.Context, req *thread.GetThreadRequest) (*thread.ThreadResponse, error) {
	t, err := h.threadService.GetThreadByID(ctx, req.ThreadId)
	if err != nil {
		return nil, err
	}

	// Convert thread to response
	response, err := h.convertThreadToResponse(ctx, t)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to convert thread to response: %v", err)
	}

	return response, nil
}

// GetThreadsByUser retrieves threads by a user with pagination
func (h *ThreadHandler) GetThreadsByUser(ctx context.Context, req *thread.GetThreadsByUserRequest) (*thread.ThreadsResponse, error) {
	// Get threads
	threads, err := h.threadService.GetThreadsByUserID(ctx, req.UserId, int(req.Page), int(req.Limit))
	if err != nil {
		return nil, err
	}

	// Convert threads to response
	threadResponses := make([]*thread.ThreadResponse, 0, len(threads))
	for _, t := range threads {
		response, err := h.convertThreadToResponse(ctx, t)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Failed to convert thread to response: %v", err)
		}
		threadResponses = append(threadResponses, response)
	}

	return &thread.ThreadsResponse{
		Threads: threadResponses,
		Total:   int32(len(threads)),
	}, nil
}

// GetAllThreads retrieves all threads with pagination
func (h *ThreadHandler) GetAllThreads(ctx context.Context, req *thread.GetAllThreadsRequest) (*thread.ThreadsResponse, error) {
	// Get threads using the service layer method
	threads, err := h.threadService.GetAllThreads(ctx, int(req.Page), int(req.Limit))
	if err != nil {
		return nil, err
	}

	// Convert threads to response
	threadResponses := make([]*thread.ThreadResponse, 0, len(threads))
	for _, t := range threads {
		response, err := h.convertThreadToResponse(ctx, t)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Failed to convert thread to response: %v", err)
		}
		threadResponses = append(threadResponses, response)
	}

	return &thread.ThreadsResponse{
		Threads: threadResponses,
		Total:   int32(len(threads)),
	}, nil
}

// UpdateThread updates a thread
func (h *ThreadHandler) UpdateThread(ctx context.Context, req *thread.UpdateThreadRequest) (*thread.ThreadResponse, error) {
	t, err := h.threadService.UpdateThread(ctx, req)
	if err != nil {
		return nil, err
	}

	// Convert thread to response
	response, err := h.convertThreadToResponse(ctx, t)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to convert thread to response: %v", err)
	}

	return response, nil
}

// DeleteThread deletes a thread
func (h *ThreadHandler) DeleteThread(ctx context.Context, req *thread.DeleteThreadRequest) (*emptypb.Empty, error) {
	err := h.threadService.DeleteThread(ctx, req.ThreadId, req.UserId)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// CreateReply creates a new reply to a thread or another reply
func (h *ThreadHandler) CreateReply(ctx context.Context, req *thread.CreateReplyRequest) (*thread.ReplyResponse, error) {
	// Call the reply service to create a reply
	reply, err := h.replyService.CreateReply(ctx, req)
	if err != nil {
		return nil, err
	}

	// Convert reply to response
	response, err := h.convertReplyToResponse(ctx, reply)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to convert reply to response: %v", err)
	}

	return response, nil
}

// GetRepliesByThread retrieves replies to a thread with pagination
func (h *ThreadHandler) GetRepliesByThread(ctx context.Context, req *thread.GetRepliesByThreadRequest) (*thread.RepliesResponse, error) {
	// Get replies using the reply service
	replies, err := h.replyService.GetRepliesByThreadID(ctx, req.ThreadId, int(req.Page), int(req.Limit))
	if err != nil {
		return nil, err
	}

	// Convert replies to response format
	replyResponses := make([]*thread.ReplyResponse, 0, len(replies))
	for _, reply := range replies {
		response, err := h.convertReplyToResponse(ctx, reply)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Failed to convert reply to response: %v", err)
		}
		replyResponses = append(replyResponses, response)
	}

	// Calculate total count (could be optimized with a separate count query)
	totalCount := len(replies)

	return &thread.RepliesResponse{
		Replies: replyResponses,
		Total:   int32(totalCount),
	}, nil
}

// UpdateReply updates a reply
func (h *ThreadHandler) UpdateReply(ctx context.Context, req *thread.UpdateReplyRequest) (*thread.ReplyResponse, error) {
	// Call the reply service to update the reply
	reply, err := h.replyService.UpdateReply(ctx, req)
	if err != nil {
		return nil, err
	}

	// Convert reply to response
	response, err := h.convertReplyToResponse(ctx, reply)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to convert reply to response: %v", err)
	}

	return response, nil
}

// DeleteReply deletes a reply
func (h *ThreadHandler) DeleteReply(ctx context.Context, req *thread.DeleteReplyRequest) (*emptypb.Empty, error) {
	// Call the reply service to delete the reply
	err := h.replyService.DeleteReply(ctx, req.ReplyId, req.UserId)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// LikeThread adds a like to a thread
func (h *ThreadHandler) LikeThread(ctx context.Context, req *thread.LikeThreadRequest) (*emptypb.Empty, error) {
	// Call the interaction service to like the thread
	err := h.interactionService.LikeThread(ctx, req.UserId, req.ThreadId)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// UnlikeThread removes a like from a thread
func (h *ThreadHandler) UnlikeThread(ctx context.Context, req *thread.UnlikeThreadRequest) (*emptypb.Empty, error) {
	// Call the interaction service to unlike the thread
	err := h.interactionService.UnlikeThread(ctx, req.UserId, req.ThreadId)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// LikeReply adds a like to a reply
func (h *ThreadHandler) LikeReply(ctx context.Context, req *thread.LikeReplyRequest) (*emptypb.Empty, error) {
	// Call the interaction service to like the reply
	err := h.interactionService.LikeReply(ctx, req.UserId, req.ReplyId)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// UnlikeReply removes a like from a reply
func (h *ThreadHandler) UnlikeReply(ctx context.Context, req *thread.UnlikeReplyRequest) (*emptypb.Empty, error) {
	// Call the interaction service to unlike the reply
	err := h.interactionService.UnlikeReply(ctx, req.UserId, req.ReplyId)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// RepostThread reposts a thread
func (h *ThreadHandler) RepostThread(ctx context.Context, req *thread.RepostThreadRequest) (*emptypb.Empty, error) {
	// Call the interaction service to repost the thread
	var content *string
	if req.AddedContent != nil {
		content = req.AddedContent
	}

	err := h.interactionService.RepostThread(ctx, req.UserId, req.ThreadId, content)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// RemoveRepost removes a repost
func (h *ThreadHandler) RemoveRepost(ctx context.Context, req *thread.RemoveRepostRequest) (*emptypb.Empty, error) {
	// Call the interaction service to remove the repost
	err := h.interactionService.RemoveRepost(ctx, req.UserId, req.ThreadId)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// BookmarkThread bookmarks a thread
func (h *ThreadHandler) BookmarkThread(ctx context.Context, req *thread.BookmarkThreadRequest) (*emptypb.Empty, error) {
	// Call the interaction service to bookmark the thread
	err := h.interactionService.BookmarkThread(ctx, req.UserId, req.ThreadId)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// RemoveBookmark removes a bookmark
func (h *ThreadHandler) RemoveBookmark(ctx context.Context, req *thread.RemoveBookmarkRequest) (*emptypb.Empty, error) {
	// Call the interaction service to remove the bookmark from the thread
	err := h.interactionService.RemoveBookmark(ctx, req.UserId, req.ThreadId)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// CreatePoll creates a poll for a thread
func (h *ThreadHandler) CreatePoll(ctx context.Context, req *thread.CreatePollRequest) (*thread.PollResponse, error) {
	// Implementation will be added when PollService is created
	return nil, status.Error(codes.Unimplemented, "Method not implemented")
}

// VotePoll adds a vote to a poll option
func (h *ThreadHandler) VotePoll(ctx context.Context, req *thread.VotePollRequest) (*emptypb.Empty, error) {
	// Implementation will be added when PollService is created
	return nil, status.Error(codes.Unimplemented, "Method not implemented")
}

// GetPollResults gets the results of a poll
func (h *ThreadHandler) GetPollResults(ctx context.Context, req *thread.GetPollResultsRequest) (*thread.PollResultsResponse, error) {
	// Implementation will be added when PollService is created
	return nil, status.Error(codes.Unimplemented, "Method not implemented")
}

// GetTrendingHashtags returns a list of trending hashtags
func (h *ThreadHandler) GetTrendingHashtags(ctx context.Context, req *thread.GetTrendingHashtagsRequest) (*thread.GetTrendingHashtagsResponse, error) {
	// Get limit from request, default to 10 if not specified
	limit := int(req.Limit)
	if limit <= 0 {
		limit = 10 // Default limit
	}

	// Get trending hashtags from repository
	hashtags, err := h.hashtagRepo.GetTrendingHashtags(limit)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to retrieve trending hashtags: %v", err)
	}

	// Convert to response format
	hashtagResponses := make([]*thread.HashtagResponse, 0, len(hashtags))
	for _, hashtag := range hashtags {
		// Get thread count for this hashtag
		count, err := h.hashtagRepo.CountThreadsWithHashtag(hashtag.HashtagID.String())
		if err != nil {
			// Continue with count 0 rather than failing the whole request
			count = 0
		}

		hashtagResponses = append(hashtagResponses, &thread.HashtagResponse{
			Name:  hashtag.Text,
			Count: int64(count),
		})
	}

	return &thread.GetTrendingHashtagsResponse{
		Hashtags: hashtagResponses,
	}, nil
}

// PinThread implements the PinThread gRPC method
func (h *ThreadHandler) PinThread(ctx context.Context, req *thread.PinThreadRequest) (*emptypb.Empty, error) {
	if req.ThreadId == "" {
		return nil, status.Error(codes.InvalidArgument, "Thread ID is required")
	}

	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "User ID is required")
	}

	// Call the thread service to pin the thread
	err := h.threadService.PinThread(ctx, req.ThreadId, req.UserId)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// UnpinThread implements the UnpinThread gRPC method
func (h *ThreadHandler) UnpinThread(ctx context.Context, req *thread.UnpinThreadRequest) (*emptypb.Empty, error) {
	if req.ThreadId == "" {
		return nil, status.Error(codes.InvalidArgument, "Thread ID is required")
	}

	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "User ID is required")
	}

	// Call the thread service to unpin the thread
	err := h.threadService.UnpinThread(ctx, req.ThreadId, req.UserId)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// PinReply implements the PinReply gRPC method
func (h *ThreadHandler) PinReply(ctx context.Context, req *thread.PinReplyRequest) (*emptypb.Empty, error) {
	if req.ReplyId == "" {
		return nil, status.Error(codes.InvalidArgument, "Reply ID is required")
	}

	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "User ID is required")
	}

	// Call the thread service to pin the reply
	err := h.threadService.PinReply(ctx, req.ReplyId, req.UserId)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// UnpinReply implements the UnpinReply gRPC method
func (h *ThreadHandler) UnpinReply(ctx context.Context, req *thread.UnpinReplyRequest) (*emptypb.Empty, error) {
	if req.ReplyId == "" {
		return nil, status.Error(codes.InvalidArgument, "Reply ID is required")
	}

	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "User ID is required")
	}

	// Call the thread service to unpin the reply
	err := h.threadService.UnpinReply(ctx, req.ReplyId, req.UserId)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// Helper function to convert a Thread model to a ThreadResponse proto
func (h *ThreadHandler) convertThreadToResponse(ctx context.Context, threadModel *model.Thread) (*thread.ThreadResponse, error) {
	// Create the nested Thread structure
	protoThread := &thread.Thread{
		Id:        threadModel.ThreadID.String(),
		UserId:    threadModel.UserID.String(),
		Content:   threadModel.Content,
		CreatedAt: timestamppb.New(threadModel.CreatedAt),
		UpdatedAt: timestamppb.New(threadModel.UpdatedAt),
	}

	// Set optional fields
	if threadModel.IsPinned {
		protoThread.IsPinned = &threadModel.IsPinned
	}

	if threadModel.WhoCanReply != "" {
		protoThread.WhoCanReply = &threadModel.WhoCanReply
	}

	if threadModel.ScheduledAt != nil {
		protoThread.ScheduledAt = timestamppb.New(*threadModel.ScheduledAt)
	}

	if threadModel.CommunityID != nil {
		communityId := threadModel.CommunityID.String()
		protoThread.CommunityId = &communityId
	}

	if threadModel.IsAdvertisement {
		protoThread.IsAdvertisement = &threadModel.IsAdvertisement
	}

	// Create the full response with the thread and its stats
	response := &thread.ThreadResponse{
		Thread: protoThread,
	}

	// Get thread stats
	if h.interactionRepo != nil {
		threadID := threadModel.ThreadID.String()

		// Calculate like count
		likeCount, err := h.interactionRepo.CountThreadLikes(threadID)
		if err == nil {
			response.LikesCount = likeCount
		}

		// Calculate repost count
		repostCount, err := h.interactionRepo.CountThreadReposts(threadID)
		if err == nil {
			response.RepostsCount = repostCount
		}

		// Calculate bookmark count
		bookmarkCount, err := h.interactionRepo.CountThreadBookmarks(threadID)
		if err == nil {
			response.BookmarkCount = bookmarkCount
			// Also set ViewCount for backward compatibility
			protoThread.ViewCount = bookmarkCount
		}
	}

	// Fetch user information if available
	if h.userClient != nil {
		userInfo, err := h.userClient.GetUserById(ctx, threadModel.UserID.String())
		if err == nil && userInfo != nil {
			// Create User object for the thread
			response.User = &thread.User{
				Id:                userInfo.Id,
				Name:              userInfo.DisplayName,
				Username:          userInfo.Username,
				ProfilePictureUrl: userInfo.ProfilePictureUrl,
				IsVerified:        userInfo.IsVerified,
			}
		} else {
			log.Printf("Could not fetch user info for thread %s by user %s: %v",
				threadModel.ThreadID.String(), threadModel.UserID.String(), err)
		}
	}

	return response, nil
}

func (h *ThreadHandler) convertReplyToResponse(ctx context.Context, reply *model.Reply) (*thread.ReplyResponse, error) {
	protoReply := &thread.Reply{
		Id:        reply.ReplyID.String(),
		ThreadId:  reply.ThreadID.String(),
		UserId:    reply.UserID.String(),
		Content:   reply.Content,
		CreatedAt: timestamppb.New(reply.CreatedAt),
		UpdatedAt: timestamppb.New(reply.UpdatedAt),
	}

	if reply.ParentReplyID != nil {
		protoReply.ParentId = reply.ParentReplyID.String()
	}

	response := &thread.ReplyResponse{
		Reply: protoReply,
	}

	if h.interactionRepo != nil {
		replyID := reply.ReplyID.String()

		// Calculate like count
		likeCount, err := h.interactionRepo.CountReplyLikes(replyID)
		if err == nil {
			response.LikesCount = likeCount
		}

		// Calculate bookmark count - just log it for now as the field doesn't exist in proto
		bookmarkCount, err := h.interactionRepo.CountReplyBookmarks(replyID)
		if err == nil {
			log.Printf("Reply %s has %d bookmarks", replyID, bookmarkCount)
			// Can't set this as the field doesn't exist in the proto definition
			// response.BookmarkCount = bookmarkCount
		}

		// Get metadata from context if available
		md, ok := metadata.FromIncomingContext(ctx)
		if ok {
			// Check if user ID is in metadata
			userIDs := md.Get("user_id")
			if len(userIDs) > 0 {
				userID := userIDs[0]

				// Check if user has liked this reply
				hasLiked, err := h.interactionService.HasUserLikedReply(ctx, userID, replyID)
				if err == nil {
					response.LikedByUser = hasLiked
				}

				// Check if user has bookmarked this reply - just log it for now
				hasBookmarked, err := h.interactionService.HasUserBookmarkedReply(ctx, userID, replyID)
				if err == nil {
					log.Printf("Reply %s is bookmarked by user %s: %v", replyID, userID, hasBookmarked)
					// Can't set this as the field doesn't exist in the proto definition
					// response.BookmarkedByUser = hasBookmarked
				}
			}
		}
	}

	// Fetch user information if available
	if h.userClient != nil {
		userInfo, err := h.userClient.GetUserById(ctx, reply.UserID.String())
		if err == nil && userInfo != nil {
			log.Printf("User info retrieved for reply %s by user %s (username: %s)",
				reply.ReplyID.String(), userInfo.Id, userInfo.Username)

			// Create User object for the reply
			response.User = &thread.User{
				Id:                userInfo.Id,
				Name:              userInfo.DisplayName,
				Username:          userInfo.Username,
				ProfilePictureUrl: userInfo.ProfilePictureUrl,
				IsVerified:        userInfo.IsVerified,
			}
		} else {
			log.Printf("Could not fetch user info for reply %s by user %s: %v",
				reply.ReplyID.String(), reply.UserID.String(), err)
		}
	}

	return response, nil
}
