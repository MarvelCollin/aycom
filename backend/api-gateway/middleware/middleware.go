package middleware

import (
	"net/http"
	"strings"

	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// CORS middleware for handling Cross-Origin Resource Sharing
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin == "" {
			origin = "http://localhost:3000" // Default to frontend origin if not specified
		}

		// When credentials are included, Access-Control-Allow-Origin cannot be '*'
		// It must be set to a specific origin
		c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		// Allow credentials (cookies, authorization headers)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// JWTAuth middleware for JWT authentication
func JWTAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Printf("Processing JWT authentication for: %s %s", c.Request.Method, c.Request.URL.Path)

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			log.Printf("No Authorization header found for %s %s", c.Request.Method, c.Request.URL.Path)
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "User not authenticated",
				"code":    "UNAUTHORIZED",
			})
			c.Abort()
			return
		}

		log.Printf("Auth header found: %s...", authHeader[:min(len(authHeader), 15)])

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			log.Printf("Invalid Authorization header format: %s", authHeader[:min(len(authHeader), 30)])
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Authorization header format must be Bearer {token}",
				"code":    "UNAUTHORIZED",
			})
			c.Abort()
			return
		}

		token := parts[1]
		log.Printf("Token length: %d chars", len(token))

		if token == "" {
			log.Printf("Empty token provided")
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "User not authenticated",
				"code":    "UNAUTHORIZED",
			})
			c.Abort()
			return
		}

		claims := jwt.MapClaims{}

		parsedToken, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
			// Check signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(secret), nil
		})

		if err != nil {
			log.Printf("JWT parse error: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "User not authenticated",
				"code":    "UNAUTHORIZED",
			})
			c.Abort()
			return
		}

		if !parsedToken.Valid {
			log.Printf("Invalid JWT token")
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "User not authenticated",
				"code":    "UNAUTHORIZED",
			})
			c.Abort()
			return
		}

		// Log the extracted user ID for debugging
		userIdClaim := claims["user_id"]
		log.Printf("JWT Middleware: Extracted user_id claim: %v (Type: %T)", userIdClaim, userIdClaim)

		// Ensure the claim is a string before setting
		userIdStr, ok := userIdClaim.(string)
		if !ok {
			log.Printf("JWT Middleware: user_id claim is not a string: %v", userIdClaim)
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Invalid token claims",
				"code":    "UNAUTHORIZED",
			})
			c.Abort()
			return
		}

		// Set both userID and userId for compatibility
		c.Set("userId", userIdStr)
		c.Set("userID", userIdStr)
		log.Printf("JWT Middleware: Successfully validated token for user %s", userIdStr)
		c.Next()
	}
}

// OptionalJWTAuth middleware for optional JWT authentication
// If the user is authenticated, set the userId in the context
// If not, allow the request to continue anyway
func OptionalJWTAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Printf("Processing optional JWT authentication for: %s %s", c.Request.Method, c.Request.URL.Path)

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			log.Printf("No Authorization header found for %s %s - continuing as anonymous", c.Request.Method, c.Request.URL.Path)
			c.Next()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			log.Printf("Invalid Authorization header format: %s - continuing as anonymous", authHeader[:min(len(authHeader), 30)])
			c.Next()
			return
		}

		token := parts[1]
		if token == "" {
			log.Printf("Empty token provided - continuing as anonymous")
			c.Next()
			return
		}

		claims := jwt.MapClaims{}

		parsedToken, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
			// Check signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(secret), nil
		})

		if err != nil || !parsedToken.Valid {
			log.Printf("JWT parse error or invalid token - continuing as anonymous")
			c.Next()
			return
		}

		// Extract user ID from token
		userIdClaim := claims["user_id"]
		userIdStr, ok := userIdClaim.(string)
		if !ok {
			log.Printf("JWT Middleware: user_id claim is not a string: %v - continuing as anonymous", userIdClaim)
			c.Next()
			return
		}

		// Set both userID and userId for compatibility
		c.Set("userId", userIdStr)
		c.Set("userID", userIdStr)
		log.Printf("JWT Middleware: Successfully validated token for user %s", userIdStr)
		c.Next()
	}
}

// Helper function to get the minimum of two ints
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func AuthMiddleware(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Authorization header is required",
			"code":    "UNAUTHORIZED",
		})
		c.Abort()
		return
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Authorization header format must be Bearer {token}",
			"code":    "UNAUTHORIZED",
		})
		c.Abort()
		return
	}

	c.Next()
}
