package routes

import (
	"github.com/gin-gonic/gin"
)

// RegisterAuthRoutes registers additional auth routes if needed.
// Note: Primary auth routes are already defined in routes.go - use this function
// only for special cases or extensions not covered in the main routes.
func RegisterAuthRoutes(router *gin.RouterGroup) {
	// Example for additional auth routes that aren't in main routes.go
	// Uncomment and customize as needed:

	/*
		auth := router.Group("/auth")
		{
			// Add any specialized auth routes here - ensure they don't duplicate routes.go

			// Protected auth routes
			authorized := auth.Group("")
			authorized.Use(middleware.AuthMiddleware)
			{
				authorized.POST("/logout", handlers.Logout)
				// Add other protected auth routes
			}
		}
	*/
}
