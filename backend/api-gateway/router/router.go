package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"aycom/backend/api-gateway/config"
	"aycom/backend/api-gateway/docs"
	"aycom/backend/api-gateway/middleware"
	"aycom/backend/api-gateway/routes"
)

func SetupRouter(cfg *config.Config) *gin.Engine {
	r := gin.Default()

	// Enable CORS for all origins - must be first middleware
	r.Use(middleware.CORS())

	// Add debug middleware after CORS
	r.Use(middleware.CORSDebug())

	// Add global OPTIONS handler to handle preflight requests that don't match specific routes
	r.NoRoute(func(c *gin.Context) {
		if c.Request.Method == "OPTIONS" {
			origin := c.Request.Header.Get("Origin")
			if origin == "" {
				origin = "http://localhost:3000"
			}

			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
			c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS, HEAD")
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Admin-Request, X-Debug-Panel, Accept, Cache-Control, X-Requested-With, X-Api-Key, X-Auth-Token, Pragma, Expires, Connection, User-Agent, Host, Referer, Cookie, Set-Cookie, *")
			c.Writer.Header().Set("Access-Control-Max-Age", "86400")
			c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Type, Authorization, X-Powered-By")

			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.JSON(http.StatusNotFound, gin.H{"message": "Not Found"})
	})

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
