package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/Acad600-TPA/WEB-MV-242/auth/internal/repository/postgres"
	"github.com/Acad600-TPA/WEB-MV-242/auth/internal/server"
	pb "github.com/Acad600-TPA/WEB-MV-242/auth/proto"
	"google.golang.org/grpc"
)

// go:generate protoc --go_out=. --go-grpc_out=. ../proto/auth.proto

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "9090"
	}

	// Get JWT key from environment variable or use default for development
	jwtKey := os.Getenv("JWT_SECRET")
	if jwtKey == "" {
		jwtKey = "your-default-secret-key-for-development"
		log.Println("Warning: Using default JWT secret key. This is insecure for production.")
	}

	// Database connection
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	// Initialize repository
	userRepo, err := postgres.NewUserRepository(
		fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			dbHost, dbPort, dbUser, dbPass, dbName),
	)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Start gRPC server
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	// Register auth service
	authServer := server.NewAuthServer(userRepo, jwtKey)
	pb.RegisterAuthServiceServer(grpcServer, authServer)

	log.Printf("Auth service starting on port %s...", port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
