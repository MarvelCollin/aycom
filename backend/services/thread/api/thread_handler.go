package handlers

import (
	"aycom/backend/proto/thread"
	"aycom/backend/services/thread/model"
	"aycom/backend/services/thread/repository"
	"aycom/backend/services/thread/service"
	"context"
	"log"

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
}

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

	threads, err := h.threadService.GetThreadsByUserID(ctx, req.UserId, int(req.Page), int(req.Limit))
	if err != nil {
		return nil, err
	}

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

	return nil, status.Error(codes.Unimplemented, "Method not implemented")
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

	threadIDs, err := h.interactionService.GetLikedThreadsByUserID(ctx, req.UserId, page, limit)
	if err != nil {
		return nil, err
	}

	threadResponses := make([]*thread.ThreadResponse, 0, len(threadIDs))
	for _, id := range threadIDs {

		t, err := h.threadService.GetThreadByID(ctx, id)
		if err != nil {
			log.Printf("Failed to get thread %s: %v", id, err)
			continue
		}

		response, err := h.convertThreadToResponse(ctx, t)
		if err != nil {
			log.Printf("Failed to convert thread %s to response: %v", id, err)
			continue
		}

		response.LikedByUser = true
		threadResponses = append(threadResponses, response)
	}

	return &thread.ThreadsResponse{
		Threads: threadResponses,
		Total:   int32(len(threadResponses)),
	}, nil
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

func (h *ThreadHandler) convertThreadToResponse(ctx context.Context, threadModel *model.Thread) (*thread.ThreadResponse, error) {

	protoThread := &thread.Thread{
		Id:        threadModel.ThreadID.String(),
		UserId:    threadModel.UserID.String(),
		Content:   threadModel.Content,
		CreatedAt: timestamppb.New(threadModel.CreatedAt),
		UpdatedAt: timestamppb.New(threadModel.UpdatedAt),
	}

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

	response := &thread.ThreadResponse{
		Thread: protoThread,
	}

	if h.interactionRepo != nil {
		threadID := threadModel.ThreadID.String()

		likeCount, err := h.interactionRepo.CountThreadLikes(threadID)
		if err == nil {
			response.LikesCount = likeCount
		}

		repostCount, err := h.interactionRepo.CountThreadReposts(threadID)
		if err == nil {
			response.RepostsCount = repostCount
		}

		bookmarkCount, err := h.interactionRepo.CountThreadBookmarks(threadID)
		if err == nil {
			response.BookmarkCount = bookmarkCount

			protoThread.ViewCount = bookmarkCount
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if ok {

			userIDs := md.Get("user_id")
			if len(userIDs) > 0 {
				userID := userIDs[0]

				hasLiked, err := h.interactionService.HasUserLikedThread(ctx, userID, threadID)
				if err == nil {
					response.LikedByUser = hasLiked
				}

				hasReposted, err := h.interactionService.HasUserReposted(ctx, userID, threadID)
				if err == nil {
					response.RepostedByUser = hasReposted
				}

				hasBookmarked, err := h.interactionService.HasUserBookmarked(ctx, userID, threadID)
				if err == nil {
					response.BookmarkedByUser = hasBookmarked
				}
			}
		}
	}

	if h.userClient != nil {
		userInfo, err := h.userClient.GetUserById(ctx, threadModel.UserID.String())
		if err == nil && userInfo != nil {

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
