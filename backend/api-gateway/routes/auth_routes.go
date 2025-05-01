package routes

import (
	"github.com/Acad600-Tpa/WEB-MV-242/backend/api-gateway/handlers"
	"github.com/Acad600-Tpa/WEB-MV-242/backend/api-gateway/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterAuthRoutes(router *gin.RouterGroup) {
	auth := router.Group("/auth")
	{
		auth.POST("/login", handlers.Login)
		auth.POST("/register", handlers.Register)
		auth.POST("/refresh-token", handlers.RefreshToken)
		auth.POST("/verify-email", handlers.VerifyEmail)
		auth.POST("/resend-verification", handlers.ResendVerificationCode)
		auth.POST("/google", handlers.GoogleAuth)

		authorized := auth.Group("")
		authorized.Use(middleware.AuthMiddleware)
		{
			authorized.POST("/logout", handlers.Logout)
		}
	}
}
