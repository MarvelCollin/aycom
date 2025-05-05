package handlers

import (
	"context"
	"log"

	"aycom/backend/proto/thread"
	"aycom/backend/services/thread/model"
	"aycom/backend/services/thread/repository"
	"aycom/backend/services/thread/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
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
		Threads:    threadResponses,
		TotalCount: int32(len(threads)),
		Page:       req.Page,
		Limit:      req.Limit,
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
		Threads:    threadResponses,
		TotalCount: int32(len(threads)),
		Page:       req.Page,
		Limit:      req.Limit,
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
		Replies:    replyResponses,
		TotalCount: int32(totalCount),
		Page:       req.Page,
		Limit:      req.Limit,
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
	err := h.interactionService.RepostThread(ctx, req.UserId, req.ThreadId, &req.Content)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// RemoveRepost removes a repost
func (h *ThreadHandler) RemoveRepost(ctx context.Context, req *thread.RemoveRepostRequest) (*emptypb.Empty, error) {
	// Call the interaction service to remove the repost
	err := h.interactionService.RemoveRepost(ctx, req.UserId, req.RepostId)
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
	// Call the interaction service to remove the bookmark
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

// GetTrendingHashtags returns the most popular hashtags
func (h *ThreadHandler) GetTrendingHashtags(ctx context.Context, req *thread.GetTrendingHashtagsRequest) (*thread.GetTrendingHashtagsResponse, error) {
	log.Printf("GetTrendingHashtags called with limit: %d", req.Limit)

	limit := int(req.Limit)
	if limit <= 0 {
		limit = 10 // Default limit
	}

	// Get trending hashtags from repository
	hashtags, err := h.hashtagRepo.GetTrendingHashtags(limit)
	if err != nil {
		log.Printf("Error getting trending hashtags: %v", err)
		return nil, status.Errorf(codes.Internal, "Failed to get trending hashtags: %v", err)
	}

	// Convert hashtags to HashtagResponse objects
	hashtagResponses := make([]*thread.HashtagResponse, 0, len(hashtags))
	for _, hashtag := range hashtags {
		// Get thread count for this hashtag
		count, err := h.hashtagRepo.CountThreadsWithHashtag(hashtag.HashtagID.String())
		if err != nil {
			log.Printf("Error counting threads for hashtag %s: %v", hashtag.Text, err)
			// Continue with count 0 rather than failing the whole request
			count = 0
		}

		// Create a hashtag response
		hashtagResponses = append(hashtagResponses, &thread.HashtagResponse{
			Id:          hashtag.HashtagID.String(),
			Text:        hashtag.Text,
			ThreadCount: int64(count),
		})
	}

	return &thread.GetTrendingHashtagsResponse{
		Hashtags: hashtagResponses,
	}, nil
}

// Helper function to convert a Thread model to a ThreadResponse proto
func (h *ThreadHandler) convertThreadToResponse(ctx context.Context, thread *model.Thread) (*thread.ThreadResponse, error) {
	// Create a basic response with available data
	response := &thread.ThreadResponse{
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

	// Fetch media for this thread - implement in appropriate repository if needed
	// This is left as a TODO since we don't have direct access to thread.Media
	// response.Media = ...

	// Fetch poll for this thread - implement in appropriate repository if needed
	// This is left as a TODO since we don't have direct access to thread.Poll
	// response.Poll = ...

	// Get thread stats
	if h.interactionRepo != nil {
		threadID := thread.ThreadID.String()

		// Calculate reply count - if available through replyRepo
		// We don't have CountRepliesByThreadID in ReplyService, so maybe use replyRepo directly
		// For now we'll leave it at 0
		// replyCount, err := h.replyService.CountRepliesByThreadID(ctx, threadID)
		// if err == nil {
		//     response.ReplyCount = replyCount
		// }

		// Calculate like count
		likeCount, err := h.interactionRepo.CountThreadLikes(threadID)
		if err == nil {
			response.LikeCount = likeCount
		}

		// Calculate repost count
		repostCount, err := h.interactionRepo.CountThreadReposts(threadID)
		if err == nil {
			response.RepostCount = repostCount
		}

		// Calculate bookmark count
		bookmarkCount, err := h.interactionRepo.CountThreadBookmarks(threadID)
		if err == nil {
			// Store in ViewCount since BookmarkCount might not exist in this proto version
			response.ViewCount = bookmarkCount

			// For newer proto versions that have BookmarkCount field:
			// Commented out until proto definitions are synced properly
			// response.BookmarkCount = bookmarkCount

			log.Printf("Thread %s has %d bookmarks (stored in ViewCount field)", threadID, bookmarkCount)
		}

		// Check if thread is liked by user
		// This would need the current user ID, which we don't have in this context
		// If needed, add a parameter for current user ID and implement this

		// Check if thread is bookmarked by user
		// This would need the current user ID, which we don't have in this context
		// If needed, add a parameter for current user ID and implement this
	}

	// Fetch user information if available
	if h.userClient != nil {
		userInfo, err := h.userClient.GetUserById(ctx, thread.UserID.String())
		if err == nil && userInfo != nil {
			// As a workaround, embed user info in the content field since proto doesn't have the User field
			log.Printf("User info retrieved for %s (username: %s) - embedding in content as workaround",
				userInfo.Id, userInfo.Username)

			// Format: [USER:username@displayName@profilePictureUrl]content
			// Include profile picture URL in the metadata for frontend to use
			response.Content = "[USER:" + userInfo.Username + "@" + userInfo.DisplayName + "@" + userInfo.ProfilePictureUrl + "]" + response.Content
		} else {
			log.Printf("Could not fetch user info for thread %s by user %s: %v",
				thread.ThreadID.String(), thread.UserID.String(), err)
		}
	} else {
		log.Printf("No user client available - thread %s by user %s will not have user details",
			thread.ThreadID.String(), thread.UserID.String())
	}

	return response, nil
}

// Helper function to convert a Reply model to a ReplyResponse proto
func (h *ThreadHandler) convertReplyToResponse(ctx context.Context, reply *model.Reply) (*thread.ReplyResponse, error) {
	// Create a basic response with available data
	response := &thread.ReplyResponse{
		Id:        reply.ReplyID.String(),
		ThreadId:  reply.ThreadID.String(),
		UserId:    reply.UserID.String(),
		Content:   reply.Content,
		CreatedAt: timestamppb.New(reply.CreatedAt),
		UpdatedAt: timestamppb.New(reply.UpdatedAt),
	}

	// Set parent ID if exists
	if reply.ParentReplyID != nil {
		response.ParentId = reply.ParentReplyID.String()
	}

	// Get reply stats if interaction repo is available
	if h.interactionRepo != nil {
		replyID := reply.ReplyID.String()

		// Calculate like count
		likeCount, err := h.interactionRepo.CountReplyLikes(replyID)
		if err == nil {
			response.LikeCount = likeCount
		}

		// Calculate reply count (for nested replies)
		replyCount, err := h.replyService.CountRepliesByParentID(ctx, replyID)
		if err == nil {
			response.ReplyCount = int64(replyCount)
		}
	}

	// Fetch user information if available
	if h.userClient != nil {
		userInfo, err := h.userClient.GetUserById(ctx, reply.UserID.String())
		if err == nil && userInfo != nil {
			log.Printf("User info retrieved for reply %s by user %s (username: %s)",
				reply.ReplyID.String(), userInfo.Id, userInfo.Username)

			// As a workaround, embed user info in the content field
			// Format: [USER:username@displayName@profilePictureUrl]content
			response.Content = "[USER:" + userInfo.Username + "@" + userInfo.DisplayName + "@" + userInfo.ProfilePictureUrl + "]" + response.Content
		} else {
			log.Printf("Could not fetch user info for reply %s by user %s: %v",
				reply.ReplyID.String(), reply.UserID.String(), err)
		}
	}

	return response, nil
}
