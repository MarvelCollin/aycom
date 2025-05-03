package routes

import (
	"aycom/backend/api-gateway/config"
	"aycom/backend/api-gateway/handlers"
	"aycom/backend/api-gateway/middleware"

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
	auth.Use(handlers.RateLimitMiddleware)
	{
		auth.GET("/oauth-config", handlers.GetOAuthConfig)
		// auth.POST("/login", handlers.Login)
		// auth.POST("/register", handlers.Register)
		// auth.POST("/register-with-media", handlers.RegisterWithMedia)
		// auth.POST("/refresh", handlers.RefreshToken)
	}

	// Public user registration and login
	publicUsers := v1.Group("/users")
	{
		publicUsers.POST("/register", handlers.RegisterUser)
		publicUsers.POST("/login", handlers.LoginUser)
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
	// payments := protected.Group("/payments")
	// payments.POST("", handlers.CreatePayment)
	// payments.GET(":id", handlers.GetPayment)
	// payments.GET("/history", handlers.GetPaymentHistory)
}
