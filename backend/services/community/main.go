package main

import (
	"aycom/backend/proto/community"
	"log"
	"net"
	"os"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"

	"aycom/backend/services/community/api"
	"aycom/backend/services/community/model"
	"aycom/backend/services/community/repository"
	"aycom/backend/services/community/service"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	db, err := initDatabase()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Initialize repositories
	communityRepo := repository.NewCommunityRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	memberRepo := repository.NewCommunityMemberRepository(db)
	joinRequestRepo := repository.NewCommunityJoinRequestRepository(db)
	ruleRepo := repository.NewCommunityRuleRepository(db)
	chatRepo := repository.NewChatRepository(db)
	participantRepo := repository.NewParticipantRepository(db)
	messageRepo := repository.NewMessageRepository(db)
	deletedChatRepo := repository.NewDeletedChatRepository(db)

	// Initialize services
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

	// Initialize gRPC handler
	communityHandler := api.NewCommunityHandler(
		communityService,
		chatService,
		memberRepo,
		joinRequestRepo,
		ruleRepo,
	)

	// Start gRPC server
	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	community.RegisterCommunityServiceServer(grpcServer, communityHandler)

	log.Println("Community service running on :50052")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

func initDatabase() (*gorm.DB, error) {
	dsn := "host=localhost user=postgres password=postgres dbname=aycom port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Auto-migrate models
	err = db.AutoMigrate(
		&model.Community{},
		&model.Category{},
		&model.CommunityCategory{},
		&model.CommunityMember{},
		&model.CommunityJoinRequest{},
		&model.CommunityRule{},
		&model.Chat{},
		&model.ChatParticipant{},
		&model.Message{},
	)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
