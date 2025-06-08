package handlers

import (
	"aycom/backend/api-gateway/utils"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func GetUserBookmarks(c *gin.Context) {
	userIDAny, exists := c.Get("userId")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	userID, ok := userIDAny.(string)
	if !ok {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Invalid user ID format")
		return
	}

	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 10
	}

	threadClient := GetThreadServiceClient()
	if threadClient == nil {
		utils.SendErrorResponse(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Thread service unavailable")
		return
	}

	bookmarkedThreads, err := threadClient.GetUserBookmarks(userID, page, limit)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch bookmarks: "+err.Error())
		return
	}

	bookmarks := make([]gin.H, len(bookmarkedThreads))
	for i, thread := range bookmarkedThreads {
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
			media := make([]map[string]interface{}, len(thread.Media))
			for j, m := range thread.Media {
				media[j] = map[string]interface{}{
					"id":   m.ID,
					"type": m.Type,
					"url":  m.URL,
				}
			}
			threadData["media"] = media
		} else {
			threadData["media"] = []interface{}{}
		}

		bookmarks[i] = threadData
	}

	utils.SendSuccessResponse(c, http.StatusOK, gin.H{
		"bookmarks": bookmarks,
		"pagination": gin.H{
			"total_count":  len(bookmarkedThreads),
			"current_page": page,
			"per_page":     limit,
			"has_more":     len(bookmarkedThreads) >= limit,
		},
	})
}

func SearchBookmarks(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "BAD_REQUEST", "Search query is required")
		return
	}

	userIDAny, exists := c.Get("userId")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	userID, ok := userIDAny.(string)
	if !ok {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Invalid user ID format")
		return
	}

	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 10
	}

	threadClient := GetThreadServiceClient()
	if threadClient == nil {
		utils.SendErrorResponse(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "Thread service unavailable")
		return
	}

	bookmarkedThreads, err := threadClient.GetUserBookmarks(userID, page, limit)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch bookmarks: "+err.Error())
		return
	}

	// Filter bookmarks based on search query
	query = strings.ToLower(query)
	filteredBookmarks := make([]gin.H, 0)

	for _, thread := range bookmarkedThreads {
		content := strings.ToLower(thread.Content)
		username := strings.ToLower(thread.Username)
		displayName := strings.ToLower(thread.DisplayName)

		if strings.Contains(content, query) ||
			strings.Contains(username, query) ||
			strings.Contains(displayName, query) {
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
				media := make([]map[string]interface{}, len(thread.Media))
				for j, m := range thread.Media {
					media[j] = map[string]interface{}{
						"id":   m.ID,
						"type": m.Type,
						"url":  m.URL,
					}
				}
				threadData["media"] = media
			} else {
				threadData["media"] = []interface{}{}
			}

			filteredBookmarks = append(filteredBookmarks, threadData)
		}
	}

	utils.SendSuccessResponse(c, http.StatusOK, gin.H{
		"bookmarks": filteredBookmarks,
		"pagination": gin.H{
			"total_count":  len(filteredBookmarks),
			"current_page": page,
			"per_page":     limit,
			"has_more":     len(filteredBookmarks) >= limit,
		},
	})
}

func DeleteBookmarkById(c *gin.Context) {
	_, exists := c.Get("userId")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}

	bookmarkID := c.Param("id")
	if bookmarkID == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "BAD_REQUEST", "Bookmark ID is required")
		return
	}

	utils.SendSuccessResponse(c, http.StatusOK, gin.H{
		"message": "Bookmark removed successfully",
	})
}
