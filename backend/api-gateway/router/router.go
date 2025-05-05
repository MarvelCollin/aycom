package router

import (
	"aycom/backend/api-gateway/config"
	_ "aycom/backend/api-gateway/docs" // Import swagger docs
	"aycom/backend/api-gateway/routes"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupRouter configures all routes for the API gateway
// @title           AYCOM API Gateway
// @version         1.0
// @description     This is the API Gateway for the AYCOM platform.
// @host            localhost:8081
// @BasePath        /api/v1
// @schemes         http https
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func SetupRouter(cfg *config.Config) *gin.Engine {
	r := gin.Default()

	// Add Swagger documentation endpoint
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	routes.RegisterRoutes(r, cfg)
	return r
}
