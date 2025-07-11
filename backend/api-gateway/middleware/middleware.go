package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
	userpb "aycom/backend/proto/user"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"aycom/backend/api-gateway/handlers"
	"aycom/backend/api-gateway/utils"
)


type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

func isDevelopment() bool {
	return os.Getenv("GIN_MODE") != "release"
}

func CORSDebug() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Printf("CORS Debug: %s request to %s", c.Request.Method, c.Request.URL.Path)
		log.Printf("CORS Debug: Origin header: %s", c.Request.Header.Get("Origin"))

		for name, values := range c.Request.Header {
			log.Printf("CORS Debug: Header %s: %v", name, values)
		}

		if c.Request.Method == "OPTIONS" {
			log.Printf("CORS Debug: Handling OPTIONS request")
		}

		c.Next()

		log.Printf("CORS Debug: Response status: %d", c.Writer.Status())
		for name, values := range c.Writer.Header() {
			log.Printf("CORS Debug: Response header %s: %v", name, values)
		}
	}
}

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin == "" {
			origin = "http:
		}

		c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH, HEAD")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Admin-Request, X-Debug-Panel, Accept, Cache-Control, X-Requested-With, X-Api-Key, X-Auth-Token, Pragma, Expires, Connection, User-Agent, Host, Referer, Cookie, Set-Cookie, *")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Type, Authorization, X-Powered-By")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		origWriter := c.Writer

		c.Writer = &corsResponseWriter{ResponseWriter: origWriter, origin: origin}

		c.Next()
	}
}

type corsResponseWriter struct {
	gin.ResponseWriter
	origin string
}

func (w *corsResponseWriter) Write(data []byte) (int, error) {
	w.ensureCORSHeaders()
	return w.ResponseWriter.Write(data)
}

func (w *corsResponseWriter) WriteHeader(statusCode int) {
	w.ensureCORSHeaders()
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *corsResponseWriter) ensureCORSHeaders() {
	w.Header().Set("Access-Control-Allow-Origin", w.origin)
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH, HEAD")
	w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Admin-Request, X-Debug-Panel, Accept, Cache-Control, X-Requested-With, X-Api-Key, X-Auth-Token, Pragma, Expires, Connection, User-Agent, Host, Referer, Cookie, Set-Cookie, *")
	w.Header().Set("Access-Control-Max-Age", "86400")
	w.Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Type, Authorization, X-Powered-By")
}

func JWTAuth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Printf("JWTAuth: Processing request path: %s", c.Request.URL.Path)

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			log.Printf("JWTAuth: No Authorization header found")
			utils.SendErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authorization header is required")
			c.Abort()
			return
		}

		log.Printf("JWTAuth: Found Authorization header (length: %d)", len(authHeader))

		bearerToken := strings.Split(authHeader, " ")
		if len(bearerToken) != 2 || bearerToken[0] != "Bearer" {
			log.Printf("JWTAuth: Invalid Authorization format: %s", authHeader)
			utils.SendErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid token format")
			c.Abort()
			return
		}

		tokenStr := bearerToken[1]
		if tokenStr == "" {
			log.Printf("JWTAuth: Empty token")
			utils.SendErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid token")
			c.Abort()
			return
		}

		
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				log.Printf("JWTAuth: Unexpected signing method: %v", token.Header["alg"])
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(jwtSecret), nil
		})

		if err != nil {
			log.Printf("JWTAuth: Error parsing token: %v", err)
			if err == jwt.ErrSignatureInvalid {
				utils.SendErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid token signature")
			} else {
				utils.SendErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
			}
			c.Abort()
			return
		}

		if !token.Valid {
			log.Printf("JWTAuth: Invalid token")
			utils.SendErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid token")
			c.Abort()
			return
		}

		log.Printf("JWTAuth: Valid token for user ID: %s", claims.Subject)
		
		c.Set("userID", claims.Subject)
		c.Set("userId", claims.Subject) 
		c.Set("userClaims", claims)

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

		userIdClaim := claims["sub"]
		userIdStr, ok := userIdClaim.(string)
		if !ok {

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
		userID, exists := c.Get("userID")
		log.Printf("AdminOnly: Processing request path: %s for user ID: %v", c.Request.URL.Path, userID)

		if !exists {
			log.Printf("AdminOnly: No userID found in context")
			utils.SendErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not authenticated")
			c.Abort()
			return
		}

		
		if handlers.UserClient == nil {
			log.Printf("AdminOnly: UserClient is nil")
			utils.SendErrorResponse(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "User service is unavailable")
			c.Abort()
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		log.Printf("AdminOnly: Checking if user %v is admin", userID)

		response, err := handlers.UserClient.GetUser(ctx, &userpb.GetUserRequest{
			UserId: userID.(string),
		})

		if err != nil {
			log.Printf("AdminOnly: Error getting user: %v", err)
			utils.SendErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to verify user")
			c.Abort()
			return
		}

		if response == nil || response.User == nil {
			log.Printf("AdminOnly: User not found or nil response")
			utils.SendErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not found")
			c.Abort()
			return
		}

		if !response.User.IsAdmin {
			log.Printf("AdminOnly: User %v is not an admin", userID)
			utils.SendErrorResponse(c, http.StatusForbidden, "FORBIDDEN", "Admin access required")
			c.Abort()
			return
		}

		log.Printf("AdminOnly: User %v is authenticated as admin", userID)
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
