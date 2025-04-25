package middleware

import (
	"net/http"
	"strings"

	"github.com/Acad600-TPA/WEB-MV-242/backend/gateway/internal/config"
	"github.com/gin-gonic/gin"
)

// JWTAuth middleware handles JWT authentication
func JWTAuth(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			return
		}

		token := parts[1]

		// In a real implementation, you would validate the token
		// This is a placeholder for now
		// For example, you might call the auth service via gRPC:
		// client := auth.NewClient(cfg.Services.Auth)
		// valid, userData, err := client.ValidateToken(token)

		// If token is invalid
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		// Set user data in context
		c.Set("user_id", "user123")
		c.Set("user_role", "user")

		c.Next()
	}
}
