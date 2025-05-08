package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/status"
)

// SearchUsers handles searching for users
// @Summary Search for users
// @Description Search for users by name, username or other criteria
// @Tags Users
// @Accept json
// @Produce json
// @Param q query string true "Search query"
// @Param page query int false "Page number for pagination"
// @Param limit query int false "Number of results per page"
// @Param filter query string false "Filter results (all, following, verified)"
// @Param sort query string false "Sort results (relevance, newest, followers)"
// @Success 200 {object} models.UserSearchResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/users/search [get]
func SearchUsers(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		SendErrorResponse(c, http.StatusBadRequest, "BAD_REQUEST", "Search query is required")
		return
	}

	// Get pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	filter := c.DefaultQuery("filter", "all")

	// Check if service client is initialized
	if userServiceClient == nil {
		SendErrorResponse(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "User service client not initialized")
		return
	}

	// Call the service client method
	users, totalCount, err := userServiceClient.SearchUsers(query, filter, page, limit)
	if err != nil {
		// Handle errors
		st, ok := status.FromError(err)
		if ok {
			SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to search users: "+st.Message())
		} else {
			SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error while searching users")
		}
		log.Printf("Error searching users: %v", err)
		return
	}

	// Convert to response format
	var userResults []gin.H
	for _, user := range users {
		userResults = append(userResults, gin.H{
			"id":                  user.ID,
			"username":            user.Username,
			"name":                user.Name,
			"profile_picture_url": user.ProfilePictureURL,
			"bio":                 user.Bio,
			"is_verified":         user.IsVerified,
			"follower_count":      user.FollowerCount,
			"is_following":        user.IsFollowing,
		})
	}

	// Return successful response with users
	SendSuccessResponse(c, http.StatusOK, gin.H{
		"users": userResults,
		"pagination": gin.H{
			"total":   totalCount,
			"page":    page,
			"limit":   limit,
			"hasMore": len(users) == limit && (page*limit) < totalCount,
		},
	})
}

// SearchThreads handles searching for threads/posts
// @Summary Search for threads
// @Description Search for threads by content, hashtags or other criteria
// @Tags Threads
// @Accept json
// @Produce json
// @Param q query string true "Search query"
// @Param page query int false "Page number for pagination"
// @Param limit query int false "Number of results per page"
// @Param filter query string false "Filter results (all, media, links, etc)"
// @Param sort query string false "Sort results (relevance, newest, popular)"
// @Success 200 {object} models.ThreadSearchResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/threads/search [get]
func SearchThreads(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Search query is required"})
		return
	}

	// TODO: Implement actual search logic by calling thread service
	c.JSON(http.StatusOK, gin.H{
		"threads": []gin.H{},
		"pagination": gin.H{
			"total":   0,
			"page":    1,
			"limit":   10,
			"hasMore": false,
		},
	})
}

// SearchCommunities handles searching for communities
// @Summary Search for communities
// @Description Search for communities by name, description or other criteria
// @Tags Communities
// @Accept json
// @Produce json
// @Param q query string true "Search query"
// @Param page query int false "Page number for pagination"
// @Param limit query int false "Number of results per page"
// @Param filter query string false "Filter results (all, joined, public, etc)"
// @Param sort query string false "Sort results (relevance, newest, members)"
// @Success 200 {object} models.CommunitySearchResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/communities/search [get]
func SearchCommunities(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Search query is required"})
		return
	}

	// TODO: Implement actual search logic by calling community service
	c.JSON(http.StatusOK, gin.H{
		"communities": []gin.H{},
		"pagination": gin.H{
			"total":   0,
			"page":    1,
			"limit":   10,
			"hasMore": false,
		},
	})
}

// GetUserRecommendations returns recommended users for the current user
// @Summary Get user recommendations
// @Description Get a list of recommended users to follow
// @Tags Users
// @Accept json
// @Produce json
// @Success 200 {object} models.UserRecommendationsResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/v1/users/recommendations [get]
func GetUserRecommendations(c *gin.Context) {
	// Get current user ID from JWT token
	_, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// TODO: Implement actual recommendation logic by calling user service
	c.JSON(http.StatusOK, gin.H{
		"users": []gin.H{},
	})
}
