package routes

import (
	"github.com/Acad600-Tpa/WEB-MV-242/backend/api-gateway/config"
	"github.com/Acad600-Tpa/WEB-MV-242/backend/api-gateway/handlers"
	"github.com/Acad600-Tpa/WEB-MV-242/backend/api-gateway/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes sets up all the routes for the API Gateway
func RegisterRoutes(router *gin.Engine, cfg *config.Config) {
	// Set the config for handlers
	handlers.Config = cfg

	// Add global middleware
	router.Use(middleware.Logger())
	router.Use(middleware.CORS())

	// Health check
	router.GET("/health", handlers.HealthCheck)

	// API v1 group
	v1 := router.Group("/api/v1")

	// Public routes with rate limiting
	auth := v1.Group("/auth")
	auth.Use(handlers.RateLimitMiddleware())
	{
		auth.GET("/oauth-config", handlers.GetOAuthConfig)
		auth.POST("/login", handlers.Login)
		auth.POST("/register", handlers.Register)
		auth.POST("/register-with-media", handlers.RegisterWithMedia)
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
		users.GET("/suggestions", handlers.GetSuggestedUsers)
		users.GET("/check-username", handlers.CheckUsernameAvailability)
		users.POST("/:id/follow", handlers.FollowUser)
		users.POST("/:id/unfollow", handlers.UnfollowUser)
		users.GET("/:id/followers", handlers.GetUserFollowers)
		users.GET("/:id/following", handlers.GetUserFollowing)
	}

	// Thread routes
	threads := protected.Group("/threads")
	{
		threads.POST("", handlers.CreateThread)
		threads.GET("/:id", handlers.GetThread)
		threads.GET("/user/:id", handlers.GetThreadsByUser)
		threads.PUT("/:id", handlers.UpdateThread)
		threads.DELETE("/:id", handlers.DeleteThread)
		threads.POST("/media", handlers.UploadThreadMedia)
	}

	// Trend routes
	trends := protected.Group("/trends")
	{
		trends.GET("", handlers.GetTrends)
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
