package main

import (
	"aycom/backend/proto/community"
	"fmt"
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

	// Get port from environment variable or use default
	port := getEnv("COMMUNITY_SERVICE_PORT", "9093")

	// Start gRPC server
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	community.RegisterCommunityServiceServer(grpcServer, communityHandler)

	log.Printf("Community service started on port %s", port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

func initDatabase() (*gorm.DB, error) {
	// Get database connection parameters from environment variables
	dbHost := getEnv("DATABASE_HOST", "community_db")
	dbPort := getEnv("DATABASE_PORT", "5432")
	dbUser := getEnv("DATABASE_USER", "kolin")
	dbPassword := getEnv("DATABASE_PASSWORD", "kolin")
	dbName := getEnv("DATABASE_NAME", "community_db")

	// Format the database connection string
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	log.Printf("Connecting to database: %s:%s/%s", dbHost, dbPort, dbName)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Auto-migrate models
	log.Println("Community service database migrations started")
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
		&model.DeletedChat{},
	)
	if err != nil {
		return nil, err
	}
	log.Println("Community service database migrations completed successfully")

	// Seed the database with sample data
	seeder := db.NewCommunitySeeder(db)
	if err := seeder.SeedAll(); err != nil {
		log.Printf("Warning: Failed to seed database: %v", err)
	} else {
		log.Println("Database seeded successfully")
	}

	return db, nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
