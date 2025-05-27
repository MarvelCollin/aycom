package router

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"aycom/backend/api-gateway/config"
	"aycom/backend/api-gateway/docs"
	"aycom/backend/api-gateway/routes"
)

func SetupRouter(cfg *config.Config) *gin.Engine {
	r := gin.Default()

	docs.SwaggerInfo.Host = "localhost:8083"
	docs.SwaggerInfo.BasePath = "/api/v1"
	docs.SwaggerInfo.Title = "AYCOM API"
	docs.SwaggerInfo.Description = "API Gateway for AYCOM platform"
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Schemes = []string{"http", "https"}

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	routes.RegisterRoutes(r, cfg)
	return r
}
