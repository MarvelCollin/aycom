package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// CheckUsernameAvailability checks if a username is available
// @Summary Check username availability
// @Description Checks if a username is available for registration
// @Tags Users
// @Produce json
// @Router /api/v1/users/check-username [get]
func CheckUsernameAvailability(c *gin.Context) {
	username := c.Query("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Username parameter is required",
		})
		return
	}

	log.Printf("Username availability check for: %s", username)

	// For testing, you might want to have some "taken" usernames
	isAvailable := true
	if username == "admin" || username == "system" || username == "root" {
		isAvailable = false
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"available": isAvailable,
	})
}

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

	// TODO: Implement actual follow logic with user service
	// For now, just return success for frontend testing
	log.Printf("User %s following user %s", currentUserID, targetUserID)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Successfully followed user",
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

	// TODO: Implement actual unfollow logic with user service
	// For now, just return success for frontend testing
	log.Printf("User %s unfollowing user %s", currentUserID, targetUserID)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Successfully unfollowed user",
	})
}

// GetFollowers returns the followers of a user
// @Summary Get followers
// @Description Gets a list of followers for a user
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

	// TODO: Implement actual get followers logic with user service
	// For now, just return dummy data for frontend testing
	log.Printf("Getting followers for user %s, page %d, limit %d", targetUserID, page, limit)

	// Create dummy followers
	followers := []gin.H{
		{
			"id":              "follower1",
			"name":            "Follower One",
			"username":        "follower1",
			"profile_picture": "https://via.placeholder.com/150",
			"verified":        true,
			"is_following":    false,
		},
		{
			"id":              "follower2",
			"name":            "Follower Two",
			"username":        "follower2",
			"profile_picture": "https://via.placeholder.com/150",
			"verified":        false,
			"is_following":    true,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"followers": followers,
		"page":      page,
		"limit":     limit,
		"total":     2,
	})
}

// GetFollowing returns the users a user is following
// @Summary Get following
// @Description Gets a list of users a user is following
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

	// TODO: Implement actual get following logic with user service
	// For now, just return dummy data for frontend testing
	log.Printf("Getting following for user %s, page %d, limit %d", targetUserID, page, limit)

	// Create dummy following list
	following := []gin.H{
		{
			"id":              "following1",
			"name":            "Following One",
			"username":        "following1",
			"profile_picture": "https://via.placeholder.com/150",
			"verified":        true,
			"is_following":    true,
		},
		{
			"id":              "following2",
			"name":            "Following Two",
			"username":        "following2",
			"profile_picture": "https://via.placeholder.com/150",
			"verified":        false,
			"is_following":    true,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"following": following,
		"page":      page,
		"limit":     limit,
		"total":     2,
	})
}
