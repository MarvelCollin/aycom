package router

import (
	"github.com/Acad600-Tpa/WEB-MV-242/backend/api-gateway/handlers"
	"github.com/Acad600-Tpa/WEB-MV-242/backend/api-gateway/middleware"
	"github.com/gin-gonic/gin"
)

// SetupRouter configures all routes for the API gateway
func SetupRouter() *gin.Engine {
	r := gin.Default()

	// Middleware
	r.Use(gin.Recovery())
	r.Use(handlers.RateLimitMiddleware())

	// Health check
	r.GET("/health", handlers.HealthCheck)

	// Auth routes - no authentication required
	auth := r.Group("/auth")
	{
		auth.POST("/register", handlers.Register)
		auth.POST("/register-with-media", handlers.RegisterWithMedia)
		auth.POST("/login", handlers.Login)
		auth.POST("/refresh-token", handlers.RefreshToken)
		auth.POST("/verify-email", handlers.VerifyEmail)
		auth.POST("/resend-verification", handlers.ResendVerificationCode)
		auth.POST("/google", handlers.GoogleAuth)
		auth.GET("/oauth-config", handlers.GetOAuthConfig)
	}

	// Protected routes - authentication required
	protected := r.Group("/api")
	protected.Use(middleware.AuthMiddleware())

	// User routes
	users := protected.Group("/users")
	{
		users.GET("/profile", handlers.GetUserProfile)
		users.PUT("/profile", handlers.UpdateUserProfile)
	}

	// Thread routes
	threads := protected.Group("/threads")
	{
		threads.POST("", handlers.CreateThread)
		threads.GET("/:id", handlers.GetThread)
		threads.GET("/user/:userId", handlers.GetThreadsByUser)
		threads.GET("/me", handlers.GetThreadsByUser) // Current user's threads
		threads.PUT("/:id", handlers.UpdateThread)
		threads.DELETE("/:id", handlers.DeleteThread)
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

	return r
}
