package handlers

import (
	threadProto "aycom/backend/proto/thread"
	"context"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"aycom/backend/api-gateway/utils"
)

// Helper function to safely extract bookmark count using reflection
func extractBookmarkCount(t *threadProto.ThreadResponse) int64 {
	if t == nil {
		return 0
	}

	// Try direct method call first if it exists
	tValue := reflect.ValueOf(t)
	getBookmarkCountMethod := tValue.MethodByName("GetBookmarkCount")
	if getBookmarkCountMethod.IsValid() {
		result := getBookmarkCountMethod.Call(nil)
		if len(result) > 0 {
			return result[0].Int()
		}
	}

	// Try to access field directly with reflection as fallback
	tElem := reflect.ValueOf(t).Elem()
	field := tElem.FieldByName("BookmarkCount")
	if field.IsValid() {
		return field.Int()
	}

	// Return default if nothing works
	return 0
}

func CreateThread(c *gin.Context) {

	origin := c.Request.Header.Get("Origin")
	if origin == "" {
		origin = "http://localhost:3000"
	}
	c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Admin-Request, X-Debug-Panel, Accept, Cache-Control, X-Requested-With, X-Api-Key, X-Auth-Token, Pragma, Expires, Connection, User-Agent, Host, Referer, Cookie, Set-Cookie, *")

	log.Printf("CreateThread: Origin: %s, Content-Type: %s, Auth header length: %d",
		c.GetHeader("Origin"),
		c.GetHeader("Content-Type"),
		len(c.GetHeader("Authorization")))

	if c.Request.Method == "OPTIONS" {
		c.AbortWithStatus(http.StatusNoContent)
		return
	}

	log.Printf("Processing CreateThread request from origin: %s", origin)

	userIDAny, exists := c.Get("userId")
	if !exists {

		userIDAny, exists = c.Get("userID")
		if !exists {
			log.Printf("CreateThread: No user ID found in context (tried both 'userId' and 'userID')")
			utils.SendErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User ID not found in token")
			return
		}
		log.Printf("CreateThread: Found user ID using 'userID' key")
	} else {
		log.Printf("CreateThread: Found user ID using 'userId' key")
	}

	userID, ok := userIDAny.(string)
	if !ok {
		log.Printf("CreateThread: User ID is not a string: %v (type %T)", userIDAny, userIDAny)
		utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Invalid User ID format in token")
		return
	}

	log.Printf("Creating thread for user ID: %s", userID)

	var request threadProto.CreateThreadRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		log.Printf("Error binding JSON for CreateThread: %v", err)
		utils.SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request: "+err.Error())
		return
	}

	request.UserId = userID

	conn, err := threadConnPool.Get()
	if err != nil {
		log.Printf("Error connecting to thread service: %v", err)
		utils.SendErrorResponse(c, http.StatusInternalServerError, "SERVICE_UNAVAILABLE", "Failed to connect to thread service: "+err.Error())
		return
	}
	defer threadConnPool.Put(conn)

	client := threadProto.NewThreadServiceClient(conn)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	md := metadata.New(map[string]string{
		"authorization": c.GetHeader("Authorization"),
	})
	ctx = metadata.NewOutgoingContext(ctx, md)

	log.Printf("CreateThread: Sending request to thread service. Content length: %d, Has media: %v",
		len(request.Content), len(request.Media) > 0)

	resp, err := client.CreateThread(ctx, &request)
	if err != nil {
		log.Printf("Error creating thread: %v", err)
		if st, ok := status.FromError(err); ok {
			httpStatus := http.StatusInternalServerError
			if st.Code() == codes.InvalidArgument {
				httpStatus = http.StatusBadRequest
			}
			utils.SendErrorResponse(c, httpStatus, st.Code().String(), st.Message())
		} else {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create thread: "+err.Error())
		}
		return
	}

	log.Printf("Successfully created thread ID: %s", resp.Thread.Id)

	c.Header("Content-Type", "application/json")

	c.JSON(http.StatusCreated, gin.H{
		"id":         resp.Thread.Id,
		"content":    resp.Thread.Content,
		"media":      resp.Thread.Media,
		"media_urls": resp.Thread.Media,
		"success":    true,
		"thread":     resp.Thread,
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
		"bookmark_count":      thread.BookmarkCount,
		"views_count":         0,
		"is_liked":            thread.IsLiked,
		"is_reposted":         thread.IsReposted,
		"is_bookmarked":       thread.IsBookmarked,
		"is_pinned":           thread.IsPinned,
		"is_verified":         false,
		"user_id":             thread.UserID,
		"username":            thread.Username,
		"name":                thread.DisplayName,
		"profile_picture_url": thread.ProfilePicture,
		"community_id":        nil,
		"community_name":      nil,
		"parent_id":           nil,
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

		log.Printf("No authenticated user for GetThreadsByUser, proceeding as guest")
	}

	authenticatedUserIDStr := ""
	if exists {
		var ok bool
		authenticatedUserIDStr, ok = authenticatedUserID.(string)
		if !ok {
			log.Printf("Invalid user ID format in token, proceeding as guest")
		} else {
			log.Printf("Authenticated user ID: %s", authenticatedUserIDStr)
		}
	}

	userID := c.Param("id")

	if userID == "me" {
		if authenticatedUserIDStr == "" {
			utils.SendErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required to view your own profile")
			return
		}
		userID = authenticatedUserIDStr
		log.Printf("Using authenticated user ID for 'me' parameter: %s", userID)
	}

	if userID == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "User ID is required")
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

	_, uuidErr := uuid.Parse(userID)
	if uuidErr != nil {
		log.Printf("UserID '%s' is not a valid UUID, attempting to resolve as username", userID)

		if userServiceClient == nil {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "SERVICE_UNAVAILABLE", "User service client not initialized")
			return
		}

		user, err := userServiceClient.GetUserByUsername(userID)
		if err != nil {
			utils.SendErrorResponse(c, http.StatusNotFound, "NOT_FOUND", fmt.Sprintf("User with username '%s' not found", userID))
			log.Printf("Failed to resolve username '%s' to UUID: %v", userID, err)
			return
		}

		userID = user.ID
		log.Printf("Resolved username '%s' to UUID '%s'", c.Param("id"), userID)
	}

	log.Printf("Connecting to thread service for user %s", userID)
	conn, err := threadConnPool.Get()
	if err != nil {
		log.Printf("Failed to connect to thread service: %v", err)
		utils.SendErrorResponse(c, http.StatusInternalServerError, "SERVICE_UNAVAILABLE", "Failed to connect to thread service: "+err.Error())
		return
	}
	defer threadConnPool.Put(conn)

	client := threadProto.NewThreadServiceClient(conn)
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	if authenticatedUserIDStr != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "user_id", authenticatedUserIDStr)
	}

	log.Printf("Requesting threads for user %s, page %d, limit %d", userID, page, limit)
	resp, err := client.GetThreadsByUser(ctx, &threadProto.GetThreadsByUserRequest{
		UserId: userID,
		Page:   int32(page),
		Limit:  int32(limit),
	})

	if err != nil {
		log.Printf("Error getting threads for user %s: %v", userID, err)
		if st, ok := status.FromError(err); ok {
			httpStatus := http.StatusInternalServerError
			if st.Code() == codes.NotFound {
				httpStatus = http.StatusNotFound
			}
			log.Printf("gRPC status error: code=%v, message=%s", st.Code(), st.Message())
			utils.SendErrorResponse(c, httpStatus, st.Code().String(), st.Message())
		} else {
			log.Printf("Non-status error: %v", err)
			utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get threads: "+err.Error())
		}
		return
	}

	log.Printf("Retrieved %d threads for user %s", len(resp.Threads), userID)
	threads := make([]map[string]interface{}, len(resp.Threads))
	for i, t := range resp.Threads {
		thread := map[string]interface{}{
			"id":            t.Thread.Id,
			"thread_id":     t.Thread.Id,
			"content":       t.Thread.Content,
			"user_id":       t.Thread.UserId,
			"likes_count":   t.LikesCount,
			"replies_count": t.RepliesCount,
			"reposts_count": t.RepostsCount,
			// Initialize with default value since the field is not directly accessible
			"bookmark_count": extractBookmarkCount(t),
			"views_count":    t.Thread.ViewCount,
			"is_liked":       t.LikedByUser,
			"is_reposted":    t.RepostedByUser,
			"is_bookmarked":  t.BookmarkedByUser,
			"is_pinned":      t.Thread.IsPinned != nil && *t.Thread.IsPinned,

			"username":            "anonymous",
			"name":                "User",
			"profile_picture_url": "",
		}

		if t.Thread.CreatedAt != nil {
			thread["created_at"] = t.Thread.CreatedAt.AsTime()
		} else {
			thread["created_at"] = time.Now()
		}

		if t.Thread.UpdatedAt != nil {
			thread["updated_at"] = t.Thread.UpdatedAt.AsTime()
		} else {
			thread["updated_at"] = time.Now()
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

	utils.SendSuccessResponse(c, http.StatusOK, gin.H{
		"threads": threads,
		"total":   resp.Total,
		"success": true,
	})
}

func UpdateThread(c *gin.Context) {
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

	log.Printf("User %s is updating a thread", userID)

	threadID := c.Param("id")
	if threadID == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Thread ID is required")
		return
	}

	var request threadProto.UpdateThreadRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request: "+err.Error())
		return
	}

	request.ThreadId = threadID

	conn, err := threadConnPool.Get()
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "SERVICE_UNAVAILABLE", "Failed to connect to thread service: "+err.Error())
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
			utils.SendErrorResponse(c, httpStatus, st.Code().String(), st.Message())
		} else {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update thread: "+err.Error())
		}
		return
	}

	c.JSON(http.StatusOK, resp)
}

func DeleteThread(c *gin.Context) {
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
			utils.SendErrorResponse(c, httpStatus, st.Code().String(), st.Message())
		} else {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete thread: "+err.Error())
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
		thread["bookmark_count"] = extractBookmarkCount(t)
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
	log.Printf("GetAllThreads endpoint called")

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

	var userID string
	if userIDAny, exists := c.Get("userId"); exists {
		if userIDStr, ok := userIDAny.(string); ok {
			userID = userIDStr
		}
	}

	if threadServiceClient == nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "SERVICE_UNAVAILABLE", "Thread service client not initialized")
		return
	}

	threads, err := threadServiceClient.GetAllThreads(userID, page, limit)
	if err != nil {
		log.Printf("Error getting threads from service: %v", err)
		utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch threads: "+err.Error())
		return
	}

	log.Printf("Retrieved %d threads from thread service", len(threads))

	threadList := make([]map[string]interface{}, len(threads))
	for i, thread := range threads {
		threadList[i] = map[string]interface{}{
			"id":                  thread.ID,
			"content":             thread.Content,
			"created_at":          thread.CreatedAt.Format(time.RFC3339),
			"updated_at":          thread.UpdatedAt.Format(time.RFC3339),
			"user_id":             thread.UserID,
			"username":            thread.Username,
			"name":                thread.DisplayName,
			"profile_picture_url": thread.ProfilePicture,
			"likes_count":         thread.LikeCount,
			"replies_count":       thread.ReplyCount,
			"reposts_count":       thread.RepostCount,
			"bookmark_count":      thread.BookmarkCount,
			"is_liked":            thread.IsLiked,
			"is_reposted":         thread.IsReposted,
			"is_bookmarked":       thread.IsBookmarked,
			"is_pinned":           thread.IsPinned,
		}

		if len(thread.Media) > 0 {
			mediaList := make([]map[string]interface{}, len(thread.Media))
			for j, m := range thread.Media {
				mediaList[j] = map[string]interface{}{
					"id":   m.ID,
					"url":  m.URL,
					"type": m.Type,
				}
			}
			threadList[i]["media"] = mediaList
		} else {
			threadList[i]["media"] = []interface{}{}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"threads":     threadList,
		"total_count": len(threads),
	})
}

func GetUserReplies(c *gin.Context) {
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

	userID := c.Param("id")

	if userID == "me" {
		userID = authenticatedUserIDStr
		log.Printf("Using authenticated user ID for 'me' parameter: %s", userID)
	}

	if userID == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "User ID is required")
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
		utils.SendErrorResponse(c, http.StatusInternalServerError, "SERVICE_UNAVAILABLE", "Thread service client not initialized")
		return
	}

	_, uuidErr := uuid.Parse(userID)
	if uuidErr != nil {

		log.Printf("UserID '%s' is not a valid UUID, attempting to resolve as username", userID)

		if userServiceClient == nil {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "SERVICE_UNAVAILABLE", "User service client not initialized")
			return
		}

		user, err := userServiceClient.GetUserByUsername(userID)
		if err != nil {

			utils.SendErrorResponse(c, http.StatusNotFound, "NOT_FOUND", fmt.Sprintf("User with username '%s' not found", userID))
			log.Printf("Failed to resolve username '%s' to UUID: %v", userID, err)
			return
		}

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
			utils.SendErrorResponse(c, httpStatus, st.Code().String(), st.Message())
		} else {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get user replies: "+err.Error())
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
		utils.SendErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User ID not found in token")
		return
	}

	authenticatedUserIDStr, ok := authenticatedUserID.(string)
	if !ok {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Invalid User ID format in token")
		return
	}

	userID := c.Param("id")

	if userID == "me" {
		userID = authenticatedUserIDStr
		log.Printf("Using authenticated user ID for 'me' parameter: %s", userID)
	}

	if userID == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "User ID is required")
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
		utils.SendErrorResponse(c, http.StatusInternalServerError, "SERVICE_UNAVAILABLE", "Thread service client not initialized")
		return
	}

	_, uuidErr := uuid.Parse(userID)
	if uuidErr != nil {

		log.Printf("UserID '%s' is not a valid UUID, attempting to resolve as username", userID)

		if userServiceClient == nil {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "SERVICE_UNAVAILABLE", "User service client not initialized")
			return
		}

		user, err := userServiceClient.GetUserByUsername(userID)
		if err != nil {

			utils.SendErrorResponse(c, http.StatusNotFound, "NOT_FOUND", fmt.Sprintf("User with username '%s' not found", userID))
			log.Printf("Failed to resolve username '%s' to UUID: %v", userID, err)
			return
		}

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
			utils.SendErrorResponse(c, httpStatus, st.Code().String(), st.Message())
		} else {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get user liked threads: "+err.Error())
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
		utils.SendErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User ID not found in token")
		return
	}

	authenticatedUserIDStr, ok := authenticatedUserID.(string)
	if !ok {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Invalid User ID format in token")
		return
	}

	userID := c.Param("id")

	if userID == "me" {
		userID = authenticatedUserIDStr
		log.Printf("Using authenticated user ID for 'me' parameter: %s", userID)
	}

	if userID == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "User ID is required")
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
		utils.SendErrorResponse(c, http.StatusInternalServerError, "SERVICE_UNAVAILABLE", "Thread service client not initialized")
		return
	}

	_, uuidErr := uuid.Parse(userID)
	if uuidErr != nil {

		log.Printf("UserID '%s' is not a valid UUID, attempting to resolve as username", userID)

		if userServiceClient == nil {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "SERVICE_UNAVAILABLE", "User service client not initialized")
			return
		}

		user, err := userServiceClient.GetUserByUsername(userID)
		if err != nil {

			utils.SendErrorResponse(c, http.StatusNotFound, "NOT_FOUND", fmt.Sprintf("User with username '%s' not found", userID))
			log.Printf("Failed to resolve username '%s' to UUID: %v", userID, err)
			return
		}

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
			utils.SendErrorResponse(c, httpStatus, st.Code().String(), st.Message())
		} else {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get user media: "+err.Error())
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

	log.Printf("Pinning thread %s for user %s", threadID, userID)

	if threadServiceClient == nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "SERVICE_UNAVAILABLE", "Thread service client not initialized")
		return
	}

	err := threadServiceClient.PinThread(threadID, userID)
	if err != nil {
		log.Printf("Error pinning thread: %v", err)
		utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to pin thread: "+err.Error())
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

	log.Printf("Unpinning thread %s for user %s", threadID, userID)

	if threadServiceClient == nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "SERVICE_UNAVAILABLE", "Thread service client not initialized")
		return
	}

	err := threadServiceClient.UnpinThread(threadID, userID)
	if err != nil {
		log.Printf("Error unpinning thread: %v", err)
		utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to unpin thread: "+err.Error())
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

		if thread.UserID != userIDStr {

			var username string
			var profilePic string
			var displayName string

			if userServiceClient != nil {
				userInfo, err := userServiceClient.GetUserById(userIDStr)
				if err == nil && userInfo != nil {
					username = userInfo.Username
					displayName = userInfo.DisplayName
					profilePic = userInfo.ProfilePictureURL
				}
			}

			notificationData := map[string]interface{}{
				"thread_id":      threadID,
				"user_id":        userIDStr,
				"username":       username,
				"display_name":   displayName,
				"avatar":         profilePic,
				"thread_content": thread.Content,
				"timestamp":      time.Now().Format(time.RFC3339),
				"is_read":        false,
			}

			content := "bookmarked your post"
			if displayName != "" {
				content = displayName + " " + content
			} else {
				content = "Someone " + content
			}

			notificationID, err := SendNotification(thread.UserID, NotificationTypeLike, content, notificationData)
			if err != nil {
				log.Printf("BookmarkThreadHandler: Error creating notification: %v", err)
			} else {
				log.Printf("BookmarkThreadHandler: Created notification %s for user %s", notificationID, thread.UserID)
			}
		} else {
			log.Printf("BookmarkThreadHandler: Skipping notification for self-bookmark")
		}
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

func LikeThreadHandler(c *gin.Context) {
	threadID := c.Param("id")
	userID, exists := c.Get("userId")
	if !exists {
		log.Printf("LikeThreadHandler: No user ID found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		log.Printf("LikeThreadHandler: User ID is not a string: %v", userID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
		return
	}

	if threadID == "" {
		log.Printf("LikeThreadHandler: No thread ID provided")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Thread ID is required"})
		return
	}

	log.Printf("LikeThreadHandler: Request received - threadID=%s, userID=%s", threadID, userIDStr)

	if threadServiceClient == nil {
		log.Printf("LikeThreadHandler: Thread service client is nil")
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Thread service unavailable",
		})
		return
	}

	log.Printf("LikeThreadHandler: Calling threadServiceClient.LikeThread")
	err := threadServiceClient.LikeThread(threadID, userIDStr)
	if err != nil {
		log.Printf("LikeThreadHandler: Error from thread service: %v", err)

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
			"error":   "Failed to like thread: " + err.Error(),
		})
		return
	}

	log.Printf("LikeThreadHandler: Successfully liked thread %s for user %s", threadID, userIDStr)

	thread, err := threadServiceClient.GetThreadByID(threadID, userIDStr)
	if err == nil && thread != nil {
		log.Printf("LikeThreadHandler: Verification - Thread %s for user %s: like status is now %v",
			threadID, userIDStr, thread.IsLiked)

		if thread.UserID != userIDStr {

			var username string
			var profilePic string
			var displayName string

			if userServiceClient != nil {
				userInfo, err := userServiceClient.GetUserById(userIDStr)
				if err == nil && userInfo != nil {
					username = userInfo.Username
					displayName = userInfo.DisplayName
					profilePic = userInfo.ProfilePictureURL
				}
			}

			notificationData := map[string]interface{}{
				"thread_id":      threadID,
				"user_id":        userIDStr,
				"username":       username,
				"display_name":   displayName,
				"avatar":         profilePic,
				"thread_content": thread.Content,
				"timestamp":      time.Now().Format(time.RFC3339),
				"is_read":        false,
			}

			content := "liked your post"
			if displayName != "" {
				content = displayName + " " + content
			} else {
				content = "Someone " + content
			}

			notificationID, err := SendNotification(thread.UserID, NotificationTypeLike, content, notificationData)
			if err != nil {
				log.Printf("LikeThreadHandler: Error creating notification: %v", err)
			} else {
				log.Printf("LikeThreadHandler: Created notification %s for user %s", notificationID, thread.UserID)
			}
		} else {
			log.Printf("LikeThreadHandler: Skipping notification for self-like")
		}
	} else if err != nil {
		log.Printf("LikeThreadHandler: Error verifying like: %v", err)
	} else {
		log.Printf("LikeThreadHandler: Thread not found during verification")
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Thread liked successfully",
	})
}

func UnlikeThreadHandler(c *gin.Context) {
	threadID := c.Param("id")
	userID, exists := c.Get("userId")
	if !exists {
		log.Printf("UnlikeThreadHandler: No user ID found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		log.Printf("UnlikeThreadHandler: User ID is not a string: %v", userID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
		return
	}

	if threadID == "" {
		log.Printf("UnlikeThreadHandler: No thread ID provided")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Thread ID is required"})
		return
	}

	log.Printf("UnlikeThreadHandler: Request received - threadID=%s, userID=%s", threadID, userIDStr)

	if threadServiceClient == nil {
		log.Printf("UnlikeThreadHandler: Thread service client is nil")
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Thread service unavailable",
		})
		return
	}

	log.Printf("UnlikeThreadHandler: Calling threadServiceClient.UnlikeThread")
	err := threadServiceClient.UnlikeThread(threadID, userIDStr)
	if err != nil {
		log.Printf("UnlikeThreadHandler: Error from thread service: %v", err)

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
			"error":   "Failed to unlike thread: " + err.Error(),
		})
		return
	}

	log.Printf("UnlikeThreadHandler: Successfully unliked thread %s for user %s", threadID, userIDStr)

	thread, err := threadServiceClient.GetThreadByID(threadID, userIDStr)
	if err == nil && thread != nil {
		log.Printf("UnlikeThreadHandler: Verification - Thread %s for user %s: like status is now %v",
			threadID, userIDStr, thread.IsLiked)
	} else if err != nil {
		log.Printf("UnlikeThreadHandler: Error verifying unlike: %v", err)
	} else {
		log.Printf("UnlikeThreadHandler: Thread not found during verification")
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Thread unliked successfully",
	})
}

func RemoveBookmarkHandler(c *gin.Context) {
	threadID := c.Param("id")
	userID, exists := c.Get("userId")
	if !exists {
		log.Printf("RemoveBookmarkHandler: No user ID found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		log.Printf("RemoveBookmarkHandler: User ID is not a string: %v", userID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
		return
	}

	if threadID == "" {
		log.Printf("RemoveBookmarkHandler: No thread ID provided")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Thread ID is required"})
		return
	}

	log.Printf("RemoveBookmarkHandler: Request received - threadID=%s, userID=%s", threadID, userIDStr)

	if threadServiceClient == nil {
		log.Printf("RemoveBookmarkHandler: Thread service client is nil")
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Thread service unavailable",
		})
		return
	}

	log.Printf("RemoveBookmarkHandler: Calling threadServiceClient.RemoveBookmark")
	err := threadServiceClient.RemoveBookmark(threadID, userIDStr)
	if err != nil {
		log.Printf("RemoveBookmarkHandler: Error from thread service: %v", err)

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
			"error":   "Failed to remove bookmark: " + err.Error(),
		})
		return
	}

	log.Printf("RemoveBookmarkHandler: Successfully removed bookmark for thread %s for user %s", threadID, userIDStr)

	thread, err := threadServiceClient.GetThreadByID(threadID, userIDStr)
	if err == nil && thread != nil {
		log.Printf("RemoveBookmarkHandler: Verification - Thread %s for user %s: bookmark status is now %v",
			threadID, userIDStr, thread.IsBookmarked)
	} else if err != nil {
		log.Printf("RemoveBookmarkHandler: Error verifying bookmark removal: %v", err)
	} else {
		log.Printf("RemoveBookmarkHandler: Thread not found during verification")
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Bookmark removed successfully",
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

	var request struct {
		MediaUrls []string `json:"media_urls"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(request.MediaUrls) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No media URLs provided"})
		return
	}

	for _, url := range request.MediaUrls {
		if !strings.Contains(url, ".supabase.co") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid URL source"})
			return
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

	mediaObjects := make([]*threadProto.Media, len(request.MediaUrls))
	for i, url := range request.MediaUrls {

		mediaType := "image"
		if strings.Contains(url, ".mp4") || strings.Contains(url, ".webm") || strings.Contains(url, ".mov") {
			mediaType = "video"
		} else if strings.Contains(url, ".gif") {
			mediaType = "gif"
		}

		mediaObjects[i] = &threadProto.Media{
			Id:   uuid.New().String(),
			Url:  url,
			Type: mediaType,
		}
	}

	resp, err := client.UpdateThread(ctx, &threadProto.UpdateThreadRequest{
		ThreadId: threadID,
		UserId:   userID,
		Media:    mediaObjects,
	})

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
			utils.SendErrorResponse(c, httpStatus, st.Code().String(), st.Message())
		} else {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update thread media: "+err.Error())
		}
		return
	}

	threadData := safeExtractThreadData(resp)

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"thread":     threadData,
		"media_urls": request.MediaUrls,
	})
}

// standardizeReplyResponse creates a standardized reply response object
func standardizeReplyResponse(reply *Thread) map[string]interface{} {
	replyData := map[string]interface{}{
		"id":                  reply.ID,
		"content":             reply.Content,
		"created_at":          reply.CreatedAt,
		"updated_at":          reply.UpdatedAt,
		"thread_id":           reply.ParentID, // Parent thread ID
		"parent_id":           nil,            // Default null for top-level replies
		"likes_count":         reply.LikeCount,
		"replies_count":       reply.ReplyCount,
		"reposts_count":       reply.RepostCount,
		"bookmark_count":      reply.BookmarkCount,
		"views_count":         0, // Default value if not available
		"is_liked":            reply.IsLiked,
		"is_reposted":         reply.IsReposted,
		"is_bookmarked":       reply.IsBookmarked,
		"is_pinned":           reply.IsPinned,
		"is_verified":         false, // Default value if not available
		"user_id":             reply.UserID,
		"username":            reply.Username,
		"name":                reply.DisplayName,
		"profile_picture_url": reply.ProfilePicture,
	}

	if len(reply.Media) > 0 {
		mediaList := make([]map[string]interface{}, len(reply.Media))
		for i, m := range reply.Media {
			mediaList[i] = map[string]interface{}{
				"id":   m.ID,
				"url":  m.URL,
				"type": m.Type,
			}
		}
		replyData["media"] = mediaList
	} else {
		replyData["media"] = []map[string]interface{}{}
	}

	return replyData
}
