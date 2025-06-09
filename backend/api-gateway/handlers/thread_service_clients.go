package handlers

import (
	threadProto "aycom/backend/proto/thread"
	"context"
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"aycom/backend/api-gateway/config"
)

type ThreadServiceClient interface {
	CreateThread(userID, content string, mediaIDs []string) (string, error)
	GetThreadByID(threadID string, userID string) (*Thread, error)
	GetThreadsByUserID(userID string, requestingUserID string, page, limit int) ([]*Thread, error)
	GetAllThreads(userID string, page, limit int) ([]*Thread, error)
	UpdateThread(threadID, userID, content string) (*Thread, error)
	DeleteThread(threadID, userID string) error

	SearchThreads(query string, userID string, page, limit int) ([]*Thread, error)

	LikeThread(threadID, userID string) error
	UnlikeThread(threadID, userID string) error
	ReplyToThread(threadID, userID, content string, mediaIDs []string) (string, error)
	GetThreadReplies(threadID string, userID string, page, limit int) ([]*Thread, error)
	RepostThread(threadID, userID string) error
	RemoveRepost(threadID, userID string) error

	BookmarkThread(threadID, userID string) error
	RemoveBookmark(threadID, userID string) error
	GetUserBookmarks(userID string, page, limit int) ([]*Thread, error)
	SearchUserBookmarks(userID, query string, page, limit int) ([]*Thread, error)

	GetRepliesByUser(userID string, page, limit int) ([]*Thread, error)
	GetLikedThreadsByUser(userID string, page, limit int) ([]*Thread, error)
	GetMediaByUser(userID string, page, limit int) ([]Media, error)

	PinThread(threadID, userID string) error
	UnpinThread(threadID, userID string) error
	PinReply(replyID, userID string) error
	UnpinReply(replyID, userID string) error

	GetTrendingHashtags(limit int) ([]string, error)
}

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
	BookmarkCount  int
	IsLiked        bool
	IsReposted     bool
	IsBookmarked   bool
	IsPinned       bool
	Media          []Media
	ParentID       string
}

type Media struct {
	ID        string
	Type      string
	URL       string
	Thumbnail string
}

type GRPCThreadServiceClient struct {
	client threadProto.ThreadServiceClient
	conn   *grpc.ClientConn
}

var threadServiceClient ThreadServiceClient

func InitThreadServiceClient(cfg *config.Config) {
	log.Println("Initializing thread service client...")

	if cfg.Services.ThreadService == "" {
		log.Fatal("Error: Thread service address is not configured")
		return
	}
	log.Printf("Attempting to connect to Thread service at %s", cfg.Services.ThreadService)

	// Add retry logic for more robust connection
	var conn *grpc.ClientConn
	var err error
	maxRetries := 5
	retryDelay := 3 * time.Second

	for i := 0; i < maxRetries; i++ {
		conn, err = grpc.NewClient(
			cfg.Services.ThreadService,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithDefaultCallOptions(grpc.WaitForReady(true)),
		)

		if err == nil {
			break
		}

		log.Printf("Attempt %d: Failed to connect to Thread service: %v. Retrying in %v...",
			i+1, err, retryDelay)

		if i < maxRetries-1 {
			time.Sleep(retryDelay)
		}
	}

	if err != nil {
		log.Fatalf("ERROR: Failed to connect to Thread service at %s after %d attempts: %v",
			cfg.Services.ThreadService, maxRetries, err)
		return
	}

	log.Printf("Successfully connected to Thread service at %s", cfg.Services.ThreadService)
	grpcClient := threadProto.NewThreadServiceClient(conn)
	threadServiceClient = &GRPCThreadServiceClient{
		client: grpcClient,
		conn:   conn,
	}

	// Test the connection
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

func GetThreadServiceClient() ThreadServiceClient {
	return threadServiceClient
}

func (c *GRPCThreadServiceClient) CreateThread(userID, content string, mediaIDs []string) (string, error) {
	if c.client == nil {
		return "", fmt.Errorf("thread service client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	mediaObjects := make([]*threadProto.Media, len(mediaIDs))
	for i, mediaID := range mediaIDs {
		mediaObjects[i] = &threadProto.Media{
			Id:   mediaID,
			Url:  "",
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

func (c *GRPCThreadServiceClient) GetThreadByID(threadID string, userID string) (*Thread, error) {
	if c.client == nil {
		return nil, fmt.Errorf("thread service client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if userID != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "user_id", userID)
	}

	resp, err := c.client.GetThreadById(ctx, &threadProto.GetThreadRequest{
		ThreadId: threadID,
	})
	if err != nil {
		return nil, err
	}

	thread := convertProtoToThread(resp.Thread)

	if resp.BookmarkedByUser {
		thread.IsBookmarked = true
	}

	return thread, nil
}

func (c *GRPCThreadServiceClient) GetThreadsByUserID(userID string, requestingUserID string, page, limit int) ([]*Thread, error) {
	if c.client == nil {
		return nil, fmt.Errorf("thread service client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Printf("GetThreadsByUserID: Fetching threads for userID=%s, requestingUserID=%s, page=%d, limit=%d",
		userID, requestingUserID, page, limit)

	// Add user ID to context metadata if available
	if requestingUserID != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "user_id", requestingUserID)
	}

	// Make the gRPC call with retry mechanism
	var resp *threadProto.ThreadsResponse
	var err error
	maxRetries := 3

	for attempt := 1; attempt <= maxRetries; attempt++ {
		if attempt > 1 {
			log.Printf("Retry attempt %d for GetThreadsByUser for userID=%s", attempt, userID)
			// Short sleep before retry
			time.Sleep(time.Duration(attempt*300) * time.Millisecond)
		}

		resp, err = c.client.GetThreadsByUser(ctx, &threadProto.GetThreadsByUserRequest{
			UserId: userID,
			Page:   int32(page),
			Limit:  int32(limit),
		})

		// If successful or context deadline exceeded, break the retry loop
		if err == nil || status.Code(err) == codes.DeadlineExceeded {
			break
		}
	}

	// Handle errors
	if err != nil {
		log.Printf("ERROR in GetThreadsByUserID for userID=%s: %v (error code: %s)",
			userID, err, status.Code(err))
		return nil, err
	}

	// Convert thread proto objects to Thread structs
	threads := make([]*Thread, len(resp.Threads))
	for i, t := range resp.Threads {
		threads[i] = convertProtoToThread(t)
	}

	log.Printf("Successfully fetched %d threads for userID=%s", len(threads), userID)
	return threads, nil
}

func (c *GRPCThreadServiceClient) GetAllThreads(userID string, page, limit int) ([]*Thread, error) {
	if c.client == nil {
		return nil, fmt.Errorf("thread service client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if userID != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "user_id", userID)
	}

	resp, err := c.client.GetAllThreads(ctx, &threadProto.GetAllThreadsRequest{
		Page:  int32(page),
		Limit: int32(limit),
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

func (c *GRPCThreadServiceClient) SearchThreads(query string, userID string, page, limit int) ([]*Thread, error) {

	threads, err := c.GetAllThreads(userID, page, limit)
	if err != nil {
		return nil, err
	}

	if query == "" {
		return threads, nil
	}

	var filteredThreads []*Thread
	queryLower := strings.ToLower(query)
	for _, thread := range threads {
		if strings.Contains(strings.ToLower(thread.Content), queryLower) {
			filteredThreads = append(filteredThreads, thread)
		}
	}

	return filteredThreads, nil
}

func (c *GRPCThreadServiceClient) LikeThread(threadID, userID string) error {
	if c.client == nil {
		return fmt.Errorf("thread service client not initialized")
	}

	maxRetries := 3
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

		ctx = metadata.AppendToOutgoingContext(ctx, "user_id", userID)

		log.Printf("Attempt %d: Liking thread %s for user %s", attempt, threadID, userID)

		_, err := c.client.LikeThread(ctx, &threadProto.LikeThreadRequest{
			ThreadId: threadID,
			UserId:   userID,
		})

		cancel()

		if err == nil {
			log.Printf("Successfully liked thread %s for user %s", threadID, userID)

			verifyCtx, verifyCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer verifyCancel()

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

		time.Sleep(time.Duration(attempt*500) * time.Millisecond)
	}

	log.Printf("Failed to like thread after %d attempts: %v", maxRetries, lastErr)
	return lastErr
}

func (c *GRPCThreadServiceClient) UnlikeThread(threadID, userID string) error {
	if c.client == nil {
		return fmt.Errorf("thread service client not initialized")
	}

	maxRetries := 3
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

		ctx = metadata.AppendToOutgoingContext(ctx, "user_id", userID)

		log.Printf("Attempt %d: Unliking thread %s for user %s", attempt, threadID, userID)

		_, err := c.client.UnlikeThread(ctx, &threadProto.UnlikeThreadRequest{
			ThreadId: threadID,
			UserId:   userID,
		})

		cancel()

		if err == nil {
			log.Printf("Successfully unliked thread %s for user %s", threadID, userID)
			return nil
		}

		// Check for rate limiting error
		if st, ok := status.FromError(err); ok && st.Code() == codes.ResourceExhausted {
			log.Printf("Rate limiting detected when unliking thread: %v", err)
			return err // Don't retry rate limiting errors
		}

		lastErr = err
		log.Printf("Error unliking thread (attempt %d): %v", attempt, err)

		// Use exponential backoff for retries
		backoffTime := time.Duration(attempt*attempt*250) * time.Millisecond
		time.Sleep(backoffTime)
	}

	log.Printf("Failed to unlike thread after %d attempts: %v", maxRetries, lastErr)
	return lastErr
}

func (c *GRPCThreadServiceClient) ReplyToThread(threadID, userID, content string, mediaIDs []string) (string, error) {
	if c.client == nil {
		return "", fmt.Errorf("thread service client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	mediaObjects := make([]*threadProto.Media, len(mediaIDs))
	for i, mediaID := range mediaIDs {
		mediaObjects[i] = &threadProto.Media{
			Id:   mediaID,
			Url:  "",
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

func (c *GRPCThreadServiceClient) GetThreadReplies(threadID string, userID string, page, limit int) ([]*Thread, error) {
	if c.client == nil {
		return nil, fmt.Errorf("thread service client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

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

	replies := make([]*Thread, len(resp.Replies))
	for i, replyResp := range resp.Replies {
		if replyResp.Reply == nil {
			continue
		}

		reply := replyResp.Reply

		username := "anonymous"
		displayName := "User"
		profilePicURL := "https://secure.gravatar.com/avatar/0?d=mp"

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

		// Default reply count to 0
		repliesCount := 0

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
			ReplyCount:     repliesCount,
			IsLiked:        replyResp.LikedByUser,
			IsBookmarked:   replyResp.BookmarkedByUser,
			ParentID:       threadID,
		}

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

func (c *GRPCThreadServiceClient) BookmarkThread(threadID, userID string) error {
	if c.client == nil {
		return fmt.Errorf("thread service client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Printf("Sending BookmarkThread request to thread service - threadID: %s, userID: %s", threadID, userID)

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

func (c *GRPCThreadServiceClient) RemoveBookmark(threadID, userID string) error {
	if c.client == nil {
		return fmt.Errorf("thread service client not initialized")
	}

	maxRetries := 3
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

		ctx = metadata.AppendToOutgoingContext(ctx, "user_id", userID)

		log.Printf("Attempt %d: Removing bookmark for thread %s by user %s", attempt, threadID, userID)

		_, err := c.client.RemoveBookmark(ctx, &threadProto.RemoveBookmarkRequest{
			ThreadId: threadID,
			UserId:   userID,
		})

		cancel()

		if err == nil {
			log.Printf("Successfully removed bookmark for thread %s by user %s", threadID, userID)

			verifyCtx, verifyCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer verifyCancel()

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

		time.Sleep(time.Duration(attempt*500) * time.Millisecond)
	}

	log.Printf("Failed to remove bookmark after %d attempts: %v", maxRetries, lastErr)
	return lastErr
}

func (c *GRPCThreadServiceClient) GetUserBookmarks(userID string, page, limit int) ([]*Thread, error) {
	log.Printf("GetUserBookmarks client called with userID: %s, page: %d, limit: %d", userID, page, limit)

	if c.client == nil {
		log.Printf("GetUserBookmarks: thread service client is nil")
		return nil, fmt.Errorf("thread service client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	md := metadata.New(map[string]string{
		"user_id": userID,
	})
	ctx = metadata.NewOutgoingContext(ctx, md)
	log.Printf("GetUserBookmarks: Created context with user_id: %s in metadata", userID)

	bookmarksMethod := reflect.ValueOf(c.client).MethodByName("GetBookmarksByUser")
	if bookmarksMethod.IsValid() {
		log.Printf("GetUserBookmarks: Found GetBookmarksByUser method on client")

		ctxArg := reflect.ValueOf(ctx)
		reqArg := reflect.New(bookmarksMethod.Type().In(1).Elem()).Interface()

		reqVal := reflect.ValueOf(reqArg).Elem()
		reqVal.FieldByName("UserId").SetString(userID)
		reqVal.FieldByName("Page").SetInt(int64(page))
		reqVal.FieldByName("Limit").SetInt(int64(limit))

		results := bookmarksMethod.Call([]reflect.Value{ctxArg, reflect.ValueOf(reqArg)})
		if !results[1].IsNil() {
			err := results[1].Interface().(error)
			log.Printf("GetUserBookmarks: Error from GetBookmarksByUser: %v", err)
			return nil, fmt.Errorf("failed to get bookmarks: %w", err)
		}

		resp := results[0].Interface()
		threadsResp := resp.(*threadProto.ThreadsResponse)

		threads := make([]*Thread, len(threadsResp.Threads))
		for i, t := range threadsResp.Threads {
			thread := convertProtoToThread(t)

			thread.IsBookmarked = true
			threads[i] = thread
		}

		log.Printf("Successfully retrieved %d bookmarks using GetBookmarksByUser", len(threads))
		return threads, nil
	}

	log.Printf("GetUserBookmarks: GetBookmarksByUser method not found, returning empty bookmarks list")
	return []*Thread{}, nil
}

func (c *GRPCThreadServiceClient) SearchUserBookmarks(userID, query string, page, limit int) ([]*Thread, error) {

	return []*Thread{}, nil
}

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

	hashtags := make([]string, len(resp.Hashtags))
	for i, h := range resp.Hashtags {
		hashtags[i] = h.Name
	}

	return hashtags, nil
}

func (c *GRPCThreadServiceClient) PinThread(threadID, userID string) error {
	if c.client == nil {
		return fmt.Errorf("thread service client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	method := reflect.ValueOf(c.client).MethodByName("PinThread")
	if !method.IsValid() {
		return fmt.Errorf("method PinThread not found on client")
	}

	ctxArg := reflect.ValueOf(ctx)
	reqArg := reflect.New(method.Type().In(1).Elem()).Interface()

	reqVal := reflect.ValueOf(reqArg).Elem()
	reqVal.FieldByName("ThreadId").SetString(threadID)
	reqVal.FieldByName("UserId").SetString(userID)

	results := method.Call([]reflect.Value{ctxArg, reflect.ValueOf(reqArg)})
	if !results[1].IsNil() {
		return results[1].Interface().(error)
	}

	return nil
}

func (c *GRPCThreadServiceClient) UnpinThread(threadID, userID string) error {
	if c.client == nil {
		return fmt.Errorf("thread service client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	method := reflect.ValueOf(c.client).MethodByName("UnpinThread")
	if !method.IsValid() {
		return fmt.Errorf("method UnpinThread not found on client")
	}

	ctxArg := reflect.ValueOf(ctx)
	reqArg := reflect.New(method.Type().In(1).Elem()).Interface()

	reqVal := reflect.ValueOf(reqArg).Elem()
	reqVal.FieldByName("ThreadId").SetString(threadID)
	reqVal.FieldByName("UserId").SetString(userID)

	results := method.Call([]reflect.Value{ctxArg, reflect.ValueOf(reqArg)})
	if !results[1].IsNil() {
		return results[1].Interface().(error)
	}

	return nil
}

func (c *GRPCThreadServiceClient) PinReply(replyID, userID string) error {
	if c.client == nil {
		return fmt.Errorf("thread service client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	method := reflect.ValueOf(c.client).MethodByName("PinReply")
	if !method.IsValid() {
		return fmt.Errorf("method PinReply not found on client")
	}

	ctxArg := reflect.ValueOf(ctx)
	reqArg := reflect.New(method.Type().In(1).Elem()).Interface()

	reqVal := reflect.ValueOf(reqArg).Elem()
	reqVal.FieldByName("ReplyId").SetString(replyID)
	reqVal.FieldByName("UserId").SetString(userID)

	results := method.Call([]reflect.Value{ctxArg, reflect.ValueOf(reqArg)})
	if !results[1].IsNil() {
		return results[1].Interface().(error)
	}

	return nil
}

func (c *GRPCThreadServiceClient) UnpinReply(replyID, userID string) error {
	if c.client == nil {
		return fmt.Errorf("thread service client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	method := reflect.ValueOf(c.client).MethodByName("UnpinReply")
	if !method.IsValid() {
		return fmt.Errorf("method UnpinReply not found on client")
	}

	ctxArg := reflect.ValueOf(ctx)
	reqArg := reflect.New(method.Type().In(1).Elem()).Interface()

	reqVal := reflect.ValueOf(reqArg).Elem()
	reqVal.FieldByName("ReplyId").SetString(replyID)
	reqVal.FieldByName("UserId").SetString(userID)

	results := method.Call([]reflect.Value{ctxArg, reflect.ValueOf(reqArg)})
	if !results[1].IsNil() {
		return results[1].Interface().(error)
	}

	return nil
}

func (c *GRPCThreadServiceClient) GetLikedThreadsByUser(userID string, page, limit int) ([]*Thread, error) {
	if c.client == nil {
		return nil, fmt.Errorf("thread service client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	method := reflect.ValueOf(c.client).MethodByName("GetLikedThreadsByUser")
	if !method.IsValid() {
		return nil, fmt.Errorf("method GetLikedThreadsByUser not found on client")
	}

	ctxArg := reflect.ValueOf(ctx)
	reqArg := reflect.New(method.Type().In(1).Elem()).Interface()

	reqVal := reflect.ValueOf(reqArg).Elem()
	reqVal.FieldByName("UserId").SetString(userID)
	reqVal.FieldByName("Page").SetInt(int64(page))
	reqVal.FieldByName("Limit").SetInt(int64(limit))

	results := method.Call([]reflect.Value{ctxArg, reflect.ValueOf(reqArg)})
	if !results[1].IsNil() {
		return nil, results[1].Interface().(error)
	}

	resp := results[0].Interface()

	threadsVal := reflect.ValueOf(resp).Elem().FieldByName("Threads")

	threads := make([]*Thread, threadsVal.Len())
	for i := 0; i < threadsVal.Len(); i++ {
		threadResp := threadsVal.Index(i).Interface()
		threads[i] = convertProtoToThread(threadResp)
	}

	return threads, nil
}

func (c *GRPCThreadServiceClient) GetMediaByUser(userID string, page, limit int) ([]Media, error) {
	if c.client == nil {
		return nil, fmt.Errorf("thread service client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	method := reflect.ValueOf(c.client).MethodByName("GetMediaByUser")
	if !method.IsValid() {
		return nil, fmt.Errorf("method GetMediaByUser not found on client")
	}

	ctxArg := reflect.ValueOf(ctx)
	reqArg := reflect.New(method.Type().In(1).Elem()).Interface()

	reqVal := reflect.ValueOf(reqArg).Elem()
	reqVal.FieldByName("UserId").SetString(userID)
	reqVal.FieldByName("Page").SetInt(int64(page))
	reqVal.FieldByName("Limit").SetInt(int64(limit))

	results := method.Call([]reflect.Value{ctxArg, reflect.ValueOf(reqArg)})
	if !results[1].IsNil() {
		return nil, results[1].Interface().(error)
	}

	resp := results[0].Interface()

	mediaVal := reflect.ValueOf(resp).Elem().FieldByName("Media")

	media := make([]Media, mediaVal.Len())
	for i := 0; i < mediaVal.Len(); i++ {
		m := mediaVal.Index(i).Interface()

		mVal := reflect.ValueOf(m).Elem()
		media[i] = Media{
			ID:   mVal.FieldByName("Id").String(),
			URL:  mVal.FieldByName("Url").String(),
			Type: mVal.FieldByName("Type").String(),
		}
	}

	return media, nil
}

func (c *GRPCThreadServiceClient) GetRepliesByUser(userID string, page, limit int) ([]*Thread, error) {
	if c.client == nil {
		return nil, fmt.Errorf("thread service client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	method := reflect.ValueOf(c.client).MethodByName("GetRepliesByUser")
	if !method.IsValid() {
		return nil, fmt.Errorf("method GetRepliesByUser not found on client")
	}

	ctxArg := reflect.ValueOf(ctx)
	reqArg := reflect.New(method.Type().In(1).Elem()).Interface()

	reqVal := reflect.ValueOf(reqArg).Elem()
	reqVal.FieldByName("UserId").SetString(userID)
	reqVal.FieldByName("Page").SetInt(int64(page))
	reqVal.FieldByName("Limit").SetInt(int64(limit))

	results := method.Call([]reflect.Value{ctxArg, reflect.ValueOf(reqArg)})
	if !results[1].IsNil() {
		return nil, results[1].Interface().(error)
	}

	resp := results[0].Interface()
	repliesVal := reflect.ValueOf(resp).Elem().FieldByName("Replies")

	replies := make([]*Thread, repliesVal.Len())
	for i := 0; i < repliesVal.Len(); i++ {
		r := repliesVal.Index(i).Interface()
		rVal := reflect.ValueOf(r).Elem()

		replyObj := rVal.FieldByName("Reply").Interface()
		userObj := rVal.FieldByName("User").Interface()

		replyVal := reflect.ValueOf(replyObj).Elem()
		userVal := reflect.ValueOf(userObj).Elem()

		isPinned := false
		if replyVal.FieldByName("IsPinned").IsValid() && !replyVal.FieldByName("IsPinned").IsNil() {
			isPinned = replyVal.FieldByName("IsPinned").Elem().Bool()
		}

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
			ReplyCount:     0, // Default to 0 for now
			IsLiked:        rVal.FieldByName("LikedByUser").Bool(),
			IsBookmarked:   rVal.FieldByName("BookmarkedByUser").Bool(),
			ParentID:       replyVal.FieldByName("ThreadId").String(),
			IsPinned:       isPinned,
		}

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

func convertProtoToThread(t any) *Thread {
	if t == nil {
		return nil
	}

	thread := &Thread{
		ID:          "unknown",
		Content:     "",
		UserID:      "",
		Username:    "anonymous",
		DisplayName: "User",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if tr, ok := t.(*threadProto.ThreadResponse); ok && tr != nil {
		if tr.Thread != nil {
			thread.ID = tr.Thread.Id
			thread.Content = tr.Thread.Content
			thread.UserID = tr.Thread.UserId
			if tr.Thread.CreatedAt != nil {
				thread.CreatedAt = tr.Thread.CreatedAt.AsTime()
			}
			if tr.Thread.UpdatedAt != nil {
				thread.UpdatedAt = tr.Thread.UpdatedAt.AsTime()
			}
			thread.LikeCount = int(tr.LikesCount)
			thread.ReplyCount = int(tr.RepliesCount)
			thread.RepostCount = int(tr.RepostsCount)
			thread.BookmarkCount = int(tr.BookmarkCount)
			thread.IsLiked = tr.LikedByUser
			thread.IsReposted = tr.RepostedByUser
			thread.IsBookmarked = tr.BookmarkedByUser

			if tr.Thread.IsPinned != nil {
				thread.IsPinned = *tr.Thread.IsPinned
			}

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

		if tr.User != nil {
			thread.Username = tr.User.Username
			thread.DisplayName = tr.User.Name
			thread.ProfilePicture = tr.User.ProfilePictureUrl
		}

		return thread
	}

	log.Printf("Thread type conversion: received type %T", t)

	v := reflect.ValueOf(t)

	if v.Kind() == reflect.Ptr && !v.IsNil() {
		v = v.Elem()
	}

	if v.Kind() == reflect.Struct {

		threadField := v.FieldByName("Thread")
		if threadField.IsValid() && !threadField.IsNil() {
			threadVal := threadField.Elem()

			idField := threadVal.FieldByName("Id")
			if idField.IsValid() {
				thread.ID = idField.String()
			}

			contentField := threadVal.FieldByName("Content")
			if contentField.IsValid() {
				thread.Content = contentField.String()
			}

			userIDField := threadVal.FieldByName("UserId")
			if userIDField.IsValid() {
				thread.UserID = userIDField.String()
			}

			createdAtField := threadVal.FieldByName("CreatedAt")
			if createdAtField.IsValid() && !createdAtField.IsNil() {

				asTimeMethod := createdAtField.MethodByName("AsTime")
				if asTimeMethod.IsValid() {
					result := asTimeMethod.Call(nil)
					if len(result) > 0 {
						thread.CreatedAt = result[0].Interface().(time.Time)
					}
				}
			}

			updatedAtField := threadVal.FieldByName("UpdatedAt")
			if updatedAtField.IsValid() && !updatedAtField.IsNil() {

				asTimeMethod := updatedAtField.MethodByName("AsTime")
				if asTimeMethod.IsValid() {
					result := asTimeMethod.Call(nil)
					if len(result) > 0 {
						thread.UpdatedAt = result[0].Interface().(time.Time)
					}
				}
			}

			isPinnedField := threadVal.FieldByName("IsPinned")
			if isPinnedField.IsValid() && !isPinnedField.IsNil() {
				thread.IsPinned = isPinnedField.Elem().Bool()
			}

			mediaField := threadVal.FieldByName("Media")
			if mediaField.IsValid() && mediaField.Kind() == reflect.Slice {
				mediaCount := mediaField.Len()
				if mediaCount > 0 {
					thread.Media = make([]Media, mediaCount)
					for i := 0; i < mediaCount; i++ {
						mediaItem := mediaField.Index(i)
						if !mediaItem.IsNil() {
							mediaItemVal := mediaItem.Elem()

							var media Media

							idField := mediaItemVal.FieldByName("Id")
							if idField.IsValid() {
								media.ID = idField.String()
							}

							typeField := mediaItemVal.FieldByName("Type")
							if typeField.IsValid() {
								media.Type = typeField.String()
							}

							urlField := mediaItemVal.FieldByName("Url")
							if urlField.IsValid() {
								media.URL = urlField.String()
							}

							thread.Media[i] = media
						}
					}
				}
			}
		}

		likesCountField := v.FieldByName("LikesCount")
		if likesCountField.IsValid() && likesCountField.Kind() == reflect.Int64 {
			thread.LikeCount = int(likesCountField.Int())
		}

		repliesCountField := v.FieldByName("RepliesCount")
		if repliesCountField.IsValid() && repliesCountField.Kind() == reflect.Int64 {
			thread.ReplyCount = int(repliesCountField.Int())
		}

		repostsCountField := v.FieldByName("RepostsCount")
		if repostsCountField.IsValid() && repostsCountField.Kind() == reflect.Int64 {
			thread.RepostCount = int(repostsCountField.Int())
		}

		bookmarkCountField := v.FieldByName("BookmarkCount")
		if bookmarkCountField.IsValid() && bookmarkCountField.Kind() == reflect.Int64 {
			thread.BookmarkCount = int(bookmarkCountField.Int())
		}

		likedByUserField := v.FieldByName("LikedByUser")
		if likedByUserField.IsValid() {
			thread.IsLiked = likedByUserField.Bool()
		}

		repostedByUserField := v.FieldByName("RepostedByUser")
		if repostedByUserField.IsValid() {
			thread.IsReposted = repostedByUserField.Bool()
		}

		bookmarkedByUserField := v.FieldByName("BookmarkedByUser")
		if bookmarkedByUserField.IsValid() {
			thread.IsBookmarked = bookmarkedByUserField.Bool()
		}

		userField := v.FieldByName("User")
		if userField.IsValid() && !userField.IsNil() {
			userVal := userField.Elem()

			usernameField := userVal.FieldByName("Username")
			if usernameField.IsValid() {
				thread.Username = usernameField.String()
			}

			nameField := userVal.FieldByName("Name")
			if nameField.IsValid() {
				thread.DisplayName = nameField.String()
			}

			profilePicField := userVal.FieldByName("ProfilePictureUrl")
			if profilePicField.IsValid() {
				thread.ProfilePicture = profilePicField.String()
			}
		}
	}

	return thread
}
