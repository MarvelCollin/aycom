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
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
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

	c.JSON(http.StatusCreated, gin.H{
		"id":         resp.Thread.Id,
		"content":    resp.Thread.Content,
		"media":      resp.Thread.Media,
		"media_urls": resp.Thread.Media,
		"success":    true,
	})
}

func GetThread(c *gin.Context) {
	threadID := c.Param("id")
	if threadID == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Thread ID is required")
		return
	}

	userIDAny, exists := c.Get("userID")
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
		utils.SendErrorResponse(c, http.StatusInternalServerError, "SERVICE_UNAVAILABLE", "Thread service client not initialized")
		return
	}

	thread, err := threadServiceClient.GetThreadByID(threadID, userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			utils.SendErrorResponse(c, http.StatusNotFound, "THREAD_NOT_FOUND", "Thread not found")
			return
		}

		utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Error retrieving thread: "+err.Error())
		return
	}

	threadData := gin.H{
		"id":                  thread.ID,
		"content":             thread.Content,
		"created_at":          thread.CreatedAt,
		"updated_at":          thread.UpdatedAt,
		"likes_count":         thread.LikeCount,
		"replies_count":       thread.ReplyCount,
		"reposts_count":       thread.RepostCount,
		"is_liked":            thread.IsLiked,
		"is_reposted":         thread.IsReposted,
		"is_bookmarked":       thread.IsBookmarked,
		"is_pinned":           thread.IsPinned,
		"user_id":             thread.UserID,
		"username":            thread.Username,
		"name":                thread.DisplayName,
		"profile_picture_url": thread.ProfilePicture,
	}

	if len(thread.Media) > 0 {
		mediaList := make([]map[string]interface{}, len(thread.Media))
		for i, m := range thread.Media {
			mediaList[i] = map[string]interface{}{
				"id":   m.ID,
				"url":  m.URL,
				"type": m.Type,
			}
		}
		threadData["media"] = mediaList
	}

	log.Printf("Thread %s bookmark status for user %s: %v", threadID, userID, thread.IsBookmarked)

	utils.SendSuccessResponse(c, http.StatusOK, threadData)
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

	// Check if userID is a valid UUID
	_, uuidErr := uuid.Parse(userID)
	if uuidErr != nil {
		// Not a UUID, try to resolve as username
		log.Printf("UserID '%s' is not a valid UUID, attempting to resolve as username", userID)

		if userServiceClient == nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Success: false,
				Message: "User service client not initialized",
				Code:    "SERVICE_UNAVAILABLE",
			})
			return
		}

		user, err := userServiceClient.GetUserByUsername(userID)
		if err != nil {
			// Failed to find user by username
			c.JSON(http.StatusNotFound, ErrorResponse{
				Success: false,
				Message: fmt.Sprintf("User with username '%s' not found", userID),
				Code:    "NOT_FOUND",
			})
			log.Printf("Failed to resolve username '%s' to UUID: %v", userID, err)
			return
		}

		// Use the resolved UUID
		userID = user.ID
		log.Printf("Resolved username '%s' to UUID '%s'", c.Param("id"), userID)
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
			"likes_count":    t.LikesCount,
			"replies_count":  t.RepliesCount,
			"reposts_count":  t.RepostsCount,
			"bookmark_count": t.BookmarkCount,
			"views_count":    t.Thread.ViewCount,
			"is_liked":       t.LikedByUser,
			"is_reposted":    t.RepostedByUser,
			"is_bookmarked":  t.BookmarkedByUser,
			"is_pinned":      t.Thread.IsPinned != nil && *t.Thread.IsPinned,

			"username":            "anonymous",
			"name":                "User",
			"profile_picture_url": "",
		}

		if t.User != nil {
			thread["username"] = t.User.Username
			thread["name"] = t.User.Name
			thread["profile_picture_url"] = t.User.ProfilePictureUrl
			thread["is_verified"] = t.User.IsVerified
		}

		if len(t.Thread.Media) > 0 {
			media := make([]map[string]interface{}, len(t.Thread.Media))
			for j, m := range t.Thread.Media {
				if m != nil {
					media[j] = map[string]interface{}{
						"id":   m.Id,
						"type": m.Type,
						"url":  m.Url,
					}
				} else {
					media[j] = map[string]interface{}{
						"id":   "",
						"type": "",
						"url":  "",
					}
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

func safeExtractThreadData(t *threadProto.ThreadResponse) map[string]interface{} {
	thread := map[string]interface{}{
		"id":             "",
		"thread_id":      "",
		"content":        "",
		"user_id":        "",
		"created_at":     time.Now(),
		"updated_at":     time.Now(),
		"likes_count":    0,
		"replies_count":  0,
		"reposts_count":  0,
		"bookmark_count": 0,
		"views_count":    0,
		"is_liked":       false,
		"is_reposted":    false,
		"is_bookmarked":  false,
		"is_pinned":      false,

		"username":            "anonymous",
		"name":                "User",
		"profile_picture_url": "",
	}

	if t != nil {
		if t.Thread != nil {
			thread["id"] = t.Thread.Id
			thread["thread_id"] = t.Thread.Id
			thread["content"] = t.Thread.Content
			thread["user_id"] = t.Thread.UserId

			if t.Thread.CreatedAt != nil {
				thread["created_at"] = t.Thread.CreatedAt.AsTime()
			}

			if t.Thread.UpdatedAt != nil {
				thread["updated_at"] = t.Thread.UpdatedAt.AsTime()
			}

			thread["views_count"] = t.Thread.ViewCount

			if t.Thread.IsPinned != nil {
				thread["is_pinned"] = *t.Thread.IsPinned
			}

			if len(t.Thread.Media) > 0 {
				media := make([]map[string]interface{}, len(t.Thread.Media))
				for j, m := range t.Thread.Media {
					if m != nil {
						media[j] = map[string]interface{}{
							"id":   m.Id,
							"type": m.Type,
							"url":  m.Url,
						}
					} else {
						media[j] = map[string]interface{}{
							"id":   "",
							"type": "",
							"url":  "",
						}
					}
				}
				thread["media"] = media
			} else {
				thread["media"] = []interface{}{}
			}
		}

		thread["likes_count"] = t.LikesCount
		thread["replies_count"] = t.RepliesCount
		thread["reposts_count"] = t.RepostsCount
		thread["bookmark_count"] = t.BookmarkCount
		thread["is_liked"] = t.LikedByUser
		thread["is_reposted"] = t.RepostedByUser
		thread["is_bookmarked"] = t.BookmarkedByUser

		if t.User != nil {
			thread["username"] = t.User.Username
			thread["name"] = t.User.Name
			thread["profile_picture_url"] = t.User.ProfilePictureUrl
			thread["is_verified"] = t.User.IsVerified
		}
	}

	return thread
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

	log.Printf("GetAllThreads - page: %d, limit: %d, userID: %s", page, limit, userID)

	if threadServiceClient != nil {

		log.Printf("Attempting to get threads using threadServiceClient")
		threads, err := threadServiceClient.GetAllThreads(userID, page, limit)
		if err == nil {

			log.Printf("Successfully retrieved %d threads using threadServiceClient", len(threads))
			c.JSON(http.StatusOK, gin.H{
				"threads": threads,
				"total":   len(threads),
			})
			return
		}

		log.Printf("Error using threadServiceClient.GetAllThreads: %v", err)
	}

	conn, err := threadConnPool.Get()
	if err != nil {
		log.Printf("Failed to get thread service connection: %v", err)
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

	if userID != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "user_id", userID)
	}

	log.Printf("Calling thread service GetAllThreads RPC")
	resp, err := client.GetAllThreads(ctx, &threadProto.GetAllThreadsRequest{
		Page:  int32(page),
		Limit: int32(limit),
	})

	if err != nil {
		log.Printf("Error from GetAllThreads RPC: %v", err)
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

	log.Printf("Successfully received %d threads from thread service", len(resp.Threads))

	threads := make([]map[string]interface{}, len(resp.Threads))
	for i, t := range resp.Threads {

		log.Printf("Thread %d type: %T", i, t)

		threads[i] = safeExtractThreadData(t)
	}

	log.Printf("Returning %d threads to client", len(threads))
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

	// Check if userID is a valid UUID
	_, uuidErr := uuid.Parse(userID)
	if uuidErr != nil {
		// Not a UUID, try to resolve as username
		log.Printf("UserID '%s' is not a valid UUID, attempting to resolve as username", userID)

		if userServiceClient == nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Success: false,
				Message: "User service client not initialized",
				Code:    "SERVICE_UNAVAILABLE",
			})
			return
		}

		user, err := userServiceClient.GetUserByUsername(userID)
		if err != nil {
			// Failed to find user by username
			c.JSON(http.StatusNotFound, ErrorResponse{
				Success: false,
				Message: fmt.Sprintf("User with username '%s' not found", userID),
				Code:    "NOT_FOUND",
			})
			log.Printf("Failed to resolve username '%s' to UUID: %v", userID, err)
			return
		}

		// Use the resolved UUID
		userID = user.ID
		log.Printf("Resolved username '%s' to UUID '%s'", c.Param("id"), userID)
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
			"id":                  reply.ID,
			"reply_id":            reply.ID,
			"thread_id":           reply.ParentID,
			"content":             reply.Content,
			"user_id":             reply.UserID,
			"created_at":          reply.CreatedAt,
			"updated_at":          reply.UpdatedAt,
			"likes_count":         reply.LikeCount,
			"is_liked":            reply.IsLiked,
			"is_bookmarked":       reply.IsBookmarked,
			"username":            reply.Username,
			"name":                reply.DisplayName,
			"profile_picture_url": reply.ProfilePicture,
			"thread_author":       "unknown",
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

	// Check if userID is a valid UUID
	_, uuidErr := uuid.Parse(userID)
	if uuidErr != nil {
		// Not a UUID, try to resolve as username
		log.Printf("UserID '%s' is not a valid UUID, attempting to resolve as username", userID)

		if userServiceClient == nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Success: false,
				Message: "User service client not initialized",
				Code:    "SERVICE_UNAVAILABLE",
			})
			return
		}

		user, err := userServiceClient.GetUserByUsername(userID)
		if err != nil {
			// Failed to find user by username
			c.JSON(http.StatusNotFound, ErrorResponse{
				Success: false,
				Message: fmt.Sprintf("User with username '%s' not found", userID),
				Code:    "NOT_FOUND",
			})
			log.Printf("Failed to resolve username '%s' to UUID: %v", userID, err)
			return
		}

		// Use the resolved UUID
		userID = user.ID
		log.Printf("Resolved username '%s' to UUID '%s'", c.Param("id"), userID)
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
				Message: "Failed to get user liked threads: " + err.Error(),
				Code:    "INTERNAL_ERROR",
			})
		}
		return
	}

	threadItems := make([]map[string]interface{}, len(threads))
	for i, thread := range threads {
		threadItems[i] = map[string]interface{}{
			"id":                  thread.ID,
			"content":             thread.Content,
			"user_id":             thread.UserID,
			"thread_id":           thread.ID,
			"created_at":          thread.CreatedAt,
			"updated_at":          thread.UpdatedAt,
			"likes_count":         thread.LikeCount,
			"replies_count":       thread.ReplyCount,
			"reposts_count":       thread.RepostCount,
			"is_liked":            thread.IsLiked,
			"is_reposted":         thread.IsReposted,
			"is_bookmarked":       thread.IsBookmarked,
			"username":            thread.Username,
			"name":                thread.DisplayName,
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
			threadItems[i]["media"] = media
		} else {
			threadItems[i]["media"] = []interface{}{}
		}
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

	// Check if userID is a valid UUID
	_, uuidErr := uuid.Parse(userID)
	if uuidErr != nil {
		// Not a UUID, try to resolve as username
		log.Printf("UserID '%s' is not a valid UUID, attempting to resolve as username", userID)

		if userServiceClient == nil {
			c.JSON(http.StatusInternalServerError, ErrorResponse{
				Success: false,
				Message: "User service client not initialized",
				Code:    "SERVICE_UNAVAILABLE",
			})
			return
		}

		user, err := userServiceClient.GetUserByUsername(userID)
		if err != nil {
			// Failed to find user by username
			c.JSON(http.StatusNotFound, ErrorResponse{
				Success: false,
				Message: fmt.Sprintf("User with username '%s' not found", userID),
				Code:    "NOT_FOUND",
			})
			log.Printf("Failed to resolve username '%s' to UUID: %v", userID, err)
			return
		}

		// Use the resolved UUID
		userID = user.ID
		log.Printf("Resolved username '%s' to UUID '%s'", c.Param("id"), userID)
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
			"thread_id": m.Thumbnail,
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

	if threadServiceClient == nil {
		log.Printf("BookmarkThreadHandler: Thread service client is nil")
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Thread service unavailable",
		})
		return
	}

	log.Printf("BookmarkThreadHandler: Calling threadServiceClient.BookmarkThread")
	err := threadServiceClient.BookmarkThread(threadID, userIDStr)
	if err != nil {
		log.Printf("BookmarkThreadHandler: Error from thread service: %v", err)

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

func UpdateThreadMediaURLsHandler(c *gin.Context) {
	threadID := c.Param("id")
	if threadID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Thread ID is required"})
		return
	}

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

	// Completely rewritten struct with clean JSON tag
	var request struct {
		MediaUrls []string `json:"media_urls"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for _, url := range request.MediaUrls {
		if !strings.Contains(url, ".supabase.co") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid URL source"})
			return
		}
	}

	threadClient := GetThreadServiceClient()
	if threadClient == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Thread service unavailable"})
		return
	}

	updatedThread, err := threadClient.UpdateThread(threadID, userID, "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to update thread media: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"thread":     updatedThread,
		"media_urls": request.MediaUrls,
	})
}
