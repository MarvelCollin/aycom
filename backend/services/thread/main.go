package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"thread-service/internal/handlers"
	"thread-service/model"
	"thread-service/pkg/db"
	"thread-service/proto"
	"thread-service/repository"
	"thread-service/service"

	"google.golang.org/grpc"
	"gorm.io/gorm"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Database connection
	dbConn, err := db.ConnectDatabaseWithRetry()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Println("Successfully connected to database")

	// Run migrations
	if err := migrateModels(dbConn); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
	log.Println("Database migration completed")

	// Dependency Injection
	threadRepo := repository.NewThreadRepository(dbConn)
	replyRepo := repository.NewReplyRepository(dbConn)
	likeRepo := repository.NewLikeRepository(dbConn)
	repostRepo := repository.NewRepostRepository(dbConn)
	bookmarkRepo := repository.NewBookmarkRepository(dbConn)
	mediaRepo := repository.NewMediaRepository(dbConn)
	hashtagRepo := repository.NewHashtagRepository(dbConn)
	mentionRepo := repository.NewMentionRepository(dbConn)
	pollRepo := repository.NewPollRepository(dbConn)

	// Initialize services
	threadService := service.NewThreadService(
		threadRepo,
		mediaRepo,
		hashtagRepo,
		mentionRepo,
	)

	replyService := service.NewReplyService(
		replyRepo,
		threadRepo,
		mediaRepo,
		mentionRepo,
	)

	interactionService := service.NewInteractionService(
		likeRepo,
		repostRepo,
		bookmarkRepo,
	)

	pollService := service.NewPollService(pollRepo)

	// Initialize handler
	threadHandler := handlers.NewThreadHandler(
		threadService,
		replyService,
		interactionService,
		pollService,
	)

	// Handle seed command
	if len(os.Args) > 1 && os.Args[1] == "seed" {
		log.Println("Seeding thread data...")
		seeder := repository.NewThreadSeeder(dbConn)
		if err := seeder.SeedThreads(); err != nil {
			log.Fatalf("Failed to seed threads: %v", err)
		}
		log.Println("Seeding completed.")
		return
	}

	// Setup gRPC server
	grpcPort := os.Getenv("THREAD_SERVICE_PORT")
	if grpcPort == "" {
		grpcPort = "9092" // Default gRPC port
	}

	listener, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", grpcPort, err)
	}

	grpcServer := grpc.NewServer()
	proto.RegisterThreadServiceServer(grpcServer, threadHandler)

	log.Printf("Thread gRPC server starting on port %s...", grpcPort)

	// Start gRPC server in a goroutine
	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			log.Printf("gRPC server failed to serve: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down thread service...")

	// Gracefully stop the gRPC server
	grpcServer.GracefulStop()
	log.Println("gRPC server stopped.")

	// Close database connection
	if sqlDB, err := dbConn.DB(); err == nil {
		sqlDB.Close()
		log.Println("Database connection closed.")
	}

	log.Println("Thread service stopped gracefully.")
}

// migrateModels ensures all database tables are created and migrated
func migrateModels(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.Thread{},
		&model.Reply{},
		&model.Media{},
		&model.Like{},
		&model.Repost{},
		&model.Bookmark{},
		&model.Hashtag{},
		&model.ThreadHashtag{},
		&model.UserMention{},
		&model.Poll{},
		&model.PollOption{},
		&model.PollVote{},
		&model.Category{},
		&model.ThreadCategory{},
	)
}
