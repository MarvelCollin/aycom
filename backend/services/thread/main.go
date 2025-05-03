package main

import (
	"fmt"
	"log"
	"net"
	"os"

	handlers "aycom/backend/services/thread/api"
	"aycom/backend/services/thread/db"
	"aycom/backend/services/thread/proto"
	"aycom/backend/services/thread/service"

	"google.golang.org/grpc"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "9092"
	}
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		getEnv("DATABASE_HOST", "thread_db"),
		getEnv("DATABASE_PORT", "5432"),
		getEnv("DATABASE_USER", "kolin"),
		getEnv("DATABASE_PASSWORD", "kolin"),
		getEnv("DATABASE_NAME", "thread_db"),
	)
	dbConn, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	threadRepo := db.NewThreadRepository(dbConn)
	mediaRepo := db.NewMediaRepository(dbConn)
	hashtagRepo := db.NewHashtagRepository(dbConn)
	mentionRepo := db.NewMentionRepository(dbConn)

	threadService := service.NewThreadService(threadRepo, mediaRepo, hashtagRepo, mentionRepo)
	// TODO: Initialize replyService, interactionService, pollService as needed

	handler := handlers.NewThreadHandler(threadService, nil, nil, nil)

	grpcServer := grpc.NewServer()
	proto.RegisterThreadServiceServer(grpcServer, handler)

	log.Printf("Thread service started on port %s", port)
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
