package main

import (
	"log"
	"net"
	"os"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	handlers "aycom/backend/services/thread/api"
	"aycom/backend/services/thread/db"
	"aycom/backend/services/thread/proto"
	"aycom/backend/services/thread/repository"
	"aycom/backend/services/thread/service"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found or cannot be loaded: %v", err)
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()

		// Initialize thread service
		initThreadService()
	}()

	wg.Wait()
}

// initThreadService initializes and starts the thread gRPC service
func initThreadService() {
	port := getEnv("PORT", "9092")
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// Initialize database connection
	dbConn := db.InitDB()

	// Run database migrations
	if err := db.RunMigrations(dbConn); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Seed database if in development mode
	environment := getEnv("ENVIRONMENT", "development")
	if environment == "development" {
		if err := db.SeedDatabase(dbConn); err != nil {
			log.Printf("Warning: Failed to seed database: %v", err)
		}
	}

	// Initialize repositories using the new structure
	threadRepo := repository.NewThreadRepository(dbConn)
	mediaRepo := repository.NewMediaRepository(dbConn)
	hashtagRepo := repository.NewHashtagRepository(dbConn)
	replyRepo := repository.NewReplyRepository(dbConn)
	interactionRepo := repository.NewInteractionRepository(dbConn)
	pollRepo := repository.NewPollRepository(dbConn)

	// Initialize user client for fetching user data
	userClient := connectToUserService()

	// Initialize services with the new repositories
	threadService := service.NewThreadService(threadRepo, mediaRepo, hashtagRepo, replyRepo)
	replyService := service.NewReplyService(replyRepo, threadRepo, mediaRepo)
	interactionService := service.NewInteractionService(interactionRepo)
	pollService := service.NewPollService(pollRepo)

	// Initialize the gRPC handler
	handler := handlers.NewThreadHandler(
		threadService,
		replyService,
		interactionService,
		pollService,
		interactionRepo,
		userClient, // Pass the user client to the handler
	)

	// Configure gRPC server with potential TLS settings
	var opts []grpc.ServerOption
	tlsEnabled := getEnv("TLS_ENABLED", "false") == "true"
	if tlsEnabled {
		// TLS configuration would go here if needed
		log.Println("TLS is enabled but not configured. Please update the code to configure TLS.")
	}

	grpcServer := grpc.NewServer(opts...)
	proto.RegisterThreadServiceServer(grpcServer, handler)
	log.Printf("Thread service started on port %s, environment: %s", port, environment)
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

// connectToUserService establishes a connection to the user service with retries
func connectToUserService() service.UserClient {
	userServiceHost := getEnv("USER_SERVICE_HOST", "user_service")
	userServicePort := getEnv("USER_SERVICE_PORT", "9091")
	userServiceAddr := userServiceHost + ":" + userServicePort

	log.Printf("Connecting to User service at %s", userServiceAddr)

	// Retry parameters
	maxRetries := 5
	retryDelay := 2 * time.Second

	// Try to connect with retries
	var userConn *grpc.ClientConn
	var err error
	var userClient service.UserClient

	for i := 0; i < maxRetries; i++ {
		log.Printf("Attempting to connect to User service at %s (attempt %d/%d)", userServiceAddr, i+1, maxRetries)

		userConn, err = grpc.Dial(
			userServiceAddr,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithBlock(),
			grpc.WithTimeout(5*time.Second),
		)

		if err == nil {
			log.Printf("Successfully connected to User service at %s", userServiceAddr)
			userClient = service.NewUserClient(userConn)
			return userClient
		}

		log.Printf("Failed to connect to User service (attempt %d/%d): %v", i+1, maxRetries, err)

		if i < maxRetries-1 {
			log.Printf("Retrying in %v...", retryDelay)
			time.Sleep(retryDelay)
		}
	}

	log.Printf("CRITICAL: Could not connect to User service after %d attempts: %v", maxRetries, err)
	log.Printf("Thread service will not be able to retrieve real user data!")
	return nil
}

// getEnv retrieves an environment variable value or returns a default if not set
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
