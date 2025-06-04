package handlers

import (
	"aycom/backend/api-gateway/utils"
	threadProto "aycom/backend/proto/thread"
	userProto "aycom/backend/proto/user"
	"context"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func FollowUser(c *gin.Context) {
	targetUserID := c.Param("userId")
	if targetUserID == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "BAD_REQUEST", "User ID parameter is required")
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}
	currentUserID := userID.(string)

	if targetUserID == currentUserID {
		utils.SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Users cannot follow themselves")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resolvedUserID, err := utils.ResolveUserIdentifier(ctx, UserClient, targetUserID)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusNotFound, "NOT_FOUND", fmt.Sprintf("Target user not found: %v", err))
		log.Printf("FollowUser: Failed to resolve user identifier: %v", err)
		return
	}

	isFollowing, err := utils.CheckFollowStatus(ctx, UserClient, currentUserID, resolvedUserID)
	if err != nil {
		log.Printf("Error checking follow status: %v", err)

	}

	if isFollowing {
		log.Printf("FollowUser: User %s is already following user %s", currentUserID, resolvedUserID)
		utils.SendSuccessResponse(c, http.StatusOK, gin.H{
			"success":               true,
			"action":                "already_following",
			"is_following":          true,
			"message":               "Already following this user",
			"was_already_following": true,
			"is_now_following":      true,
		})
		return
	}

	followRequest := &userProto.FollowUserRequest{
		FollowerId: currentUserID,
		FollowedId: resolvedUserID,
	}
	followResp, err := UserClient.FollowUser(ctx, followRequest)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to follow user: "+err.Error())
		log.Printf("Error following user: %v", err)
		return
	}

	log.Printf("User %s successfully followed user %s", currentUserID, resolvedUserID)

	utils.SendSuccessResponse(c, http.StatusOK, gin.H{
		"success":               true,
		"action":                "followed",
		"is_following":          true,
		"message":               followResp.Message,
		"was_already_following": followResp.WasAlreadyFollowing,
		"is_now_following":      followResp.IsNowFollowing,
	})
}

func UnfollowUser(c *gin.Context) {
	targetUserID := c.Param("userId")
	if targetUserID == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "BAD_REQUEST", "User ID parameter is required")
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}
	currentUserID := userID.(string)

	if targetUserID == currentUserID {
		utils.SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Users cannot unfollow themselves")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resolvedUserID, err := utils.ResolveUserIdentifier(ctx, UserClient, targetUserID)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusNotFound, "NOT_FOUND", fmt.Sprintf("Target user not found: %v", err))
		log.Printf("UnfollowUser: Failed to resolve user identifier: %v", err)
		return
	}

	isFollowing, err := utils.CheckFollowStatus(ctx, UserClient, currentUserID, resolvedUserID)
	if err != nil {
		log.Printf("Error checking follow status: %v", err)

	}

	if !isFollowing {
		log.Printf("UnfollowUser: User %s is already not following user %s", currentUserID, resolvedUserID)
		utils.SendSuccessResponse(c, http.StatusOK, gin.H{
			"success":          true,
			"action":           "not_following",
			"message":          "User was not following this user",
			"is_following":     false,
			"was_following":    false,
			"is_now_following": false,
		})
		return
	}

	unfollowRequest := &userProto.UnfollowUserRequest{
		FollowerId: currentUserID,
		FollowedId: resolvedUserID,
	}
	_, err = UserClient.UnfollowUser(ctx, unfollowRequest)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to unfollow user: "+err.Error())
		log.Printf("Error unfollowing user: %v", err)
		return
	}

	log.Printf("User %s successfully unfollowed user %s", currentUserID, resolvedUserID)

	utils.SendSuccessResponse(c, http.StatusOK, gin.H{
		"success":          true,
		"message":          "Successfully unfollowed user",
		"action":           "unfollowed",
		"is_following":     false,
		"was_following":    true,
		"is_now_following": false,
	})
}

func GetFollowers(c *gin.Context) {
	targetUserID := c.Param("userId")
	if targetUserID == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "BAD_REQUEST", "User ID parameter is required")
		return
	}

	page := 1
	limit := 20

	pageStr := c.Query("page")
	if pageStr != "" {
		if val, err := strconv.Atoi(pageStr); err == nil && val > 0 {
			page = val
		}
	}

	limitStr := c.Query("limit")
	if limitStr != "" {
		if val, err := strconv.Atoi(limitStr); err == nil && val > 0 && val <= 100 {
			limit = val
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resolvedUserID, err := utils.ResolveUserIdentifier(ctx, UserClient, targetUserID)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusNotFound, "NOT_FOUND", fmt.Sprintf("User not found: %v", err))
		log.Printf("GetFollowers: Failed to resolve user identifier: %v", err)
		return
	}

	followersReq := &userProto.GetFollowersRequest{
		UserId: resolvedUserID,
		Page:   int32(page),
		Limit:  int32(limit),
	}

	followersResp, err := UserClient.GetFollowers(ctx, followersReq)
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			utils.SendErrorResponse(c, http.StatusNotFound, "NOT_FOUND", "User not found")
			return
		}

		log.Printf("Error getting followers: %v", err)
		utils.SendSuccessResponse(c, http.StatusOK, gin.H{
			"followers": []gin.H{},
			"pagination": gin.H{
				"total_count":  0,
				"current_page": page,
				"per_page":     limit,
				"has_more":     false,
			},
		})
		return
	}

	followers := make([]gin.H, 0, len(followersResp.GetFollowers()))
	for _, follower := range followersResp.GetFollowers() {
		followers = append(followers, gin.H{
			"id":                  follower.Id,
			"username":            follower.Username,
			"name":                follower.Name,
			"bio":                 follower.Bio,
			"profile_picture_url": follower.ProfilePictureUrl,
			"is_verified":         follower.IsVerified,
			"is_following":        follower.IsFollowing,
			"follower_count":      follower.FollowerCount,
			"following_count":     follower.FollowingCount,
		})
	}

	utils.SendSuccessResponse(c, http.StatusOK, gin.H{
		"followers": followers,
		"pagination": gin.H{
			"total_count":  followersResp.TotalCount,
			"current_page": page,
			"per_page":     limit,
			"has_more":     followersResp.TotalCount > int32(page*limit),
		},
	})
}

func GetFollowing(c *gin.Context) {
	targetUserID := c.Param("userId")
	if targetUserID == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "BAD_REQUEST", "User ID parameter is required")
		return
	}

	page := 1
	limit := 20

	pageStr := c.Query("page")
	if pageStr != "" {
		if val, err := strconv.Atoi(pageStr); err == nil && val > 0 {
			page = val
		}
	}

	limitStr := c.Query("limit")
	if limitStr != "" {
		if val, err := strconv.Atoi(limitStr); err == nil && val > 0 && val <= 100 {
			limit = val
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resolvedUserID, err := utils.ResolveUserIdentifier(ctx, UserClient, targetUserID)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusNotFound, "NOT_FOUND", fmt.Sprintf("User not found: %v", err))
		log.Printf("GetFollowing: Failed to resolve user identifier: %v", err)
		return
	}

	followingReq := &userProto.GetFollowingRequest{
		UserId: resolvedUserID,
		Page:   int32(page),
		Limit:  int32(limit),
	}

	followingResp, err := UserClient.GetFollowing(ctx, followingReq)
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			utils.SendErrorResponse(c, http.StatusNotFound, "NOT_FOUND", "User not found")
			return
		}

		log.Printf("Error getting following: %v", err)
		utils.SendSuccessResponse(c, http.StatusOK, gin.H{
			"following": []gin.H{},
			"pagination": gin.H{
				"total_count":  0,
				"current_page": page,
				"per_page":     limit,
				"has_more":     false,
			},
		})
		return
	}

	following := make([]gin.H, 0, len(followingResp.GetFollowing()))
	for _, user := range followingResp.GetFollowing() {
		following = append(following, gin.H{
			"id":                  user.Id,
			"username":            user.Username,
			"name":                user.Name,
			"bio":                 user.Bio,
			"profile_picture_url": user.ProfilePictureUrl,
			"is_verified":         user.IsVerified,
			"is_following":        true,
			"follower_count":      user.FollowerCount,
			"following_count":     user.FollowingCount,
		})
	}

	utils.SendSuccessResponse(c, http.StatusOK, gin.H{
		"following": following,
		"pagination": gin.H{
			"total_count":  followingResp.TotalCount,
			"current_page": page,
			"per_page":     limit,
			"has_more":     followingResp.TotalCount > int32(page*limit),
		},
	})
}

func LikeThread(c *gin.Context) {
	threadID := c.Param("id")
	if threadID == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "BAD_REQUEST", "Thread ID parameter is required")
		return
	}

	// Always succeed - mock like feature without authentication
	utils.SendSuccessResponse(c, http.StatusOK, gin.H{
		"message":      "Thread liked successfully",
		"thread_id":    threadID,
		"is_now_liked": true,
	})
}

func UnlikeThread(c *gin.Context) {
	threadID := c.Param("id")
	if threadID == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "BAD_REQUEST", "Thread ID parameter is required")
		return
	}

	// Always succeed - mock unlike feature without authentication
	utils.SendSuccessResponse(c, http.StatusOK, gin.H{
		"message":      "Thread unliked successfully",
		"thread_id":    threadID,
		"is_now_liked": false,
	})
}

// Helper function to standardize userID extraction from context
func getUserIDFromContext(c *gin.Context) (string, bool) {
	// Try "userId" first (newer convention)
	if userID, exists := c.Get("userId"); exists {
		if userIDStr, ok := userID.(string); ok && userIDStr != "" {
			return userIDStr, true
		}
	}

	// Fall back to "userID" (older convention)
	if userID, exists := c.Get("userID"); exists {
		if userIDStr, ok := userID.(string); ok && userIDStr != "" {
			return userIDStr, true
		}
	}

	return "", false
}

func ReplyToThread(c *gin.Context) {

	userIDAny, exists := c.Get("userId")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User ID not found in token")
		return
	}

	userID, ok := userIDAny.(string)
	if !ok {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Invalid User ID format in token")
		return
	}

	threadID := c.Param("id")
	if threadID == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Thread ID is required")
		return
	}

	var request struct {
		Content          string               `json:"content" binding:"required"`
		Media            []*threadProto.Media `json:"media,omitempty"`
		ParentReplyID    string               `json:"parent_reply_id,omitempty"`
		MentionedUserIDs []string             `json:"mentioned_user_ids,omitempty"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request: "+err.Error())
		return
	}

	createReplyRequest := &threadProto.CreateReplyRequest{
		ThreadId:         threadID,
		UserId:           userID,
		Content:          request.Content,
		Media:            request.Media,
		MentionedUserIds: request.MentionedUserIDs,
	}

	if request.ParentReplyID != "" {
		createReplyRequest.ParentId = request.ParentReplyID
	}

	conn, err := threadConnPool.Get()
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "SERVICE_UNAVAILABLE", "Failed to connect to thread service: "+err.Error())
		return
	}
	defer threadConnPool.Put(conn)

	client := threadProto.NewThreadServiceClient(conn)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, err := client.CreateReply(ctx, createReplyRequest)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			httpStatus := http.StatusInternalServerError
			if st.Code() == codes.InvalidArgument {
				httpStatus = http.StatusBadRequest
			} else if st.Code() == codes.NotFound {
				httpStatus = http.StatusNotFound
			}
			utils.SendErrorResponse(c, httpStatus, st.Code().String(), st.Message())
		} else {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create reply: "+err.Error())
		}
		return
	}

	c.JSON(http.StatusCreated, resp)
}

func GetThreadReplies(c *gin.Context) {

	threadID := c.Param("id")
	if threadID == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Thread ID is required")
		return
	}

	page := 1
	limit := 20

	pageStr := c.Query("page")
	if pageStr != "" {
		if val, err := strconv.Atoi(pageStr); err == nil && val > 0 {
			page = val
		}
	}

	limitStr := c.Query("limit")
	if limitStr != "" {
		if val, err := strconv.Atoi(limitStr); err == nil && val > 0 && val <= 100 {
			limit = val
		}
	}

	conn, err := threadConnPool.Get()
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "SERVICE_UNAVAILABLE", "Failed to connect to thread service: "+err.Error())
		return
	}
	defer threadConnPool.Put(conn)

	client := threadProto.NewThreadServiceClient(conn)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, err := client.GetRepliesByThread(ctx, &threadProto.GetRepliesByThreadRequest{
		ThreadId: threadID,
		Page:     int32(page),
		Limit:    int32(limit),
	})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			httpStatus := http.StatusInternalServerError
			if st.Code() == codes.NotFound {
				httpStatus = http.StatusNotFound
			}
			utils.SendErrorResponse(c, httpStatus, st.Code().String(), st.Message())
		} else {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get thread replies: "+err.Error())
		}
		return
	}

	c.JSON(http.StatusOK, resp)
}

func RepostThread(c *gin.Context) {

	userIDAny, exists := c.Get("userId")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User ID not found in token")
		return
	}

	userID, ok := userIDAny.(string)
	if !ok {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Invalid User ID format in token")
		return
	}

	threadID := c.Param("id")
	if threadID == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Thread ID is required")
		return
	}

	var request struct {
		Content string `json:"content"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {

		log.Printf("Warning: Failed to parse request body: %v", err)
	}

	conn, err := threadConnPool.Get()
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "SERVICE_UNAVAILABLE", "Failed to connect to thread service: "+err.Error())
		return
	}
	defer threadConnPool.Put(conn)

	client := threadProto.NewThreadServiceClient(conn)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	_, err = client.RepostThread(ctx, &threadProto.RepostThreadRequest{
		ThreadId:     threadID,
		UserId:       userID,
		AddedContent: &request.Content,
	})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			httpStatus := http.StatusInternalServerError
			if st.Code() == codes.NotFound {
				httpStatus = http.StatusNotFound
			}
			utils.SendErrorResponse(c, httpStatus, st.Code().String(), st.Message())
		} else {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to repost thread: "+err.Error())
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Thread reposted successfully",
	})
}

func RemoveRepost(c *gin.Context) {

	userIDAny, exists := c.Get("userId")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User ID not found in token")
		return
	}

	userID, ok := userIDAny.(string)
	if !ok {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Invalid User ID format in token")
		return
	}

	threadID := c.Param("id")
	if threadID == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Thread ID is required")
		return
	}

	conn, err := threadConnPool.Get()
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "SERVICE_UNAVAILABLE", "Failed to connect to thread service: "+err.Error())
		return
	}
	defer threadConnPool.Put(conn)

	client := threadProto.NewThreadServiceClient(conn)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	_, err = client.RemoveRepost(ctx, &threadProto.RemoveRepostRequest{
		ThreadId: threadID,
		UserId:   userID,
	})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			httpStatus := http.StatusInternalServerError
			if st.Code() == codes.NotFound {
				httpStatus = http.StatusNotFound
			}
			utils.SendErrorResponse(c, httpStatus, st.Code().String(), st.Message())
		} else {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to remove repost: "+err.Error())
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Repost removed successfully",
	})
}

func BookmarkThread(c *gin.Context) {
	threadID := c.Param("id")
	if threadID == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Thread ID is required")
		return
	}

	// Always succeed - mock bookmark feature without authentication
	utils.SendSuccessResponse(c, http.StatusOK, gin.H{
		"message": "Thread bookmarked successfully",
	})
}

func RemoveBookmark(c *gin.Context) {
	threadID := c.Param("id")
	if threadID == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Thread ID is required")
		return
	}

	// Always succeed - mock remove bookmark feature without authentication
	utils.SendSuccessResponse(c, http.StatusOK, gin.H{
		"message": "Bookmark removed successfully",
	})
}

func GetThreadsFromFollowing(c *gin.Context) {
	authenticatedUserID, exists := c.Get("userId")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User ID not found in token")
		return
	}

	authenticatedUserIDStr, ok := authenticatedUserID.(string)
	if !ok {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Invalid User ID format in token")
		return
	}

	page := 1
	limit := 10

	pageStr := c.Query("page")
	if pageStr != "" {
		pageInt, err := strconv.Atoi(pageStr)
		if err == nil && pageInt > 0 {
			page = pageInt
		}
	}

	limitStr := c.Query("limit")
	if limitStr != "" {
		limitInt, err := strconv.Atoi(limitStr)
		if err == nil && limitInt > 0 && limitInt <= 50 {
			limit = limitInt
		}
	}

	log.Printf("Getting following threads for user: %s, page: %d, limit: %d", authenticatedUserIDStr, page, limit)

	if userServiceClient == nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "SERVICE_UNAVAILABLE", "User service client not initialized")
		return
	}

	followingList, err := userServiceClient.GetFollowing(authenticatedUserIDStr, page, limit)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Error retrieving following users: "+err.Error())
		return
	}

	if len(followingList) == 0 {
		utils.SendSuccessResponse(c, http.StatusOK, gin.H{
			"threads": []gin.H{},
			"pagination": gin.H{
				"total_count":  0,
				"current_page": page,
				"per_page":     limit,
			},
		})
		return
	}

	followingIDs := make([]string, len(followingList))
	for i, user := range followingList {
		followingIDs[i] = user.ID
	}

	if threadServiceClient == nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "SERVICE_UNAVAILABLE", "Thread service client not initialized")
		return
	}

	var allThreads []*Thread
	for _, followedUserID := range followingIDs {
		userThreads, err := threadServiceClient.GetThreadsByUserID(followedUserID, authenticatedUserIDStr, 1, 10)
		if err != nil {
			log.Printf("Error getting threads for user %s: %v", followedUserID, err)
			continue
		}
		allThreads = append(allThreads, userThreads...)
	}

	sort.Slice(allThreads, func(i, j int) bool {
		return allThreads[i].CreatedAt.After(allThreads[j].CreatedAt)
	})

	startIdx := (page - 1) * limit
	endIdx := startIdx + limit
	if startIdx >= len(allThreads) {

		utils.SendSuccessResponse(c, http.StatusOK, gin.H{
			"threads": []gin.H{},
			"pagination": gin.H{
				"total_count":  len(allThreads),
				"current_page": page,
				"per_page":     limit,
			},
		})
		return
	}

	if endIdx > len(allThreads) {
		endIdx = len(allThreads)
	}

	pagedThreads := allThreads[startIdx:endIdx]

	utils.SendSuccessResponse(c, http.StatusOK, gin.H{
		"threads": pagedThreads,
		"pagination": gin.H{
			"total_count":  len(allThreads),
			"current_page": page,
			"per_page":     limit,
		},
	})
}

func LikeReply(c *gin.Context) {

	userIDAny, exists := c.Get("userId")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User ID not found in token")
		return
	}

	userID, ok := userIDAny.(string)
	if !ok {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Invalid User ID format in token")
		return
	}

	replyID := c.Param("id")
	if replyID == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Reply ID is required")
		return
	}

	conn, err := threadConnPool.Get()
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "SERVICE_UNAVAILABLE", "Failed to connect to thread service: "+err.Error())
		return
	}
	defer threadConnPool.Put(conn)

	client := threadProto.NewThreadServiceClient(conn)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	_, err = client.LikeReply(ctx, &threadProto.LikeReplyRequest{
		ReplyId: replyID,
		UserId:  userID,
	})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			httpStatus := http.StatusInternalServerError
			if st.Code() == codes.NotFound {
				httpStatus = http.StatusNotFound
			}
			utils.SendErrorResponse(c, httpStatus, st.Code().String(), st.Message())
		} else {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to like reply: "+err.Error())
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Reply liked successfully",
	})
}

func UnlikeReply(c *gin.Context) {

	userIDAny, exists := c.Get("userId")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User ID not found in token")
		return
	}

	userID, ok := userIDAny.(string)
	if !ok {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Invalid User ID format in token")
		return
	}

	replyID := c.Param("id")
	if replyID == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Reply ID is required")
		return
	}

	conn, err := threadConnPool.Get()
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "SERVICE_UNAVAILABLE", "Failed to connect to thread service: "+err.Error())
		return
	}
	defer threadConnPool.Put(conn)

	client := threadProto.NewThreadServiceClient(conn)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	_, err = client.UnlikeReply(ctx, &threadProto.UnlikeReplyRequest{
		ReplyId: replyID,
		UserId:  userID,
	})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			httpStatus := http.StatusInternalServerError
			if st.Code() == codes.NotFound {
				httpStatus = http.StatusNotFound
			}
			utils.SendErrorResponse(c, httpStatus, st.Code().String(), st.Message())
		} else {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to unlike reply: "+err.Error())
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Reply unliked successfully",
	})
}

func BookmarkReply(c *gin.Context) {

	userIDAny, exists := c.Get("userId")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User ID not found in token")
		return
	}

	userID, ok := userIDAny.(string)
	if !ok {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Invalid User ID format in token")
		return
	}

	replyID := c.Param("id")
	if replyID == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Reply ID is required")
		return
	}

	log.Printf("BookmarkReply: Attempting to bookmark reply %s for user %s", replyID, userID)

	conn, err := threadConnPool.Get()
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "SERVICE_UNAVAILABLE", "Failed to connect to thread service: "+err.Error())
		return
	}
	defer threadConnPool.Put(conn)

	client := threadProto.NewThreadServiceClient(conn)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	_, err = client.BookmarkReply(ctx, &threadProto.BookmarkReplyRequest{
		ReplyId: replyID,
		UserId:  userID,
	})

	if err != nil {

		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.NotFound:
				utils.SendErrorResponse(c, http.StatusNotFound, "NOT_FOUND", "Reply not found")
				return
			case codes.AlreadyExists:

				utils.SendSuccessResponse(c, http.StatusOK, gin.H{
					"success": true,
					"message": "Reply already bookmarked",
					"code":    "ALREADY_BOOKMARKED",
				})
				return
			default:
				utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to bookmark reply: "+st.Message())
				return
			}
		}

		utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to bookmark reply: "+err.Error())
		return
	}

	log.Printf("Successfully bookmarked reply %s for user %s", replyID, userID)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Reply bookmarked successfully",
	})
}

func RemoveReplyBookmark(c *gin.Context) {

	userIDAny, exists := c.Get("userId")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User ID not found in token")
		return
	}

	userID, ok := userIDAny.(string)
	if !ok {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Invalid User ID format in token")
		return
	}

	replyID := c.Param("id")
	if replyID == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Reply ID is required")
		return
	}

	conn, err := threadConnPool.Get()
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "SERVICE_UNAVAILABLE", "Failed to connect to thread service: "+err.Error())
		return
	}
	defer threadConnPool.Put(conn)

	client := threadProto.NewThreadServiceClient(conn)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	_, err = client.RemoveReplyBookmark(ctx, &threadProto.RemoveReplyBookmarkRequest{
		ReplyId: replyID,
		UserId:  userID,
	})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			httpStatus := http.StatusInternalServerError
			if st.Code() == codes.NotFound {
				httpStatus = http.StatusNotFound
			}
			utils.SendErrorResponse(c, httpStatus, st.Code().String(), st.Message())
		} else {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to remove reply bookmark: "+err.Error())
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Reply bookmark removed successfully",
	})
}

func SearchSocialUsers(c *gin.Context) {

	query := c.Query("q")
	if query == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "BAD_REQUEST", "Search query is required")
		return
	}

	filter := c.DefaultQuery("filter", "")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 10
	}

	users, totalCount, err := userServiceClient.SearchUsers(query, filter, page, limit)
	if err != nil {
		log.Printf("Error searching users: %v", err)
		utils.SendErrorResponse(c, http.StatusInternalServerError, "SERVER_ERROR", "Failed to search users")
		return
	}

	utils.SendSuccessResponse(c, http.StatusOK, gin.H{
		"users": users,
		"pagination": gin.H{
			"total_count":  totalCount,
			"current_page": page,
			"per_page":     limit,
			"has_more":     len(users) == limit && (page*limit) < totalCount,
		},
	})
}

func PinReply(c *gin.Context) {

	userIDAny, exists := c.Get("userId")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User ID not found in token")
		return
	}

	userID, ok := userIDAny.(string)
	if !ok {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Invalid User ID format in token")
		return
	}

	replyID := c.Param("id")
	if replyID == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Reply ID is required")
		return
	}

	if threadServiceClient == nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "SERVICE_UNAVAILABLE", "Thread service client not initialized")
		return
	}

	err := threadServiceClient.PinReply(replyID, userID)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			httpStatus := http.StatusInternalServerError
			if st.Code() == codes.NotFound {
				httpStatus = http.StatusNotFound
			}
			utils.SendErrorResponse(c, httpStatus, st.Code().String(), st.Message())
		} else {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to pin reply: "+err.Error())
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Reply pinned successfully",
	})
}

func UnpinReply(c *gin.Context) {

	userIDAny, exists := c.Get("userId")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User ID not found in token")
		return
	}

	userID, ok := userIDAny.(string)
	if !ok {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Invalid User ID format in token")
		return
	}

	replyID := c.Param("id")
	if replyID == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Reply ID is required")
		return
	}

	if threadServiceClient == nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "SERVICE_UNAVAILABLE", "Thread service client not initialized")
		return
	}

	err := threadServiceClient.UnpinReply(replyID, userID)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			httpStatus := http.StatusInternalServerError
			if st.Code() == codes.NotFound {
				httpStatus = http.StatusNotFound
			}
			utils.SendErrorResponse(c, httpStatus, st.Code().String(), st.Message())
		} else {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to unpin reply: "+err.Error())
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Reply unpinned successfully",
	})
}

func GetRepliesByParentReply(c *gin.Context) {

	parentReplyID := c.Param("id")
	if parentReplyID == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Parent Reply ID is required")
		return
	}

	log.Printf("GetRepliesByParentReply: Fetching replies for parent reply ID %s", parentReplyID)

	page := 1
	limit := 20

	pageStr := c.Query("page")
	if pageStr != "" {
		if val, err := strconv.Atoi(pageStr); err == nil && val > 0 {
			page = val
		}
	}

	limitStr := c.Query("limit")
	if limitStr != "" {
		if val, err := strconv.Atoi(limitStr); err == nil && val > 0 && val <= 100 {
			limit = val
		}
	}

	conn, err := threadConnPool.Get()
	if err != nil {
		log.Printf("GetRepliesByParentReply: Failed to connect to thread service: %v", err)
		utils.SendErrorResponse(c, http.StatusInternalServerError, "SERVICE_UNAVAILABLE", "Failed to connect to thread service: "+err.Error())
		return
	}
	defer threadConnPool.Put(conn)

	client := threadProto.NewThreadServiceClient(conn)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	req := &threadProto.GetRepliesByParentReplyRequest{
		ParentReplyId: parentReplyID,
		Page:          int32(page),
		Limit:         int32(limit),
	}

	resp, err := client.GetRepliesByParentReply(ctx, req)
	if err != nil {
		st, ok := status.FromError(err)

		log.Printf("GetRepliesByParentReply: Error from service: %v (status code: %v)", err, st.Code())

		if ok {
			httpStatus := http.StatusInternalServerError

			switch st.Code() {
			case codes.NotFound:
				httpStatus = http.StatusNotFound
			case codes.InvalidArgument:
				httpStatus = http.StatusBadRequest
			case codes.Unavailable:
				httpStatus = http.StatusServiceUnavailable
			case codes.DeadlineExceeded, codes.Canceled:
				httpStatus = http.StatusGatewayTimeout
			}

			utils.SendErrorResponse(c, httpStatus, st.Code().String(), st.Message())
		} else {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get reply replies: "+err.Error())
		}
		return
	}

	c.Header("Cache-Control", "public, max-age=10")

	c.JSON(http.StatusOK, resp)
}

func CheckFollowStatus(c *gin.Context) {
	targetUserID := c.Param("userId")
	if targetUserID == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "BAD_REQUEST", "User ID parameter is required")
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}
	currentUserID := userID.(string)

	log.Printf("CheckFollowStatus: Checking if user %s is following user %s", currentUserID, targetUserID)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resolvedUserID, err := utils.ResolveUserIdentifier(ctx, UserClient, targetUserID)
	if err != nil {
		log.Printf("CheckFollowStatus: Failed to resolve user identifier %s: %v", targetUserID, err)
		utils.SendErrorResponse(c, http.StatusNotFound, "NOT_FOUND", fmt.Sprintf("Target user not found: %v", err))
		return
	}

	log.Printf("CheckFollowStatus: Resolved user ID %s to %s", targetUserID, resolvedUserID)

	isFollowingReq := &userProto.IsFollowingRequest{
		FollowerId: currentUserID,
		FollowedId: resolvedUserID,
	}

	isFollowingResp, err := UserClient.IsFollowing(ctx, isFollowingReq)
	if err != nil {
		log.Printf("CheckFollowStatus: Error checking follow status: %v", err)
		utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to check follow status: "+err.Error())
		return
	}

	isFollowing := isFollowingResp.IsFollowing
	log.Printf("CheckFollowStatus: User %s following user %s: %v", currentUserID, resolvedUserID, isFollowing)

	utils.SendSuccessResponse(c, http.StatusOK, gin.H{
		"success":      true,
		"is_following": isFollowing,
		"follower_id":  currentUserID,
		"followed_id":  resolvedUserID,
	})
}
