package handlers

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"aycom/backend/api-gateway/config"
	threadProto "aycom/backend/proto/thread"
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
}

// Global instance of the thread service client
var threadServiceClient ThreadServiceClient

// InitThreadServiceClient initializes the thread service client
func InitThreadServiceClient(cfg *config.Config) {
	log.Println("Thread service client initialization skipped - using direct client")
	// Actual implementation depends on your thread service architecture
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

	resp, err := c.client.GetThreadById(ctx, &threadProto.GetThreadRequest{
		ThreadId: threadID,
		// Note: ThreadId is the only field in GetThreadRequest according to the proto
	})
	if err != nil {
		return nil, err
	}

	// Convert proto thread to Thread struct
	return convertProtoToThread(resp.Thread), nil
}

// GetThreadsByUserID implements ThreadServiceClient
func (c *GRPCThreadServiceClient) GetThreadsByUserID(userID string, requestingUserID string, page, limit int) ([]*Thread, error) {
	if c.client == nil {
		return nil, fmt.Errorf("thread service client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := c.client.LikeThread(ctx, &threadProto.LikeThreadRequest{
		ThreadId: threadID,
		UserId:   userID,
	})
	return err
}

// UnlikeThread implements ThreadServiceClient
func (c *GRPCThreadServiceClient) UnlikeThread(threadID, userID string) error {
	if c.client == nil {
		return fmt.Errorf("thread service client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := c.client.UnlikeThread(ctx, &threadProto.UnlikeThreadRequest{
		ThreadId: threadID,
		UserId:   userID,
	})
	return err
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

	_, err := c.client.BookmarkThread(ctx, &threadProto.BookmarkThreadRequest{
		ThreadId: threadID,
		UserId:   userID,
	})
	return err
}

// RemoveBookmark implements ThreadServiceClient
func (c *GRPCThreadServiceClient) RemoveBookmark(threadID, userID string) error {
	if c.client == nil {
		return fmt.Errorf("thread service client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := c.client.RemoveBookmark(ctx, &threadProto.RemoveBookmarkRequest{
		ThreadId: threadID,
		UserId:   userID,
	})
	return err
}

// GetUserBookmarks implements ThreadServiceClient
func (c *GRPCThreadServiceClient) GetUserBookmarks(userID string, page, limit int) ([]*Thread, error) {
	// Note: This is a placeholder. The Thread service would need to implement a method to fetch bookmarks
	// For now, we'll just return an empty list
	return []*Thread{}, nil
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

// GetRepliesByUser implements ThreadServiceClient
func (c *GRPCThreadServiceClient) GetRepliesByUser(userID string, page, limit int) ([]*Thread, error) {
	if c.client == nil {
		return nil, fmt.Errorf("thread service client not initialized")
	}

	// In a real implementation, this context would be used for the gRPC call
	_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// This is a mock implementation - in a real system, you would call the actual
	// thread service with a method like GetRepliesByUser that returns this info
	// For now, we'll just return an empty list to avoid errors
	return []*Thread{}, nil
}

// GetLikedThreadsByUser implements ThreadServiceClient
func (c *GRPCThreadServiceClient) GetLikedThreadsByUser(userID string, page, limit int) ([]*Thread, error) {
	if c.client == nil {
		return nil, fmt.Errorf("thread service client not initialized")
	}

	// In a real implementation, this context would be used for the gRPC call
	_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// This is a mock implementation - in a real system, you would call the actual
	// thread service with a method like GetLikedThreadsByUser
	// For now, we'll just return an empty list to avoid errors
	return []*Thread{}, nil
}

// GetMediaByUser implements ThreadServiceClient
func (c *GRPCThreadServiceClient) GetMediaByUser(userID string, page, limit int) ([]Media, error) {
	if c.client == nil {
		return nil, fmt.Errorf("thread service client not initialized")
	}

	// In a real implementation, this context would be used for the gRPC call
	_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// This is a mock implementation
	// For now, we'll just return an empty list to avoid errors
	return []Media{}, nil
}

// PinThread implements ThreadServiceClient
func (c *GRPCThreadServiceClient) PinThread(threadID, userID string) error {
	if c.client == nil {
		return fmt.Errorf("thread service client not initialized")
	}

	// In a real implementation, this context would be used for the gRPC call
	_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Mock implementation
	return nil
}

// UnpinThread implements ThreadServiceClient
func (c *GRPCThreadServiceClient) UnpinThread(threadID, userID string) error {
	if c.client == nil {
		return fmt.Errorf("thread service client not initialized")
	}

	// In a real implementation, this context would be used for the gRPC call
	_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Mock implementation
	return nil
}

// PinReply implements ThreadServiceClient
func (c *GRPCThreadServiceClient) PinReply(replyID, userID string) error {
	if c.client == nil {
		return fmt.Errorf("thread service client not initialized")
	}

	// In a real implementation, this context would be used for the gRPC call
	_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Mock implementation
	return nil
}

// UnpinReply implements ThreadServiceClient
func (c *GRPCThreadServiceClient) UnpinReply(replyID, userID string) error {
	if c.client == nil {
		return fmt.Errorf("thread service client not initialized")
	}

	// In a real implementation, this context would be used for the gRPC call
	_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Mock implementation
	return nil
}

// Helper function to convert proto thread to Thread struct
func convertProtoToThread(t any) *Thread {
	if t == nil {
		return nil
	}

	var thread *Thread

	// Handle both Thread and ThreadResponse types
	switch v := t.(type) {
	case *threadProto.Thread:
		thread = &Thread{
			ID:        v.Id,
			Content:   v.Content,
			UserID:    v.UserId,
			CreatedAt: v.CreatedAt.AsTime(),
			UpdatedAt: v.UpdatedAt.AsTime(),
			ParentID:  "", // No parent for regular threads
			// Default values for user data if not available
			Username:    "anonymous",
			DisplayName: "User",
		}

		// Convert media
		if len(v.Media) > 0 {
			thread.Media = make([]Media, len(v.Media))
			for i, m := range v.Media {
				thread.Media[i] = Media{
					ID:   m.Id,
					Type: m.Type,
					URL:  m.Url,
				}
			}
		}

	case *threadProto.ThreadResponse:
		if v.Thread == nil {
			return nil
		}

		thread = &Thread{
			ID:           v.Thread.Id,
			Content:      v.Thread.Content,
			UserID:       v.Thread.UserId,
			CreatedAt:    v.Thread.CreatedAt.AsTime(),
			UpdatedAt:    v.Thread.UpdatedAt.AsTime(),
			ParentID:     "", // No parent for regular threads
			LikeCount:    int(v.LikesCount),
			ReplyCount:   int(v.RepliesCount),
			RepostCount:  int(v.RepostsCount),
			IsLiked:      v.LikedByUser,
			IsReposted:   v.RepostedByUser,
			IsBookmarked: v.BookmarkedByUser,
			// Default values for user data if User not available
			Username:    "anonymous",
			DisplayName: "User",
		}

		// Set user data if available
		if v.User != nil {
			thread.Username = v.User.Username
			thread.DisplayName = v.User.Name
			thread.ProfilePicture = v.User.ProfilePictureUrl
		}

		// Convert media
		if len(v.Thread.Media) > 0 {
			thread.Media = make([]Media, len(v.Thread.Media))
			for i, m := range v.Thread.Media {
				thread.Media[i] = Media{
					ID:   m.Id,
					Type: m.Type,
					URL:  m.Url,
				}
			}
		}
	}

	return thread
}
