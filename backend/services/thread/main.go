package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/Acad600-Tpa/WEB-MV-242/backend/services/thread/internal/handlers"
	"github.com/Acad600-Tpa/WEB-MV-242/backend/services/thread/pkg/db"
	"github.com/Acad600-Tpa/WEB-MV-242/backend/services/thread/proto/thread-service/proto"
	"github.com/Acad600-Tpa/WEB-MV-242/backend/services/thread/repository"
	"github.com/Acad600-Tpa/WEB-MV-242/backend/services/thread/service"

	"google.golang.org/grpc"
	"gorm.io/gorm"
)

func main() {
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
	log.Println("Starting database migration...")

	// Step 1: Create all tables without foreign key constraints
	tableMigrations := []string{
		// Create base tables first
		`CREATE TABLE IF NOT EXISTS threads (
			thread_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL,
			content TEXT NOT NULL,
			is_pinned BOOLEAN DEFAULT FALSE,
			who_can_reply VARCHAR(20) NOT NULL,
			scheduled_at TIMESTAMP WITH TIME ZONE,
			community_id UUID,
			is_advertisement BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
			deleted_at TIMESTAMP WITH TIME ZONE
		)`,

		`CREATE TABLE IF NOT EXISTS replies (
			reply_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			thread_id UUID NOT NULL,
			user_id UUID NOT NULL,
			content TEXT NOT NULL,
			is_pinned BOOLEAN DEFAULT FALSE,
			parent_reply_id UUID,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
			deleted_at TIMESTAMP WITH TIME ZONE
		)`,

		`CREATE TABLE IF NOT EXISTS media (
			media_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			thread_id UUID,
			reply_id UUID,
			type VARCHAR(10) NOT NULL,
			url VARCHAR(512) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
			deleted_at TIMESTAMP WITH TIME ZONE
		)`,

		`CREATE TABLE IF NOT EXISTS likes (
			user_id UUID NOT NULL,
			thread_id UUID,
			reply_id UUID,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
			deleted_at TIMESTAMP WITH TIME ZONE,
			PRIMARY KEY (user_id, thread_id, reply_id)
		)`,

		`CREATE TABLE IF NOT EXISTS reposts (
			user_id UUID NOT NULL,
			thread_id UUID NOT NULL,
			repost_text TEXT,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
			deleted_at TIMESTAMP WITH TIME ZONE,
			PRIMARY KEY (user_id, thread_id)
		)`,

		`CREATE TABLE IF NOT EXISTS bookmarks (
			user_id UUID NOT NULL,
			thread_id UUID NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
			deleted_at TIMESTAMP WITH TIME ZONE,
			PRIMARY KEY (user_id, thread_id)
		)`,

		`CREATE TABLE IF NOT EXISTS hashtags (
			hashtag_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			text VARCHAR(50) NOT NULL UNIQUE,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
			deleted_at TIMESTAMP WITH TIME ZONE
		)`,

		`CREATE TABLE IF NOT EXISTS thread_hashtags (
			thread_id UUID NOT NULL,
			hashtag_id UUID NOT NULL,
			deleted_at TIMESTAMP WITH TIME ZONE,
			PRIMARY KEY (thread_id, hashtag_id)
		)`,

		`CREATE TABLE IF NOT EXISTS user_mentions (
			mention_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			mentioned_user_id UUID NOT NULL,
			thread_id UUID,
			reply_id UUID,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
			deleted_at TIMESTAMP WITH TIME ZONE
		)`,

		`CREATE TABLE IF NOT EXISTS categories (
			category_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(50) NOT NULL,
			type VARCHAR(10) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
			deleted_at TIMESTAMP WITH TIME ZONE
		)`,

		`CREATE TABLE IF NOT EXISTS thread_categories (
			thread_id UUID NOT NULL,
			category_id UUID NOT NULL,
			deleted_at TIMESTAMP WITH TIME ZONE,
			PRIMARY KEY (thread_id, category_id)
		)`,

		`CREATE TABLE IF NOT EXISTS polls (
			poll_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			thread_id UUID NOT NULL UNIQUE,
			question TEXT NOT NULL,
			closes_at TIMESTAMP WITH TIME ZONE NOT NULL,
			who_can_vote VARCHAR(20) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
			deleted_at TIMESTAMP WITH TIME ZONE
		)`,

		`CREATE TABLE IF NOT EXISTS poll_options (
			option_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			poll_id UUID NOT NULL,
			text VARCHAR(100) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
			deleted_at TIMESTAMP WITH TIME ZONE
		)`,

		`CREATE TABLE IF NOT EXISTS poll_votes (
			vote_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			poll_id UUID NOT NULL,
			option_id UUID NOT NULL,
			user_id UUID NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
			deleted_at TIMESTAMP WITH TIME ZONE
		)`,
	}

	// Step 2: Add foreign key constraints after all tables are created
	constraintMigrations := []string{
		`ALTER TABLE replies ADD CONSTRAINT fk_threads_replies
			FOREIGN KEY (thread_id) REFERENCES threads(thread_id) ON DELETE CASCADE`,

		`ALTER TABLE replies ADD CONSTRAINT fk_replies_parent
			FOREIGN KEY (parent_reply_id) REFERENCES replies(reply_id) ON DELETE SET NULL`,

		`ALTER TABLE media ADD CONSTRAINT fk_threads_media
			FOREIGN KEY (thread_id) REFERENCES threads(thread_id) ON DELETE CASCADE`,

		`ALTER TABLE media ADD CONSTRAINT fk_replies_media
			FOREIGN KEY (reply_id) REFERENCES replies(reply_id) ON DELETE CASCADE`,

		`ALTER TABLE likes ADD CONSTRAINT fk_threads_likes
			FOREIGN KEY (thread_id) REFERENCES threads(thread_id) ON DELETE CASCADE`,

		`ALTER TABLE likes ADD CONSTRAINT fk_replies_likes
			FOREIGN KEY (reply_id) REFERENCES replies(reply_id) ON DELETE CASCADE`,

		`ALTER TABLE reposts ADD CONSTRAINT fk_threads_reposts
			FOREIGN KEY (thread_id) REFERENCES threads(thread_id) ON DELETE CASCADE`,

		`ALTER TABLE bookmarks ADD CONSTRAINT fk_threads_bookmarks
			FOREIGN KEY (thread_id) REFERENCES threads(thread_id) ON DELETE CASCADE`,

		`ALTER TABLE thread_hashtags ADD CONSTRAINT fk_threads_thread_hashtags
			FOREIGN KEY (thread_id) REFERENCES threads(thread_id) ON DELETE CASCADE`,

		`ALTER TABLE thread_hashtags ADD CONSTRAINT fk_hashtags_thread_hashtags
			FOREIGN KEY (hashtag_id) REFERENCES hashtags(hashtag_id) ON DELETE CASCADE`,

		`ALTER TABLE user_mentions ADD CONSTRAINT fk_threads_user_mentions
			FOREIGN KEY (thread_id) REFERENCES threads(thread_id) ON DELETE CASCADE`,

		`ALTER TABLE user_mentions ADD CONSTRAINT fk_replies_user_mentions
			FOREIGN KEY (reply_id) REFERENCES replies(reply_id) ON DELETE CASCADE`,

		`ALTER TABLE thread_categories ADD CONSTRAINT fk_threads_thread_categories
			FOREIGN KEY (thread_id) REFERENCES threads(thread_id) ON DELETE CASCADE`,

		`ALTER TABLE thread_categories ADD CONSTRAINT fk_categories_thread_categories
			FOREIGN KEY (category_id) REFERENCES categories(category_id) ON DELETE CASCADE`,

		`ALTER TABLE polls ADD CONSTRAINT fk_threads_polls
			FOREIGN KEY (thread_id) REFERENCES threads(thread_id) ON DELETE CASCADE`,

		`ALTER TABLE poll_options ADD CONSTRAINT fk_polls_poll_options
			FOREIGN KEY (poll_id) REFERENCES polls(poll_id) ON DELETE CASCADE`,

		`ALTER TABLE poll_votes ADD CONSTRAINT fk_polls_poll_votes
			FOREIGN KEY (poll_id) REFERENCES polls(poll_id) ON DELETE CASCADE`,

		`ALTER TABLE poll_votes ADD CONSTRAINT fk_poll_options_poll_votes
			FOREIGN KEY (option_id) REFERENCES poll_options(option_id) ON DELETE CASCADE`,
	}

	// Execute table creation migrations
	log.Println("Creating tables...")
	for _, migration := range tableMigrations {
		if err := executeMigration(db, migration); err != nil {
			return err
		}
	}

	// Execute constraint migrations
	log.Println("Adding foreign key constraints...")
	for _, migration := range constraintMigrations {
		if err := executeMigration(db, migration); err != nil {
			// If there's an error, log it but continue with the next constraint
			log.Printf("Warning: Failed to add constraint: %v", err)
		}
	}

	log.Println("Migration completed successfully")
	return nil
}

// executeMigration helper function to execute a single migration statement
func executeMigration(db *gorm.DB, migration string) error {
	tx := db.Begin()
	if err := tx.Exec(migration).Error; err != nil {
		tx.Rollback()
		// Skip errors for existing objects
		if strings.Contains(err.Error(), "already exists") ||
			strings.Contains(err.Error(), "duplicate key") ||
			strings.Contains(err.Error(), "relation") {
			log.Printf("Note: %v", err)
			return nil
		}
		return err
	}
	return tx.Commit().Error
}
