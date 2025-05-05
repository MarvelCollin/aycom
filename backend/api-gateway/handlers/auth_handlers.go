package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// AuthHandler handles authentication
// @Summary Authentication related endpoints
// @Description Provides authentication services for the API
// @Tags Authentication
func AuthHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Auth handler working",
		})
	}
}

// RefreshToken handles token refresh
// @Summary Refresh access token
// @Description Refreshes the access token using a valid refresh token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param refresh_token body string true "Refresh token"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/auth/refresh [post]
func RefreshToken(c *gin.Context) {
	// TODO: Implement token refresh
	c.JSON(http.StatusOK, gin.H{
		"message": "Token refreshed successfully",
	})
}
