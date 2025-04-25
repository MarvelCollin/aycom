package main

import (
	"fmt"
	"log"

	"github.com/Acad600-TPA/WEB-MV-242/backend/gateway/internal/api/router"
	"github.com/Acad600-TPA/WEB-MV-242/backend/gateway/internal/config"
	_ "github.com/Acad600-TPA/WEB-MV-242/backend/gateway/internal/docs" // Import generated Swagger docs
)

// @title AYCOM API Gateway
// @version 1.0
// @description API Gateway for AYCOM microservices architecture

// @contact.name API Support
// @contact.email support@aycom.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/v1
// @schemes http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token

func main() {
	// Load configuration
	cfg := config.Load()

	// Setup router
	r := router.SetupRouter(cfg)

	// Start server
	serverAddr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Starting API Gateway on %s", serverAddr)
	if err := r.Run(serverAddr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
