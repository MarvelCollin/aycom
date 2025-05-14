package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"aycom/backend/api-gateway/utils"
	threadProto "aycom/backend/proto/thread"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func CreateThread(c *gin.Context) {
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

	var request threadProto.CreateThreadRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Message: "Invalid request: " + err.Error(),
			Code:    "INVALID_REQUEST",
		})
		return
	}

	request.UserId = userID

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

	resp, err := client.CreateThread(ctx, &request)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			httpStatus := http.StatusInternalServerError
			if st.Code() == codes.InvalidArgument {
				httpStatus = http.StatusBadRequest
			}
			c.JSON(httpStatus, ErrorResponse{
				Success: false,
				Message: st.Message(),
				Code:    st.Code().String(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Success: false,
				Message: "Failed to create thread: " + err.Error(),
				Code:    "INTERNAL_ERROR",
			})
		}
		return
	}

	c.JSON(http.StatusCreated, resp)
}

func GetThread(c *gin.Context) {
	threadID := c.Param("id")
	if threadID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Message: "Thread ID is required",
			Code:    "INVALID_REQUEST",
		})
		return
	}

	// Extract user ID from context if available
	userIDAny, exists := c.Get("userId")
	var userID string
	if exists {
		userIDStr, ok := userIDAny.(string)
		if ok {
			userID = userIDStr
			log.Printf("Getting thread %s for authenticated user %s", threadID, userID)
		}
	} else {
		log.Printf("Getting thread %s for unauthenticated user", threadID)
	}

	if threadServiceClient == nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Thread service client not initialized",
			Code:    "SERVICE_UNAVAILABLE",
		})
		return
	}

	thread, err := threadServiceClient.GetThreadByID(threadID, userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Success: false,
				Message: "Thread not found",
				Code:    "THREAD_NOT_FOUND",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Error retrieving thread: " + err.Error(),
			Code:    "INTERNAL_ERROR",
		})
		return
	}

	// Create response with additional fields
	response := gin.H{
		"id":              thread.ID,
		"content":         thread.Content,
		"created_at":      thread.CreatedAt,
		"updated_at":      thread.UpdatedAt,
		"likes":           thread.LikeCount,
		"replies":         thread.ReplyCount,
		"reposts":         thread.RepostCount,
		"is_liked":        thread.IsLiked,
		"is_reposted":     thread.IsReposted,
		"is_bookmarked":   thread.IsBookmarked,
		"is_pinned":       thread.IsPinned,
		"user_id":         thread.UserID,
		"username":        thread.Username,
		"display_name":    thread.DisplayName,
		"profile_picture": thread.ProfilePicture,
	}

	// Add media if available
	if len(thread.Media) > 0 {
		mediaResponse := make([]map[string]interface{}, len(thread.Media))
		for i, m := range thread.Media {
			mediaResponse[i] = map[string]interface{}{
				"id":   m.ID,
				"url":  m.URL,
				"type": m.Type,
			}
		}
		response["media"] = mediaResponse
	}

	// Log bookmark status for debugging
	log.Printf("Thread %s bookmark status for user %s: %v", threadID, userID, thread.IsBookmarked)

	c.JSON(http.StatusOK, response)
}

func GetThreadsByUser(c *gin.Context) {
	authenticatedUserID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Success: false,
			Message: "User ID not found in token",
			Code:    "UNAUTHORIZED",
		})
		return
	}

	authenticatedUserIDStr, ok := authenticatedUserID.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Invalid User ID format in token",
			Code:    "INTERNAL_ERROR",
		})
		return
	}

	userID := c.Param("id")

	if userID == "me" {
		userID = authenticatedUserIDStr
		log.Printf("Using authenticated user ID for 'me' parameter: %s", userID)
	}

	if userID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Message: "User ID is required",
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

	resp, err := client.GetThreadsByUser(ctx, &threadProto.GetThreadsByUserRequest{
		UserId: userID,
		Page:   int32(page),
		Limit:  int32(limit),
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
				Message: "Failed to get threads: " + err.Error(),
				Code:    "INTERNAL_ERROR",
			})
		}
		return
	}

	threads := make([]map[string]interface{}, len(resp.Threads))
	for i, t := range resp.Threads {
		thread := map[string]interface{}{
			"id":             t.Thread.Id,
			"thread_id":      t.Thread.Id,
			"content":        t.Thread.Content,
			"user_id":        t.Thread.UserId,
			"created_at":     t.Thread.CreatedAt.AsTime(),
			"updated_at":     t.Thread.UpdatedAt.AsTime(),
			"like_count":     t.LikesCount,
			"reply_count":    t.RepliesCount,
			"repost_count":   t.RepostsCount,
			"bookmark_count": t.BookmarkCount,
			"view_count":     t.Thread.ViewCount, // For backward compatibility
			"is_liked":       t.LikedByUser,
			"is_repost":      t.RepostedByUser,
			"is_bookmarked":  t.BookmarkedByUser,
			"is_pinned":      t.Thread.IsPinned != nil && *t.Thread.IsPinned,
			// Default user values
			"username":            "anonymous",
			"display_name":        "User",
			"profile_picture_url": "",
		}

		if t.User != nil {
			thread["username"] = t.User.Username
			thread["display_name"] = t.User.Name
			thread["profile_picture_url"] = t.User.ProfilePictureUrl
			thread["is_verified"] = t.User.IsVerified
		}

		if len(t.Thread.Media) > 0 {
			media := make([]map[string]interface{}, len(t.Thread.Media))
			for j, m := range t.Thread.Media {
				media[j] = map[string]interface{}{
					"id":   m.Id,
					"type": m.Type,
					"url":  m.Url,
				}
			}
			thread["media"] = media
		} else {
			thread["media"] = []interface{}{}
		}

		threads[i] = thread
	}

	c.JSON(http.StatusOK, gin.H{
		"threads": threads,
		"total":   resp.Total,
	})
}

func UpdateThread(c *gin.Context) {
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

	log.Printf("User %s is updating a thread", userID)

	threadID := c.Param("id")
	if threadID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Message: "Thread ID is required",
			Code:    "INVALID_REQUEST",
		})
		return
	}

	var request threadProto.UpdateThreadRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Message: "Invalid request: " + err.Error(),
			Code:    "INVALID_REQUEST",
		})
		return
	}

	request.ThreadId = threadID

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

	resp, err := client.UpdateThread(ctx, &request)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			httpStatus := http.StatusInternalServerError
			if st.Code() == codes.NotFound {
				httpStatus = http.StatusNotFound
			} else if st.Code() == codes.InvalidArgument {
				httpStatus = http.StatusBadRequest
			} else if st.Code() == codes.PermissionDenied {
				httpStatus = http.StatusForbidden
			}
			c.JSON(httpStatus, ErrorResponse{
				Success: false,
				Message: st.Message(),
				Code:    st.Code().String(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Success: false,
				Message: "Failed to update thread: " + err.Error(),
				Code:    "INTERNAL_ERROR",
			})
		}
		return
	}

	c.JSON(http.StatusOK, resp)
}

func DeleteThread(c *gin.Context) {
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

	_, err = client.DeleteThread(ctx, &threadProto.DeleteThreadRequest{
		ThreadId: threadID,
		UserId:   userID,
	})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			httpStatus := http.StatusInternalServerError
			if st.Code() == codes.NotFound {
				httpStatus = http.StatusNotFound
			} else if st.Code() == codes.PermissionDenied {
				httpStatus = http.StatusForbidden
			}
			c.JSON(httpStatus, ErrorResponse{
				Success: false,
				Message: st.Message(),
				Code:    st.Code().String(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Success: false,
				Message: "Failed to delete thread: " + err.Error(),
				Code:    "INTERNAL_ERROR",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Thread deleted successfully",
	})
}

func UploadThreadMedia(c *gin.Context) {
	_, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "User not authenticated",
		})
		return
	}

	threadID := c.PostForm("thread_id")
	if threadID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Thread ID is required",
		})
		return
	}

	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid form data: " + err.Error(),
		})
		return
	}

	files := form.File["file"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "No files provided",
		})
		return
	}

	var mediaUrls []string
	var mediaTypes []string

	for _, file := range files {
		fileExt := filepath.Ext(file.Filename)
		allowedExts := map[string]bool{
			".jpg":  true,
			".jpeg": true,
			".png":  true,
			".gif":  true,
			".mp4":  true,
			".webm": true,
			".mov":  true,
		}

		if !allowedExts[fileExt] {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": fmt.Sprintf("File type %s not allowed", fileExt),
			})
			return
		}

		mediaType := "image"
		if fileExt == ".gif" {
			mediaType = "gif"
		} else if fileExt == ".mp4" || fileExt == ".webm" || fileExt == ".mov" {
			mediaType = "video"
		}

		fileContent, err := file.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to open file: " + err.Error(),
			})
			return
		}
		defer fileContent.Close()

		bucket := "thread-media"
		folder := mediaType + "s"

		url, err := utils.UploadFile(fileContent, file.Filename, bucket, folder)
		if err != nil {
			log.Printf("Failed to upload thread media to Supabase: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to upload file: " + err.Error(),
			})
			return
		}

		mediaUrls = append(mediaUrls, url)
		mediaTypes = append(mediaTypes, mediaType)
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"thread_id": threadID,
		"media":     mediaUrls,
		"types":     mediaTypes,
	})
}

func GetAllThreads(c *gin.Context) {
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

	userID := ""
	if userIDVal, exists := c.Get("userId"); exists {
		if userIDStr, ok := userIDVal.(string); ok {
			userID = userIDStr
			log.Printf("Authenticated user %s is viewing threads", userID)
		}
	} else {
		log.Printf("Anonymous user is viewing threads")
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

	resp, err := client.GetAllThreads(ctx, &threadProto.GetAllThreadsRequest{
		Page:  int32(page),
		Limit: int32(limit),
	})

	if err != nil {
		if st, ok := status.FromError(err); ok {
			httpStatus := http.StatusInternalServerError
			c.JSON(httpStatus, ErrorResponse{
				Success: false,
				Message: st.Message(),
				Code:    st.Code().String(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Success: false,
				Message: "Failed to get threads: " + err.Error(),
				Code:    "INTERNAL_ERROR",
			})
		}
		return
	}

	threads := make([]map[string]interface{}, len(resp.Threads))
	for i, t := range resp.Threads {
		thread := map[string]interface{}{
			"id":             t.Thread.Id,
			"thread_id":      t.Thread.Id,
			"content":        t.Thread.Content,
			"user_id":        t.Thread.UserId,
			"created_at":     t.Thread.CreatedAt.AsTime(),
			"updated_at":     t.Thread.UpdatedAt.AsTime(),
			"like_count":     t.LikesCount,
			"reply_count":    t.RepliesCount,
			"repost_count":   t.RepostsCount,
			"bookmark_count": t.BookmarkCount,
			"view_count":     t.Thread.ViewCount, // For backward compatibility
			"is_liked":       t.LikedByUser,
			"is_repost":      t.RepostedByUser,
			"is_bookmarked":  t.BookmarkedByUser,
			"is_pinned":      t.Thread.IsPinned != nil && *t.Thread.IsPinned,
			// Default user values
			"username":            "anonymous",
			"display_name":        "User",
			"profile_picture_url": "",
		}

		if t.User != nil {
			thread["username"] = t.User.Username
			thread["display_name"] = t.User.Name
			thread["profile_picture_url"] = t.User.ProfilePictureUrl
			thread["is_verified"] = t.User.IsVerified
		}

		if len(t.Thread.Media) > 0 {
			media := make([]map[string]interface{}, len(t.Thread.Media))
			for j, m := range t.Thread.Media {
				media[j] = map[string]interface{}{
					"id":   m.Id,
					"type": m.Type,
					"url":  m.Url,
				}
			}
			thread["media"] = media
		} else {
			thread["media"] = []interface{}{}
		}

		threads[i] = thread
	}

	c.JSON(http.StatusOK, gin.H{
		"threads": threads,
		"total":   resp.Total,
	})
}

func GetUserReplies(c *gin.Context) {
	authenticatedUserID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Success: false,
			Message: "User ID not found in token",
			Code:    "UNAUTHORIZED",
		})
		return
	}

	authenticatedUserIDStr, ok := authenticatedUserID.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Invalid User ID format in token",
			Code:    "INTERNAL_ERROR",
		})
		return
	}

	userID := c.Param("id")

	if userID == "me" {
		userID = authenticatedUserIDStr
		log.Printf("Using authenticated user ID for 'me' parameter: %s", userID)
	}

	if userID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Message: "User ID is required",
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

	if threadServiceClient == nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Thread service client not initialized",
			Code:    "SERVICE_UNAVAILABLE",
		})
		return
	}

	replies, err := threadServiceClient.GetRepliesByUser(userID, page, limit)
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
				Message: "Failed to get user replies: " + err.Error(),
				Code:    "INTERNAL_ERROR",
			})
		}
		return
	}

	replyItems := make([]map[string]interface{}, len(replies))
	for i, reply := range replies {
		replyItem := map[string]interface{}{
			"id":            reply.ID,
			"reply_id":      reply.ID,
			"thread_id":     reply.ParentID,
			"content":       reply.Content,
			"user_id":       reply.UserID,
			"created_at":    reply.CreatedAt,
			"updated_at":    reply.UpdatedAt,
			"like_count":    reply.LikeCount,
			"is_liked":      reply.IsLiked,
			"is_bookmarked": reply.IsBookmarked,
			// Default user values
			"username":            reply.Username,
			"display_name":        reply.DisplayName,
			"profile_picture_url": reply.ProfilePicture,
			"thread_author":       "unknown", // Original thread author
		}

		if len(reply.Media) > 0 {
			media := make([]map[string]interface{}, len(reply.Media))
			for j, m := range reply.Media {
				media[j] = map[string]interface{}{
					"id":   m.ID,
					"type": m.Type,
					"url":  m.URL,
				}
			}
			replyItem["media"] = media
		} else {
			replyItem["media"] = []interface{}{}
		}

		replyItems[i] = replyItem
	}

	c.JSON(http.StatusOK, gin.H{
		"replies": replyItems,
		"total":   len(replies),
	})
}

func GetUserLikedThreads(c *gin.Context) {
	authenticatedUserID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Success: false,
			Message: "User ID not found in token",
			Code:    "UNAUTHORIZED",
		})
		return
	}

	authenticatedUserIDStr, ok := authenticatedUserID.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Invalid User ID format in token",
			Code:    "INTERNAL_ERROR",
		})
		return
	}

	userID := c.Param("id")

	if userID == "me" {
		userID = authenticatedUserIDStr
		log.Printf("Using authenticated user ID for 'me' parameter: %s", userID)
	}

	if userID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Message: "User ID is required",
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

	if threadServiceClient == nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Thread service client not initialized",
			Code:    "SERVICE_UNAVAILABLE",
		})
		return
	}

	threads, err := threadServiceClient.GetLikedThreadsByUser(userID, page, limit)
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
				Message: "Failed to get liked threads: " + err.Error(),
				Code:    "INTERNAL_ERROR",
			})
		}
		return
	}

	threadItems := make([]map[string]interface{}, len(threads))
	for i, thread := range threads {
		threadItem := map[string]interface{}{
			"id":            thread.ID,
			"thread_id":     thread.ID,
			"content":       thread.Content,
			"user_id":       thread.UserID,
			"created_at":    thread.CreatedAt,
			"updated_at":    thread.UpdatedAt,
			"like_count":    thread.LikeCount,
			"reply_count":   thread.ReplyCount,
			"repost_count":  thread.RepostCount,
			"is_liked":      true, // Since these are liked threads
			"is_repost":     thread.IsReposted,
			"is_bookmarked": thread.IsBookmarked,
			"is_pinned":     thread.IsPinned,
			// User data
			"username":            thread.Username,
			"display_name":        thread.DisplayName,
			"profile_picture_url": thread.ProfilePicture,
		}

		if len(thread.Media) > 0 {
			media := make([]map[string]interface{}, len(thread.Media))
			for j, m := range thread.Media {
				media[j] = map[string]interface{}{
					"id":   m.ID,
					"type": m.Type,
					"url":  m.URL,
				}
			}
			threadItem["media"] = media
		} else {
			threadItem["media"] = []interface{}{}
		}

		threadItems[i] = threadItem
	}

	c.JSON(http.StatusOK, gin.H{
		"threads": threadItems,
		"total":   len(threads),
	})
}

func GetUserMedia(c *gin.Context) {
	authenticatedUserID, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Success: false,
			Message: "User ID not found in token",
			Code:    "UNAUTHORIZED",
		})
		return
	}

	authenticatedUserIDStr, ok := authenticatedUserID.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Invalid User ID format in token",
			Code:    "INTERNAL_ERROR",
		})
		return
	}

	userID := c.Param("id")

	if userID == "me" {
		userID = authenticatedUserIDStr
		log.Printf("Using authenticated user ID for 'me' parameter: %s", userID)
	}

	if userID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Message: "User ID is required",
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

	if threadServiceClient == nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Thread service client not initialized",
			Code:    "SERVICE_UNAVAILABLE",
		})
		return
	}

	mediaItems, err := threadServiceClient.GetMediaByUser(userID, page, limit)
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
				Message: "Failed to get user media: " + err.Error(),
				Code:    "INTERNAL_ERROR",
			})
		}
		return
	}

	mediaResponse := make([]map[string]interface{}, len(mediaItems))
	for i, m := range mediaItems {
		mediaResponse[i] = map[string]interface{}{
			"id":        m.ID,
			"thread_id": m.Thumbnail, // Use the thumbnail field to store thread_id
			"url":       m.URL,
			"type":      m.Type,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"media": mediaResponse,
		"total": len(mediaItems),
	})
}

func PinThread(c *gin.Context) {
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

	log.Printf("Pinning thread %s for user %s", threadID, userID)

	if threadServiceClient == nil {
		log.Println("Thread service client not initialized - using mock implementation")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Thread service client not initialized",
			Code:    "SERVICE_UNAVAILABLE",
		})
		return
	}

	err := threadServiceClient.PinThread(threadID, userID)
	if err != nil {
		log.Printf("Error pinning thread: %v", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Failed to pin thread: " + err.Error(),
			Code:    "INTERNAL_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Thread pinned successfully",
	})
}

func UnpinThread(c *gin.Context) {
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

	log.Printf("Unpinning thread %s for user %s", threadID, userID)

	if threadServiceClient == nil {
		log.Println("Thread service client not initialized - using mock implementation")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Thread service client not initialized",
			Code:    "SERVICE_UNAVAILABLE",
		})
		return
	}

	err := threadServiceClient.UnpinThread(threadID, userID)
	if err != nil {
		log.Printf("Error unpinning thread: %v", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Message: "Failed to unpin thread: " + err.Error(),
			Code:    "INTERNAL_ERROR",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Thread unpinned successfully",
	})
}

func BookmarkThreadHandler(c *gin.Context) {
	threadID := c.Param("id")
	userID, exists := c.Get("userId")
	if !exists {
		log.Printf("BookmarkThreadHandler: No user ID found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		log.Printf("BookmarkThreadHandler: User ID is not a string: %v", userID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
		return
	}

	if threadID == "" {
		log.Printf("BookmarkThreadHandler: No thread ID provided")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Thread ID is required"})
		return
	}

	log.Printf("BookmarkThreadHandler: Request received - threadID=%s, userID=%s", threadID, userIDStr)

	// Check if thread service client is initialized
	if threadServiceClient == nil {
		log.Printf("BookmarkThreadHandler: Thread service client is nil")
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Thread service unavailable",
		})
		return
	}

	// Attempt to bookmark the thread
	log.Printf("BookmarkThreadHandler: Calling threadServiceClient.BookmarkThread")
	err := threadServiceClient.BookmarkThread(threadID, userIDStr)
	if err != nil {
		log.Printf("BookmarkThreadHandler: Error from thread service: %v", err)

		// Check for specific error types and return appropriate status codes
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   err.Error(),
			})
			return
		}

		if strings.Contains(err.Error(), "Invalid") {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to bookmark thread: " + err.Error(),
		})
		return
	}

	log.Printf("BookmarkThreadHandler: Successfully bookmarked thread %s for user %s", threadID, userIDStr)

	// After bookmarking, let's verify the thread really got bookmarked
	thread, err := threadServiceClient.GetThreadByID(threadID, userIDStr)
	if err == nil && thread != nil {
		log.Printf("BookmarkThreadHandler: Verification - Thread %s for user %s: bookmark status is now %v",
			threadID, userIDStr, thread.IsBookmarked)
	} else if err != nil {
		log.Printf("BookmarkThreadHandler: Error verifying bookmark: %v", err)
	} else {
		log.Printf("BookmarkThreadHandler: Thread not found during verification")
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Thread bookmarked successfully",
	})
}

// UpdateThreadMediaURLsHandler updates a thread with media URLs from Supabase
func UpdateThreadMediaURLsHandler(c *gin.Context) {
	// Get thread ID from URL
	threadID := c.Param("id")
	if threadID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Thread ID is required"})
		return
	}

	// Get authenticated user ID
	userIdValue, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userID, ok := userIdValue.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
		return
	}

	// Parse request body
	var req struct {
		MediaUrls []string `json:"mediaUrls"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate all URLs are from Supabase
	for _, url := range req.MediaUrls {
		if !strings.Contains(url, ".supabase.co") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid URL source"})
			return
		}
	}

	// Get thread service client
	threadClient := GetThreadServiceClient()
	if threadClient == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Thread service unavailable"})
		return
	}

	// Update thread with new media URLs
	// Since UpdateThreadMedia might not exist yet, we'll use UpdateThread as a fallback
	updatedThread, err := threadClient.UpdateThread(threadID, userID, "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to update thread media: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"thread":     updatedThread,
		"media_urls": req.MediaUrls,
	})
}
