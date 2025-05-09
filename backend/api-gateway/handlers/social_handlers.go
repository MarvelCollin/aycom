package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	threadProto "aycom/backend/proto/thread"
	userProto "aycom/backend/proto/user"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// FollowUser handles a user following another user
// @Summary Follow user
// @Description Follows a user
// @Tags Users
// @Produce json
// @Router /api/v1/users/{userId}/follow [post]
func FollowUser(c *gin.Context) {
	targetUserID := c.Param("userId")
	if targetUserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "User ID parameter is required",
		})
		return
	}

	// Get current user ID from context (set by JWT middleware)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "User not authenticated",
		})
		return
	}
	currentUserID := userID.(string)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// First, check if both users exist
	_, err := UserClient.GetUser(ctx, &userProto.GetUserRequest{
		UserId: targetUserID,
	})

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "Target user not found",
		})
		log.Printf("Error following user: target user %s not found: %v", targetUserID, err)
		return
	}

	// Create the follow relationship using the FollowUser endpoint
	followRequest := &userProto.FollowUserRequest{
		FollowerId: currentUserID,
		FollowedId: targetUserID,
	}

	followResp, err := UserClient.FollowUser(ctx, followRequest)
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.AlreadyExists {
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "Already following this user",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to follow user: " + err.Error(),
		})
		log.Printf("Error following user: %v", err)
		return
	}

	log.Printf("User %s successfully followed user %s", currentUserID, targetUserID)

	c.JSON(http.StatusOK, gin.H{
		"success": followResp.Success,
		"message": followResp.Message,
	})
}

// UnfollowUser handles a user unfollowing another user
// @Summary Unfollow user
// @Description Unfollows a user
// @Tags Users
// @Produce json
// @Router /api/v1/users/{userId}/unfollow [post]
func UnfollowUser(c *gin.Context) {
	targetUserID := c.Param("userId")
	if targetUserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "User ID parameter is required",
		})
		return
	}

	// Get current user ID from context (set by JWT middleware)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "User not authenticated",
		})
		return
	}
	currentUserID := userID.(string)

	// First, check if both users exist
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Verify target user exists
	_, err := UserClient.GetUser(ctx, &userProto.GetUserRequest{
		UserId: targetUserID,
	})

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "Target user not found",
		})
		log.Printf("Error unfollowing user: target user %s not found: %v", targetUserID, err)
		return
	}

	// Remove the follow relationship using the UnfollowUser endpoint
	unfollowRequest := &userProto.UnfollowUserRequest{
		FollowerId: currentUserID,
		FollowedId: targetUserID,
	}

	unfollowResp, err := UserClient.UnfollowUser(ctx, unfollowRequest)
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "You are not following this user",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to unfollow user: " + err.Error(),
		})
		log.Printf("Error unfollowing user: %v", err)
		return
	}

	log.Printf("User %s successfully unfollowed user %s", currentUserID, targetUserID)

	c.JSON(http.StatusOK, gin.H{
		"success": unfollowResp.Success,
		"message": unfollowResp.Message,
	})
}

// GetFollowers returns a user's followers
// @Summary Get followers
// @Description Gets a list of users that follow a specific user
// @Tags Users
// @Produce json
// @Router /api/v1/users/{userId}/followers [get]
func GetFollowers(c *gin.Context) {
	targetUserID := c.Param("userId")
	if targetUserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "User ID parameter is required",
		})
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Call the user service to get followers
	followersReq := &userProto.GetFollowersRequest{
		UserId: targetUserID,
		Page:   int32(page),
		Limit:  int32(limit),
	}

	followersResp, err := UserClient.GetFollowers(ctx, followersReq)
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "User not found",
			})
			return
		}

		log.Printf("Error getting followers: %v", err)
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"followers": []gin.H{},
				"page":      page,
				"limit":     limit,
				"total":     0,
			},
		})
		return
	}

	// Convert the response to the expected format
	followers := make([]gin.H, 0, len(followersResp.GetFollowers()))
	for _, user := range followersResp.GetFollowers() {
		followers = append(followers, gin.H{
			"id":                  user.GetId(),
			"username":            user.GetUsername(),
			"name":                user.GetName(),
			"profile_picture_url": user.GetProfilePictureUrl(),
			"is_following":        user.GetIsFollowing(),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"followers": followers,
			"page":      page,
			"limit":     limit,
			"total":     followersResp.GetTotalCount(),
		},
	})
}

// GetFollowing returns the users a user is following
// @Summary Get following
// @Description Gets a list of users that a user is following
// @Tags Users
// @Produce json
// @Router /api/v1/users/{userId}/following [get]
func GetFollowing(c *gin.Context) {
	targetUserID := c.Param("userId")
	if targetUserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "User ID parameter is required",
		})
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Call the user service to get following users
	followingReq := &userProto.GetFollowingRequest{
		UserId: targetUserID,
		Page:   int32(page),
		Limit:  int32(limit),
	}

	followingResp, err := UserClient.GetFollowing(ctx, followingReq)
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": "User not found",
			})
			return
		}

		log.Printf("Error getting following users: %v", err)
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"following": []gin.H{},
				"page":      page,
				"limit":     limit,
				"total":     0,
			},
		})
		return
	}

	// Convert the response to the expected format
	following := make([]gin.H, 0, len(followingResp.GetFollowing()))
	for _, user := range followingResp.GetFollowing() {
		following = append(following, gin.H{
			"id":                  user.GetId(),
			"username":            user.GetUsername(),
			"name":                user.GetName(),
			"profile_picture_url": user.GetProfilePictureUrl(),
			"is_following":        true, // By definition, we're following these users
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"following": following,
			"page":      page,
			"limit":     limit,
			"total":     followingResp.GetTotalCount(),
		},
	})
}

// LikeThread handles the API request to like a thread
// @Summary Like a thread
// @Description Adds a like to a thread
// @Tags Social
// @Accept json
// @Produce json
// @Param id path string true "Thread ID"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/threads/{id}/like [post]
func LikeThread(c *gin.Context) {
	// Get user ID from token
	userIDAny, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Success: false,
			Message: "User ID not found in token",
			Code:    "UNAUTHORIZED",
		})
		return
	}

	userID, ok := userIDAny.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Invalid User ID format in token",
			Code:    "INTERNAL_ERROR",
		})
		return
	}

	// Get thread ID from URL
	threadID := c.Param("id")
	if threadID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Message: "Thread ID is required",
			Code:    "INVALID_REQUEST",
		})
		return
	}

	// Get connection to thread service
	conn, err := threadConnPool.Get()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Failed to connect to thread service: " + err.Error(),
			Code:    "SERVICE_UNAVAILABLE",
		})
		return
	}
	defer threadConnPool.Put(conn)

	// Create thread service client
	client := threadProto.NewThreadServiceClient(conn)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Call thread service
	_, err = client.LikeThread(ctx, &threadProto.LikeThreadRequest{
		ThreadId: threadID,
		UserId:   userID,
	})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			httpStatus := http.StatusInternalServerError
			if st.Code() == codes.NotFound {
				httpStatus = http.StatusNotFound
			}
			c.JSON(httpStatus, ErrorResponse{
				Success: false,
				Message: st.Message(),
				Code:    st.Code().String(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Success: false,
				Message: "Failed to like thread: " + err.Error(),
				Code:    "INTERNAL_ERROR",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Thread liked successfully",
	})
}

// UnlikeThread handles the API request to unlike a thread
// @Summary Unlike a thread
// @Description Removes a like from a thread
// @Tags Social
// @Accept json
// @Produce json
// @Param id path string true "Thread ID"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/threads/{id}/like [delete]
func UnlikeThread(c *gin.Context) {
	// Get user ID from token
	userIDAny, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Success: false,
			Message: "User ID not found in token",
			Code:    "UNAUTHORIZED",
		})
		return
	}

	userID, ok := userIDAny.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Invalid User ID format in token",
			Code:    "INTERNAL_ERROR",
		})
		return
	}

	// Get thread ID from URL
	threadID := c.Param("id")
	if threadID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Message: "Thread ID is required",
			Code:    "INVALID_REQUEST",
		})
		return
	}

	// Get connection to thread service
	conn, err := threadConnPool.Get()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Failed to connect to thread service: " + err.Error(),
			Code:    "SERVICE_UNAVAILABLE",
		})
		return
	}
	defer threadConnPool.Put(conn)

	// Create thread service client
	client := threadProto.NewThreadServiceClient(conn)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Call thread service
	_, err = client.UnlikeThread(ctx, &threadProto.UnlikeThreadRequest{
		ThreadId: threadID,
		UserId:   userID,
	})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			httpStatus := http.StatusInternalServerError
			if st.Code() == codes.NotFound {
				httpStatus = http.StatusNotFound
			}
			c.JSON(httpStatus, ErrorResponse{
				Success: false,
				Message: st.Message(),
				Code:    st.Code().String(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Success: false,
				Message: "Failed to unlike thread: " + err.Error(),
				Code:    "INTERNAL_ERROR",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Thread unliked successfully",
	})
}

// ReplyToThread handles the API request to create a reply to a thread
// @Summary Reply to thread
// @Description Creates a new reply to a thread
// @Tags Social
// @Accept json
// @Produce json
// @Param id path string true "Thread ID"
// @Success 201 {object} threadProto.ReplyResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/threads/{id}/replies [post]
func ReplyToThread(c *gin.Context) {
	// Get user ID from token
	userIDAny, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Success: false,
			Message: "User ID not found in token",
			Code:    "UNAUTHORIZED",
		})
		return
	}

	userID, ok := userIDAny.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Invalid User ID format in token",
			Code:    "INTERNAL_ERROR",
		})
		return
	}

	// Get thread ID from URL
	threadID := c.Param("id")
	if threadID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Message: "Thread ID is required",
			Code:    "INVALID_REQUEST",
		})
		return
	}

	// Parse request
	var request struct {
		Content          string               `json:"content" binding:"required"`
		Media            []*threadProto.Media `json:"media,omitempty"`
		ParentReplyID    string               `json:"parent_reply_id,omitempty"`
		MentionedUserIDs []string             `json:"mentioned_user_ids,omitempty"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Message: "Invalid request: " + err.Error(),
			Code:    "INVALID_REQUEST",
		})
		return
	}

	// Prepare the CreateReplyRequest
	createReplyRequest := &threadProto.CreateReplyRequest{
		ThreadId:         threadID,
		UserId:           userID,
		Content:          request.Content,
		Media:            request.Media,
		MentionedUserIds: request.MentionedUserIDs,
	}

	// Add parent reply ID if provided
	if request.ParentReplyID != "" {
		createReplyRequest.ParentId = request.ParentReplyID
	}

	// Get connection to thread service
	conn, err := threadConnPool.Get()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Failed to connect to thread service: " + err.Error(),
			Code:    "SERVICE_UNAVAILABLE",
		})
		return
	}
	defer threadConnPool.Put(conn)

	// Create thread service client
	client := threadProto.NewThreadServiceClient(conn)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Call thread service
	resp, err := client.CreateReply(ctx, createReplyRequest)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			httpStatus := http.StatusInternalServerError
			if st.Code() == codes.InvalidArgument {
				httpStatus = http.StatusBadRequest
			} else if st.Code() == codes.NotFound {
				httpStatus = http.StatusNotFound
			}
			c.JSON(httpStatus, ErrorResponse{
				Success: false,
				Message: st.Message(),
				Code:    st.Code().String(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Success: false,
				Message: "Failed to create reply: " + err.Error(),
				Code:    "INTERNAL_ERROR",
			})
		}
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// GetThreadReplies handles the API request to get replies to a thread
// @Summary Get thread replies
// @Description Returns all replies for a thread
// @Tags Social
// @Produce json
// @Param id path string true "Thread ID"
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Success 200 {object} threadProto.RepliesResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/threads/{id}/replies [get]
func GetThreadReplies(c *gin.Context) {
	// Get thread ID from URL
	threadID := c.Param("id")
	if threadID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Message: "Thread ID is required",
			Code:    "INVALID_REQUEST",
		})
		return
	}

	// Get pagination parameters
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

	// Get connection to thread service
	conn, err := threadConnPool.Get()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Failed to connect to thread service: " + err.Error(),
			Code:    "SERVICE_UNAVAILABLE",
		})
		return
	}
	defer threadConnPool.Put(conn)

	// Create thread service client
	client := threadProto.NewThreadServiceClient(conn)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Call thread service
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
			c.JSON(httpStatus, ErrorResponse{
				Success: false,
				Message: st.Message(),
				Code:    st.Code().String(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Success: false,
				Message: "Failed to get thread replies: " + err.Error(),
				Code:    "INTERNAL_ERROR",
			})
		}
		return
	}

	c.JSON(http.StatusOK, resp)
}

// RepostThread handles the API request to repost a thread
// @Summary Repost a thread
// @Description Creates a repost of a thread
// @Tags Social
// @Accept json
// @Produce json
// @Param id path string true "Thread ID"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/threads/{id}/repost [post]
func RepostThread(c *gin.Context) {
	// Get user ID from token
	userIDAny, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Success: false,
			Message: "User ID not found in token",
			Code:    "UNAUTHORIZED",
		})
		return
	}

	userID, ok := userIDAny.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Invalid User ID format in token",
			Code:    "INTERNAL_ERROR",
		})
		return
	}

	// Get thread ID from URL
	threadID := c.Param("id")
	if threadID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Message: "Thread ID is required",
			Code:    "INVALID_REQUEST",
		})
		return
	}

	// Parse request
	var request struct {
		Content string `json:"content"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		// Content is optional, so just log the error and continue
		log.Printf("Warning: Failed to parse request body: %v", err)
	}

	// Get connection to thread service
	conn, err := threadConnPool.Get()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Failed to connect to thread service: " + err.Error(),
			Code:    "SERVICE_UNAVAILABLE",
		})
		return
	}
	defer threadConnPool.Put(conn)

	// Create thread service client
	client := threadProto.NewThreadServiceClient(conn)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Call thread service
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
			c.JSON(httpStatus, ErrorResponse{
				Success: false,
				Message: st.Message(),
				Code:    st.Code().String(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Success: false,
				Message: "Failed to repost thread: " + err.Error(),
				Code:    "INTERNAL_ERROR",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Thread reposted successfully",
	})
}

// RemoveRepost handles the API request to remove a repost
// @Summary Remove a repost
// @Description Removes a repost of a thread
// @Tags Social
// @Produce json
// @Param id path string true "Thread ID"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/threads/{id}/repost [delete]
func RemoveRepost(c *gin.Context) {
	// Get user ID from token
	userIDAny, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Success: false,
			Message: "User ID not found in token",
			Code:    "UNAUTHORIZED",
		})
		return
	}

	userID, ok := userIDAny.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Invalid User ID format in token",
			Code:    "INTERNAL_ERROR",
		})
		return
	}

	// Get thread ID from URL
	threadID := c.Param("id")
	if threadID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Message: "Thread ID is required",
			Code:    "INVALID_REQUEST",
		})
		return
	}

	// Get connection to thread service
	conn, err := threadConnPool.Get()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Failed to connect to thread service: " + err.Error(),
			Code:    "SERVICE_UNAVAILABLE",
		})
		return
	}
	defer threadConnPool.Put(conn)

	// Create thread service client
	client := threadProto.NewThreadServiceClient(conn)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Call thread service
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
			c.JSON(httpStatus, ErrorResponse{
				Success: false,
				Message: st.Message(),
				Code:    st.Code().String(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Success: false,
				Message: "Failed to remove repost: " + err.Error(),
				Code:    "INTERNAL_ERROR",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Repost removed successfully",
	})
}

// BookmarkThread adds a bookmark for a thread
// @Summary Bookmark a thread
// @Description Add a bookmark for a thread
// @Tags Social
// @Accept json
// @Produce json
// @Param id path string true "Thread ID"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/threads/{id}/bookmark [post]
func BookmarkThread(c *gin.Context) {
	// Get thread ID from path parameter
	threadID := c.Param("id")
	if threadID == "" {
		SendErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Thread ID is required")
		return
	}

	// Get user ID from token
	userID, exists := c.Get("userId")
	if !exists {
		SendErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User must be authenticated")
		return
	}

	// Call thread service client
	err := threadServiceClient.BookmarkThread(threadID, userID.(string))

	// Handle errors
	if err != nil {
		// Check for specific error types from the gRPC service
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.NotFound:
				SendErrorResponse(c, http.StatusNotFound, "NOT_FOUND", "Thread not found")
				return
			case codes.AlreadyExists:
				// This is not truly an error - return success with a note
				c.JSON(http.StatusOK, gin.H{
					"success": true,
					"message": "Thread was already bookmarked",
				})
				return
			default:
				SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", fmt.Sprintf("Error bookmarking thread: %v", st.Message()))
				return
			}
		}

		SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Error bookmarking thread")
		return
	}

	// Return success
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Thread bookmarked successfully",
	})
}

// RemoveBookmark removes a bookmark from a thread
// @Summary Remove a thread bookmark
// @Description Remove a bookmark from a thread
// @Tags Social
// @Accept json
// @Produce json
// @Param id path string true "Thread ID"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/threads/{id}/bookmark [delete]
func RemoveBookmark(c *gin.Context) {
	// Get thread ID from path parameter
	threadID := c.Param("id")
	if threadID == "" {
		SendErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Thread ID is required")
		return
	}

	// Get user ID from token
	userID, exists := c.Get("userId")
	if !exists {
		SendErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User must be authenticated")
		return
	}

	// Call thread service client
	err := threadServiceClient.RemoveBookmark(threadID, userID.(string))

	// Handle errors
	if err != nil {
		// Check for specific error types from the gRPC service
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.NotFound:
				// This is not truly an error for removing a bookmark - return success with a note
				c.JSON(http.StatusOK, gin.H{
					"success": true,
					"message": "Thread was not bookmarked",
				})
				return
			default:
				SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", fmt.Sprintf("Error removing bookmark: %v", st.Message()))
				return
			}
		}

		SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Error removing bookmark")
		return
	}

	// Return success
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Bookmark removed successfully",
	})
}

// GetThreadsFromFollowing retrieves threads from users that the authenticated user follows
// @Summary Get following threads
// @Description Gets threads from users that the authenticated user follows
// @Tags Threads
// @Produce json
// @Param page query int false "Page number"
// @Param limit query int false "Number of items per page"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/threads/following [get]
func GetThreadsFromFollowing(c *gin.Context) {
	// Get authenticated user ID from context
	authenticatedUserID, exists := c.Get("userId")
	if !exists {
		log.Printf("GetThreadsFromFollowing: No userId in context, returning empty results")
		// Return empty results instead of error to be more resilient
		c.JSON(http.StatusOK, gin.H{
			"threads": []gin.H{},
			"total":   0,
		})
		return
	}

	authenticatedUserIDStr, ok := authenticatedUserID.(string)
	if !ok {
		log.Printf("GetThreadsFromFollowing: Invalid userId type, returning empty results")
		// Return empty results instead of error to be more resilient
		c.JSON(http.StatusOK, gin.H{
			"threads": []gin.H{},
			"total":   0,
		})
		return
	}

	// Get pagination parameters
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

	// Return empty array since this endpoint is not yet fully implemented
	c.JSON(http.StatusOK, gin.H{
		"threads": []gin.H{},
		"total":   0,
		"page":    page,
		"limit":   limit,
	})
}

// LikeReply handles the API request to like a reply
// @Summary Like a reply
// @Description Adds a like to a reply
// @Tags Social
// @Produce json
// @Param id path string true "Reply ID"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/replies/{id}/like [post]
func LikeReply(c *gin.Context) {
	// Get user ID from token
	userIDAny, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Success: false,
			Message: "User ID not found in token",
			Code:    "UNAUTHORIZED",
		})
		return
	}

	userID, ok := userIDAny.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Invalid User ID format in token",
			Code:    "INTERNAL_ERROR",
		})
		return
	}

	// Get reply ID from URL
	replyID := c.Param("id")
	if replyID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Message: "Reply ID is required",
			Code:    "INVALID_REQUEST",
		})
		return
	}

	// Get connection to thread service
	conn, err := threadConnPool.Get()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Failed to connect to thread service: " + err.Error(),
			Code:    "SERVICE_UNAVAILABLE",
		})
		return
	}
	defer threadConnPool.Put(conn)

	// Create thread service client
	client := threadProto.NewThreadServiceClient(conn)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Call thread service
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
			c.JSON(httpStatus, ErrorResponse{
				Success: false,
				Message: st.Message(),
				Code:    st.Code().String(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Success: false,
				Message: "Failed to like reply: " + err.Error(),
				Code:    "INTERNAL_ERROR",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Reply liked successfully",
	})
}

// UnlikeReply handles the API request to unlike a reply
// @Summary Unlike a reply
// @Description Removes a like from a reply
// @Tags Social
// @Produce json
// @Param id path string true "Reply ID"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/replies/{id}/like [delete]
func UnlikeReply(c *gin.Context) {
	// Get user ID from token
	userIDAny, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Success: false,
			Message: "User ID not found in token",
			Code:    "UNAUTHORIZED",
		})
		return
	}

	userID, ok := userIDAny.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Invalid User ID format in token",
			Code:    "INTERNAL_ERROR",
		})
		return
	}

	// Get reply ID from URL
	replyID := c.Param("id")
	if replyID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Message: "Reply ID is required",
			Code:    "INVALID_REQUEST",
		})
		return
	}

	// Get connection to thread service
	conn, err := threadConnPool.Get()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Failed to connect to thread service: " + err.Error(),
			Code:    "SERVICE_UNAVAILABLE",
		})
		return
	}
	defer threadConnPool.Put(conn)

	// Create thread service client
	client := threadProto.NewThreadServiceClient(conn)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Call thread service
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
			c.JSON(httpStatus, ErrorResponse{
				Success: false,
				Message: st.Message(),
				Code:    st.Code().String(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Success: false,
				Message: "Failed to unlike reply: " + err.Error(),
				Code:    "INTERNAL_ERROR",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Reply unliked successfully",
	})
}

// BookmarkReply handles the API request to bookmark a reply
// @Summary Bookmark a reply
// @Description Adds a bookmark for a reply
// @Tags Social
// @Produce json
// @Param id path string true "Reply ID"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/replies/{id}/bookmark [post]
func BookmarkReply(c *gin.Context) {
	// Get user ID from token
	userIDAny, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Success: false,
			Message: "User ID not found in token",
			Code:    "UNAUTHORIZED",
		})
		return
	}

	userID, ok := userIDAny.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Invalid User ID format in token",
			Code:    "INTERNAL_ERROR",
		})
		return
	}

	// Get reply ID from URL
	replyID := c.Param("id")
	if replyID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Message: "Reply ID is required",
			Code:    "INVALID_REQUEST",
		})
		return
	}

	// Get connection to thread service
	conn, err := threadConnPool.Get()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Failed to connect to thread service: " + err.Error(),
			Code:    "SERVICE_UNAVAILABLE",
		})
		return
	}
	defer threadConnPool.Put(conn)

	// Create thread service client
	client := threadProto.NewThreadServiceClient(conn)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Call thread service
	_, err = client.BookmarkReply(ctx, &threadProto.BookmarkReplyRequest{
		ReplyId: replyID,
		UserId:  userID,
	})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			httpStatus := http.StatusInternalServerError
			if st.Code() == codes.NotFound {
				httpStatus = http.StatusNotFound
			}
			c.JSON(httpStatus, ErrorResponse{
				Success: false,
				Message: st.Message(),
				Code:    st.Code().String(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Success: false,
				Message: "Failed to bookmark reply: " + err.Error(),
				Code:    "INTERNAL_ERROR",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Reply bookmarked successfully",
	})
}

// RemoveReplyBookmark handles the API request to remove a bookmark from a reply
// @Summary Remove a reply bookmark
// @Description Removes a bookmark for a reply
// @Tags Social
// @Produce json
// @Param id path string true "Reply ID"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/replies/{id}/bookmark [delete]
func RemoveReplyBookmark(c *gin.Context) {
	// Get user ID from token
	userIDAny, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Success: false,
			Message: "User ID not found in token",
			Code:    "UNAUTHORIZED",
		})
		return
	}

	userID, ok := userIDAny.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Invalid User ID format in token",
			Code:    "INTERNAL_ERROR",
		})
		return
	}

	// Get reply ID from URL
	replyID := c.Param("id")
	if replyID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Message: "Reply ID is required",
			Code:    "INVALID_REQUEST",
		})
		return
	}

	// Get connection to thread service
	conn, err := threadConnPool.Get()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Failed to connect to thread service: " + err.Error(),
			Code:    "SERVICE_UNAVAILABLE",
		})
		return
	}
	defer threadConnPool.Put(conn)

	// Create thread service client
	client := threadProto.NewThreadServiceClient(conn)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Call thread service
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
			c.JSON(httpStatus, ErrorResponse{
				Success: false,
				Message: st.Message(),
				Code:    st.Code().String(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Success: false,
				Message: "Failed to remove reply bookmark: " + err.Error(),
				Code:    "INTERNAL_ERROR",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Reply bookmark removed successfully",
	})
}

// SearchSocialUsers searches for users based on query and filters for social contexts
// @Summary Search users for social features
// @Description Search for users by name, username, or email in a social context
// @Tags Users,Social
// @Accept json
// @Produce json
// @Router /api/v1/social/users/search [get]
func SearchSocialUsers(c *gin.Context) {
	// Get search parameters from query
	query := c.Query("q")
	if query == "" {
		SendErrorResponse(c, http.StatusBadRequest, "BAD_REQUEST", "Search query is required")
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

	// Use the user service client to search users
	users, totalCount, err := userServiceClient.SearchUsers(query, filter, page, limit)
	if err != nil {
		log.Printf("Error searching users: %v", err)
		SendErrorResponse(c, http.StatusInternalServerError, "SERVER_ERROR", "Failed to search users")
		return
	}

	SendSuccessResponse(c, http.StatusOK, gin.H{
		"users": users,
		"pagination": gin.H{
			"total":   totalCount,
			"page":    page,
			"limit":   limit,
			"hasMore": len(users) == limit && (page*limit) < totalCount,
		},
	})
}

// @Summary Pin reply to profile
// @Description Pins a reply to the user's profile
// @Tags Social
// @Produce json
// @Param id path string true "Reply ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/replies/{id}/pin [post]
func PinReply(c *gin.Context) {
	// Get user ID from token
	userIDAny, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Success: false,
			Message: "User ID not found in token",
			Code:    "UNAUTHORIZED",
		})
		return
	}

	userID, ok := userIDAny.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Invalid User ID format in token",
			Code:    "INTERNAL_ERROR",
		})
		return
	}

	// Get reply ID from URL
	replyID := c.Param("id")
	if replyID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Message: "Reply ID is required",
			Code:    "INVALID_REQUEST",
		})
		return
	}

	// Check if the thread service client is available
	if threadServiceClient == nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Thread service client not initialized",
			Code:    "SERVICE_UNAVAILABLE",
		})
		return
	}

	// Use the interface implementation
	err := threadServiceClient.PinReply(replyID, userID)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			httpStatus := http.StatusInternalServerError
			if st.Code() == codes.NotFound {
				httpStatus = http.StatusNotFound
			}
			c.JSON(httpStatus, ErrorResponse{
				Success: false,
				Message: st.Message(),
				Code:    st.Code().String(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Success: false,
				Message: "Failed to pin reply: " + err.Error(),
				Code:    "INTERNAL_ERROR",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Reply pinned successfully",
	})
}

// @Summary Unpin reply from profile
// @Description Unpins a reply from the user's profile
// @Tags Social
// @Produce json
// @Param id path string true "Reply ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/replies/{id}/pin [delete]
func UnpinReply(c *gin.Context) {
	// Get user ID from token
	userIDAny, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Success: false,
			Message: "User ID not found in token",
			Code:    "UNAUTHORIZED",
		})
		return
	}

	userID, ok := userIDAny.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Invalid User ID format in token",
			Code:    "INTERNAL_ERROR",
		})
		return
	}

	// Get reply ID from URL
	replyID := c.Param("id")
	if replyID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Message: "Reply ID is required",
			Code:    "INVALID_REQUEST",
		})
		return
	}

	// Check if the thread service client is available
	if threadServiceClient == nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Thread service client not initialized",
			Code:    "SERVICE_UNAVAILABLE",
		})
		return
	}

	// Use the interface implementation
	err := threadServiceClient.UnpinReply(replyID, userID)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			httpStatus := http.StatusInternalServerError
			if st.Code() == codes.NotFound {
				httpStatus = http.StatusNotFound
			}
			c.JSON(httpStatus, ErrorResponse{
				Success: false,
				Message: st.Message(),
				Code:    st.Code().String(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Success: false,
				Message: "Failed to unpin reply: " + err.Error(),
				Code:    "INTERNAL_ERROR",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Reply unpinned successfully",
	})
}

// GetRepliesByParentReply handles the API request to get replies to a specific reply
// @Summary Get replies to a reply
// @Description Returns all replies for a specific parent reply
// @Tags Social
// @Produce json
// @Param id path string true "Parent Reply ID"
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Success 200 {object} threadProto.RepliesResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/replies/{id}/replies [get]
func GetRepliesByParentReply(c *gin.Context) {
	// Get reply ID from URL
	parentReplyID := c.Param("id")
	if parentReplyID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Message: "Parent Reply ID is required",
			Code:    "INVALID_REQUEST",
		})
		return
	}

	// Get pagination parameters
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

	// Get connection to thread service
	conn, err := threadConnPool.Get()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Failed to connect to thread service: " + err.Error(),
			Code:    "SERVICE_UNAVAILABLE",
		})
		return
	}
	defer threadConnPool.Put(conn)

	// Create thread service client
	client := threadProto.NewThreadServiceClient(conn)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Call thread service
	resp, err := client.GetRepliesByParentReply(ctx, &threadProto.GetRepliesByParentReplyRequest{
		ParentReplyId: parentReplyID,
		Page:          int32(page),
		Limit:         int32(limit),
	})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			httpStatus := http.StatusInternalServerError
			if st.Code() == codes.NotFound {
				httpStatus = http.StatusNotFound
			}
			c.JSON(httpStatus, ErrorResponse{
				Success: false,
				Message: st.Message(),
				Code:    st.Code().String(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Success: false,
				Message: "Failed to get reply replies: " + err.Error(),
				Code:    "INTERNAL_ERROR",
			})
		}
		return
	}

	c.JSON(http.StatusOK, resp)
}
