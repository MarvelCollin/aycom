package main

import (
	"aycom/backend/proto/community"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"

	"aycom/backend/services/community/api"
	"aycom/backend/services/community/model"
	"aycom/backend/services/community/repository"
	"aycom/backend/services/community/service"

	"github.com/google/uuid"
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

	communityRepo := repository.NewCommunityRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	memberRepo := repository.NewCommunityMemberRepository(db)
	joinRequestRepo := repository.NewCommunityJoinRequestRepository(db)
	ruleRepo := repository.NewCommunityRuleRepository(db)
	chatRepo := repository.NewChatRepository(db)
	participantRepo := repository.NewParticipantRepository(db)
	messageRepo := repository.NewMessageRepository(db)
	deletedChatRepo := repository.NewDeletedChatRepository(db)

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

	communityHandler := api.NewCommunityHandler(
		communityService,
		chatService,
		memberRepo,
		joinRequestRepo,
		ruleRepo,
	)

	port := getEnv("COMMUNITY_SERVICE_PORT", "9093")

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

	dbHost := getEnv("DATABASE_HOST", "community_db")
	dbPort := getEnv("DATABASE_PORT", "5432")
	dbUser := getEnv("DATABASE_USER", "kolin")
	dbPassword := getEnv("DATABASE_PASSWORD", "kolin")
	dbName := getEnv("DATABASE_NAME", "community_db")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	log.Printf("Connecting to database: %s:%s/%s", dbHost, dbPort, dbName)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

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

	if err := seedDatabase(db); err != nil {
		log.Printf("Warning: Failed to seed database: %v", err)
	} else {
		log.Println("Database seeded successfully")
	}

	return db, nil
}

func seedDatabase(db *gorm.DB) error {

	var communityCount int64
	if err := db.Model(&model.Community{}).Count(&communityCount).Error; err != nil {
		return fmt.Errorf("failed to count communities: %w", err)
	}

	if communityCount > 0 {
		log.Println("Communities already exist, skipping seeding")
		return nil
	}

	userIDs := []uuid.UUID{
		uuid.MustParse("91df5727-a9c5-427e-94ce-e0486e3bfdb7"), 
		uuid.MustParse("fd434c0e-95de-41d0-a576-9d4ea2fed7e9"), 
	}

	communities := []model.Community{
		{
			CommunityID: uuid.New(),
			Name:        "Tech Enthusiasts",
			Description: "A community for technology lovers and early adopters. We discuss the latest gadgets, software releases, and tech trends.",
			LogoURL:     "https://via.placeholder.com/150",
			BannerURL:   "https://via.placeholder.com/600x200",
			CreatorID:   userIDs[0],
			IsApproved:  true,
			CreatedAt:   time.Now().Add(-60 * 24 * time.Hour),
			UpdatedAt:   time.Now().Add(-60 * 24 * time.Hour),
		},
		{
			CommunityID: uuid.New(),
			Name:        "Fitness & Health",
			Description: "Join us to discuss fitness routines, health tips, nutrition advice, and wellness strategies.",
			LogoURL:     "https://via.placeholder.com/150",
			BannerURL:   "https://via.placeholder.com/600x200",
			CreatorID:   userIDs[0],
			IsApproved:  true,
			CreatedAt:   time.Now().Add(-55 * 24 * time.Hour),
			UpdatedAt:   time.Now().Add(-55 * 24 * time.Hour),
		},
		{
			CommunityID: uuid.New(),
			Name:        "Developers Hub",
			Description: "A community for software developers to share knowledge, discuss programming languages, and collaborate on projects.",
			LogoURL:     "https://via.placeholder.com/150",
			BannerURL:   "https://via.placeholder.com/600x200",
			CreatorID:   userIDs[0],
			IsApproved:  true,
			CreatedAt:   time.Now().Add(-50 * 24 * time.Hour),
			UpdatedAt:   time.Now().Add(-50 * 24 * time.Hour),
		},
	}

	if err := db.Create(&communities).Error; err != nil {
		return fmt.Errorf("failed to seed communities: %w", err)
	}

	log.Printf("Created %d communities", len(communities))

	members := []model.CommunityMember{}

	for _, community := range communities {
		members = append(members, model.CommunityMember{
			CommunityID: community.CommunityID,
			UserID:      community.CreatorID,
			Role:        "admin",
			CreatedAt:   time.Now().Add(-60 * 24 * time.Hour),
			UpdatedAt:   time.Now().Add(-60 * 24 * time.Hour),
		})
	}

	if len(communities) >= 2 && len(userIDs) >= 2 {
		members = append(members, model.CommunityMember{
			CommunityID: communities[0].CommunityID,
			UserID:      userIDs[1],
			Role:        "member",
			CreatedAt:   time.Now().Add(-55 * 24 * time.Hour),
			UpdatedAt:   time.Now().Add(-55 * 24 * time.Hour),
		})

		members = append(members, model.CommunityMember{
			CommunityID: communities[1].CommunityID,
			UserID:      userIDs[1],
			Role:        "moderator",
			CreatedAt:   time.Now().Add(-50 * 24 * time.Hour),
			UpdatedAt:   time.Now().Add(-50 * 24 * time.Hour),
		})
	}

	if err := db.Create(&members).Error; err != nil {
		return fmt.Errorf("failed to seed community members: %w", err)
	}

	log.Printf("Created %d community members", len(members))

	if len(communities) >= 3 && len(userIDs) >= 2 {
		joinRequest := model.CommunityJoinRequest{
			RequestID:   uuid.New(),
			CommunityID: communities[2].CommunityID,
			UserID:      userIDs[1],
			Status:      "pending",
			CreatedAt:   time.Now().Add(-10 * 24 * time.Hour),
			UpdatedAt:   time.Now().Add(-10 * 24 * time.Hour),
		}

		if err := db.Create(&joinRequest).Error; err != nil {
			return fmt.Errorf("failed to seed join request: %w", err)
		}

		log.Println("Created community join request")
	}

	return nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}