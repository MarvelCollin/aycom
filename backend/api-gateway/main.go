package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "api-gateway/docs" // Import generated docs
)

// @title           AYCOM API Gateway
// @version         1.0
// @description     This is the API Gateway for AYCOM application
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.aycom.example.com/support
// @contact.email  support@aycom.example.com

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

func main() {
	// Get API Gateway port from environment variable
	port := os.Getenv("API_GATEWAY_PORT")
	if port == "" {
		port = "8080" // Default port
	}

	// Initialize Gin router
	r := gin.Default()

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API routes
	api := r.Group("/api/v1")
	{
		// User service routes
		userRoutes := api.Group("/users")
		{
			userRoutes.GET("/", func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "List users endpoint"})
			})
			userRoutes.GET("/:id", func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "Get user by ID: " + c.Param("id")})
			})
		}

		// Product service routes
		productRoutes := api.Group("/products")
		{
			productRoutes.GET("/", func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "List products endpoint"})
			})
			productRoutes.GET("/:id", func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "Get product by ID: " + c.Param("id")})
			})
		}

		// Auth service routes
		authRoutes := api.Group("/auth")
		{
			authRoutes.POST("/login", func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "Login endpoint"})
			})
			authRoutes.POST("/register", func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "Register endpoint"})
			})
		}
	}

	// Swagger documentation
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Start the server
	serverAddr := ":" + port
	fmt.Printf("API Gateway started on port: %s\n", port)
	fmt.Printf("Swagger UI available at: http://localhost:%s/swagger/index.html\n", port)

	if err := r.Run(serverAddr); err != nil {
		log.Fatalf("Failed to start API Gateway: %v", err)
	}
}
