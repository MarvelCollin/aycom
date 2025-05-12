package service

import (
	"context"
	"log"

	"aycom/backend/services/thread/model"
	"aycom/backend/services/thread/repository"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Thread is an alias for the model.Thread to avoid circular imports
type Thread = model.Thread

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

	// Reply bookmark operations
	BookmarkReply(ctx context.Context, userID, replyID string) error
	RemoveReplyBookmark(ctx context.Context, userID, replyID string) error
	HasUserBookmarkedReply(ctx context.Context, userID, replyID string) (bool, error)

	// New method
	GetLikedThreadsByUserID(ctx context.Context, userID string, page, limit int) ([]string, error)
}

// interactionService implements the InteractionService interface
type interactionService struct {
	interactionRepo repository.InteractionRepository
	threadRepo      repository.ThreadRepository
	userRepo        repository.UserRepository
}

// NewInteractionService creates a new interaction service
func NewInteractionService(
	interactionRepo repository.InteractionRepository,
	threadRepo repository.ThreadRepository,
	userRepo repository.UserRepository,
) InteractionService {
	// Log a warning if userRepo is nil, but allow the service to be created
	if userRepo == nil {
		log.Printf("Warning: UserRepository is nil in InteractionService constructor")
	}

	return &interactionService{
		interactionRepo: interactionRepo,
		threadRepo:      threadRepo,
		userRepo:        userRepo,
	}
}

// LikeThread adds a like to a thread
func (s *interactionService) LikeThread(ctx context.Context, userID, threadID string) error {
	if userID == "" || threadID == "" {
		return status.Error(codes.InvalidArgument, "User ID and Thread ID are required")
	}

	// Check if user has already liked this thread
	hasLiked, err := s.interactionRepo.IsThreadLikedByUser(userID, threadID)
	if err != nil {
		return status.Errorf(codes.Internal, "Failed to check if user has liked thread: %v", err)
	}

	if hasLiked {
		return status.Error(codes.AlreadyExists, "User has already liked this thread")
	}

	// Add like
	if err := s.interactionRepo.LikeThread(userID, threadID); err != nil {
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
	hasLiked, err := s.interactionRepo.IsThreadLikedByUser(userID, threadID)
	if err != nil {
		return status.Errorf(codes.Internal, "Failed to check if user has liked thread: %v", err)
	}

	if !hasLiked {
		return status.Error(codes.NotFound, "User has not liked this thread")
	}

	// Remove like
	if err := s.interactionRepo.UnlikeThread(userID, threadID); err != nil {
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
	hasLiked, err := s.interactionRepo.IsReplyLikedByUser(userID, replyID)
	if err != nil {
		return status.Errorf(codes.Internal, "Failed to check if user has liked reply: %v", err)
	}

	if hasLiked {
		return status.Error(codes.AlreadyExists, "User has already liked this reply")
	}

	// Add like
	if err := s.interactionRepo.LikeReply(userID, replyID); err != nil {
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
	hasLiked, err := s.interactionRepo.IsReplyLikedByUser(userID, replyID)
	if err != nil {
		return status.Errorf(codes.Internal, "Failed to check if user has liked reply: %v", err)
	}

	if !hasLiked {
		return status.Error(codes.NotFound, "User has not liked this reply")
	}

	// Remove like
	if err := s.interactionRepo.UnlikeReply(userID, replyID); err != nil {
		return status.Errorf(codes.Internal, "Failed to unlike reply: %v", err)
	}

	return nil
}

// HasUserLikedThread checks if a user has liked a thread
func (s *interactionService) HasUserLikedThread(ctx context.Context, userID, threadID string) (bool, error) {
	if userID == "" || threadID == "" {
		return false, status.Error(codes.InvalidArgument, "User ID and Thread ID are required")
	}

	return s.interactionRepo.IsThreadLikedByUser(userID, threadID)
}

// HasUserLikedReply checks if a user has liked a reply
func (s *interactionService) HasUserLikedReply(ctx context.Context, userID, replyID string) (bool, error) {
	if userID == "" || replyID == "" {
		return false, status.Error(codes.InvalidArgument, "User ID and Reply ID are required")
	}

	return s.interactionRepo.IsReplyLikedByUser(userID, replyID)
}

// RepostThread reposts a thread
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

	// Create repost model
	userUUID, err := s.parseUUID(userID)
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "Invalid user ID: %v", err)
	}

	threadUUID, err := s.parseUUID(threadID)
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "Invalid thread ID: %v", err)
	}

	repost := &model.Repost{
		UserID:   userUUID,
		ThreadID: threadUUID,
	}

	if repostText != nil {
		repost.RepostText = repostText
	}

	// Add repost
	if err := s.interactionRepo.RepostThread(repost); err != nil {
		return status.Errorf(codes.Internal, "Failed to repost thread: %v", err)
	}

	return nil
}

// parseUUID is a helper function to parse UUID strings
func (s *interactionService) parseUUID(id string) (uuid.UUID, error) {
	return uuid.Parse(id)
}

// RemoveRepost removes a repost
func (s *interactionService) RemoveRepost(ctx context.Context, userID, threadID string) error {
	if userID == "" || threadID == "" {
		return status.Error(codes.InvalidArgument, "User ID and Thread ID are required")
	}

	// Check if user has reposted this thread
	hasReposted, err := s.interactionRepo.IsThreadRepostedByUser(userID, threadID)
	if err != nil {
		return status.Errorf(codes.Internal, "Failed to check if user has reposted thread: %v", err)
	}

	if !hasReposted {
		return status.Error(codes.NotFound, "User has not reposted this thread")
	}

	// Remove repost
	if err := s.interactionRepo.RemoveRepost(userID, threadID); err != nil {
		return status.Errorf(codes.Internal, "Failed to remove repost: %v", err)
	}

	return nil
}

// HasUserReposted checks if a user has reposted a thread
func (s *interactionService) HasUserReposted(ctx context.Context, userID, threadID string) (bool, error) {
	if userID == "" || threadID == "" {
		return false, status.Error(codes.InvalidArgument, "User ID and Thread ID are required")
	}

	return s.interactionRepo.IsThreadRepostedByUser(userID, threadID)
}

// BookmarkThread bookmarks a thread
func (s *interactionService) BookmarkThread(ctx context.Context, userID, threadID string) error {
	if userID == "" || threadID == "" {
		return status.Error(codes.InvalidArgument, "User ID and Thread ID are required")
	}

	// Check if user has already bookmarked this thread
	hasBookmarked, err := s.interactionRepo.IsThreadBookmarkedByUser(userID, threadID)
	if err != nil {
		return status.Errorf(codes.Internal, "Failed to check if user has bookmarked thread: %v", err)
	}

	if hasBookmarked {
		return status.Error(codes.AlreadyExists, "User has already bookmarked this thread")
	}

	// Add bookmark
	if err := s.interactionRepo.BookmarkThread(userID, threadID); err != nil {
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
	hasBookmarked, err := s.interactionRepo.IsThreadBookmarkedByUser(userID, threadID)
	if err != nil {
		return status.Errorf(codes.Internal, "Failed to check if user has bookmarked thread: %v", err)
	}

	// If not bookmarked, just return success - this makes the API idempotent
	if !hasBookmarked {
		return nil // Return success instead of an error
	}

	// Remove bookmark
	if err := s.interactionRepo.RemoveBookmark(userID, threadID); err != nil {
		return status.Errorf(codes.Internal, "Failed to remove bookmark: %v", err)
	}

	return nil
}

// HasUserBookmarked checks if a user has bookmarked a thread
func (s *interactionService) HasUserBookmarked(ctx context.Context, userID, threadID string) (bool, error) {
	if userID == "" || threadID == "" {
		return false, status.Error(codes.InvalidArgument, "User ID and Thread ID are required")
	}

	return s.interactionRepo.IsThreadBookmarkedByUser(userID, threadID)
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

	threads, err := s.interactionRepo.GetUserBookmarks(userID, page, limit)
	if err != nil {
		return nil, 0, status.Errorf(codes.Internal, "Failed to retrieve bookmarks: %v", err)
	}

	// Count total bookmarks for pagination
	// This would need a separate method in the repository
	totalCount := int64(len(threads)) // Temporary solution

	return threads, totalCount, nil
}

// BookmarkReply bookmarks a reply
func (s *interactionService) BookmarkReply(ctx context.Context, userID, replyID string) error {
	if userID == "" || replyID == "" {
		return status.Error(codes.InvalidArgument, "User ID and Reply ID are required")
	}

	// Check if user has already bookmarked this reply
	hasBookmarked, err := s.interactionRepo.IsReplyBookmarkedByUser(userID, replyID)
	if err != nil {
		return status.Errorf(codes.Internal, "Failed to check if user has bookmarked reply: %v", err)
	}

	if hasBookmarked {
		return status.Error(codes.AlreadyExists, "User has already bookmarked this reply")
	}

	// Add bookmark
	if err := s.interactionRepo.BookmarkReply(userID, replyID); err != nil {
		return status.Errorf(codes.Internal, "Failed to bookmark reply: %v", err)
	}

	return nil
}

// RemoveReplyBookmark removes a bookmark from a reply
func (s *interactionService) RemoveReplyBookmark(ctx context.Context, userID, replyID string) error {
	if userID == "" || replyID == "" {
		return status.Error(codes.InvalidArgument, "User ID and Reply ID are required")
	}

	// Check if user has bookmarked this reply
	hasBookmarked, err := s.interactionRepo.IsReplyBookmarkedByUser(userID, replyID)
	if err != nil {
		return status.Errorf(codes.Internal, "Failed to check if user has bookmarked reply: %v", err)
	}

	if !hasBookmarked {
		return status.Error(codes.NotFound, "User has not bookmarked this reply")
	}

	// Remove bookmark
	if err := s.interactionRepo.RemoveReplyBookmark(userID, replyID); err != nil {
		return status.Errorf(codes.Internal, "Failed to remove reply bookmark: %v", err)
	}

	return nil
}

// HasUserBookmarkedReply checks if a user has bookmarked a reply
func (s *interactionService) HasUserBookmarkedReply(ctx context.Context, userID, replyID string) (bool, error) {
	if userID == "" || replyID == "" {
		return false, status.Error(codes.InvalidArgument, "User ID and Reply ID are required")
	}

	return s.interactionRepo.IsReplyBookmarkedByUser(userID, replyID)
}

// GetLikedThreadsByUserID retrieves thread IDs liked by a specific user with pagination
func (s *interactionService) GetLikedThreadsByUserID(ctx context.Context, userID string, page, limit int) ([]string, error) {
	if userID == "" {
		return nil, status.Error(codes.InvalidArgument, "User ID is required")
	}

	// Default pagination values if not provided
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	// Get thread IDs from repository
	threadIDs, err := s.interactionRepo.FindLikedThreadsByUserID(userID, page, limit)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to retrieve liked threads: %v", err)
	}

	return threadIDs, nil
}
