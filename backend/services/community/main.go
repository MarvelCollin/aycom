package main

import (
	"aycom/backend/proto/community"
	"log"
	"net"
	"os"
	"path/filepath"
	"sync"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"

	"aycom/backend/services/community/api"
	"aycom/backend/services/community/db"
	"aycom/backend/services/community/repository"
	"aycom/backend/services/community/service"
)

func main() {
	if err := godotenv.Load(); err != nil {
		rootEnvPath := filepath.Join("..", "..", "..", ".env")
		if err := godotenv.Load(rootEnvPath); err != nil {
			log.Printf("Warning: .env file not found or cannot be loaded: %v", err)
		} else {
			log.Printf("Loaded .env from root directory: %s", rootEnvPath)
		}
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		port := getEnv("PORT", "9093")
		listener, err := net.Listen("tcp", ":"+port)
		if err != nil {
			log.Fatalf("Failed to listen: %v", err)
		}

		dbConn := db.InitDB()
		if err := db.RunMigrations(dbConn); err != nil {
			log.Fatalf("Failed to run migrations: %v", err)
		}

		communityRepo := repository.NewCommunityRepository(dbConn)
		categoryRepo := repository.NewCategoryRepository(dbConn)
		memberRepo := repository.NewCommunityMemberRepository(dbConn)
		joinRequestRepo := repository.NewCommunityJoinRequestRepository(dbConn)
		ruleRepo := repository.NewCommunityRuleRepository(dbConn)
		chatRepo := repository.NewChatRepository(dbConn)
		messageRepo := repository.NewMessageRepository(dbConn)
		participantRepo := repository.NewParticipantRepository(dbConn)
		deletedChatRepo := repository.NewDeletedChatRepository(dbConn)

		communityService := service.NewCommunityService(
			communityRepo,
			categoryRepo,
			memberRepo,
			joinRequestRepo,
			ruleRepo,
		)
		chatService := service.NewChatService(
			chatRepo,
			participantRepo,
			messageRepo,
			deletedChatRepo,
		)

		handler := api.NewCommunityHandler(communityService, chatService, memberRepo, joinRequestRepo)

		grpcServer := grpc.NewServer()
		community.RegisterCommunityServiceServer(grpcServer, handler)
		log.Printf("Community service started on port %s", port)
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	wg.Wait()
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
