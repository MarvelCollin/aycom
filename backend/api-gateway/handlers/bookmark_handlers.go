package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetUserBookmarks handles fetching all bookmarks for the current user
// @Summary Get user bookmarks
// @Description Get all threads bookmarked by the current user
// @Tags Bookmarks
// @Accept json
// @Produce json
// @Param page query int false "Page number for pagination"
// @Param limit query int false "Number of results per page"
// @Success 200 {object} models.BookmarksResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/bookmarks [get]
func GetUserBookmarks(c *gin.Context) {
	// Get current user ID from JWT token
	userIDAny, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userID, ok := userIDAny.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
		return
	}

	// Parse pagination parameters
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

	// Get thread service client
	threadClient := GetThreadServiceClient()
	if threadClient == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Thread service unavailable"})
		return
	}

	// Get user bookmarks
	bookmarkedThreads, err := threadClient.GetUserBookmarks(userID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch bookmarks: " + err.Error(),
		})
		return
	}

	// Convert to response format
	bookmarks := make([]gin.H, len(bookmarkedThreads))
	for i, thread := range bookmarkedThreads {
		bookmarks[i] = gin.H{
			"id":              thread.ID,
			"thread_id":       thread.ID,
			"content":         thread.Content,
			"user_id":         thread.UserID,
			"username":        thread.Username,
			"display_name":    thread.DisplayName,
			"profile_picture": thread.ProfilePicture,
			"created_at":      thread.CreatedAt,
			"updated_at":      thread.UpdatedAt,
			"like_count":      thread.LikeCount,
			"reply_count":     thread.ReplyCount,
			"repost_count":    thread.RepostCount,
			"is_liked":        thread.IsLiked,
			"is_repost":       thread.IsReposted,
			"is_bookmarked":   true, // Since these are bookmarks
			"is_pinned":       thread.IsPinned,
		}

		// Add media if available
		if thread.Media != nil && len(thread.Media) > 0 {
			media := make([]map[string]interface{}, len(thread.Media))
			for j, m := range thread.Media {
				media[j] = map[string]interface{}{
					"id":   m.ID,
					"type": m.Type,
					"url":  m.URL,
				}
			}
			bookmarks[i]["media"] = media
		} else {
			bookmarks[i]["media"] = []interface{}{}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"bookmarks": bookmarks,
		"pagination": gin.H{
			"total":   len(bookmarkedThreads), // Ideally we'd get a total count from the service
			"page":    page,
			"limit":   limit,
			"hasMore": len(bookmarkedThreads) >= limit, // Simple estimation
		},
	})
}

// SearchBookmarks handles searching through a user's bookmarks
// @Summary Search user bookmarks
// @Description Search through the current user's bookmarked threads
// @Tags Bookmarks
// @Accept json
// @Produce json
// @Param q query string true "Search query"
// @Param page query int false "Page number for pagination"
// @Param limit query int false "Number of results per page"
// @Success 200 {object} models.BookmarksResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/bookmarks/search [get]
func SearchBookmarks(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Search query is required"})
		return
	}

	// Get current user ID from JWT token
	_, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// TODO: Search bookmarks from thread service
	// This is a placeholder implementation
	c.JSON(http.StatusOK, gin.H{
		"bookmarks": []gin.H{},
		"pagination": gin.H{
			"total":   0,
			"page":    1,
			"limit":   10,
			"hasMore": false,
		},
	})
}

// RemoveBookmark removes a bookmark
// @Summary Remove a bookmark
// @Description Remove a bookmark from the current user's list
// @Tags Bookmarks
// @Accept json
// @Produce json
// @Param id path string true "Bookmark ID"
// @Success 200 {object} models.SuccessResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/bookmarks/{id} [delete]
func RemoveBookmark(c *gin.Context) {
	// Get current user ID from JWT token
	_, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get bookmark ID from path parameter
	bookmarkID := c.Param("id")
	if bookmarkID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bookmark ID is required"})
		return
	}

	// TODO: Implement actual bookmark removal from service
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Bookmark removed successfully",
	})
}
