package handlers

import (
	"aycom/backend/api-gateway/utils"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/status"
)

func SearchUsers(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "BAD_REQUEST", "Search query is required")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	filter := c.DefaultQuery("filter", "all")

	if userServiceClient == nil {
		utils.SendErrorResponse(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "User service client not initialized")
		return
	}

	users, totalCount, err := userServiceClient.SearchUsers(query, filter, page, limit)
	if err != nil {

		st, ok := status.FromError(err)
		if ok {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to search users: "+st.Message())
		} else {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error while searching users")
		}
		log.Printf("Error searching users: %v", err)
		return
	}

	var userResults []gin.H
	for _, user := range users {
		userResults = append(userResults, gin.H{
			"id":                  user.ID,
			"username":            user.Username,
			"name":                user.Name,
			"profile_picture_url": user.ProfilePictureURL,
			"bio":                 user.Bio,
			"is_verified":         user.IsVerified,
			"is_admin":            user.IsAdmin,
			"follower_count":      user.FollowerCount,
			"is_following":        user.IsFollowing,
		})
	}

	utils.SendSuccessResponse(c, http.StatusOK, gin.H{
		"users": userResults,
		"pagination": gin.H{
			"total_count":  totalCount,
			"current_page": page,
			"per_page":     limit,
			"has_more":     len(users) == limit && (page*limit) < totalCount,
		},
	})
}

func SearchThreads(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "BAD_REQUEST", "Search query is required")
		return
	}

	utils.SendSuccessResponse(c, http.StatusOK, gin.H{
		"threads": []gin.H{},
		"pagination": gin.H{
			"total_count":  0,
			"current_page": 1,
			"per_page":     10,
			"has_more":     false,
		},
	})
}

func SearchCommunities(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "BAD_REQUEST", "Search query is required")
		return
	}

	utils.SendSuccessResponse(c, http.StatusOK, gin.H{
		"communities": []gin.H{},
		"pagination": gin.H{
			"total_count":  0,
			"current_page": 1,
			"per_page":     10,
			"has_more":     false,
		},
	})
}

func GetUserRecommendations(c *gin.Context) {

	userID, exists := c.Get("userID")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}
	userIDStr := userID.(string)

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "3"))

	if userServiceClient == nil {
		utils.SendErrorResponse(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "User service client not initialized")
		return
	}

	users, err := userServiceClient.GetUserRecommendations(userIDStr, limit)
	if err != nil {

		st, ok := status.FromError(err)
		if ok {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get user recommendations: "+st.Message())
		} else {
			utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error while getting user recommendations")
		}
		log.Printf("Error getting user recommendations: %v", err)
		return
	}

	var userResults []gin.H
	for _, user := range users {
		userResults = append(userResults, gin.H{
			"id":                  user.ID,
			"username":            user.Username,
			"name":                user.Name,
			"profile_picture_url": user.ProfilePictureURL,
			"bio":                 user.Bio,
			"is_verified":         user.IsVerified,
			"is_admin":            user.IsAdmin,
			"follower_count":      user.FollowerCount,
		})
	}

	utils.SendSuccessResponse(c, http.StatusOK, gin.H{
		"users": userResults,
	})
}
