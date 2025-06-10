package handlers

import (
	"aycom/backend/proto/thread"
	"aycom/backend/services/thread/model"
	"aycom/backend/services/thread/repository"
	"aycom/backend/services/thread/service"
	"context"
	"log"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ThreadHandler struct {
	thread.UnimplementedThreadServiceServer
	threadService      service.ThreadService
	replyService       service.ReplyService
	interactionService service.InteractionService
	pollService        service.PollService
	interactionRepo    repository.InteractionRepository
	userClient         service.UserClient
	hashtagRepo        repository.HashtagRepository
	threadRepo         repository.ThreadRepository
	mediaRepo          repository.MediaRepository
}

func NewThreadHandler(
	threadService service.ThreadService,
	replyService service.ReplyService,
	interactionService service.InteractionService,
	pollService service.PollService,
	interactionRepo repository.InteractionRepository,
	userClient service.UserClient,
	hashtagRepo repository.HashtagRepository,
	threadRepo repository.ThreadRepository,
	mediaRepo repository.MediaRepository,
) *ThreadHandler {
	// TODO: The UserRelationService should be created in main.go and injected into ReplyService
	// For proper implementation of reply permissions based on who_can_reply setting

	return &ThreadHandler{
		threadService:      threadService,
		replyService:       replyService,
		interactionService: interactionService,
		pollService:        pollService,
		interactionRepo:    interactionRepo,
		userClient:         userClient,
		hashtagRepo:        hashtagRepo,
		threadRepo:         threadRepo,
		mediaRepo:          mediaRepo,
	}
}

func (h *ThreadHandler) CreateThread(ctx context.Context, req *thread.CreateThreadRequest) (*thread.ThreadResponse, error) {
	t, err := h.threadService.CreateThread(ctx, req)
	if err != nil {
		return nil, err
	}

	response, err := h.convertThreadToResponse(ctx, t)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to convert thread to response: %v", err)
	}

	return response, nil
}

func (h *ThreadHandler) GetThreadById(ctx context.Context, req *thread.GetThreadRequest) (*thread.ThreadResponse, error) {
	threadID := req.ThreadId

	threadModel, err := h.threadService.GetThreadByID(ctx, threadID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to get thread: %v", err)
	}

	if threadModel == nil {
		return nil, status.Error(codes.NotFound, "Thread not found")
	}

	response, err := h.convertThreadToResponse(ctx, threadModel)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to convert thread: %v", err)
	}

	var userID string
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		userIDs := md.Get("user_id")
		if len(userIDs) > 0 {
			userID = userIDs[0]
			log.Printf("User ID from context metadata: %s", userID)

			if userID != "" {
				hasBookmarked, err := h.interactionService.HasUserBookmarked(ctx, userID, threadID)
				if err != nil {
					log.Printf("WARNING: Error checking bookmark status in GetThreadById: %v", err)
				} else {
					response.BookmarkedByUser = hasBookmarked
					log.Printf("Thread %s bookmark status for user %s: %v", threadID, userID, hasBookmarked)
				}
			}
		}
	}

	return response, nil
}

func (h *ThreadHandler) GetThreadsByUser(ctx context.Context, req *thread.GetThreadsByUserRequest) (*thread.ThreadsResponse, error) {
	var finalError error

	// Add panic recovery
	defer func() {
		if r := recover(); r != nil {
			log.Printf("PANIC RECOVERED in GetThreadsByUser: %v", r)
			finalError = status.Error(codes.Internal, "Internal server error occurred")
		}
	}()

	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Request cannot be nil")
	}

	if req.UserId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "User ID is required")
	}

	log.Printf("GetThreadsByUser: Fetching threads for user %s, page %d, limit %d", req.UserId, req.Page, req.Limit)

	// Extract the requesting user ID from context metadata if available
	var requestingUserID string
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		userIDs := md.Get("user_id")
		if len(userIDs) > 0 {
			requestingUserID = userIDs[0]
			log.Printf("GetThreadsByUser: Request from authenticated user: %s", requestingUserID)
		}
	}

	// Setup default pagination values if not provided
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 20
	}

	threads, err := h.threadService.GetThreadsByUserID(ctx, req.UserId, int(req.Page), int(req.Limit))
	if err != nil {
		log.Printf("GetThreadsByUser: Error fetching threads: %v", err)
		// Return empty successful response instead of error for "not found" cases
		if status.Code(err) == codes.NotFound {
			log.Printf("GetThreadsByUser: No threads found for user %s", req.UserId)
			return &thread.ThreadsResponse{
				Threads: []*thread.ThreadResponse{},
				Total:   0,
			}, nil
		}
		return nil, err
	}

	log.Printf("GetThreadsByUser: Found %d threads for user %s", len(threads), req.UserId)

	// If no threads found, return empty response instead of error
	if len(threads) == 0 {
		log.Printf("GetThreadsByUser: No threads found for user %s", req.UserId)
		return &thread.ThreadsResponse{
			Threads: []*thread.ThreadResponse{},
			Total:   0,
		}, nil
	}

	threadResponses := make([]*thread.ThreadResponse, 0, len(threads))
	for _, t := range threads {
		// Skip any null threads
		if t == nil {
			continue
		}

		// Wrap conversion in try-catch to prevent individual thread errors from breaking the entire response
		func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("PANIC in thread conversion for thread %v: %v", t.ThreadID, r)
				}
			}()

			response, err := h.convertThreadToResponse(ctx, t)
			if err != nil {
				log.Printf("GetThreadsByUser: Error converting thread %s: %v", t.ThreadID.String(), err)
				// Continue processing other threads instead of failing completely
				return
			}

			// Set bookmarked status if we have a requesting user ID
			if requestingUserID != "" && h.interactionService != nil {
				// Use a separate context with timeout for this operation
				interactionCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
				defer cancel()

				// Check if user bookmarked the thread
				hasBookmarked, err := h.interactionService.HasUserBookmarked(interactionCtx, requestingUserID, t.ThreadID.String())
				if err != nil {
					log.Printf("WARNING: Failed to check if user %s bookmarked thread %s: %v",
						requestingUserID, t.ThreadID.String(), err)
				} else {
					response.BookmarkedByUser = hasBookmarked
					log.Printf("Thread %s bookmark status for user %s: %v",
						t.ThreadID.String(), requestingUserID, hasBookmarked)
				}
			}

			threadResponses = append(threadResponses, response)
		}()
	}

	return &thread.ThreadsResponse{
		Threads: threadResponses,
		Total:   int32(len(threadResponses)),
	}, finalError
}

func (h *ThreadHandler) GetAllThreads(ctx context.Context, req *thread.GetAllThreadsRequest) (*thread.ThreadsResponse, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 || req.Limit > 100 {
		req.Limit = 20
	}

	threads, err := h.threadService.GetAllThreads(ctx, int(req.Page), int(req.Limit))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to retrieve threads: %v", err)
	}

	// Get total count for pagination
	totalCount, err := h.threadService.GetTotalThreadCount(ctx)
	if err != nil {
		// Log the error but continue with the threads we have
		log.Printf("WARNING: Failed to get total thread count: %v", err)
		totalCount = int64(len(threads))
	}

	threadResponses := make([]*thread.ThreadResponse, 0, len(threads))
	for _, t := range threads {
		response, err := h.convertThreadToResponse(ctx, t)
		if err != nil {
			// Log the error but continue with other threads
			log.Printf("Error converting thread %s: %v", t.ThreadID.String(), err)
			continue
		}
		threadResponses = append(threadResponses, response)
	}

	return &thread.ThreadsResponse{
		Threads: threadResponses,
		Total:   int32(totalCount),
	}, nil
}

func (h *ThreadHandler) UpdateThread(ctx context.Context, req *thread.UpdateThreadRequest) (*thread.ThreadResponse, error) {
	t, err := h.threadService.UpdateThread(ctx, req)
	if err != nil {
		return nil, err
	}

	response, err := h.convertThreadToResponse(ctx, t)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to convert thread to response: %v", err)
	}

	return response, nil
}

func (h *ThreadHandler) DeleteThread(ctx context.Context, req *thread.DeleteThreadRequest) (*emptypb.Empty, error) {
	err := h.threadService.DeleteThread(ctx, req.ThreadId, req.UserId)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (h *ThreadHandler) CreateReply(ctx context.Context, req *thread.CreateReplyRequest) (*thread.ReplyResponse, error) {

	reply, err := h.replyService.CreateReply(ctx, req)
	if err != nil {
		return nil, err
	}

	response, err := h.convertReplyToResponse(ctx, reply)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to convert reply to response: %v", err)
	}

	return response, nil
}

func (h *ThreadHandler) GetRepliesByThread(ctx context.Context, req *thread.GetRepliesByThreadRequest) (*thread.RepliesResponse, error) {

	replies, err := h.replyService.GetRepliesByThreadID(ctx, req.ThreadId, int(req.Page), int(req.Limit))
	if err != nil {
		return nil, err
	}

	replyResponses := make([]*thread.ReplyResponse, 0, len(replies))
	for _, reply := range replies {
		response, err := h.convertReplyToResponse(ctx, reply)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Failed to convert reply to response: %v", err)
		}
		replyResponses = append(replyResponses, response)
	}

	totalCount := len(replies)

	return &thread.RepliesResponse{
		Replies: replyResponses,
		Total:   int32(totalCount),
	}, nil
}

func (h *ThreadHandler) UpdateReply(ctx context.Context, req *thread.UpdateReplyRequest) (*thread.ReplyResponse, error) {

	reply, err := h.replyService.UpdateReply(ctx, req)
	if err != nil {
		return nil, err
	}

	response, err := h.convertReplyToResponse(ctx, reply)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to convert reply to response: %v", err)
	}

	return response, nil
}

func (h *ThreadHandler) DeleteReply(ctx context.Context, req *thread.DeleteReplyRequest) (*emptypb.Empty, error) {

	err := h.replyService.DeleteReply(ctx, req.ReplyId, req.UserId)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (h *ThreadHandler) LikeThread(ctx context.Context, req *thread.LikeThreadRequest) (*emptypb.Empty, error) {

	err := h.interactionService.LikeThread(ctx, req.UserId, req.ThreadId)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (h *ThreadHandler) UnlikeThread(ctx context.Context, req *thread.UnlikeThreadRequest) (*emptypb.Empty, error) {

	err := h.interactionService.UnlikeThread(ctx, req.UserId, req.ThreadId)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (h *ThreadHandler) LikeReply(ctx context.Context, req *thread.LikeReplyRequest) (*emptypb.Empty, error) {

	err := h.interactionService.LikeReply(ctx, req.UserId, req.ReplyId)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (h *ThreadHandler) UnlikeReply(ctx context.Context, req *thread.UnlikeReplyRequest) (*emptypb.Empty, error) {

	err := h.interactionService.UnlikeReply(ctx, req.UserId, req.ReplyId)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (h *ThreadHandler) RepostThread(ctx context.Context, req *thread.RepostThreadRequest) (*emptypb.Empty, error) {

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

func (h *ThreadHandler) RemoveRepost(ctx context.Context, req *thread.RemoveRepostRequest) (*emptypb.Empty, error) {

	err := h.interactionService.RemoveRepost(ctx, req.UserId, req.ThreadId)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (h *ThreadHandler) BookmarkThread(ctx context.Context, req *thread.BookmarkThreadRequest) (*emptypb.Empty, error) {
	log.Printf("API Handler: BookmarkThread called with userID=%s, threadID=%s", req.UserId, req.ThreadId)

	if req.UserId == "" {
		log.Printf("ERROR: BookmarkThread - Missing userID")
		return nil, status.Error(codes.InvalidArgument, "User ID is required")
	}

	if req.ThreadId == "" {
		log.Printf("ERROR: BookmarkThread - Missing threadID")
		return nil, status.Error(codes.InvalidArgument, "Thread ID is required")
	}

	if md, ok := metadata.FromIncomingContext(ctx); ok {
		for key, values := range md {
			log.Printf("Context metadata: %s=%v", key, values)
		}
	}

	err := h.interactionService.BookmarkThread(ctx, req.UserId, req.ThreadId)
	if err != nil {
		log.Printf("ERROR: Failed to bookmark thread: %v", err)
		return nil, err
	}

	isBookmarked, err := h.interactionService.HasUserBookmarked(ctx, req.UserId, req.ThreadId)
	if err != nil {
		log.Printf("WARNING: Error verifying bookmark status: %v", err)
	} else if !isBookmarked {
		log.Printf("WARNING: Bookmark verification failed - bookmark not found after creation")
	} else {
		log.Printf("Bookmark verification successful - bookmark found")
	}

	log.Printf("Successfully bookmarked thread %s for user %s at API handler level", req.ThreadId, req.UserId)
	return &emptypb.Empty{}, nil
}

func (h *ThreadHandler) RemoveBookmark(ctx context.Context, req *thread.RemoveBookmarkRequest) (*emptypb.Empty, error) {

	err := h.interactionService.RemoveBookmark(ctx, req.UserId, req.ThreadId)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (h *ThreadHandler) CreatePoll(ctx context.Context, req *thread.CreatePollRequest) (*thread.PollResponse, error) {

	return nil, status.Error(codes.Unimplemented, "Method not implemented")
}

func (h *ThreadHandler) VotePoll(ctx context.Context, req *thread.VotePollRequest) (*emptypb.Empty, error) {
	log.Printf("VotePoll called for poll ID: %s, option ID: %s, user ID: %s",
		req.PollId, req.OptionId, req.UserId)

	if err := h.pollService.AddVoteToPoll(ctx, req.PollId, req.OptionId, req.UserId); err != nil {
		log.Printf("Error voting in poll: %v", err)
		return nil, status.Errorf(codes.Internal, "Failed to vote in poll: %v", err)
	}

	return &emptypb.Empty{}, nil
}

func (h *ThreadHandler) GetPollResults(ctx context.Context, req *thread.GetPollResultsRequest) (*thread.PollResultsResponse, error) {

	return nil, status.Error(codes.Unimplemented, "Method not implemented")
}

func (h *ThreadHandler) GetTrendingHashtags(ctx context.Context, req *thread.GetTrendingHashtagsRequest) (*thread.GetTrendingHashtagsResponse, error) {

	limit := int(req.Limit)
	if limit <= 0 {
		limit = 10
	}

	hashtags, err := h.hashtagRepo.GetTrendingHashtags(limit)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to retrieve trending hashtags: %v", err)
	}

	hashtagResponses := make([]*thread.HashtagResponse, 0, len(hashtags))
	for _, hashtag := range hashtags {

		count, err := h.hashtagRepo.CountThreadsWithHashtag(hashtag.HashtagID.String())
		if err != nil {

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

func (h *ThreadHandler) PinThread(ctx context.Context, req *thread.PinThreadRequest) (*emptypb.Empty, error) {
	if req.ThreadId == "" {
		return nil, status.Error(codes.InvalidArgument, "Thread ID is required")
	}

	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "User ID is required")
	}

	err := h.threadService.PinThread(ctx, req.ThreadId, req.UserId)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (h *ThreadHandler) UnpinThread(ctx context.Context, req *thread.UnpinThreadRequest) (*emptypb.Empty, error) {
	if req.ThreadId == "" {
		return nil, status.Error(codes.InvalidArgument, "Thread ID is required")
	}

	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "User ID is required")
	}

	err := h.threadService.UnpinThread(ctx, req.ThreadId, req.UserId)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (h *ThreadHandler) PinReply(ctx context.Context, req *thread.PinReplyRequest) (*emptypb.Empty, error) {
	if req.ReplyId == "" {
		return nil, status.Error(codes.InvalidArgument, "Reply ID is required")
	}

	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "User ID is required")
	}

	err := h.threadService.PinReply(ctx, req.ReplyId, req.UserId)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (h *ThreadHandler) UnpinReply(ctx context.Context, req *thread.UnpinReplyRequest) (*emptypb.Empty, error) {
	if req.ReplyId == "" {
		return nil, status.Error(codes.InvalidArgument, "Reply ID is required")
	}

	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "User ID is required")
	}

	err := h.threadService.UnpinReply(ctx, req.ReplyId, req.UserId)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (h *ThreadHandler) BookmarkReply(ctx context.Context, req *thread.BookmarkReplyRequest) (*emptypb.Empty, error) {
	if req.ReplyId == "" {
		return nil, status.Error(codes.InvalidArgument, "Reply ID is required")
	}

	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "User ID is required")
	}

	err := h.interactionService.BookmarkReply(ctx, req.UserId, req.ReplyId)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (h *ThreadHandler) RemoveReplyBookmark(ctx context.Context, req *thread.RemoveReplyBookmarkRequest) (*emptypb.Empty, error) {
	if req.ReplyId == "" {
		return nil, status.Error(codes.InvalidArgument, "Reply ID is required")
	}

	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "User ID is required")
	}

	err := h.interactionService.RemoveReplyBookmark(ctx, req.UserId, req.ReplyId)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (h *ThreadHandler) GetRepliesByParentReply(ctx context.Context, req *thread.GetRepliesByParentReplyRequest) (*thread.RepliesResponse, error) {
	if req.ParentReplyId == "" {
		return nil, status.Error(codes.InvalidArgument, "Parent reply ID is required")
	}

	page := int(req.Page)
	if page <= 0 {
		page = 1
	}

	limit := int(req.Limit)
	if limit <= 0 {
		limit = 10
	}

	replies, err := h.replyService.FindRepliesByParentID(ctx, req.ParentReplyId, page, limit)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to retrieve replies: %v", err)
	}

	replyResponses := make([]*thread.ReplyResponse, 0, len(replies))
	for _, reply := range replies {
		response, err := h.convertReplyToResponse(ctx, reply)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Failed to convert reply to response: %v", err)
		}
		replyResponses = append(replyResponses, response)
	}

	totalCount := len(replies)

	return &thread.RepliesResponse{
		Replies: replyResponses,
		Total:   int32(totalCount),
	}, nil
}

func (h *ThreadHandler) GetRepliesByUser(ctx context.Context, req *thread.GetRepliesByUserRequest) (*thread.RepliesResponse, error) {

	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "User ID is required")
	}

	page := 1
	limit := 20
	if req.Page > 0 {
		page = int(req.Page)
	}
	if req.Limit > 0 && req.Limit <= 100 {
		limit = int(req.Limit)
	}

	replies, err := h.replyService.GetRepliesByUserID(ctx, req.UserId, page, limit)
	if err != nil {
		return nil, err
	}

	replyResponses := make([]*thread.ReplyResponse, 0, len(replies))
	for _, reply := range replies {
		response, err := h.convertReplyToResponse(ctx, reply)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Failed to convert reply to response: %v", err)
		}
		replyResponses = append(replyResponses, response)
	}

	return &thread.RepliesResponse{
		Replies: replyResponses,
		Total:   int32(len(replies)),
	}, nil
}

func (h *ThreadHandler) GetLikedThreadsByUser(ctx context.Context, req *thread.GetLikedThreadsByUserRequest) (*thread.ThreadsResponse, error) {
	log.Printf("GetLikedThreadsByUser: Starting with request userId=%s, page=%d, limit=%d", req.UserId, req.Page, req.Limit)

	var finalError error

	// Add panic recovery
	defer func() {
		if r := recover(); r != nil {
			log.Printf("PANIC RECOVERED in GetLikedThreadsByUser: %v", r)
			finalError = status.Error(codes.Internal, "Internal server error occurred")
		}
	}()

	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "User ID is required")
	}

	page := 1
	limit := 20
	if req.Page > 0 {
		page = int(req.Page)
	}
	if req.Limit > 0 && req.Limit <= 100 {
		limit = int(req.Limit)
	}

	// Extract the requesting user ID from context metadata if available
	var requestingUserID string
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		userIDs := md.Get("user_id")
		if len(userIDs) > 0 {
			requestingUserID = userIDs[0]
			log.Printf("GetLikedThreadsByUser: Request from authenticated user: %s", requestingUserID)
		}
	}

	threadIDs, err := h.interactionService.GetLikedThreadsByUserID(ctx, req.UserId, page, limit)
	if err != nil {
		log.Printf("GetLikedThreadsByUser: Error getting liked threads: %v", err)
		return nil, err
	}

	log.Printf("GetLikedThreadsByUser: Found %d liked threads for user %s", len(threadIDs), req.UserId)

	threadResponses := make([]*thread.ThreadResponse, 0, len(threadIDs))
	for _, id := range threadIDs {
		// Wrap in a function to handle potential panics per thread
		func(threadID string) {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("PANIC in processing thread %s: %v", threadID, r)
				}
			}()

			t, err := h.threadService.GetThreadByID(ctx, threadID)
			if err != nil {
				log.Printf("GetLikedThreadsByUser: Error retrieving thread %s: %v", threadID, err)
				return
			}

			response, err := h.convertThreadToResponse(ctx, t)
			if err != nil {
				log.Printf("GetLikedThreadsByUser: Error converting thread %s: %v", threadID, err)
				return
			}

			// Set liked status to true since these are liked threads
			response.LikedByUser = true

			// Check bookmark status if we have a requesting user ID
			if requestingUserID != "" && h.interactionService != nil {
				// Use a separate context with timeout for this operation
				interactionCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
				defer cancel()

				// Check if user bookmarked the thread
				hasBookmarked, err := h.interactionService.HasUserBookmarked(interactionCtx, requestingUserID, threadID)
				if err != nil {
					log.Printf("WARNING: Failed to check if user %s bookmarked thread %s: %v",
						requestingUserID, threadID, err)
				} else {
					response.BookmarkedByUser = hasBookmarked
					log.Printf("Thread %s bookmark status for user %s: %v",
						threadID, requestingUserID, hasBookmarked)
				}
			}

			threadResponses = append(threadResponses, response)
		}(id)
	}

	log.Printf("GetLikedThreadsByUser: Successfully processed %d threads", len(threadResponses))

	return &thread.ThreadsResponse{
		Threads: threadResponses,
		Total:   int32(len(threadResponses)),
	}, finalError
}

func (h *ThreadHandler) GetMediaByUser(ctx context.Context, req *thread.GetMediaByUserRequest) (*thread.GetMediaByUserResponse, error) {

	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "User ID is required")
	}

	page := 1
	limit := 20
	if req.Page > 0 {
		page = int(req.Page)
	}
	if req.Limit > 0 && req.Limit <= 100 {
		limit = int(req.Limit)
	}

	mediaItems, err := h.threadService.GetMediaByUserID(ctx, req.UserId, page, limit)
	if err != nil {
		return nil, err
	}

	mediaResponses := make([]*thread.MediaItem, 0, len(mediaItems))
	for _, m := range mediaItems {
		threadID := ""
		if m.ThreadID != nil {
			threadID = m.ThreadID.String()
		}

		mediaResponses = append(mediaResponses, &thread.MediaItem{
			Id:        m.MediaID.String(),
			Url:       m.URL,
			Type:      m.Type,
			ThreadId:  threadID,
			CreatedAt: timestamppb.New(m.CreatedAt),
		})
	}

	return &thread.GetMediaByUserResponse{
		Media: mediaResponses,
		Total: int32(len(mediaResponses)),
	}, nil
}

func (h *ThreadHandler) GetBookmarksByUser(ctx context.Context, req *thread.GetBookmarksByUserRequest) (*thread.ThreadsResponse, error) {
	log.Printf("GetBookmarksByUser called with user_id=%s, page=%d, limit=%d", req.UserId, req.Page, req.Limit)

	var finalError error

	// Add panic recovery
	defer func() {
		if r := recover(); r != nil {
			log.Printf("PANIC RECOVERED in GetBookmarksByUser: %v", r)
			finalError = status.Error(codes.Internal, "Internal server error occurred")
		}
	}()

	if req.UserId == "" {
		log.Printf("ERROR: GetBookmarksByUser - Missing userID")
		return nil, status.Error(codes.InvalidArgument, "User ID is required")
	}

	// Extract the requesting user ID from context metadata if available
	var requestingUserID string
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		userIDs := md.Get("user_id")
		if len(userIDs) > 0 {
			requestingUserID = userIDs[0]
			log.Printf("GetBookmarksByUser: Request from authenticated user: %s", requestingUserID)
		}
	}

	// Set default pagination values if not provided
	page := int(req.Page)
	limit := int(req.Limit)
	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	// Get bookmarks using the interaction service
	threads, count, err := h.interactionService.GetUserBookmarks(ctx, req.UserId, page, limit)
	if err != nil {
		log.Printf("ERROR: Failed to get user bookmarks: %v", err)
		return nil, status.Errorf(codes.Internal, "Failed to get bookmarks: %v", err)
	}

	// If no bookmarks found, return empty response instead of error
	if len(threads) == 0 {
		log.Printf("GetBookmarksByUser: No bookmarks found for user %s", req.UserId)
		return &thread.ThreadsResponse{
			Threads: []*thread.ThreadResponse{},
			Total:   0,
		}, nil
	}

	// Convert thread models to thread responses
	threadResponses := make([]*thread.ThreadResponse, 0, len(threads))
	for _, t := range threads {
		// Skip any null threads
		if t == nil {
			continue
		}

		// Wrap conversion in try-catch to prevent individual thread errors from breaking the entire response
		func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("PANIC in thread conversion for thread %v: %v", t.ThreadID, r)
				}
			}()

			threadResp, err := h.convertThreadToResponse(ctx, t)
			if err != nil {
				log.Printf("WARNING: Failed to convert thread %s to response: %v", t.ThreadID.String(), err)
				return
			}

			// Always set the bookmark status to true since these are from bookmarks
			threadResp.BookmarkedByUser = true

			// Check if user liked the thread
			if requestingUserID != "" && h.interactionService != nil {
				// Use a separate context with timeout for this operation
				interactionCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
				defer cancel()

				// Check if user liked the thread
				hasLiked, err := h.interactionService.HasUserLikedThread(interactionCtx, requestingUserID, t.ThreadID.String())
				if err != nil {
					log.Printf("WARNING: Failed to check if user %s liked thread %s: %v",
						requestingUserID, t.ThreadID.String(), err)
				} else {
					threadResp.LikedByUser = hasLiked
					log.Printf("Thread %s like status for user %s: %v",
						t.ThreadID.String(), requestingUserID, hasLiked)
				}

				// Check if user reposted the thread
				hasReposted, err := h.interactionService.HasUserReposted(interactionCtx, requestingUserID, t.ThreadID.String())
				if err != nil {
					log.Printf("WARNING: Failed to check if user %s reposted thread %s: %v",
						requestingUserID, t.ThreadID.String(), err)
				} else {
					threadResp.RepostedByUser = hasReposted
					log.Printf("Thread %s repost status for user %s: %v",
						t.ThreadID.String(), requestingUserID, hasReposted)
				}
			}

			threadResponses = append(threadResponses, threadResp)
		}()
	}

	log.Printf("Successfully retrieved %d bookmarks for user %s", len(threadResponses), req.UserId)

	response := &thread.ThreadsResponse{
		Threads: threadResponses,
		Total:   int32(count),
	}

	return response, finalError
}

func (h *ThreadHandler) convertThreadToResponse(ctx context.Context, threadModel *model.Thread) (*thread.ThreadResponse, error) {
	// Check for nil thread
	if threadModel == nil {
		log.Printf("ERROR: Nil thread model passed to convertThreadToResponse")
		return nil, status.Errorf(codes.Internal, "Thread model is nil")
	}

	log.Printf("Converting thread %s to response", threadModel.ThreadID.String())

	// Extract user ID from context if available
	var requestingUserID string
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		userIDs := md.Get("user_id")
		if len(userIDs) > 0 {
			requestingUserID = userIDs[0]
		}
	}

	// Default values in case user data cannot be fetched
	userName := "Unknown User"
	userDisplayName := "Unknown"
	profilePictureURL := ""

	// Try to get user data if possible
	if h.userClient != nil {
		userData, err := h.userClient.GetUserById(ctx, threadModel.UserID.String())
		if err != nil {
			// Log error but continue with default values
			log.Printf("WARNING: Failed to get user data for thread %s: %v", threadModel.ThreadID, err)
		} else if userData != nil {
			userName = userData.Username
			userDisplayName = userData.DisplayName
			profilePictureURL = userData.ProfilePictureUrl
		}
	}

	// Create user object for the response
	user := &thread.User{
		Id:                threadModel.UserID.String(),
		Username:          userName,
		Name:              userDisplayName,
		ProfilePictureUrl: profilePictureURL,
		IsVerified:        false,
	}

	// Create thread protobuf object with required fields
	protoThread := &thread.Thread{
		Id:        threadModel.ThreadID.String(),
		UserId:    threadModel.UserID.String(),
		Content:   threadModel.Content,
		CreatedAt: timestamppb.New(threadModel.CreatedAt),
		UpdatedAt: timestamppb.New(threadModel.UpdatedAt),
	}

	// Handle optional IsPinned field as a pointer type
	isPinned := threadModel.IsPinned
	protoThread.IsPinned = &isPinned

	// Get media for the thread
	mediaList := make([]*thread.Media, 0)
	if h.mediaRepo != nil {
		threadID := threadModel.ThreadID.String()
		mediaItems, err := h.mediaRepo.FindMediaByThreadID(threadID)
		if err == nil && len(mediaItems) > 0 {
			for _, mediaItem := range mediaItems {
				mediaList = append(mediaList, &thread.Media{
					Id:   mediaItem.MediaID.String(),
					Url:  mediaItem.URL,
					Type: mediaItem.Type,
				})
			}
		}
	}
	protoThread.Media = mediaList

	// Get interaction counts
	likesCount := int64(0)
	repliesCount := int64(0)
	repostsCount := int64(0)
	bookmarkCount := int64(0)

	if h.interactionRepo != nil {
		// Get like count
		lCount, err := h.interactionRepo.CountThreadLikes(threadModel.ThreadID.String())
		if err == nil {
			likesCount = lCount
		}

		// Get repost count
		rCount, err := h.interactionRepo.CountThreadReposts(threadModel.ThreadID.String())
		if err == nil {
			repostsCount = rCount
		}

		// Get bookmark count
		bCount, err := h.interactionRepo.CountThreadBookmarks(threadModel.ThreadID.String())
		if err == nil {
			bookmarkCount = bCount
		}
	}

	// Create thread response
	response := &thread.ThreadResponse{
		Thread:           protoThread,
		User:             user,
		LikesCount:       likesCount,
		RepliesCount:     repliesCount,
		RepostsCount:     repostsCount,
		BookmarkCount:    bookmarkCount,
		LikedByUser:      false,
		RepostedByUser:   false,
		BookmarkedByUser: false,
	}

	// Check user interaction status if user ID is in context and interaction service is available
	if requestingUserID != "" && h.interactionService != nil {
		// Use a separate context with timeout for these operations
		interactionCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()

		// Check if user liked the thread
		hasLiked, err := h.interactionService.HasUserLikedThread(interactionCtx, requestingUserID, threadModel.ThreadID.String())
		if err != nil {
			log.Printf("WARNING: Failed to check if user %s liked thread %s: %v",
				requestingUserID, threadModel.ThreadID.String(), err)
		} else {
			response.LikedByUser = hasLiked
		}

		// Check if user reposted the thread
		hasReposted, err := h.interactionService.HasUserReposted(interactionCtx, requestingUserID, threadModel.ThreadID.String())
		if err != nil {
			log.Printf("WARNING: Failed to check if user %s reposted thread %s: %v",
				requestingUserID, threadModel.ThreadID.String(), err)
		} else {
			response.RepostedByUser = hasReposted
		}

		// Check if user bookmarked the thread
		hasBookmarked, err := h.interactionService.HasUserBookmarked(interactionCtx, requestingUserID, threadModel.ThreadID.String())
		if err != nil {
			log.Printf("WARNING: Failed to check if user %s bookmarked thread %s: %v",
				requestingUserID, threadModel.ThreadID.String(), err)
		} else {
			response.BookmarkedByUser = hasBookmarked
		}
	}

	log.Printf("Successfully converted thread %s to response", threadModel.ThreadID.String())
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

		likeCount, err := h.interactionRepo.CountReplyLikes(replyID)
		if err == nil {
			response.LikesCount = likeCount
		}

		bookmarkCount, err := h.interactionRepo.CountReplyBookmarks(replyID)
		if err == nil {
			response.BookmarkCount = bookmarkCount
		}

		// Count replies to this reply
		repliesCount, err := h.replyService.CountRepliesByParentID(ctx, replyID)
		if err == nil {
			response.RepliesCount = repliesCount
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if ok {

			userIDs := md.Get("user_id")
			if len(userIDs) > 0 {
				userID := userIDs[0]

				hasLiked, err := h.interactionService.HasUserLikedReply(ctx, userID, replyID)
				if err == nil {
					response.LikedByUser = hasLiked
				}

				hasBookmarked, err := h.interactionService.HasUserBookmarkedReply(ctx, userID, replyID)
				if err == nil {
					response.BookmarkedByUser = hasBookmarked
				}
			}
		}
	}

	if h.userClient != nil {
		userInfo, err := h.userClient.GetUserById(ctx, reply.UserID.String())
		if err == nil && userInfo != nil {
			log.Printf("User info retrieved for reply %s by user %s (username: %s)",
				reply.ReplyID.String(), userInfo.Id, userInfo.Username)

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
