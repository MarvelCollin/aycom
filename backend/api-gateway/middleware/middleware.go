package middleware

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin == "" {
			origin = "http://localhost:3000" 
		}

		c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, Access-Control-Allow-Headers, X-Debug-Panel")

		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400") 

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

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

		userIdClaim := claims["user_id"]
		log.Printf("JWT Middleware: Extracted user_id claim: %v (Type: %T)", userIdClaim, userIdClaim)

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
		c.Set("userID", userIdStr)
		log.Printf("JWT Middleware: Successfully validated token for user %s", userIdStr)
		c.Next()
	}
}

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

		userIdClaim := claims["user_id"]
		userIdStr, ok := userIdClaim.(string)
		if !ok {
			log.Printf("JWT Middleware: user_id claim is not a string: %v - continuing as anonymous", userIdClaim)
			c.Next()
			return
		}

		c.Set("userId", userIdStr)
		c.Set("userID", userIdStr)
		log.Printf("JWT Middleware: Successfully validated token for user %s", userIdStr)
		c.Next()
	}
}

func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Printf("Checking admin privileges for: %s %s", c.Request.Method, c.Request.URL.Path)

		userID, exists := c.Get("userID")
		if !exists {
			log.Printf("No userID found in context - admin check failed")
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Authentication required",
				"code":    "UNAUTHORIZED",
			})
			c.Abort()
			return
		}

		authHeader := c.GetHeader("Authorization")
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			log.Printf("Invalid Authorization header format: %s", authHeader[:min(len(authHeader), 30)])
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Authentication required",
				"code":    "UNAUTHORIZED",
			})
			c.Abort()
			return
		}

		token := parts[1]
		claims := jwt.MapClaims{}

		secret := getJWTSecret() 
		_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(secret), nil
		})

		if err != nil {
			log.Printf("JWT parse error: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Authentication required",
				"code":    "UNAUTHORIZED",
			})
			c.Abort()
			return
		}

		isAdmin, exists := claims["is_admin"]
		if !exists || isAdmin != true {
			log.Printf("User %v is not an admin", userID)
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "Admin privileges required",
				"code":    "FORBIDDEN",
			})
			c.Abort()
			return
		}

		log.Printf("Admin check passed for user %v", userID)
		c.Next()
	}
}

func getJWTSecret() string {

	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		return secret
	}

	return "your-secret-key"
}

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