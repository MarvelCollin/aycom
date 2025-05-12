package handlers

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"

	"aycom/backend/api-gateway/config"
	threadProto "aycom/backend/proto/thread"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// ThreadServiceClient provides methods to interact with the Thread service
type ThreadServiceClient interface {
	// Thread operations
	CreateThread(userID, content string, mediaIDs []string) (string, error)
	GetThreadByID(threadID string, userID string) (*Thread, error)
	GetThreadsByUserID(userID string, requestingUserID string, page, limit int) ([]*Thread, error)
	GetAllThreads(userID string, page, limit int) ([]*Thread, error)
	UpdateThread(threadID, userID, content string) (*Thread, error)
	DeleteThread(threadID, userID string) error

	// Search operations
	SearchThreads(query string, userID string, page, limit int) ([]*Thread, error)

	// Interaction operations
	LikeThread(threadID, userID string) error
	UnlikeThread(threadID, userID string) error
	ReplyToThread(threadID, userID, content string, mediaIDs []string) (string, error)
	GetThreadReplies(threadID string, userID string, page, limit int) ([]*Thread, error)
	RepostThread(threadID, userID string) error
	RemoveRepost(threadID, userID string) error

	// Bookmark operations
	BookmarkThread(threadID, userID string) error
	RemoveBookmark(threadID, userID string) error
	GetUserBookmarks(userID string, page, limit int) ([]*Thread, error)
	SearchUserBookmarks(userID, query string, page, limit int) ([]*Thread, error)

	// New user content operations
	GetRepliesByUser(userID string, page, limit int) ([]*Thread, error)
	GetLikedThreadsByUser(userID string, page, limit int) ([]*Thread, error)
	GetMediaByUser(userID string, page, limit int) ([]Media, error)

	// Pinning operations
	PinThread(threadID, userID string) error
	UnpinThread(threadID, userID string) error
	PinReply(replyID, userID string) error
	UnpinReply(replyID, userID string) error

	// Trending operations
	GetTrendingHashtags(limit int) ([]string, error)
}

// Thread represents a thread (post) with all its metadata
type Thread struct {
	ID             string
	Content        string
	UserID         string
	Username       string
	DisplayName    string
	ProfilePicture string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	LikeCount      int
	ReplyCount     int
	RepostCount    int
	IsLiked        bool
	IsReposted     bool
	IsBookmarked   bool
	IsPinned       bool
	Media          []Media
	ParentID       string
}

// Media represents media attached to a thread
type Media struct {
	ID        string
	Type      string
	URL       string
	Thumbnail string
}

// GRPCThreadServiceClient is an implementation of ThreadServiceClient
// that communicates with the Thread service via gRPC
type GRPCThreadServiceClient struct {
	client threadProto.ThreadServiceClient
	conn   *grpc.ClientConn
}

// Global instance of the thread service client
var threadServiceClient ThreadServiceClient

// InitThreadServiceClient initializes the thread service client
func InitThreadServiceClient(cfg *config.Config) {
	log.Println("Initializing thread service client...")

	// Check if thread service address is configured
	if cfg.Services.ThreadService == "" {
		log.Println("Warning: Thread service address is not configured, using local implementation")
		threadServiceClient = &localThreadServiceClient{}
		return
	}

	log.Printf("Attempting to connect to Thread service at %s", cfg.Services.ThreadService)

	// Try direct connection
	conn, err := grpc.Dial(
		cfg.Services.ThreadService,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithTimeout(10*time.Second),
	)

	if err != nil {
		log.Printf("ERROR: Failed to connect to Thread service at %s: %v", cfg.Services.ThreadService, err)
		log.Println("Falling back to local implementation")
		threadServiceClient = &localThreadServiceClient{}
		return
	}

	log.Printf("Successfully connected to Thread service at %s", cfg.Services.ThreadService)

	// Initialize client with the connection
	grpcClient := threadProto.NewThreadServiceClient(conn)
	threadServiceClient = &GRPCThreadServiceClient{
		client: grpcClient,
		conn:   conn,
	}

	// Test the connection with a simple request
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, testErr := grpcClient.GetTrendingHashtags(ctx, &threadProto.GetTrendingHashtagsRequest{Limit: 1})
	if testErr != nil {
		log.Printf("WARNING: Thread service connection test failed: %v", testErr)
		log.Println("Connection established but service not responding correctly")
		log.Println("Will continue with gRPC implementation but service may not be fully operational")
	} else {
		log.Println("Thread service connection test successful - service is operational")
	}
}

// GetThreadServiceClient returns the thread service client instance
func GetThreadServiceClient() ThreadServiceClient {
	return threadServiceClient
}

// localThreadServiceClient implements ThreadServiceClient with local implementations
type localThreadServiceClient struct {
}

// Additional methods to implement ThreadServiceClient interface
// Just return placeholder implementations to satisfy the interface

// CreateThread implements ThreadServiceClient
func (c *localThreadServiceClient) CreateThread(userID, content string, mediaIDs []string) (string, error) {
	log.Printf("Mock: Creating thread for user %s", userID)
	return "mock-thread-id", nil
}

// GetThreadByID implements ThreadServiceClient
func (c *localThreadServiceClient) GetThreadByID(threadID string, userID string) (*Thread, error) {
	log.Printf("Mock: Getting thread %s for user %s", threadID, userID)
	// Return a mock thread
	return &Thread{
		ID:          threadID,
		Content:     "Mock thread content",
		UserID:      userID,
		Username:    "mockuser",
		DisplayName: "Mock User",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}, nil
}

// GetThreadsByUserID implements ThreadServiceClient
func (c *localThreadServiceClient) GetThreadsByUserID(userID string, requestingUserID string, page, limit int) ([]*Thread, error) {
	log.Printf("Mock: Getting threads for user %s", userID)
	// Return an empty array for now
	return []*Thread{}, nil
}

// GetAllThreads implements ThreadServiceClient
func (c *localThreadServiceClient) GetAllThreads(userID string, page, limit int) ([]*Thread, error) {
	log.Printf("Mock: Getting all threads page %d limit %d", page, limit)
	// Return an empty array for now
	return []*Thread{}, nil
}

// UpdateThread implements ThreadServiceClient
func (c *localThreadServiceClient) UpdateThread(threadID, userID, content string) (*Thread, error) {
	log.Printf("Mock: Updating thread %s for user %s", threadID, userID)
	// Return a mock thread
	return &Thread{
		ID:          threadID,
		Content:     content,
		UserID:      userID,
		Username:    "mockuser",
		DisplayName: "Mock User",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}, nil
}

// DeleteThread implements ThreadServiceClient
func (c *localThreadServiceClient) DeleteThread(threadID, userID string) error {
	log.Printf("Mock: Deleting thread %s for user %s", threadID, userID)
	return nil
}

// SearchThreads implements ThreadServiceClient
func (c *localThreadServiceClient) SearchThreads(query string, userID string, page, limit int) ([]*Thread, error) {
	log.Printf("Mock: Searching threads with query %s", query)
	// Return an empty array for now
	return []*Thread{}, nil
}

// LikeThread implements ThreadServiceClient
func (c *localThreadServiceClient) LikeThread(threadID, userID string) error {
	log.Printf("Mock: Liking thread %s for user %s", threadID, userID)
	return nil
}

// UnlikeThread implements ThreadServiceClient
func (c *localThreadServiceClient) UnlikeThread(threadID, userID string) error {
	log.Printf("Mock: Unliking thread %s for user %s", threadID, userID)
	return nil
}

// ReplyToThread implements ThreadServiceClient
func (c *localThreadServiceClient) ReplyToThread(threadID, userID, content string, mediaIDs []string) (string, error) {
	log.Printf("Mock: Replying to thread %s for user %s", threadID, userID)
	return "mock-reply-id", nil
}

// GetThreadReplies implements ThreadServiceClient
func (c *localThreadServiceClient) GetThreadReplies(threadID string, userID string, page, limit int) ([]*Thread, error) {
	log.Printf("Mock: Getting replies for thread %s", threadID)
	// Return an empty array for now
	return []*Thread{}, nil
}

// RepostThread implements ThreadServiceClient
func (c *localThreadServiceClient) RepostThread(threadID, userID string) error {
	log.Printf("Mock: Reposting thread %s for user %s", threadID, userID)
	return nil
}

// RemoveRepost implements ThreadServiceClient
func (c *localThreadServiceClient) RemoveRepost(threadID, userID string) error {
	log.Printf("Mock: Removing repost for thread %s for user %s", threadID, userID)
	return nil
}

// BookmarkThread implements ThreadServiceClient
func (c *localThreadServiceClient) BookmarkThread(threadID, userID string) error {
	log.Printf("Mock: Bookmarking thread %s for user %s", threadID, userID)
	return nil
}

// RemoveBookmark implements ThreadServiceClient
func (c *localThreadServiceClient) RemoveBookmark(threadID, userID string) error {
	log.Printf("Mock: Removing bookmark for thread %s for user %s", threadID, userID)
	return nil
}

// GetUserBookmarks implements ThreadServiceClient
func (c *localThreadServiceClient) GetUserBookmarks(userID string, page, limit int) ([]*Thread, error) {
	log.Printf("Mock: Getting bookmarks for user %s", userID)
	// Return an empty array for now
	return []*Thread{}, nil
}

// SearchUserBookmarks implements ThreadServiceClient
func (c *localThreadServiceClient) SearchUserBookmarks(userID, query string, page, limit int) ([]*Thread, error) {
	log.Printf("Mock: Searching bookmarks for user %s with query %s", userID, query)
	// Return an empty array for now
	return []*Thread{}, nil
}

// GetRepliesByUser implements ThreadServiceClient
func (c *localThreadServiceClient) GetRepliesByUser(userID string, page, limit int) ([]*Thread, error) {
	log.Printf("Mock: Getting replies for user %s", userID)
	// Return an empty array for now
	return []*Thread{}, nil
}

// GetLikedThreadsByUser implements ThreadServiceClient
func (c *localThreadServiceClient) GetLikedThreadsByUser(userID string, page, limit int) ([]*Thread, error) {
	log.Printf("Mock: Getting liked threads for user %s", userID)
	// Return an empty array for now
	return []*Thread{}, nil
}

// GetMediaByUser implements ThreadServiceClient
func (c *localThreadServiceClient) GetMediaByUser(userID string, page, limit int) ([]Media, error) {
	log.Printf("Mock: Getting media for user %s", userID)
	// Return an empty array for now
	return []Media{}, nil
}

// GetTrendingHashtags implements ThreadServiceClient
func (c *localThreadServiceClient) GetTrendingHashtags(limit int) ([]string, error) {
	log.Printf("Mock: Getting trending hashtags with limit %d", limit)
	// Return some mock trending hashtags
	return []string{"mock1", "mock2", "mock3"}, nil
}

// PinThread implements ThreadServiceClient for local mock
func (c *localThreadServiceClient) PinThread(threadID, userID string) error {
	log.Printf("Local implementation: Pinning thread %s for user %s", threadID, userID)
	// In a real implementation, we would update the database
	// For now, just return success
	return nil
}

// UnpinThread implements ThreadServiceClient for local mock
func (c *localThreadServiceClient) UnpinThread(threadID, userID string) error {
	log.Printf("Local implementation: Unpinning thread %s for user %s", threadID, userID)
	// In a real implementation, we would update the database
	// For now, just return success
	return nil
}

// PinReply implements ThreadServiceClient for local mock
func (c *localThreadServiceClient) PinReply(replyID, userID string) error {
	log.Printf("Local implementation: Pinning reply %s for user %s", replyID, userID)
	// In a real implementation, we would update the database
	// For now, just return success
	return nil
}

// UnpinReply implements ThreadServiceClient for local mock
func (c *localThreadServiceClient) UnpinReply(replyID, userID string) error {
	log.Printf("Local implementation: Unpinning reply %s for user %s", replyID, userID)
	// In a real implementation, we would update the database
	// For now, just return success
	return nil
}

// CreateThread implements ThreadServiceClient
func (c *GRPCThreadServiceClient) CreateThread(userID, content string, mediaIDs []string) (string, error) {
	if c.client == nil {
		return "", fmt.Errorf("thread service client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create Media objects from mediaIDs
	mediaObjects := make([]*threadProto.Media, len(mediaIDs))
	for i, mediaID := range mediaIDs {
		mediaObjects[i] = &threadProto.Media{
			Id:   mediaID,
			Url:  "", // URL will be filled by the thread service
			Type: "",
		}
	}

	resp, err := c.client.CreateThread(ctx, &threadProto.CreateThreadRequest{
		UserId:  userID,
		Content: content,
		Media:   mediaObjects,
	})
	if err != nil {
		return "", err
	}

	return resp.Thread.Id, nil
}

// GetThreadByID implements ThreadServiceClient
func (c *GRPCThreadServiceClient) GetThreadByID(threadID string, userID string) (*Thread, error) {
	if c.client == nil {
		return nil, fmt.Errorf("thread service client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Add user ID to the context metadata to pass to service for auth checks
	if userID != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "user_id", userID)
	}

	resp, err := c.client.GetThreadById(ctx, &threadProto.GetThreadRequest{
		ThreadId: threadID,
		// Note: ThreadId is the only field in GetThreadRequest according to the proto
	})
	if err != nil {
		return nil, err
	}

	// Convert proto thread to Thread struct
	thread := convertProtoToThread(resp.Thread)

	// Set bookmarked status based on response
	if resp.BookmarkedByUser {
		thread.IsBookmarked = true
	}

	return thread, nil
}

// GetThreadsByUserID implements ThreadServiceClient
func (c *GRPCThreadServiceClient) GetThreadsByUserID(userID string, requestingUserID string, page, limit int) ([]*Thread, error) {
	if c.client == nil {
		return nil, fmt.Errorf("thread service client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Add requesting user ID to the context metadata to pass to service for auth checks
	if requestingUserID != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "user_id", requestingUserID)
	}

	resp, err := c.client.GetThreadsByUser(ctx, &threadProto.GetThreadsByUserRequest{
		UserId: userID,
		Page:   int32(page),
		Limit:  int32(limit),
		// Note: RequestingUserId isn't in the proto, so we're not including it
	})
	if err != nil {
		return nil, err
	}

	threads := make([]*Thread, len(resp.Threads))
	for i, t := range resp.Threads {
		threads[i] = convertProtoToThread(t)
	}

	return threads, nil
}

// GetAllThreads implements ThreadServiceClient
func (c *GRPCThreadServiceClient) GetAllThreads(userID string, page, limit int) ([]*Thread, error) {
	if c.client == nil {
		return nil, fmt.Errorf("thread service client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Add user ID to the context metadata to pass to service for auth checks
	if userID != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "user_id", userID)
	}

	resp, err := c.client.GetAllThreads(ctx, &threadProto.GetAllThreadsRequest{
		Page:  int32(page),
		Limit: int32(limit),
		// Note: UserId isn't in the proto for GetAllThreadsRequest
	})
	if err != nil {
		return nil, err
	}

	threads := make([]*Thread, len(resp.Threads))
	for i, t := range resp.Threads {
		threads[i] = convertProtoToThread(t)
	}

	return threads, nil
}

// UpdateThread implements ThreadServiceClient
func (c *GRPCThreadServiceClient) UpdateThread(threadID, userID, content string) (*Thread, error) {
	if c.client == nil {
		return nil, fmt.Errorf("thread service client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := c.client.UpdateThread(ctx, &threadProto.UpdateThreadRequest{
		ThreadId: threadID,
		UserId:   userID,
		Content:  content,
	})
	if err != nil {
		return nil, err
	}

	return convertProtoToThread(resp.Thread), nil
}

// DeleteThread implements ThreadServiceClient
func (c *GRPCThreadServiceClient) DeleteThread(threadID, userID string) error {
	if c.client == nil {
		return fmt.Errorf("thread service client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := c.client.DeleteThread(ctx, &threadProto.DeleteThreadRequest{
		ThreadId: threadID,
		UserId:   userID,
	})
	return err
}

// SearchThreads implements ThreadServiceClient
func (c *GRPCThreadServiceClient) SearchThreads(query string, userID string, page, limit int) ([]*Thread, error) {
	// This would require a new method in the Thread service proto
	// For now, we can implement a basic search using GetAllThreads and filtering
	threads, err := c.GetAllThreads(userID, page, limit)
	if err != nil {
		return nil, err
	}

	if query == "" {
		return threads, nil
	}

	// Filter threads that contain the query string
	var filteredThreads []*Thread
	queryLower := strings.ToLower(query)
	for _, thread := range threads {
		if strings.Contains(strings.ToLower(thread.Content), queryLower) {
			filteredThreads = append(filteredThreads, thread)
		}
	}

	return filteredThreads, nil
}

// LikeThread implements ThreadServiceClient
func (c *GRPCThreadServiceClient) LikeThread(threadID, userID string) error {
	if c.client == nil {
		return fmt.Errorf("thread service client not initialized")
	}

	// Maximum retry attempts
	maxRetries := 3
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // Increased timeout

		// Add user ID to context metadata
		ctx = metadata.AppendToOutgoingContext(ctx, "user_id", userID)

		log.Printf("Attempt %d: Liking thread %s for user %s", attempt, threadID, userID)

		_, err := c.client.LikeThread(ctx, &threadProto.LikeThreadRequest{
			ThreadId: threadID,
			UserId:   userID,
		})

		cancel() // Cancel context immediately after request

		if err == nil {
			log.Printf("Successfully liked thread %s for user %s", threadID, userID)

			// Verify like was actually created
			verifyCtx, verifyCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer verifyCancel()

			// Pass user ID via metadata for verification
			verifyCtx = metadata.AppendToOutgoingContext(verifyCtx, "user_id", userID)

			resp, verifyErr := c.client.GetThreadById(verifyCtx, &threadProto.GetThreadRequest{
				ThreadId: threadID,
			})

			if verifyErr != nil {
				log.Printf("Warning: Verification check error after liking thread: %v", verifyErr)
			} else if resp != nil && resp.LikedByUser {
				log.Printf("Verified thread %s is liked by user %s", threadID, userID)
			} else {
				log.Printf("Warning: Thread %s shows as NOT liked after operation", threadID)
			}

			return nil
		}

		lastErr = err
		log.Printf("Error liking thread (attempt %d): %v", attempt, err)

		// Wait before retrying
		time.Sleep(time.Duration(attempt*500) * time.Millisecond)
	}

	log.Printf("Failed to like thread after %d attempts: %v", maxRetries, lastErr)
	return lastErr
}

// UnlikeThread implements ThreadServiceClient
func (c *GRPCThreadServiceClient) UnlikeThread(threadID, userID string) error {
	if c.client == nil {
		return fmt.Errorf("thread service client not initialized")
	}

	// Maximum retry attempts
	maxRetries := 3
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // Increased timeout

		// Add user ID to context metadata
		ctx = metadata.AppendToOutgoingContext(ctx, "user_id", userID)

		log.Printf("Attempt %d: Unliking thread %s for user %s", attempt, threadID, userID)

		_, err := c.client.UnlikeThread(ctx, &threadProto.UnlikeThreadRequest{
			ThreadId: threadID,
			UserId:   userID,
		})

		cancel() // Cancel context immediately after request

		if err == nil {
			log.Printf("Successfully unliked thread %s for user %s", threadID, userID)

			// Verify like was actually removed
			verifyCtx, verifyCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer verifyCancel()

			// Pass user ID via metadata for verification
			verifyCtx = metadata.AppendToOutgoingContext(verifyCtx, "user_id", userID)

			resp, verifyErr := c.client.GetThreadById(verifyCtx, &threadProto.GetThreadRequest{
				ThreadId: threadID,
			})

			if verifyErr != nil {
				log.Printf("Warning: Verification check error after unliking thread: %v", verifyErr)
			} else if resp != nil && !resp.LikedByUser {
				log.Printf("Verified thread %s is not liked by user %s", threadID, userID)
			} else {
				log.Printf("Warning: Thread %s still shows as liked after unlike operation", threadID)
			}

			return nil
		}

		lastErr = err
		log.Printf("Error unliking thread (attempt %d): %v", attempt, err)

		// Wait before retrying
		time.Sleep(time.Duration(attempt*500) * time.Millisecond)
	}

	log.Printf("Failed to unlike thread after %d attempts: %v", maxRetries, lastErr)
	return lastErr
}

// ReplyToThread implements ThreadServiceClient
func (c *GRPCThreadServiceClient) ReplyToThread(threadID, userID, content string, mediaIDs []string) (string, error) {
	if c.client == nil {
		return "", fmt.Errorf("thread service client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create Media objects from mediaIDs
	mediaObjects := make([]*threadProto.Media, len(mediaIDs))
	for i, mediaID := range mediaIDs {
		mediaObjects[i] = &threadProto.Media{
			Id:   mediaID,
			Url:  "", // URL will be filled by the thread service
			Type: "",
		}
	}

	resp, err := c.client.CreateReply(ctx, &threadProto.CreateReplyRequest{
		ThreadId: threadID,
		UserId:   userID,
		Content:  content,
		Media:    mediaObjects,
	})
	if err != nil {
		return "", err
	}

	return resp.Reply.Id, nil
}

// GetThreadReplies implements ThreadServiceClient
func (c *GRPCThreadServiceClient) GetThreadReplies(threadID string, userID string, page, limit int) ([]*Thread, error) {
	if c.client == nil {
		return nil, fmt.Errorf("thread service client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Add user ID to the context metadata to pass to service for auth checks
	if userID != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "user_id", userID)
	}

	resp, err := c.client.GetRepliesByThread(ctx, &threadProto.GetRepliesByThreadRequest{
		ThreadId: threadID,
		Page:     int32(page),
		Limit:    int32(limit),
	})
	if err != nil {
		return nil, err
	}

	// Process the replies
	replies := make([]*Thread, len(resp.Replies))
	for i, replyResp := range resp.Replies {
		if replyResp.Reply == nil {
			// Skip invalid replies
			continue
		}

		reply := replyResp.Reply

		// Get user data
		username := "anonymous"
		displayName := "User"
		profilePicURL := "https://secure.gravatar.com/avatar/0?d=mp" // Default avatar

		// Use user data from response if available
		if replyResp.User != nil {
			user := replyResp.User
			if user.Username != "" {
				username = user.Username
			}
			if user.Name != "" {
				displayName = user.Name
			}
			if user.ProfilePictureUrl != "" {
				profilePicURL = user.ProfilePictureUrl
			}
		}

		// Create a Thread object with the Reply data
		replies[i] = &Thread{
			ID:             reply.Id,
			Content:        reply.Content,
			UserID:         reply.UserId,
			Username:       username,
			DisplayName:    displayName,
			ProfilePicture: profilePicURL,
			CreatedAt:      reply.CreatedAt.AsTime(),
			UpdatedAt:      reply.UpdatedAt.AsTime(),
			LikeCount:      int(replyResp.LikesCount),
			ReplyCount:     0, // Replies don't have nested replies in this implementation
			IsLiked:        replyResp.LikedByUser,
			IsBookmarked:   replyResp.BookmarkedByUser,
			ParentID:       threadID,
		}

		// Convert media
		if len(reply.Media) > 0 {
			replies[i].Media = make([]Media, len(reply.Media))
			for j, m := range reply.Media {
				replies[i].Media[j] = Media{
					ID:   m.Id,
					Type: m.Type,
					URL:  m.Url,
				}
			}
		}
	}

	return replies, nil
}

// RepostThread implements ThreadServiceClient
func (c *GRPCThreadServiceClient) RepostThread(threadID, userID string) error {
	if c.client == nil {
		return fmt.Errorf("thread service client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := c.client.RepostThread(ctx, &threadProto.RepostThreadRequest{
		ThreadId: threadID,
		UserId:   userID,
	})
	return err
}

// RemoveRepost implements ThreadServiceClient
func (c *GRPCThreadServiceClient) RemoveRepost(threadID, userID string) error {
	if c.client == nil {
		return fmt.Errorf("thread service client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := c.client.RemoveRepost(ctx, &threadProto.RemoveRepostRequest{
		ThreadId: threadID,
		UserId:   userID,
	})
	return err
}

// BookmarkThread implements ThreadServiceClient
func (c *GRPCThreadServiceClient) BookmarkThread(threadID, userID string) error {
	if c.client == nil {
		return fmt.Errorf("thread service client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Debug the context and request
	log.Printf("Sending BookmarkThread request to thread service - threadID: %s, userID: %s", threadID, userID)

	// Add user_id to metadata - this matches LikeThread which works
	ctx = metadata.AppendToOutgoingContext(ctx, "user_id", userID)

	_, err := c.client.BookmarkThread(ctx, &threadProto.BookmarkThreadRequest{
		ThreadId: threadID,
		UserId:   userID,
	})

	if err != nil {
		log.Printf("Error during BookmarkThread call to thread service: %v", err)
		return err
	}

	log.Printf("Successfully sent BookmarkThread request to thread service")
	return nil
}

// RemoveBookmark implements ThreadServiceClient
func (c *GRPCThreadServiceClient) RemoveBookmark(threadID, userID string) error {
	if c.client == nil {
		return fmt.Errorf("thread service client not initialized")
	}

	// Maximum retry attempts
	maxRetries := 3
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // Increased timeout

		// Add user ID to context metadata
		ctx = metadata.AppendToOutgoingContext(ctx, "user_id", userID)

		log.Printf("Attempt %d: Removing bookmark for thread %s by user %s", attempt, threadID, userID)

		_, err := c.client.RemoveBookmark(ctx, &threadProto.RemoveBookmarkRequest{
			ThreadId: threadID,
			UserId:   userID,
		})

		cancel() // Cancel context immediately after request

		if err == nil {
			log.Printf("Successfully removed bookmark for thread %s by user %s", threadID, userID)

			// Verify bookmark was actually removed
			verifyCtx, verifyCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer verifyCancel()

			// Pass user ID via metadata for verification
			verifyCtx = metadata.AppendToOutgoingContext(verifyCtx, "user_id", userID)

			resp, verifyErr := c.client.GetThreadById(verifyCtx, &threadProto.GetThreadRequest{
				ThreadId: threadID,
			})

			if verifyErr != nil {
				log.Printf("Warning: Verification check error after removing bookmark: %v", verifyErr)
			} else if resp != nil && !resp.BookmarkedByUser {
				log.Printf("Verified thread %s is not bookmarked by user %s", threadID, userID)
			} else {
				log.Printf("Warning: Thread %s still shows as bookmarked after removal", threadID)
			}

			return nil
		}

		lastErr = err
		log.Printf("Error removing bookmark (attempt %d): %v", attempt, err)

		// Wait before retrying
		time.Sleep(time.Duration(attempt*500) * time.Millisecond)
	}

	log.Printf("Failed to remove bookmark after %d attempts: %v", maxRetries, lastErr)
	return lastErr
}

// GetUserBookmarks implements ThreadServiceClient
func (c *GRPCThreadServiceClient) GetUserBookmarks(userID string, page, limit int) ([]*Thread, error) {
	if c.client == nil {
		return nil, fmt.Errorf("thread service client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Set up metadata with user ID for authentication/tracking
	md := metadata.New(map[string]string{
		"user_id": userID,
	})
	ctx = metadata.NewOutgoingContext(ctx, md)

	// Try to use the GetBookmarksByUser method which should be available in the thread service
	bookmarksMethod := reflect.ValueOf(c.client).MethodByName("GetBookmarksByUser")
	if bookmarksMethod.IsValid() {
		// Set up arguments
		ctxArg := reflect.ValueOf(ctx)
		reqArg := reflect.New(bookmarksMethod.Type().In(1).Elem()).Interface()

		// Set fields via reflection
		reqVal := reflect.ValueOf(reqArg).Elem()
		reqVal.FieldByName("UserId").SetString(userID)
		reqVal.FieldByName("Page").SetInt(int64(page))
		reqVal.FieldByName("Limit").SetInt(int64(limit))

		// Invoke method
		results := bookmarksMethod.Call([]reflect.Value{ctxArg, reflect.ValueOf(reqArg)})
		if !results[1].IsNil() {
			return nil, results[1].Interface().(error)
		}

		resp := results[0].Interface()
		threadsResp := resp.(*threadProto.ThreadsResponse)

		// Convert ThreadResponse objects to Thread structs
		threads := make([]*Thread, len(threadsResp.Threads))
		for i, t := range threadsResp.Threads {
			thread := convertProtoToThread(t)
			// Mark as bookmarked since these are bookmarks
			thread.IsBookmarked = true
			threads[i] = thread
		}

		log.Printf("Successfully retrieved %d bookmarks using GetBookmarksByUser", len(threads))
		return threads, nil
	}

	log.Printf("GetBookmarksByUser method not found, falling back to original approach")

	// Call the method using reflection since it might not be directly available
	method := reflect.ValueOf(c.client).MethodByName("GetThreadsByUser")
	if !method.IsValid() {
		log.Printf("Method GetThreadsByUser not found on client, falling back to mock data")
		return []*Thread{}, nil
	}

	// Set up arguments
	ctxArg := reflect.ValueOf(ctx)
	reqArg := reflect.New(method.Type().In(1).Elem()).Interface()

	// Set fields via reflection
	reqVal := reflect.ValueOf(reqArg).Elem()
	reqVal.FieldByName("UserId").SetString(userID)
	reqVal.FieldByName("Page").SetInt(int64(page))
	reqVal.FieldByName("Limit").SetInt(int64(limit))

	// Invoke method
	results := method.Call([]reflect.Value{ctxArg, reflect.ValueOf(reqArg)})
	if !results[1].IsNil() {
		return nil, results[1].Interface().(error)
	}

	resp := results[0].Interface()
	threadsResp := resp.(*threadProto.ThreadsResponse)

	// Convert ThreadResponse objects to Thread structs
	threads := make([]*Thread, len(threadsResp.Threads))
	for i, t := range threadsResp.Threads {
		thread := convertProtoToThread(t)
		// Mark as bookmarked since these are bookmarks
		thread.IsBookmarked = true
		threads[i] = thread
	}

	return threads, nil
}

// SearchUserBookmarks implements ThreadServiceClient
func (c *GRPCThreadServiceClient) SearchUserBookmarks(userID, query string, page, limit int) ([]*Thread, error) {
	// Note: This is a placeholder. The Thread service would need to implement a method to search bookmarks
	// For now, we'll just return an empty list
	return []*Thread{}, nil
}

// GetTrendingHashtags implements ThreadServiceClient
func (c *GRPCThreadServiceClient) GetTrendingHashtags(limit int) ([]string, error) {
	if c.client == nil {
		return nil, fmt.Errorf("thread service client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := c.client.GetTrendingHashtags(ctx, &threadProto.GetTrendingHashtagsRequest{
		Limit: int32(limit),
	})
	if err != nil {
		return nil, err
	}

	// Convert HashtagResponse objects to strings
	hashtags := make([]string, len(resp.Hashtags))
	for i, h := range resp.Hashtags {
		hashtags[i] = h.Name
	}

	return hashtags, nil
}

// PinThread implements ThreadServiceClient for GRPC implementation
func (c *GRPCThreadServiceClient) PinThread(threadID, userID string) error {
	if c.client == nil {
		return fmt.Errorf("thread service client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Call the method using reflection
	method := reflect.ValueOf(c.client).MethodByName("PinThread")
	if !method.IsValid() {
		return fmt.Errorf("method PinThread not found on client")
	}

	// Set up arguments
	ctxArg := reflect.ValueOf(ctx)
	reqArg := reflect.New(method.Type().In(1).Elem()).Interface()

	// Set fields via reflection
	reqVal := reflect.ValueOf(reqArg).Elem()
	reqVal.FieldByName("ThreadId").SetString(threadID)
	reqVal.FieldByName("UserId").SetString(userID)

	// Invoke method
	results := method.Call([]reflect.Value{ctxArg, reflect.ValueOf(reqArg)})
	if !results[1].IsNil() {
		return results[1].Interface().(error)
	}

	return nil
}

// UnpinThread implements ThreadServiceClient for GRPC implementation
func (c *GRPCThreadServiceClient) UnpinThread(threadID, userID string) error {
	if c.client == nil {
		return fmt.Errorf("thread service client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Call the method using reflection
	method := reflect.ValueOf(c.client).MethodByName("UnpinThread")
	if !method.IsValid() {
		return fmt.Errorf("method UnpinThread not found on client")
	}

	// Set up arguments
	ctxArg := reflect.ValueOf(ctx)
	reqArg := reflect.New(method.Type().In(1).Elem()).Interface()

	// Set fields via reflection
	reqVal := reflect.ValueOf(reqArg).Elem()
	reqVal.FieldByName("ThreadId").SetString(threadID)
	reqVal.FieldByName("UserId").SetString(userID)

	// Invoke method
	results := method.Call([]reflect.Value{ctxArg, reflect.ValueOf(reqArg)})
	if !results[1].IsNil() {
		return results[1].Interface().(error)
	}

	return nil
}

// PinReply implements ThreadServiceClient for GRPC implementation
func (c *GRPCThreadServiceClient) PinReply(replyID, userID string) error {
	if c.client == nil {
		return fmt.Errorf("thread service client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Call the method using reflection
	method := reflect.ValueOf(c.client).MethodByName("PinReply")
	if !method.IsValid() {
		return fmt.Errorf("method PinReply not found on client")
	}

	// Set up arguments
	ctxArg := reflect.ValueOf(ctx)
	reqArg := reflect.New(method.Type().In(1).Elem()).Interface()

	// Set fields via reflection
	reqVal := reflect.ValueOf(reqArg).Elem()
	reqVal.FieldByName("ReplyId").SetString(replyID)
	reqVal.FieldByName("UserId").SetString(userID)

	// Invoke method
	results := method.Call([]reflect.Value{ctxArg, reflect.ValueOf(reqArg)})
	if !results[1].IsNil() {
		return results[1].Interface().(error)
	}

	return nil
}

// UnpinReply implements ThreadServiceClient for GRPC implementation
func (c *GRPCThreadServiceClient) UnpinReply(replyID, userID string) error {
	if c.client == nil {
		return fmt.Errorf("thread service client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Call the method using reflection
	method := reflect.ValueOf(c.client).MethodByName("UnpinReply")
	if !method.IsValid() {
		return fmt.Errorf("method UnpinReply not found on client")
	}

	// Set up arguments
	ctxArg := reflect.ValueOf(ctx)
	reqArg := reflect.New(method.Type().In(1).Elem()).Interface()

	// Set fields via reflection
	reqVal := reflect.ValueOf(reqArg).Elem()
	reqVal.FieldByName("ReplyId").SetString(replyID)
	reqVal.FieldByName("UserId").SetString(userID)

	// Invoke method
	results := method.Call([]reflect.Value{ctxArg, reflect.ValueOf(reqArg)})
	if !results[1].IsNil() {
		return results[1].Interface().(error)
	}

	return nil
}

// GetLikedThreadsByUser implements ThreadServiceClient for GRPC implementation
func (c *GRPCThreadServiceClient) GetLikedThreadsByUser(userID string, page, limit int) ([]*Thread, error) {
	if c.client == nil {
		return nil, fmt.Errorf("thread service client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Call the method using reflection
	method := reflect.ValueOf(c.client).MethodByName("GetLikedThreadsByUser")
	if !method.IsValid() {
		return nil, fmt.Errorf("method GetLikedThreadsByUser not found on client")
	}

	// Set up arguments
	ctxArg := reflect.ValueOf(ctx)
	reqArg := reflect.New(method.Type().In(1).Elem()).Interface()

	// Set fields via reflection
	reqVal := reflect.ValueOf(reqArg).Elem()
	reqVal.FieldByName("UserId").SetString(userID)
	reqVal.FieldByName("Page").SetInt(int64(page))
	reqVal.FieldByName("Limit").SetInt(int64(limit))

	// Invoke method
	results := method.Call([]reflect.Value{ctxArg, reflect.ValueOf(reqArg)})
	if !results[1].IsNil() {
		return nil, results[1].Interface().(error)
	}

	// Process the response
	resp := results[0].Interface()

	// Extract threads from response using reflection
	threadsVal := reflect.ValueOf(resp).Elem().FieldByName("Threads")

	// Convert to our internal thread representation
	threads := make([]*Thread, threadsVal.Len())
	for i := 0; i < threadsVal.Len(); i++ {
		threadResp := threadsVal.Index(i).Interface()
		threads[i] = convertProtoToThread(threadResp)
	}

	return threads, nil
}

// GetMediaByUser implements ThreadServiceClient for GRPC implementation
func (c *GRPCThreadServiceClient) GetMediaByUser(userID string, page, limit int) ([]Media, error) {
	if c.client == nil {
		return nil, fmt.Errorf("thread service client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Call the method using reflection
	method := reflect.ValueOf(c.client).MethodByName("GetMediaByUser")
	if !method.IsValid() {
		return nil, fmt.Errorf("method GetMediaByUser not found on client")
	}

	// Set up arguments
	ctxArg := reflect.ValueOf(ctx)
	reqArg := reflect.New(method.Type().In(1).Elem()).Interface()

	// Set fields via reflection
	reqVal := reflect.ValueOf(reqArg).Elem()
	reqVal.FieldByName("UserId").SetString(userID)
	reqVal.FieldByName("Page").SetInt(int64(page))
	reqVal.FieldByName("Limit").SetInt(int64(limit))

	// Invoke method
	results := method.Call([]reflect.Value{ctxArg, reflect.ValueOf(reqArg)})
	if !results[1].IsNil() {
		return nil, results[1].Interface().(error)
	}

	// Process the response
	resp := results[0].Interface()

	// Extract media from response using reflection
	mediaVal := reflect.ValueOf(resp).Elem().FieldByName("Media")

	// Convert to our internal media representation
	media := make([]Media, mediaVal.Len())
	for i := 0; i < mediaVal.Len(); i++ {
		m := mediaVal.Index(i).Interface()

		// Extract fields using reflection
		mVal := reflect.ValueOf(m).Elem()
		media[i] = Media{
			ID:   mVal.FieldByName("Id").String(),
			URL:  mVal.FieldByName("Url").String(),
			Type: mVal.FieldByName("Type").String(),
		}
	}

	return media, nil
}

// GetRepliesByUser implements ThreadServiceClient for GRPC implementation
func (c *GRPCThreadServiceClient) GetRepliesByUser(userID string, page, limit int) ([]*Thread, error) {
	if c.client == nil {
		return nil, fmt.Errorf("thread service client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Call the method using reflection
	method := reflect.ValueOf(c.client).MethodByName("GetRepliesByUser")
	if !method.IsValid() {
		return nil, fmt.Errorf("method GetRepliesByUser not found on client")
	}

	// Set up arguments
	ctxArg := reflect.ValueOf(ctx)
	reqArg := reflect.New(method.Type().In(1).Elem()).Interface()

	// Set fields via reflection
	reqVal := reflect.ValueOf(reqArg).Elem()
	reqVal.FieldByName("UserId").SetString(userID)
	reqVal.FieldByName("Page").SetInt(int64(page))
	reqVal.FieldByName("Limit").SetInt(int64(limit))

	// Invoke method
	results := method.Call([]reflect.Value{ctxArg, reflect.ValueOf(reqArg)})
	if !results[1].IsNil() {
		return nil, results[1].Interface().(error)
	}

	// Extract replies using reflection
	resp := results[0].Interface()
	repliesVal := reflect.ValueOf(resp).Elem().FieldByName("Replies")

	// Convert to our thread type
	replies := make([]*Thread, repliesVal.Len())
	for i := 0; i < repliesVal.Len(); i++ {
		r := repliesVal.Index(i).Interface()
		rVal := reflect.ValueOf(r).Elem()

		// Get the reply and user objects
		replyObj := rVal.FieldByName("Reply").Interface()
		userObj := rVal.FieldByName("User").Interface()

		// Extract values using reflection
		replyVal := reflect.ValueOf(replyObj).Elem()
		userVal := reflect.ValueOf(userObj).Elem()

		isPinned := false
		if replyVal.FieldByName("IsPinned").IsValid() && !replyVal.FieldByName("IsPinned").IsNil() {
			isPinned = replyVal.FieldByName("IsPinned").Elem().Bool()
		}

		// Create a thread from reply data
		replies[i] = &Thread{
			ID:             replyVal.FieldByName("Id").String(),
			Content:        replyVal.FieldByName("Content").String(),
			UserID:         replyVal.FieldByName("UserId").String(),
			Username:       userVal.FieldByName("Username").String(),
			DisplayName:    userVal.FieldByName("Name").String(),
			ProfilePicture: userVal.FieldByName("ProfilePictureUrl").String(),
			CreatedAt:      replyVal.FieldByName("CreatedAt").Interface().(interface{ AsTime() time.Time }).AsTime(),
			UpdatedAt:      replyVal.FieldByName("UpdatedAt").Interface().(interface{ AsTime() time.Time }).AsTime(),
			LikeCount:      int(rVal.FieldByName("LikesCount").Int()),
			IsLiked:        rVal.FieldByName("LikedByUser").Bool(),
			ParentID:       replyVal.FieldByName("ThreadId").String(),
			IsPinned:       isPinned,
		}

		// Convert media if present
		mediaField := replyVal.FieldByName("Media")
		if mediaField.IsValid() && mediaField.Len() > 0 {
			media := make([]Media, mediaField.Len())
			for j := 0; j < mediaField.Len(); j++ {
				m := mediaField.Index(j).Interface()
				mVal := reflect.ValueOf(m).Elem()
				media[j] = Media{
					ID:   mVal.FieldByName("Id").String(),
					URL:  mVal.FieldByName("Url").String(),
					Type: mVal.FieldByName("Type").String(),
				}
			}
			replies[i].Media = media
		}
	}

	return replies, nil
}

// Helper function to convert proto Thread to our internal Thread type
func convertProtoToThread(t any) *Thread {
	if t == nil {
		return nil
	}

	// Create a default thread
	thread := &Thread{
		ID:          "unknown",
		Content:     "",
		UserID:      "",
		Username:    "anonymous",
		DisplayName: "User",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Try to convert the ThreadResponse
	if tr, ok := t.(*threadProto.ThreadResponse); ok && tr != nil {
		if tr.Thread != nil {
			thread.ID = tr.Thread.Id
			thread.Content = tr.Thread.Content
			thread.UserID = tr.Thread.UserId
			thread.CreatedAt = tr.Thread.CreatedAt.AsTime()
			thread.UpdatedAt = tr.Thread.UpdatedAt.AsTime()
			thread.LikeCount = int(tr.LikesCount)
			thread.ReplyCount = int(tr.RepliesCount)
			thread.RepostCount = int(tr.RepostsCount)
			thread.IsLiked = tr.LikedByUser
			thread.IsReposted = tr.RepostedByUser
			thread.IsBookmarked = tr.BookmarkedByUser

			// Set is_pinned if available
			if tr.Thread.IsPinned != nil {
				thread.IsPinned = *tr.Thread.IsPinned
			}

			// Convert media
			if len(tr.Thread.Media) > 0 {
				thread.Media = make([]Media, len(tr.Thread.Media))
				for i, m := range tr.Thread.Media {
					thread.Media[i] = Media{
						ID:   m.Id,
						Type: m.Type,
						URL:  m.Url,
					}
				}
			}
		}

		// Set user data if available
		if tr.User != nil {
			thread.Username = tr.User.Username
			thread.DisplayName = tr.User.Name
			thread.ProfilePicture = tr.User.ProfilePictureUrl
		}
	}

	return thread
}
