package router

import (
	"github.com/Acad600-TPA/WEB-MV-242/backend/gateway/internal/api/handlers"
	"github.com/Acad600-TPA/WEB-MV-242/backend/gateway/internal/config"
	"github.com/Acad600-TPA/WEB-MV-242/backend/gateway/internal/middleware"
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupRouter initializes the router and registers all routes
func SetupRouter(cfg *config.Config) *gin.Engine {
	router := gin.Default()

	// Enable CORS
	router.Use(middleware.CORS())

	// Setup API documentation using Swagger
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	// Health check endpoint
	router.GET("/health", handlers.HealthCheck)

	// Create API v1 group
	v1 := router.Group("/api/v1")
	{
		// Auth routes - no authentication required
		auth := v1.Group("/auth")
		{
			auth.POST("/register", handlers.Register)
			auth.POST("/login", handlers.Login)
			auth.POST("/refresh", handlers.RefreshToken)
		}

		// Protected routes - require authentication
		protected := v1.Group("")
		protected.Use(middleware.JWTAuth(cfg))
		{
			// User routes
			user := protected.Group("/users")
			{
				user.GET("/me", handlers.GetUserProfile)
				user.PUT("/me", handlers.UpdateUserProfile)
			}

			// Product routes
			product := protected.Group("/products")
			{
				product.GET("", handlers.ListProducts)
				product.GET("/:id", handlers.GetProduct)
				product.POST("", handlers.CreateProduct)
				product.PUT("/:id", handlers.UpdateProduct)
				product.DELETE("/:id", handlers.DeleteProduct)
			}
		}
	}

	return router
}
