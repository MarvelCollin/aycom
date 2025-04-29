package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	// Adjust import paths if handlers/proto are elsewhere
	"github.com/Acad600-Tpa/WEB-MV-242/backend/services/user/internal/handlers"
	"github.com/Acad600-Tpa/WEB-MV-242/backend/services/user/model"
	"github.com/Acad600-Tpa/WEB-MV-242/backend/services/user/pkg/db"
	"github.com/Acad600-Tpa/WEB-MV-242/backend/services/user/proto"
	"github.com/Acad600-Tpa/WEB-MV-242/backend/services/user/repository"
	"github.com/Acad600-Tpa/WEB-MV-242/backend/services/user/service"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// --- Database Connection ---
	dbConn, err := db.ConnectDatabaseWithRetry()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Println("Successfully connected to database")

	// --- Run Migrations ---
	err = dbConn.AutoMigrate(&model.User{})
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
	log.Println("Database migration completed")

	// --- Dependency Injection ---
	userRepo := repository.NewPostgresUserRepository(dbConn)
	userService := service.NewUserService(userRepo)
	userHandler := handlers.NewUserHandler(userService)

	// --- Seed Data (Optional) ---
	if len(os.Args) > 1 && os.Args[1] == "seed" {
		log.Println("Seeding default users...")
		seeder := repository.NewUserSeeder(dbConn)
		if err := seeder.SeedUsers(); err != nil {
			log.Fatalf("Failed to seed users: %v", err)
		}
		log.Println("Seeding completed.")
		return // Exit after seeding
	}

	// --- gRPC Server Setup ---
	grpcPort := os.Getenv("USER_SERVICE_PORT")
	if grpcPort == "" {
		grpcPort = "9091" // Default gRPC port
	}
	listener, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", grpcPort, err)
	}

	grpcServer := grpc.NewServer()
	proto.RegisterUserServiceServer(grpcServer, userHandler)

	log.Printf("User gRPC server starting on port %s...", grpcPort)

	// Start gRPC server in a goroutine
	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			log.Printf("gRPC server failed to serve: %v", err)
		}
	}()

	// --- Graceful Shutdown ---
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down user service...")

	// Gracefully stop the gRPC server
	grpcServer.GracefulStop()
	log.Println("gRPC server stopped.")

	// Close database connection (optional, depends on context)
	if sqlDB, err := dbConn.DB(); err == nil {
		sqlDB.Close()
		log.Println("Database connection closed.")
	}

	log.Println("User service stopped gracefully.")
}
