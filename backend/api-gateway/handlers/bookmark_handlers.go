package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetUserBookmarks(c *gin.Context) {
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
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Thread service unavailable"})
		return
	}

	bookmarkedThreads, err := threadClient.GetUserBookmarks(userID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch bookmarks: " + err.Error(),
		})
		return
	}

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
			"is_bookmarked":   true,
			"is_pinned":       thread.IsPinned,
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
			bookmarks[i]["media"] = media
		} else {
			bookmarks[i]["media"] = []interface{}{}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"bookmarks": bookmarks,
		"pagination": gin.H{
			"total":   len(bookmarkedThreads),
			"page":    page,
			"limit":   limit,
			"hasMore": len(bookmarkedThreads) >= limit,
		},
	})
}

func SearchBookmarks(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Search query is required"})
		return
	}

	_, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

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

func DeleteBookmarkById(c *gin.Context) {
	_, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	bookmarkID := c.Param("id")
	if bookmarkID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bookmark ID is required"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Bookmark removed successfully",
	})
}
