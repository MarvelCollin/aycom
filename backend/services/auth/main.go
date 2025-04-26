package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/Acad600-Tpa/WEB-MV-242/services/auth/config"
	"github.com/Acad600-Tpa/WEB-MV-242/services/auth/handler"
	"github.com/Acad600-Tpa/WEB-MV-242/services/auth/repository"
	"github.com/Acad600-Tpa/WEB-MV-242/services/auth/service"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Initialize configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Set up database connection
	db, err := repository.NewPostgresConnection(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Set up repository layer
	authRepo := repository.NewAuthRepository(db)

	// Set up service layer
	authService := service.NewAuthService(authRepo, cfg.JWTSecret)

	// Set up gRPC server
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.Port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	// Register gRPC service handlers
	handler.RegisterAuthServiceServer(grpcServer, authService)

	// Start gRPC server
	log.Printf("Auth service starting on port %s", cfg.Port)
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// Set up graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down auth service...")
	grpcServer.GracefulStop()
	log.Println("Auth service stopped")
}
