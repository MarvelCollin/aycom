package routes

import (
	"github.com/Acad600-Tpa/WEB-MV-242/api-gateway/config"
	"github.com/Acad600-Tpa/WEB-MV-242/api-gateway/handlers"
	"github.com/Acad600-Tpa/WEB-MV-242/api-gateway/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes sets up all the routes for the API Gateway
func RegisterRoutes(router *gin.Engine, cfg *config.Config) {
	// Add global middleware
	router.Use(middleware.Logger())
	router.Use(middleware.CORS())

	// Health check
	router.GET("/health", handlers.HealthCheck)

	// API v1 group
	v1 := router.Group("/api/v1")

	// Public routes
	auth := v1.Group("/auth")
	{
		auth.POST("/login", handlers.Login)
		auth.POST("/register", handlers.Register)
		auth.POST("/refresh", handlers.RefreshToken)
		auth.POST("/google", handlers.GoogleAuth)
		auth.POST("/verify-email", handlers.VerifyEmail)
		auth.POST("/resend-code", handlers.ResendVerificationCode)
	}

	// Protected routes - using JWT authentication middleware
	protected := v1.Group("")
	protected.Use(middleware.JWTAuth(cfg.JWTSecret))

	// User routes
	users := protected.Group("/users")
	{
		users.GET("/profile", handlers.GetUserProfile)
		users.PUT("/profile", handlers.UpdateUserProfile)
	}

	// Product routes
	products := protected.Group("/products")
	{
		products.GET("", handlers.ListProducts)
		products.GET("/:id", handlers.GetProduct)
		products.POST("", handlers.CreateProduct)
		products.PUT("/:id", handlers.UpdateProduct)
		products.DELETE("/:id", handlers.DeleteProduct)
	}

	// Payment routes
	payments := protected.Group("/payments")
	{
		payments.POST("", handlers.CreatePayment)
		payments.GET("/:id", handlers.GetPayment)
		payments.GET("/history", handlers.GetPaymentHistory)
	}
}
