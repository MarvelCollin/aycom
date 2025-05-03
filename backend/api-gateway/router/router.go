package router

import (
	"aycom/backend/api-gateway/handlers"
	"aycom/backend/api-gateway/middleware"
	"aycom/backend/api-gateway/routes"

	"github.com/gin-gonic/gin"
)

// SetupRouter configures all routes for the API gateway
func SetupRouter() *gin.Engine {
	r := gin.Default()

	// Apply middleware first
	r.Use(gin.Recovery())
	r.Use(middleware.CORS())
	r.Use(middleware.Logger())

	// Health check
	r.GET("/health", handlers.HealthCheck)

	// API v1 group
	v1 := r.Group("/api/v1")
	v1.Use(handlers.RateLimitMiddleware)

	routes.RegisterAuthRoutes(v1)

	// Protected routes - using JWT authentication middleware
	protected := v1.Group("")
	protected.Use(middleware.JWTAuth(handlers.Config.JWTSecret))

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

	return r
}
