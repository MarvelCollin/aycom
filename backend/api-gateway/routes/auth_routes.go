package routes

import (
	"aycom/backend/api-gateway/handlers"
	"aycom/backend/api-gateway/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterAuthRoutes(router *gin.RouterGroup) {
	auth := router.Group("/auth")
	{
		auth.POST("/login", handlers.Login)
		auth.POST("/register", handlers.Register)
		auth.POST("/refresh-token", handlers.RefreshToken)
		// auth.POST("/verify-email", handlers.VerifyEmail)
		// auth.POST("/resend-verification", handlers.ResendVerificationCode)
		// auth.POST("/google", handlers.GoogleAuth)

		authorized := auth.Group("")
		authorized.Use(middleware.AuthMiddleware)
		{
			authorized.POST("/logout", handlers.Logout)
		}
	}

	// User authentication routes
	api := router.Group("/api/v1")
	authGroup := api.Group("/auth")
	{
		authGroup.POST("/register", handlers.Register)
		authGroup.POST("/login", handlers.Login) // Add Login route
		// authGroup.POST("/refresh", handlers.RefreshToken)
		// authGroup.POST("/logout", handlers.Logout)
		// Add other auth routes like verification, password reset, etc.
	}
}
