package middleware

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// CORS middleware for handling Cross-Origin Resource Sharing
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Allow specific origin (your frontend dev server)
		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
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
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header format must be Bearer {token}"})
			c.Abort()
			return
		}

		token := parts[1]
		claims := jwt.MapClaims{}

		parsedToken, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})

		if err != nil || !parsedToken.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
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

		c.Set("userId", userIdStr)
		c.Next()
	}
}

// AuthMiddleware checks if a valid JWT is present
func AuthMiddleware(c *gin.Context) {
	// Get JWT secret from environment
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "wompwomp123" // Fallback secret, should match .env JWT_SECRET
		log.Println("Warning: JWT_SECRET environment variable not set, using fallback secret")
	}

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

	tokenString := parts[1]
	claims := jwt.MapClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Make sure token uses the signing method we expect
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			log.Printf("Unexpected signing method: %v", token.Header["alg"])
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(jwtSecret), nil
	})

	if err != nil {
		log.Printf("Token validation error: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Invalid or expired token",
			"code":    "UNAUTHORIZED",
		})
		c.Abort()
		return
	}

	if !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Invalid token",
			"code":    "UNAUTHORIZED",
		})
		c.Abort()
		return
	}

	// Extract user ID from token claims
	userIdClaim, exists := claims["user_id"]
	if !exists {
		log.Printf("Token missing user_id claim")
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Invalid token: missing user ID",
			"code":    "UNAUTHORIZED",
		})
		c.Abort()
		return
	}

	// Convert claim to string
	userId, ok := userIdClaim.(string)
	if !ok {
		log.Printf("user_id claim is not a string: %v (type: %T)", userIdClaim, userIdClaim)
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Invalid token claims format",
			"code":    "UNAUTHORIZED",
		})
		c.Abort()
		return
	}

	// Set user ID in context for use in handlers
	c.Set("userId", userId)
	c.Next()
}
