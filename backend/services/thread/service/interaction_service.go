package service

import (
	"context"
	"log"
	"time"

	"aycom/backend/services/thread/model"
	"aycom/backend/services/thread/repository"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Thread = model.Thread

type InteractionService interface {
	LikeThread(ctx context.Context, userID, threadID string) error
	UnlikeThread(ctx context.Context, userID, threadID string) error
	LikeReply(ctx context.Context, userID, replyID string) error
	UnlikeReply(ctx context.Context, userID, replyID string) error
	HasUserLikedThread(ctx context.Context, userID, threadID string) (bool, error)
	HasUserLikedReply(ctx context.Context, userID, replyID string) (bool, error)
	GetThreadLikeCount(ctx context.Context, threadID string) (int64, error)

	RepostThread(ctx context.Context, userID, threadID string, repostText *string) error
	RemoveRepost(ctx context.Context, userID, threadID string) error
	HasUserReposted(ctx context.Context, userID, threadID string) (bool, error)

	BookmarkThread(ctx context.Context, userID, threadID string) error
	RemoveBookmark(ctx context.Context, userID, threadID string) error
	HasUserBookmarked(ctx context.Context, userID, threadID string) (bool, error)
	GetUserBookmarks(ctx context.Context, userID string, page, limit int) ([]*Thread, int64, error)

	BookmarkReply(ctx context.Context, userID, replyID string) error
	RemoveReplyBookmark(ctx context.Context, userID, replyID string) error
	HasUserBookmarkedReply(ctx context.Context, userID, replyID string) (bool, error)

	GetLikedThreadsByUserID(ctx context.Context, userID string, page, limit int) ([]string, error)
}

type interactionService struct {
	interactionRepo repository.InteractionRepository
	threadRepo      repository.ThreadRepository
	userRepo        repository.UserRepository
}

func NewInteractionService(
	interactionRepo repository.InteractionRepository,
	threadRepo repository.ThreadRepository,
	userRepo repository.UserRepository,
) InteractionService {

	if userRepo == nil {
		log.Printf("Warning: UserRepository is nil in InteractionService constructor")
	}

	return &interactionService{
		interactionRepo: interactionRepo,
		threadRepo:      threadRepo,
		userRepo:        userRepo,
	}
}

func (s *interactionService) LikeThread(ctx context.Context, userID, threadID string) error {
	if userID == "" || threadID == "" {
		return status.Error(codes.InvalidArgument, "User ID and Thread ID are required")
	}

	hasLiked, err := s.interactionRepo.IsThreadLikedByUser(userID, threadID)
	if err != nil {
		return status.Errorf(codes.Internal, "Failed to check if user has liked thread: %v", err)
	}

	if hasLiked {

		return nil
	}

	if err := s.interactionRepo.LikeThread(userID, threadID); err != nil {

		if err.Error() == "ERROR: duplicate key value violates unique constraint \"likes_pkey\" (SQLSTATE 23505)" {

			return nil
		}
		return status.Errorf(codes.Internal, "Failed to like thread: %v", err)
	}

	return nil
}

func (s *interactionService) UnlikeThread(ctx context.Context, userID, threadID string) error {
	if userID == "" || threadID == "" {
		return status.Error(codes.InvalidArgument, "User ID and Thread ID are required")
	}

	hasLiked, err := s.interactionRepo.IsThreadLikedByUser(userID, threadID)
	if err != nil {
		return status.Errorf(codes.Internal, "Failed to check if user has liked thread: %v", err)
	}

	if !hasLiked {

		return nil
	}

	if err := s.interactionRepo.UnlikeThread(userID, threadID); err != nil {
		return status.Errorf(codes.Internal, "Failed to unlike thread: %v", err)
	}

	return nil
}

func (s *interactionService) LikeReply(ctx context.Context, userID, replyID string) error {
	if userID == "" || replyID == "" {
		return status.Error(codes.InvalidArgument, "User ID and Reply ID are required")
	}

	hasLiked, err := s.interactionRepo.IsReplyLikedByUser(userID, replyID)
	if err != nil {
		return status.Errorf(codes.Internal, "Failed to check if user has liked reply: %v", err)
	}

	if hasLiked {

		return nil
	}

	if err := s.interactionRepo.LikeReply(userID, replyID); err != nil {

		if err.Error() == "ERROR: duplicate key value violates unique constraint \"likes_pkey\" (SQLSTATE 23505)" {

			return nil
		}
		return status.Errorf(codes.Internal, "Failed to like reply: %v", err)
	}

	return nil
}

func (s *interactionService) UnlikeReply(ctx context.Context, userID, replyID string) error {
	if userID == "" || replyID == "" {
		return status.Error(codes.InvalidArgument, "User ID and Reply ID are required")
	}

	hasLiked, err := s.interactionRepo.IsReplyLikedByUser(userID, replyID)
	if err != nil {
		return status.Errorf(codes.Internal, "Failed to check if user has liked reply: %v", err)
	}

	if !hasLiked {

		return nil
	}

	if err := s.interactionRepo.UnlikeReply(userID, replyID); err != nil {
		return status.Errorf(codes.Internal, "Failed to unlike reply: %v", err)
	}

	return nil
}

func (s *interactionService) HasUserLikedThread(ctx context.Context, userID, threadID string) (bool, error) {
	if userID == "" || threadID == "" {
		return false, status.Error(codes.InvalidArgument, "User ID and Thread ID are required")
	}

	return s.interactionRepo.IsThreadLikedByUser(userID, threadID)
}

func (s *interactionService) HasUserLikedReply(ctx context.Context, userID, replyID string) (bool, error) {
	if userID == "" || replyID == "" {
		return false, status.Error(codes.InvalidArgument, "User ID and Reply ID are required")
	}

	return s.interactionRepo.IsReplyLikedByUser(userID, replyID)
}

func (s *interactionService) GetThreadLikeCount(ctx context.Context, threadID string) (int64, error) {
	if threadID == "" {
		return 0, status.Error(codes.InvalidArgument, "Thread ID is required")
	}

	count, err := s.interactionRepo.CountThreadLikes(threadID)
	if err != nil {
		return 0, status.Errorf(codes.Internal, "Failed to count thread likes: %v", err)
	}

	return count, nil
}

func (s *interactionService) RepostThread(ctx context.Context, userID, threadID string, repostText *string) error {
	if userID == "" || threadID == "" {
		return status.Error(codes.InvalidArgument, "User ID and Thread ID are required")
	}

	// Check if user has already reposted this thread
	hasReposted, err := s.interactionRepo.IsThreadRepostedByUser(userID, threadID)
	if err != nil {
		return status.Errorf(codes.Internal, "Failed to check if user has reposted thread: %v", err)
	}

	if hasReposted {
		return status.Error(codes.AlreadyExists, "User has already reposted this thread")
	}

	userUUID, err := s.parseUUID(userID)
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "Invalid user ID: %v", err)
	}

	threadUUID, err := s.parseUUID(threadID)
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "Invalid thread ID: %v", err)
	}

	// Verify the thread exists
	_, err = s.threadRepo.FindThreadByID(threadID)
	if err != nil {
		return status.Errorf(codes.Internal, "Failed to find original thread: %v", err)
	}

	// Create new thread ID for the repost
	newThreadID := uuid.New()

	// Construct repost content with attribution
	content := ""
	if repostText != nil && *repostText != "" {
		// If repost has text, use it as content
		content = *repostText
	}

	// Create a new thread representing the repost
	newThread := &model.Thread{
		ThreadID:        newThreadID,
		UserID:          userUUID,
		Content:         content,
		WhoCanReply:     "Everyone",
		IsAdvertisement: false,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		// Set repost metadata
		OriginalThreadID: &threadUUID,
		IsRepost:         true,
	}

	if err := s.threadRepo.CreateThread(newThread); err != nil {
		return status.Errorf(codes.Internal, "Failed to create repost thread: %v", err)
	}

	// Also create an entry in the reposts table for tracking purposes
	repost := &model.Repost{
		UserID:      userUUID,
		ThreadID:    threadUUID,
		RepostText:  repostText,
		NewThreadID: &newThreadID,
	}

	if err := s.interactionRepo.RepostThread(repost); err != nil {
		// If we fail to create the repost record but already created the thread,
		// we should log the error but still consider the operation successful
		log.Printf("Warning: Created repost thread %s but failed to create repost record: %v", newThreadID, err)
	}

	return nil
}

func (s *interactionService) parseUUID(id string) (uuid.UUID, error) {
	return uuid.Parse(id)
}

func (s *interactionService) RemoveRepost(ctx context.Context, userID, threadID string) error {
	if userID == "" || threadID == "" {
		return status.Error(codes.InvalidArgument, "User ID and Thread ID are required")
	}

	hasReposted, err := s.interactionRepo.IsThreadRepostedByUser(userID, threadID)
	if err != nil {
		return status.Errorf(codes.Internal, "Failed to check if user has reposted thread: %v", err)
	}

	if !hasReposted {

		return nil
	}

	if err := s.interactionRepo.RemoveRepost(userID, threadID); err != nil {
		return status.Errorf(codes.Internal, "Failed to remove repost: %v", err)
	}

	return nil
}

func (s *interactionService) HasUserReposted(ctx context.Context, userID, threadID string) (bool, error) {
	if userID == "" || threadID == "" {
		return false, status.Error(codes.InvalidArgument, "User ID and Thread ID are required")
	}

	return s.interactionRepo.IsThreadRepostedByUser(userID, threadID)
}

func (s *interactionService) BookmarkThread(ctx context.Context, userID, threadID string) error {
	if userID == "" || threadID == "" {
		return status.Error(codes.InvalidArgument, "User ID and Thread ID are required")
	}

	hasBookmarked, err := s.interactionRepo.IsThreadBookmarkedByUser(userID, threadID)
	if err != nil {
		return status.Errorf(codes.Internal, "Failed to check if user has bookmarked thread: %v", err)
	}

	if hasBookmarked {
		return status.Error(codes.AlreadyExists, "User has already bookmarked this thread")
	}

	if err := s.interactionRepo.BookmarkThread(userID, threadID); err != nil {
		return status.Errorf(codes.Internal, "Failed to bookmark thread: %v", err)
	}

	return nil
}

func (s *interactionService) RemoveBookmark(ctx context.Context, userID, threadID string) error {
	if userID == "" || threadID == "" {
		return status.Error(codes.InvalidArgument, "User ID and Thread ID are required")
	}

	hasBookmarked, err := s.interactionRepo.IsThreadBookmarkedByUser(userID, threadID)
	if err != nil {
		return status.Errorf(codes.Internal, "Failed to check if user has bookmarked thread: %v", err)
	}

	if !hasBookmarked {
		return nil
	}

	if err := s.interactionRepo.RemoveBookmark(userID, threadID); err != nil {
		return status.Errorf(codes.Internal, "Failed to remove bookmark: %v", err)
	}

	return nil
}

func (s *interactionService) HasUserBookmarked(ctx context.Context, userID, threadID string) (bool, error) {
	if userID == "" || threadID == "" {
		return false, status.Error(codes.InvalidArgument, "User ID and Thread ID are required")
	}

	return s.interactionRepo.IsThreadBookmarkedByUser(userID, threadID)
}

func (s *interactionService) GetUserBookmarks(ctx context.Context, userID string, page, limit int) ([]*Thread, int64, error) {
	if userID == "" {
		return nil, 0, status.Error(codes.InvalidArgument, "User ID is required")
	}

	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	threads, err := s.interactionRepo.GetUserBookmarks(userID, page, limit)
	if err != nil {
		return nil, 0, status.Errorf(codes.Internal, "Failed to retrieve bookmarks: %v", err)
	}

	totalCount := int64(len(threads))

	return threads, totalCount, nil
}

func (s *interactionService) BookmarkReply(ctx context.Context, userID, replyID string) error {
	if userID == "" || replyID == "" {
		return status.Error(codes.InvalidArgument, "User ID and Reply ID are required")
	}

	hasBookmarked, err := s.interactionRepo.IsReplyBookmarkedByUser(userID, replyID)
	if err != nil {
		return status.Errorf(codes.Internal, "Failed to check if user has bookmarked reply: %v", err)
	}

	if hasBookmarked {
		return status.Error(codes.AlreadyExists, "User has already bookmarked this reply")
	}

	if err := s.interactionRepo.BookmarkReply(userID, replyID); err != nil {
		return status.Errorf(codes.Internal, "Failed to bookmark reply: %v", err)
	}

	return nil
}

func (s *interactionService) RemoveReplyBookmark(ctx context.Context, userID, replyID string) error {
	if userID == "" || replyID == "" {
		return status.Error(codes.InvalidArgument, "User ID and Reply ID are required")
	}

	hasBookmarked, err := s.interactionRepo.IsReplyBookmarkedByUser(userID, replyID)
	if err != nil {
		return status.Errorf(codes.Internal, "Failed to check if user has bookmarked reply: %v", err)
	}

	if !hasBookmarked {
		return status.Error(codes.NotFound, "User has not bookmarked this reply")
	}

	if err := s.interactionRepo.RemoveReplyBookmark(userID, replyID); err != nil {
		return status.Errorf(codes.Internal, "Failed to remove reply bookmark: %v", err)
	}

	return nil
}

func (s *interactionService) HasUserBookmarkedReply(ctx context.Context, userID, replyID string) (bool, error) {
	if userID == "" || replyID == "" {
		return false, status.Error(codes.InvalidArgument, "User ID and Reply ID are required")
	}

	return s.interactionRepo.IsReplyBookmarkedByUser(userID, replyID)
}

func (s *interactionService) GetLikedThreadsByUserID(ctx context.Context, userID string, page, limit int) ([]string, error) {
	if userID == "" {
		return nil, status.Error(codes.InvalidArgument, "User ID is required")
	}

	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	threadIDs, err := s.interactionRepo.FindLikedThreadsByUserID(userID, page, limit)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to retrieve liked threads: %v", err)
	}

	return threadIDs, nil
}
