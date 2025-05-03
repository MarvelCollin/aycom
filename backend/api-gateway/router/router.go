package router

import (
	"aycom/backend/api-gateway/config"
	"aycom/backend/api-gateway/routes"

	"github.com/gin-gonic/gin"
)

// SetupRouter configures all routes for the API gateway
func SetupRouter(cfg *config.Config) *gin.Engine {
	r := gin.Default()
	routes.RegisterRoutes(r, cfg)
	return r
}
