package handlers

import (
	threadProto "aycom/backend/proto/thread"
	userProto "aycom/backend/proto/user"
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func FollowUser(c *gin.Context) {
	targetUserID := c.Param("userId")
	if targetUserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "User ID parameter is required",
		})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "User not authenticated",
		})
		return
	}
	currentUserID := userID.(string)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

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

func UnfollowUser(c *gin.Context) {
	targetUserID := c.Param("userId")
	if targetUserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "User ID parameter is required",
		})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "User not authenticated",
		})
		return
	}
	currentUserID := userID.(string)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

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

func GetFollowers(c *gin.Context) {
	targetUserID := c.Param("userId")
	if targetUserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "User ID parameter is required",
		})
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

func GetFollowing(c *gin.Context) {
	targetUserID := c.Param("userId")
	if targetUserID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "User ID parameter is required",
		})
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

	following := make([]gin.H, 0, len(followingResp.GetFollowing()))
	for _, user := range followingResp.GetFollowing() {
		following = append(following, gin.H{
			"id":                  user.GetId(),
			"username":            user.GetUsername(),
			"name":                user.GetName(),
			"profile_picture_url": user.GetProfilePictureUrl(),
			"is_following":        true,
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

func LikeThread(c *gin.Context) {

	userIDAny, exists := c.Get("userId")
	if !exists {
		log.Printf("LikeThread: No userId in context")
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Success: false,
			Message: "User ID not found in token",
			Code:    "UNAUTHORIZED",
		})
		return
	}

	userID, ok := userIDAny.(string)
	if !ok {
		log.Printf("LikeThread: Invalid userId format in context")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Invalid User ID format in token",
			Code:    "INTERNAL_ERROR",
		})
		return
	}

	threadID := c.Param("id")
	if threadID == "" {
		log.Printf("LikeThread: Missing threadId parameter")
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Message: "Thread ID is required",
			Code:    "INVALID_REQUEST",
		})
		return
	}

	log.Printf("LikeThread: Processing like for thread %s by user %s", threadID, userID)

	conn, err := threadConnPool.Get()
	if err != nil {
		log.Printf("LikeThread: Failed to get connection to thread service: %v", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Failed to connect to thread service: " + err.Error(),
			Code:    "SERVICE_UNAVAILABLE",
		})
		return
	}
	defer threadConnPool.Put(conn)

	client := threadProto.NewThreadServiceClient(conn)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	_, err = client.LikeThread(ctx, &threadProto.LikeThreadRequest{
		ThreadId: threadID,
		UserId:   userID,
	})

	if err != nil {

		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.NotFound:
				log.Printf("LikeThread: Thread %s not found", threadID)
				c.JSON(http.StatusNotFound, ErrorResponse{
					Success: false,
					Message: "Thread not found",
					Code:    "NOT_FOUND",
				})
				return
			case codes.AlreadyExists:

				log.Printf("LikeThread: Thread %s already liked by user %s", threadID, userID)
				c.JSON(http.StatusOK, gin.H{
					"success": true,
					"message": "Thread already liked",
					"code":    "ALREADY_LIKED",
				})
				return
			case codes.InvalidArgument:
				log.Printf("LikeThread: Invalid argument - %s", st.Message())
				c.JSON(http.StatusBadRequest, ErrorResponse{
					Success: false,
					Message: "Invalid request: " + st.Message(),
					Code:    "INVALID_REQUEST",
				})
				return
			default:
				log.Printf("LikeThread: Error from thread service - %s", st.Message())
				c.JSON(http.StatusInternalServerError, ErrorResponse{
					Success: false,
					Message: "Failed to like thread: " + st.Message(),
					Code:    "INTERNAL_ERROR",
				})
				return
			}
		}

		log.Printf("LikeThread: Unclassified error - %v", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Failed to like thread: " + err.Error(),
			Code:    "INTERNAL_ERROR",
		})
		return
	}

	log.Printf("LikeThread: Successfully liked thread %s by user %s", threadID, userID)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Thread liked successfully",
		"data": gin.H{
			"thread_id": threadID,
			"user_id":   userID,
		},
	})
}

func UnlikeThread(c *gin.Context) {

	userIDAny, exists := c.Get("userId")
	if !exists {
		log.Printf("UnlikeThread: No userId in context")
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Success: false,
			Message: "User ID not found in token",
			Code:    "UNAUTHORIZED",
		})
		return
	}

	userID, ok := userIDAny.(string)
	if !ok {
		log.Printf("UnlikeThread: Invalid userId format in context")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Invalid User ID format in token",
			Code:    "INTERNAL_ERROR",
		})
		return
	}

	threadID := c.Param("id")
	if threadID == "" {
		log.Printf("UnlikeThread: Missing threadId parameter")
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Message: "Thread ID is required",
			Code:    "INVALID_REQUEST",
		})
		return
	}

	log.Printf("UnlikeThread: Processing unlike for thread %s by user %s", threadID, userID)

	conn, err := threadConnPool.Get()
	if err != nil {
		log.Printf("UnlikeThread: Failed to get connection to thread service: %v", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Failed to connect to thread service: " + err.Error(),
			Code:    "SERVICE_UNAVAILABLE",
		})
		return
	}
	defer threadConnPool.Put(conn)

	client := threadProto.NewThreadServiceClient(conn)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	_, err = client.UnlikeThread(ctx, &threadProto.UnlikeThreadRequest{
		ThreadId: threadID,
		UserId:   userID,
	})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.NotFound:
				log.Printf("UnlikeThread: Thread %s or like not found", threadID)

				c.JSON(http.StatusOK, gin.H{
					"success": true,
					"message": "Thread already not liked",
				})
				return
			case codes.InvalidArgument:
				log.Printf("UnlikeThread: Invalid argument - %s", st.Message())
				c.JSON(http.StatusBadRequest, ErrorResponse{
					Success: false,
					Message: "Invalid request: " + st.Message(),
					Code:    "INVALID_REQUEST",
				})
				return
			default:
				log.Printf("UnlikeThread: Error from thread service - %s", st.Message())
				c.JSON(http.StatusInternalServerError, ErrorResponse{
					Success: false,
					Message: "Failed to unlike thread: " + st.Message(),
					Code:    "INTERNAL_ERROR",
				})
				return
			}
		} else {
			log.Printf("UnlikeThread: Unclassified error - %v", err)
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Success: false,
				Message: "Failed to unlike thread: " + err.Error(),
				Code:    "INTERNAL_ERROR",
			})
			return
		}
	}

	log.Printf("UnlikeThread: Successfully unliked thread %s by user %s", threadID, userID)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Thread unliked successfully",
	})
}

func ReplyToThread(c *gin.Context) {

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

	threadID := c.Param("id")
	if threadID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Message: "Thread ID is required",
			Code:    "INVALID_REQUEST",
		})
		return
	}

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
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Failed to connect to thread service: " + err.Error(),
			Code:    "SERVICE_UNAVAILABLE",
		})
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

func GetThreadReplies(c *gin.Context) {

	threadID := c.Param("id")
	if threadID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Message: "Thread ID is required",
			Code:    "INVALID_REQUEST",
		})
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
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Failed to connect to thread service: " + err.Error(),
			Code:    "SERVICE_UNAVAILABLE",
		})
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

func RepostThread(c *gin.Context) {

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

	threadID := c.Param("id")
	if threadID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Message: "Thread ID is required",
			Code:    "INVALID_REQUEST",
		})
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
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Failed to connect to thread service: " + err.Error(),
			Code:    "SERVICE_UNAVAILABLE",
		})
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

func RemoveRepost(c *gin.Context) {

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

	threadID := c.Param("id")
	if threadID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Message: "Thread ID is required",
			Code:    "INVALID_REQUEST",
		})
		return
	}

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

func BookmarkThread(c *gin.Context) {

	threadID := c.Param("id")
	if threadID == "" {
		SendErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Thread ID is required")
		return
	}

	userID, exists := c.Get("userId")
	if !exists {
		SendErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User must be authenticated")
		return
	}

	log.Printf("BookmarkThread: Attempting to bookmark thread %s for user %s", threadID, userID)

	err := threadServiceClient.BookmarkThread(threadID, userID.(string))

	if err != nil {

		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.NotFound:
				SendErrorResponse(c, http.StatusNotFound, "NOT_FOUND", "Thread not found")
				return
			case codes.AlreadyExists:

				c.JSON(http.StatusOK, gin.H{
					"success": true,
					"message": "Thread was already bookmarked",
					"code":    "ALREADY_BOOKMARKED",
				})
				return
			case codes.InvalidArgument:
				SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", st.Message())
				return
			case codes.PermissionDenied:
				SendErrorResponse(c, http.StatusForbidden, "FORBIDDEN", "You do not have permission to bookmark this thread")
				return
			default:
				SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", fmt.Sprintf("Error bookmarking thread: %v", st.Message()))
				return
			}
		}

		log.Printf("Error in BookmarkThread: %v", err)
		SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Error bookmarking thread")
		return
	}

	log.Printf("Successfully bookmarked thread %s for user %s", threadID, userID)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Thread bookmarked successfully",
	})
}

func RemoveBookmark(c *gin.Context) {

	threadID := c.Param("id")
	if threadID == "" {
		SendErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Thread ID is required")
		return
	}

	userID, exists := c.Get("userId")
	if !exists {
		SendErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User must be authenticated")
		return
	}

	err := threadServiceClient.RemoveBookmark(threadID, userID.(string))

	if err != nil {

		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.NotFound:

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

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Bookmark removed successfully",
	})
}

func GetThreadsFromFollowing(c *gin.Context) {

	authenticatedUserID, exists := c.Get("userId")
	if !exists {
		log.Printf("GetThreadsFromFollowing: No userId in context, returning empty results")

		c.JSON(http.StatusOK, gin.H{
			"threads": []gin.H{},
			"total":   0,
		})
		return
	}

	authenticatedUserIDStr, ok := authenticatedUserID.(string)
	if !ok {
		log.Printf("GetThreadsFromFollowing: Invalid userId type, returning empty results")

		c.JSON(http.StatusOK, gin.H{
			"threads": []gin.H{},
			"total":   0,
		})
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

	c.JSON(http.StatusOK, gin.H{
		"threads": []gin.H{},
		"total":   0,
		"page":    page,
		"limit":   limit,
	})
}

func LikeReply(c *gin.Context) {

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

	replyID := c.Param("id")
	if replyID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Message: "Reply ID is required",
			Code:    "INVALID_REQUEST",
		})
		return
	}

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

func UnlikeReply(c *gin.Context) {

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

	replyID := c.Param("id")
	if replyID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Message: "Reply ID is required",
			Code:    "INVALID_REQUEST",
		})
		return
	}

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

func BookmarkReply(c *gin.Context) {

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

	replyID := c.Param("id")
	if replyID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Message: "Reply ID is required",
			Code:    "INVALID_REQUEST",
		})
		return
	}

	log.Printf("BookmarkReply: Attempting to bookmark reply %s for user %s", replyID, userID)

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
				c.JSON(http.StatusNotFound, ErrorResponse{
					Success: false,
					Message: "Reply not found",
					Code:    "NOT_FOUND",
				})
				return
			case codes.AlreadyExists:

				c.JSON(http.StatusOK, gin.H{
					"success": true,
					"message": "Reply already bookmarked",
					"code":    "ALREADY_BOOKMARKED",
				})
				return
			default:
				c.JSON(http.StatusInternalServerError, ErrorResponse{
					Success: false,
					Message: "Failed to bookmark reply: " + st.Message(),
					Code:    "INTERNAL_ERROR",
				})
				return
			}
		}

		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Failed to bookmark reply: " + err.Error(),
			Code:    "INTERNAL_ERROR",
		})
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

	replyID := c.Param("id")
	if replyID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Message: "Reply ID is required",
			Code:    "INVALID_REQUEST",
		})
		return
	}

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

func SearchSocialUsers(c *gin.Context) {

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

func PinReply(c *gin.Context) {

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

	replyID := c.Param("id")
	if replyID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Message: "Reply ID is required",
			Code:    "INVALID_REQUEST",
		})
		return
	}

	if threadServiceClient == nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Thread service client not initialized",
			Code:    "SERVICE_UNAVAILABLE",
		})
		return
	}

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

func UnpinReply(c *gin.Context) {

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

	replyID := c.Param("id")
	if replyID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Message: "Reply ID is required",
			Code:    "INVALID_REQUEST",
		})
		return
	}

	if threadServiceClient == nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Thread service client not initialized",
			Code:    "SERVICE_UNAVAILABLE",
		})
		return
	}

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

func GetRepliesByParentReply(c *gin.Context) {

	parentReplyID := c.Param("id")
	if parentReplyID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Message: "Parent Reply ID is required",
			Code:    "INVALID_REQUEST",
		})
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
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Failed to connect to thread service: " + err.Error(),
			Code:    "SERVICE_UNAVAILABLE",
		})
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

	c.Header("Cache-Control", "public, max-age=10")

	c.JSON(http.StatusOK, resp)
}
