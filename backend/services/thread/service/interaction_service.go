package service

import (
	"context"

	"github.com/Acad600-Tpa/WEB-MV-242/backend/services/thread/repository"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// InteractionService defines the interface for interaction operations (likes, reposts, bookmarks)
type InteractionService interface {
	// Like operations
	LikeThread(ctx context.Context, userID, threadID string) error
	UnlikeThread(ctx context.Context, userID, threadID string) error
	LikeReply(ctx context.Context, userID, replyID string) error
	UnlikeReply(ctx context.Context, userID, replyID string) error
	HasUserLikedThread(ctx context.Context, userID, threadID string) (bool, error)
	HasUserLikedReply(ctx context.Context, userID, replyID string) (bool, error)

	// Repost operations
	RepostThread(ctx context.Context, userID, threadID string, repostText *string) error
	RemoveRepost(ctx context.Context, userID, threadID string) error
	HasUserReposted(ctx context.Context, userID, threadID string) (bool, error)

	// Bookmark operations
	BookmarkThread(ctx context.Context, userID, threadID string) error
	RemoveBookmark(ctx context.Context, userID, threadID string) error
	HasUserBookmarked(ctx context.Context, userID, threadID string) (bool, error)
	GetUserBookmarks(ctx context.Context, userID string, page, limit int) ([]*Thread, int64, error)
}

// interactionService implements the InteractionService interface
type interactionService struct {
	likeRepo     repository.LikeRepository
	repostRepo   repository.RepostRepository
	bookmarkRepo repository.BookmarkRepository
}

// NewInteractionService creates a new interaction service
func NewInteractionService(
	likeRepo repository.LikeRepository,
	repostRepo repository.RepostRepository,
	bookmarkRepo repository.BookmarkRepository,
) InteractionService {
	return &interactionService{
		likeRepo:     likeRepo,
		repostRepo:   repostRepo,
		bookmarkRepo: bookmarkRepo,
	}
}

// LikeThread adds a like to a thread
func (s *interactionService) LikeThread(ctx context.Context, userID, threadID string) error {
	if userID == "" || threadID == "" {
		return status.Error(codes.InvalidArgument, "User ID and Thread ID are required")
	}

	// Check if user has already liked this thread
	hasLiked, err := s.likeRepo.HasUserLikedThread(userID, threadID)
	if err != nil {
		return status.Errorf(codes.Internal, "Failed to check if user has liked thread: %v", err)
	}

	if hasLiked {
		return status.Error(codes.AlreadyExists, "User has already liked this thread")
	}

	// Add like
	if err := s.likeRepo.CreateThreadLike(userID, threadID); err != nil {
		return status.Errorf(codes.Internal, "Failed to like thread: %v", err)
	}

	return nil
}

// UnlikeThread removes a like from a thread
func (s *interactionService) UnlikeThread(ctx context.Context, userID, threadID string) error {
	if userID == "" || threadID == "" {
		return status.Error(codes.InvalidArgument, "User ID and Thread ID are required")
	}

	// Check if user has liked this thread
	hasLiked, err := s.likeRepo.HasUserLikedThread(userID, threadID)
	if err != nil {
		return status.Errorf(codes.Internal, "Failed to check if user has liked thread: %v", err)
	}

	if !hasLiked {
		return status.Error(codes.NotFound, "User has not liked this thread")
	}

	// Remove like
	if err := s.likeRepo.DeleteThreadLike(userID, threadID); err != nil {
		return status.Errorf(codes.Internal, "Failed to unlike thread: %v", err)
	}

	return nil
}

// LikeReply adds a like to a reply
func (s *interactionService) LikeReply(ctx context.Context, userID, replyID string) error {
	if userID == "" || replyID == "" {
		return status.Error(codes.InvalidArgument, "User ID and Reply ID are required")
	}

	// Check if user has already liked this reply
	hasLiked, err := s.likeRepo.HasUserLikedReply(userID, replyID)
	if err != nil {
		return status.Errorf(codes.Internal, "Failed to check if user has liked reply: %v", err)
	}

	if hasLiked {
		return status.Error(codes.AlreadyExists, "User has already liked this reply")
	}

	// Add like
	if err := s.likeRepo.CreateReplyLike(userID, replyID); err != nil {
		return status.Errorf(codes.Internal, "Failed to like reply: %v", err)
	}

	return nil
}

// UnlikeReply removes a like from a reply
func (s *interactionService) UnlikeReply(ctx context.Context, userID, replyID string) error {
	if userID == "" || replyID == "" {
		return status.Error(codes.InvalidArgument, "User ID and Reply ID are required")
	}

	// Check if user has liked this reply
	hasLiked, err := s.likeRepo.HasUserLikedReply(userID, replyID)
	if err != nil {
		return status.Errorf(codes.Internal, "Failed to check if user has liked reply: %v", err)
	}

	if !hasLiked {
		return status.Error(codes.NotFound, "User has not liked this reply")
	}

	// Remove like
	if err := s.likeRepo.DeleteReplyLike(userID, replyID); err != nil {
		return status.Errorf(codes.Internal, "Failed to unlike reply: %v", err)
	}

	return nil
}

// HasUserLikedThread checks if a user has liked a thread
func (s *interactionService) HasUserLikedThread(ctx context.Context, userID, threadID string) (bool, error) {
	if userID == "" || threadID == "" {
		return false, status.Error(codes.InvalidArgument, "User ID and Thread ID are required")
	}

	return s.likeRepo.HasUserLikedThread(userID, threadID)
}

// HasUserLikedReply checks if a user has liked a reply
func (s *interactionService) HasUserLikedReply(ctx context.Context, userID, replyID string) (bool, error) {
	if userID == "" || replyID == "" {
		return false, status.Error(codes.InvalidArgument, "User ID and Reply ID are required")
	}

	return s.likeRepo.HasUserLikedReply(userID, replyID)
}

// RepostThread reposts a thread
func (s *interactionService) RepostThread(ctx context.Context, userID, threadID string, repostText *string) error {
	if userID == "" || threadID == "" {
		return status.Error(codes.InvalidArgument, "User ID and Thread ID are required")
	}

	// Check if user has already reposted this thread
	hasReposted, err := s.repostRepo.HasUserReposted(userID, threadID)
	if err != nil {
		return status.Errorf(codes.Internal, "Failed to check if user has reposted thread: %v", err)
	}

	if hasReposted {
		return status.Error(codes.AlreadyExists, "User has already reposted this thread")
	}

	// Add repost
	if err := s.repostRepo.CreateRepost(userID, threadID, repostText); err != nil {
		return status.Errorf(codes.Internal, "Failed to repost thread: %v", err)
	}

	return nil
}

// RemoveRepost removes a repost
func (s *interactionService) RemoveRepost(ctx context.Context, userID, threadID string) error {
	if userID == "" || threadID == "" {
		return status.Error(codes.InvalidArgument, "User ID and Thread ID are required")
	}

	// Check if user has reposted this thread
	hasReposted, err := s.repostRepo.HasUserReposted(userID, threadID)
	if err != nil {
		return status.Errorf(codes.Internal, "Failed to check if user has reposted thread: %v", err)
	}

	if !hasReposted {
		return status.Error(codes.NotFound, "User has not reposted this thread")
	}

	// Remove repost
	if err := s.repostRepo.DeleteRepost(userID, threadID); err != nil {
		return status.Errorf(codes.Internal, "Failed to remove repost: %v", err)
	}

	return nil
}

// HasUserReposted checks if a user has reposted a thread
func (s *interactionService) HasUserReposted(ctx context.Context, userID, threadID string) (bool, error) {
	if userID == "" || threadID == "" {
		return false, status.Error(codes.InvalidArgument, "User ID and Thread ID are required")
	}

	return s.repostRepo.HasUserReposted(userID, threadID)
}

// BookmarkThread bookmarks a thread
func (s *interactionService) BookmarkThread(ctx context.Context, userID, threadID string) error {
	if userID == "" || threadID == "" {
		return status.Error(codes.InvalidArgument, "User ID and Thread ID are required")
	}

	// Check if user has already bookmarked this thread
	hasBookmarked, err := s.bookmarkRepo.HasUserBookmarked(userID, threadID)
	if err != nil {
		return status.Errorf(codes.Internal, "Failed to check if user has bookmarked thread: %v", err)
	}

	if hasBookmarked {
		return status.Error(codes.AlreadyExists, "User has already bookmarked this thread")
	}

	// Add bookmark
	if err := s.bookmarkRepo.CreateBookmark(userID, threadID); err != nil {
		return status.Errorf(codes.Internal, "Failed to bookmark thread: %v", err)
	}

	return nil
}

// RemoveBookmark removes a bookmark
func (s *interactionService) RemoveBookmark(ctx context.Context, userID, threadID string) error {
	if userID == "" || threadID == "" {
		return status.Error(codes.InvalidArgument, "User ID and Thread ID are required")
	}

	// Check if user has bookmarked this thread
	hasBookmarked, err := s.bookmarkRepo.HasUserBookmarked(userID, threadID)
	if err != nil {
		return status.Errorf(codes.Internal, "Failed to check if user has bookmarked thread: %v", err)
	}

	if !hasBookmarked {
		return status.Error(codes.NotFound, "User has not bookmarked this thread")
	}

	// Remove bookmark
	if err := s.bookmarkRepo.DeleteBookmark(userID, threadID); err != nil {
		return status.Errorf(codes.Internal, "Failed to remove bookmark: %v", err)
	}

	return nil
}

// HasUserBookmarked checks if a user has bookmarked a thread
func (s *interactionService) HasUserBookmarked(ctx context.Context, userID, threadID string) (bool, error) {
	if userID == "" || threadID == "" {
		return false, status.Error(codes.InvalidArgument, "User ID and Thread ID are required")
	}

	return s.bookmarkRepo.HasUserBookmarked(userID, threadID)
}

// GetUserBookmarks gets a user's bookmarks with pagination
func (s *interactionService) GetUserBookmarks(ctx context.Context, userID string, page, limit int) ([]*Thread, int64, error) {
	if userID == "" {
		return nil, 0, status.Error(codes.InvalidArgument, "User ID is required")
	}

	// Default pagination values if not provided
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	threads, totalCount, err := s.bookmarkRepo.GetUserBookmarks(userID, page, limit)
	if err != nil {
		return nil, 0, status.Errorf(codes.Internal, "Failed to retrieve bookmarks: %v", err)
	}

	return threads, totalCount, nil
}
