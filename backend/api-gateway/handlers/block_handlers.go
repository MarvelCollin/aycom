package handlers

import (
	"aycom/backend/api-gateway/utils"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func BlockUser(c *gin.Context) {

	targetUserID := c.Param("userId")
	if targetUserID == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "User ID parameter is required")
		return
	}

	currentUserID, exists := c.Get("userID")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}
	blockerID := currentUserID.(string)

	if blockerID == targetUserID {
		utils.SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Users cannot block themselves")
		return
	}

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

func UnblockUser(c *gin.Context) {

	targetUserID := c.Param("userId")
	if targetUserID == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "User ID parameter is required")
		return
	}

	currentUserID, exists := c.Get("userID")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}
	blockerID := currentUserID.(string)

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

func GetBlockedUsers(c *gin.Context) {

	currentUserID, exists := c.Get("userID")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}
	userID := currentUserID.(string)

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

	blockedUsers, err := userServiceClient.GetBlockedUsers(userID, page, limit)
	if err != nil {
		log.Printf("Error getting blocked users for %s: %v", userID, err)
		utils.SendErrorResponse(c, http.StatusInternalServerError, "SERVER_ERROR", "Failed to retrieve blocked users")
		return
	}

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

func ReportUser(c *gin.Context) {

	targetUserID := c.Param("userId")
	if targetUserID == "" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "User ID parameter is required")
		return
	}

	currentUserID, exists := c.Get("userID")
	if !exists {
		utils.SendErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
		return
	}
	reporterID := currentUserID.(string)

	if reporterID == targetUserID {
		utils.SendErrorResponse(c, http.StatusBadRequest, "INVALID_REQUEST", "Users cannot report themselves")
		return
	}

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
