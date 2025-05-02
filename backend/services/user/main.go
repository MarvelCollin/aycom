package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	handlers "aycom/backend/services/user/api"
	"aycom/backend/services/user/db"
	"aycom/backend/services/user/proto"
	"aycom/backend/services/user/service"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "9091"
	}
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		getEnv("DATABASE_HOST", "user_db"),
		getEnv("DATABASE_PORT", "5432"),
		getEnv("DATABASE_USER", "kolin"),
		getEnv("DATABASE_PASSWORD", "kolin"),
		getEnv("DATABASE_NAME", "user_db"),
	)
	dbConn, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	repo := db.NewPostgresUserRepository(dbConn)
	svc := service.NewUserService(repo)
	handler := handlers.NewUserHandler(svc)

	grpcServer := grpc.NewServer()
	proto.RegisterUserServiceServer(grpcServer, handler)

	log.Printf("User service started on port %s", port)
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
