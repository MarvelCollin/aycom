package middleware

import (
	"log"
	"net/http"
	"os"
	"strings"

	"encoding/base64"

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

		// Set using both keys to ensure compatibility with different handlers
		c.Set("userId", userIdStr)
		c.Set("userID", userIdStr)
		c.Next()
	}
}

// AuthMiddleware checks if a valid JWT is present
func AuthMiddleware(c *gin.Context) {
	// Log the request path for debugging
	path := c.Request.URL.Path
	log.Printf("AuthMiddleware: Processing request for path: %s", path)

	// Get JWT secret from environment
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "wompwomp123" // Fallback secret, should match .env JWT_SECRET
		log.Println("Warning: JWT_SECRET environment variable not set, using fallback secret")
	}

	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		log.Printf("AuthMiddleware: Missing Authorization header for path: %s", path)
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "User not authenticated",
			"code":    "UNAUTHORIZED",
		})
		c.Abort()
		return
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		log.Printf("AuthMiddleware: Invalid Authorization format: %s for path: %s", authHeader, path)
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Authorization header format must be Bearer {token}",
			"code":    "UNAUTHORIZED",
		})
		c.Abort()
		return
	}

	tokenString := parts[1]
	// Print more info about the token
	tokenLen := len(tokenString)
	if tokenLen > 20 {
		log.Printf("AuthMiddleware: Validating token: %s...%s (len: %d) for path: %s",
			tokenString[:10], tokenString[tokenLen-5:], tokenLen, path)
	} else {
		log.Printf("AuthMiddleware: Validating token (unusually short, len: %d): %s for path: %s",
			tokenLen, tokenString, path)
	}

	claims := jwt.MapClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Make sure token uses the signing method we expect
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			log.Printf("Unexpected signing method: %v for path: %s", token.Header["alg"], path)
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(jwtSecret), nil
	})

	if err != nil {
		log.Printf("Token validation error for path %s: %v", path, err)

		// Check specific error messages without using type assertions
		errMsg := err.Error()
		if strings.Contains(errMsg, "expired") {
			log.Printf("Token has expired")
		} else if strings.Contains(errMsg, "signature") {
			log.Printf("Invalid token signature")
		} else if strings.Contains(errMsg, "malformed") {
			log.Printf("Malformed token")
		} else {
			log.Printf("Other validation error: %s", errMsg)
		}

		// Dump token info for debugging
		tokenParts := strings.Split(tokenString, ".")
		if len(tokenParts) == 3 {
			log.Printf("Token header: %s", tokenParts[0])
			// Try to decode payload for debugging
			encodedPayload := tokenParts[1]
			// Manually decode base64url encoded payload
			if payloadBytes, err := base64.RawURLEncoding.DecodeString(encodedPayload); err == nil {
				log.Printf("Token payload: %s", string(payloadBytes))
			}
		}

		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "User not authenticated",
			"code":    "UNAUTHORIZED",
		})
		c.Abort()
		return
	}

	if !token.Valid {
		log.Printf("AuthMiddleware: Token is invalid for path: %s", path)
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "User not authenticated",
			"code":    "UNAUTHORIZED",
		})
		c.Abort()
		return
	}

	// Extract user ID from token claims
	userIdClaim, exists := claims["user_id"]
	if !exists {
		log.Printf("Token missing user_id claim for path: %s", path)
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "User not authenticated",
			"code":    "UNAUTHORIZED",
		})
		c.Abort()
		return
	}

	// Convert claim to string
	userId, ok := userIdClaim.(string)
	if !ok {
		log.Printf("user_id claim is not a string: %v (type: %T) for path: %s", userIdClaim, userIdClaim, path)
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "User not authenticated",
			"code":    "UNAUTHORIZED",
		})
		c.Abort()
		return
	}

	// Examine other potentially useful claims
	if exp, ok := claims["exp"]; ok {
		log.Printf("Token expires at: %v for user: %s", exp, userId)
	}
	if iat, ok := claims["iat"]; ok {
		log.Printf("Token issued at: %v for user: %s", iat, userId)
	}
	if typ, ok := claims["type"]; ok {
		log.Printf("Token type: %v for user: %s", typ, userId)
	}

	log.Printf("AuthMiddleware: Valid token for user ID: %s on path: %s", userId, path)

	// Set using both keys to ensure compatibility with different handlers
	c.Set("userId", userId)
	c.Set("userID", userId)
	c.Next()
}
