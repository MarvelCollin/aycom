package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware verifies the JWT token and adds the user info to the context
func AuthMiddleware(authClient interface{}) gin.HandlerFunc {
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

		// Using underscore to indicate it's intentionally unused for now
		_ = parts[1] // token value, will be used in actual implementation

		// Here we would validate the token using the auth service client
		// For now, we'll just implement a placeholder
		// In a real implementation, you'd call the auth service via gRPC

		// Example of what this would look like with an actual auth client:
		// resp, err := authClient.ValidateToken(context.Background(), &pb.ValidateTokenRequest{Token: token})
		// if err != nil {
		//     c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		//     return
		// }

		// Set user information in the context for downstream handlers
		c.Set("user_id", "placeholder-user-id")
		c.Set("user_role", "user") // Default role

		c.Next()
	}
}
