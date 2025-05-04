package main

import (
	"log"
	"net"
	"os"
	"sync"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"

	"aycom/backend/services/community/api"
	"aycom/backend/services/community/db"
	"aycom/backend/services/community/proto"
	"aycom/backend/services/community/repository"
	"aycom/backend/services/community/service"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found or cannot be loaded: %v", err)
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

		// Initialize repositories
		communityRepo := repository.NewCommunityRepository(dbConn)
		memberRepo := repository.NewCommunityMemberRepository(dbConn)
		joinRequestRepo := repository.NewCommunityJoinRequestRepository(dbConn)
		ruleRepo := repository.NewCommunityRuleRepository(dbConn)
		chatRepo := repository.NewChatRepository(dbConn)
		messageRepo := repository.NewMessageRepository(dbConn)
		participantRepo := repository.NewParticipantRepository(dbConn)
		deletedChatRepo := repository.NewDeletedChatRepository(dbConn)

		// Initialize the services
		communityService := service.NewCommunityService(communityRepo, memberRepo, joinRequestRepo, ruleRepo)
		chatService := service.NewChatService(
			chatRepo,
			participantRepo,
			messageRepo,
			deletedChatRepo,
		)

		handler := api.NewCommunityHandler(communityService, chatService)

		grpcServer := grpc.NewServer()
		proto.RegisterCommunityServiceServer(grpcServer, handler)
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
