package handlers

import (
	"aycom/backend/api-gateway/utils"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// BlockUser handles blocking a user
func BlockUser(c *gin.Context) {
	// Get target user ID from route parameter
	targetUserID := c.Param("userId")
	if targetUserID == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "User ID parameter is required")
		return
	}

	// Get authenticated user ID from context
	currentUserID, exists := c.Get("userID")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}
	blockerID := currentUserID.(string)

	// Don't allow users to block themselves
	if blockerID == targetUserID {
		utils.SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Users cannot block themselves")
		return
	}

	// Block the user
	err := userServiceClient.BlockUser(blockerID, targetUserID)
	if err != nil {
		log.Printf("Error blocking user %s: %v", targetUserID, err)
		utils.SendErrorResponse(c, http.StatusInternalServerError, "SERVER_ERROR", "Failed to block user")
		return
	}

	utils.SendSuccessResponse(c, http.StatusOK, gin.H{
		"message": "User blocked successfully",
	})
}

// UnblockUser handles unblocking a user
func UnblockUser(c *gin.Context) {
	// Get target user ID from route parameter
	targetUserID := c.Param("userId")
	if targetUserID == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "User ID parameter is required")
		return
	}

	// Get authenticated user ID from context
	currentUserID, exists := c.Get("userID")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}
	blockerID := currentUserID.(string)

	// Unblock the user
	err := userServiceClient.UnblockUser(blockerID, targetUserID)
	if err != nil {
		log.Printf("Error unblocking user %s: %v", targetUserID, err)
		utils.SendErrorResponse(c, http.StatusInternalServerError, "SERVER_ERROR", "Failed to unblock user")
		return
	}

	utils.SendSuccessResponse(c, http.StatusOK, gin.H{
		"message": "User unblocked successfully",
	})
}

// GetBlockedUsers returns a list of users blocked by the current user
func GetBlockedUsers(c *gin.Context) {
	// Get authenticated user ID from context
	currentUserID, exists := c.Get("userID")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}
	userID := currentUserID.(string)

	// Parse pagination parameters
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

	// Get blocked users
	blockedUsers, err := userServiceClient.GetBlockedUsers(userID, page, limit)
	if err != nil {
		log.Printf("Error getting blocked users for %s: %v", userID, err)
		utils.SendErrorResponse(c, http.StatusInternalServerError, "SERVER_ERROR", "Failed to retrieve blocked users")
		return
	}

	// Format response
	formattedUsers := make([]gin.H, 0, len(blockedUsers))
	for _, user := range blockedUsers {
		formattedUsers = append(formattedUsers, gin.H{
			"id":                  user.ID,
			"username":            user.Username,
			"name":                user.Name,
			"profile_picture_url": user.ProfilePictureURL,
			"is_verified":         user.IsVerified,
			"created_at":          user.CreatedAt,
		})
	}

	utils.SendSuccessResponse(c, http.StatusOK, gin.H{
		"blocked_users": formattedUsers,
		"pagination": gin.H{
			"current_page": page,
			"per_page":     limit,
			"total_count":  len(blockedUsers),
		},
	})
}

// ReportUser handles reporting a user
func ReportUser(c *gin.Context) {
	// Get target user ID from route parameter
	targetUserID := c.Param("userId")
	if targetUserID == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "User ID parameter is required")
		return
	}

	// Get authenticated user ID from context
	currentUserID, exists := c.Get("userID")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}
	reporterID := currentUserID.(string)

	// Don't allow users to report themselves
	if reporterID == targetUserID {
		utils.SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Users cannot report themselves")
		return
	}

	// Parse request body
	var requestBody struct {
		Reason string `json:"reason" binding:"required"`
	}

	if err := c.BindJSON(&requestBody); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request format")
		return
	}

	if requestBody.Reason == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Reason is required")
		return
	}

	// Report the user
	err := userServiceClient.ReportUser(reporterID, targetUserID, requestBody.Reason)
	if err != nil {
		log.Printf("Error reporting user %s: %v", targetUserID, err)
		utils.SendErrorResponse(c, http.StatusInternalServerError, "SERVER_ERROR", "Failed to report user")
		return
	}

	utils.SendSuccessResponse(c, http.StatusOK, gin.H{
		"message": "User reported successfully",
	})
}
