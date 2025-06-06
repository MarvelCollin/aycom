package middleware

import (
	"aycom/backend/api-gateway/utils"
	"fmt"
	"log"
	"net/http"
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

		// First try to get user ID from the standard "sub" claim
		userIdClaim := claims["sub"]
		log.Printf("JWT Middleware: Extracted sub claim: %v (Type: %T)", userIdClaim, userIdClaim)

		userIdStr, ok := userIdClaim.(string)
		if !ok {
			// Fallback to "user_id" for backward compatibility
			userIdClaim = claims["user_id"]
			log.Printf("JWT Middleware: No valid sub claim, trying user_id claim: %v (Type: %T)", userIdClaim, userIdClaim)

			userIdStr, ok = userIdClaim.(string)
			if !ok {
				log.Printf("JWT Middleware: No valid user identifier in token claims")

				// Log all available claims for debugging
				log.Printf("JWT Middleware: Available claims:")
				for key, value := range claims {
					log.Printf("  %s: %v (Type: %T)", key, value, value)
				}

				c.JSON(http.StatusUnauthorized, gin.H{
					"success": false,
					"message": "Invalid token claims",
					"code":    "UNAUTHORIZED",
				})
				c.Abort()
				return
			}
		}

		log.Printf("JWT Middleware: Successfully extracted user ID: %s from token", userIdStr)
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

		// Try "sub" claim first (standard JWT approach)
		userIdClaim := claims["sub"]
		userIdStr, ok := userIdClaim.(string)
		if !ok {
			// Fallback to "user_id" for backward compatibility
			userIdClaim = claims["user_id"]
			userIdStr, ok = userIdClaim.(string)
			if !ok {
				log.Printf("JWT Middleware: No valid user identifier in token claims - continuing as anonymous")
				c.Next()
				return
			}
		}

		c.Set("userId", userIdStr)
		c.Set("userID", userIdStr)
		log.Printf("JWT Middleware: Successfully validated token for user %s", userIdStr)
		c.Next()
	}
}

func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from context
		_, exists := c.Get("userID")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Unauthorized, missing user identity",
				"code":    "UNAUTHORIZED",
			})
			return
		}

		// Check admin status with actual token
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Unauthorized, missing token",
				"code":    "UNAUTHORIZED",
			})
			return
		}

		if !strings.HasPrefix(tokenString, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Unauthorized, invalid token format",
				"code":    "UNAUTHORIZED",
			})
			return
		}

		tokenString = tokenString[7:] // Remove Bearer prefix

		// Actually validate the token and check isAdmin claim
		claims := jwt.MapClaims{}
		secret := utils.GetJWTSecret()

		_, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return secret, nil
		})

		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Unauthorized, invalid token",
				"code":    "UNAUTHORIZED",
			})
			return
		}

		// Make sure we have a valid user ID first
		_, userOk := claims["sub"].(string)
		if !userOk {
			// Try legacy format
			_, userOk = claims["user_id"].(string)
			if !userOk {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"success": false,
					"message": "Unauthorized, invalid token claims",
					"code":    "UNAUTHORIZED",
				})
				return
			}
		}

		// Check if user is admin in claims
		isAdmin, ok := claims["is_admin"].(bool)
		if !ok || !isAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "Forbidden, admin access required",
				"code":    "FORBIDDEN",
			})
			return
		}

		c.Next()
	}
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
