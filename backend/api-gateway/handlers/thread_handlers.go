package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"aycom/backend/api-gateway/utils"
	threadProto "aycom/backend/proto/thread"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// @Summary Create thread
// @Description Creates a new thread
// @Tags Threads
// @Accept json
// @Produce json
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/threads [post]
func CreateThread(c *gin.Context) {
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

	// Parse request
	var request threadProto.CreateThreadRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Message: "Invalid request: " + err.Error(),
			Code:    "INVALID_REQUEST",
		})
		return
	}

	// Set user ID from token
	request.UserId = userID

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

// @Summary Get thread
// @Description Returns a thread by ID
// @Tags Threads
// @Produce json
// @Param id path string true "Thread ID"
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/threads/{id} [get]
func GetThread(c *gin.Context) {
	// Get thread ID from path parameter
	threadID := c.Param("id")
	if threadID == "" {
		SendErrorResponse(c, http.StatusBadRequest, "BAD_REQUEST", "Thread ID is required")
		return
	}

	// Get current user ID from JWT token (if available)
	userID, exists := c.Get("userID")
	var userIDStr string
	if exists {
		userIDStr = userID.(string)
	}

	// Check if the service client is initialized
	if threadServiceClient == nil {
		SendErrorResponse(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Thread service client not initialized")
		return
	}

	// Call the service client method
	thread, err := threadServiceClient.GetThreadByID(threadID, userIDStr)
	if err != nil {
		// Handle errors
		st, ok := status.FromError(err)
		if ok {
			// Map gRPC status code to HTTP status code
			switch st.Code() {
			case 5: // Not found
				SendErrorResponse(c, http.StatusNotFound, "NOT_FOUND", "Thread not found")
			default:
				SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve thread: "+st.Message())
			}
		} else {
			SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error while retrieving thread")
		}
		log.Printf("Error retrieving thread: %v", err)
		return
	}

	// Create response with user data clearly exposed
	threadData := map[string]interface{}{
		"id":            thread.ID,
		"thread_id":     thread.ID,
		"content":       thread.Content,
		"user_id":       thread.UserID,
		"created_at":    thread.CreatedAt,
		"updated_at":    thread.UpdatedAt,
		"like_count":    thread.LikeCount,
		"reply_count":   thread.ReplyCount,
		"repost_count":  thread.RepostCount,
		"is_liked":      thread.IsLiked,
		"is_repost":     thread.IsReposted,
		"is_bookmarked": thread.IsBookmarked,
		// Include user data directly
		"username":            thread.Username,
		"display_name":        thread.DisplayName,
		"profile_picture_url": thread.ProfilePicture,
	}

	// Add media if available
	if len(thread.Media) > 0 {
		media := make([]map[string]interface{}, len(thread.Media))
		for i, m := range thread.Media {
			media[i] = map[string]interface{}{
				"id":   m.ID,
				"type": m.Type,
				"url":  m.URL,
			}
		}
		threadData["media"] = media
	} else {
		threadData["media"] = []interface{}{}
	}

	// Return successful response with thread
	SendSuccessResponse(c, http.StatusOK, gin.H{
		"thread": threadData,
	})
}

// @Summary Get threads by user
// @Description Returns all threads for a user
// @Tags Threads
// @Produce json
// @Param id path string true "User ID"
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/threads/user/{id} [get]
func GetThreadsByUser(c *gin.Context) {
	// Get authenticated user ID from token
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

	// Get user ID from URL path parameter
	userID := c.Param("id")

	// Handle the "me" parameter to use the authenticated user's ID
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

	// Transform the response to include user data properly
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
			// Default user values
			"username":            "anonymous",
			"display_name":        "User",
			"profile_picture_url": "",
		}

		// Add user data if available
		if t.User != nil {
			thread["username"] = t.User.Username
			thread["display_name"] = t.User.Name
			thread["profile_picture_url"] = t.User.ProfilePictureUrl
			thread["is_verified"] = t.User.IsVerified
		}

		// Add media if available
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

// @Summary Update thread
// @Description Updates a thread by ID
// @Tags Threads
// @Accept json
// @Produce json
// @Param id path string true "Thread ID"
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/threads/{id} [put]
func UpdateThread(c *gin.Context) {
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

	// Use userID to log the action
	log.Printf("User %s is updating a thread", userID)

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
	var request threadProto.UpdateThreadRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Success: false,
			Message: "Invalid request: " + err.Error(),
			Code:    "INVALID_REQUEST",
		})
		return
	}

	// Set thread ID from URL
	request.ThreadId = threadID

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

// @Summary Delete thread
// @Description Deletes a thread by ID
// @Tags Threads
// @Produce json
// @Param id path string true "Thread ID"
// @Success 204 {object} nil
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/threads/{id} [delete]
func DeleteThread(c *gin.Context) {
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

// @Summary Upload thread media
// @Description Uploads media for a thread
// @Tags Threads
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Media file"
// @Failure 400 {object} ErrorResponse
// @Router /api/v1/threads/media [post]
func UploadThreadMedia(c *gin.Context) {
	// Extract user ID from context
	_, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "User not authenticated",
		})
		return
	}

	// Extract thread ID from form
	threadID := c.PostForm("thread_id")
	if threadID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Thread ID is required",
		})
		return
	}

	// Process a single file upload or multiple files upload
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid form data: " + err.Error(),
		})
		return
	}

	// Get all files from form
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

	// Process each file
	for _, file := range files {
		// Check file type
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

		// Determine media type
		mediaType := "image"
		if fileExt == ".gif" {
			mediaType = "gif"
		} else if fileExt == ".mp4" || fileExt == ".webm" || fileExt == ".mov" {
			mediaType = "video"
		}

		// Open the file
		fileContent, err := file.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to open file: " + err.Error(),
			})
			return
		}
		defer fileContent.Close()

		// Upload to Supabase
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

		// Add to results
		mediaUrls = append(mediaUrls, url)
		mediaTypes = append(mediaTypes, mediaType)
	}

	// Return success with all media URLs
	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"thread_id": threadID,
		"media":     mediaUrls,
		"types":     mediaTypes,
	})
}

// GetAllThreads returns all threads
// @Summary Get all threads
// @Description Returns all threads with pagination
// @Tags Threads
// @Produce json
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Results per page (default: 20)"
// @Success 200 {object} threadProto.ThreadsResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/threads [get]
func GetAllThreads(c *gin.Context) {
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

	// Optional authentication check
	// If the user is authenticated, we'll include their ID for personalized data
	// If not, we'll still show threads but without personalized data
	userID := ""
	if userIDVal, exists := c.Get("userId"); exists {
		if userIDStr, ok := userIDVal.(string); ok {
			userID = userIDStr
			log.Printf("Authenticated user %s is viewing threads", userID)
		}
	} else {
		log.Printf("Anonymous user is viewing threads")
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

	// Use the proper GetAllThreads method instead of the workaround
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

	// Transform the response to include user data properly
	// This ensures the frontend receives the user information directly
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
			// Default user values
			"username":            "anonymous",
			"display_name":        "User",
			"profile_picture_url": "",
		}

		// Add user data if available
		if t.User != nil {
			thread["username"] = t.User.Username
			thread["display_name"] = t.User.Name
			thread["profile_picture_url"] = t.User.ProfilePictureUrl
			thread["is_verified"] = t.User.IsVerified
		}

		// Add media if available
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

// @Summary Get user replies
// @Description Returns all replies made by a user
// @Tags Threads
// @Produce json
// @Param id path string true "User ID"
// @Param page query int false "Page number"
// @Param limit query int false "Limit per page"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/threads/user/{id}/replies [get]
func GetUserReplies(c *gin.Context) {
	// Get authenticated user ID from token
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

	// Get user ID from URL path parameter
	userID := c.Param("id")

	// Handle the "me" parameter to use the authenticated user's ID
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

	// Call thread service - since there's no direct endpoint for user replies,
	// we'll get them through a custom RPC call
	resp, err := client.GetRepliesByUser(ctx, &threadProto.GetRepliesByUserRequest{
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
				Message: "Failed to get user replies: " + err.Error(),
				Code:    "INTERNAL_ERROR",
			})
		}
		return
	}

	// Transform the response to include user data properly
	replies := make([]map[string]interface{}, len(resp.Replies))
	for i, r := range resp.Replies {
		reply := map[string]interface{}{
			"id":             r.Reply.Id,
			"reply_id":       r.Reply.Id,
			"thread_id":      r.Reply.ThreadId,
			"content":        r.Reply.Content,
			"user_id":        r.Reply.UserId,
			"created_at":     r.Reply.CreatedAt.AsTime(),
			"updated_at":     r.Reply.UpdatedAt.AsTime(),
			"like_count":     r.LikesCount,
			"is_liked":       r.LikedByUser,
			"is_bookmarked":  r.BookmarkedByUser,
			"is_pinned":      r.Reply.IsPinned,
			// Default user values
			"username":            "anonymous",
			"display_name":        "User",
			"profile_picture_url": "",
			"thread_author":       "unknown" // Original thread author
		}

		// Add user data if available
		if r.User != nil {
			reply["username"] = r.User.Username
			reply["display_name"] = r.User.Name
			reply["profile_picture_url"] = r.User.ProfilePictureUrl
			reply["is_verified"] = r.User.IsVerified
		}

		// Add thread author data if available
		if r.ThreadAuthor != nil {
			reply["thread_author"] = r.ThreadAuthor.Username
		}

		// Add media if available
		if len(r.Reply.Media) > 0 {
			media := make([]map[string]interface{}, len(r.Reply.Media))
			for j, m := range r.Reply.Media {
				media[j] = map[string]interface{}{
					"id":   m.Id,
					"type": m.Type,
					"url":  m.Url,
				}
			}
			reply["media"] = media
		} else {
			reply["media"] = []interface{}{}
		}

		replies[i] = reply
	}

	c.JSON(http.StatusOK, gin.H{
		"replies": replies,
		"total":   resp.Total,
	})
}

// @Summary Get user liked threads
// @Description Returns all threads liked by a user
// @Tags Threads
// @Produce json
// @Param id path string true "User ID"
// @Param page query int false "Page number"
// @Param limit query int false "Limit per page"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/threads/user/{id}/likes [get]
func GetUserLikedThreads(c *gin.Context) {
	// Get authenticated user ID from token
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

	// Get user ID from URL path parameter
	userID := c.Param("id")

	// Handle the "me" parameter to use the authenticated user's ID
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
	resp, err := client.GetLikedThreadsByUser(ctx, &threadProto.GetLikedThreadsByUserRequest{
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
				Message: "Failed to get liked threads: " + err.Error(),
				Code:    "INTERNAL_ERROR",
			})
		}
		return
	}

	// Transform the response to include user data properly
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
			"view_count":     t.Thread.ViewCount,
			"is_liked":       true, // Since these are liked threads
			"is_repost":      t.RepostedByUser,
			"is_bookmarked":  t.BookmarkedByUser,
			// Default user values
			"username":            "anonymous",
			"display_name":        "User",
			"profile_picture_url": "",
		}

		// Add user data if available
		if t.User != nil {
			thread["username"] = t.User.Username
			thread["display_name"] = t.User.Name
			thread["profile_picture_url"] = t.User.ProfilePictureUrl
			thread["is_verified"] = t.User.IsVerified
		}

		// Add media if available
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

// @Summary Get user media
// @Description Returns media from threads posted by a user
// @Tags Threads
// @Produce json
// @Param id path string true "User ID"
// @Param page query int false "Page number"
// @Param limit query int false "Limit per page"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/threads/user/{id}/media [get]
func GetUserMedia(c *gin.Context) {
	// Get authenticated user ID from token
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

	// Get user ID from URL path parameter
	userID := c.Param("id")

	// Handle the "me" parameter to use the authenticated user's ID
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
	resp, err := client.GetMediaByUser(ctx, &threadProto.GetMediaByUserRequest{
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
				Message: "Failed to get user media: " + err.Error(),
				Code:    "INTERNAL_ERROR",
			})
		}
		return
	}

	// Transform the response to include media data properly
	mediaItems := make([]map[string]interface{}, len(resp.Media))
	for i, m := range resp.Media {
		mediaItem := map[string]interface{}{
			"id":        m.Id,
			"thread_id": m.ThreadId,
			"url":       m.Url,
			"type":      m.Type,
			"created_at": m.CreatedAt.AsTime(),
		}

		mediaItems[i] = mediaItem
	}

	c.JSON(http.StatusOK, gin.H{
		"media": mediaItems,
		"total": resp.Total,
	})
}

// @Summary Pin thread to profile
// @Description Pins a thread to the user's profile
// @Tags Threads
// @Produce json
// @Param id path string true "Thread ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/threads/{id}/pin [post]
func PinThread(c *gin.Context) {
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
	_, err = client.PinThread(ctx, &threadProto.PinThreadRequest{
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
				Message: "Failed to pin thread: " + err.Error(),
				Code:    "INTERNAL_ERROR",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Thread pinned successfully",
	})
}

// @Summary Unpin thread from profile
// @Description Unpins a thread from the user's profile
// @Tags Threads
// @Produce json
// @Param id path string true "Thread ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /api/v1/threads/{id}/pin [delete]
func UnpinThread(c *gin.Context) {
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
	_, err = client.UnpinThread(ctx, &threadProto.UnpinThreadRequest{
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
				Message: "Failed to unpin thread: " + err.Error(),
				Code:    "INTERNAL_ERROR",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Thread unpinned successfully",
	})
}
