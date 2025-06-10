package main

import (
	"aycom/backend/proto/thread"
	"log"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	handlers "aycom/backend/services/thread/api"
	"aycom/backend/services/thread/db"
	"aycom/backend/services/thread/repository"
	"aycom/backend/services/thread/service"
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

		initThreadService()
	}()

	wg.Wait()
}

func initThreadService() {
	port := "9092"
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	dbConn := db.InitDB()

	if err := db.RunMigrations(dbConn); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	environment := getEnv("ENVIRONMENT", "development")
	if environment == "development" {
		if err := db.SeedDatabase(dbConn); err != nil {
			log.Printf("Warning: Failed to seed database: %v", err)
		}
	}

	threadRepo := repository.NewThreadRepository(dbConn)
	mediaRepo := repository.NewMediaRepository(dbConn)
	hashtagRepo := repository.NewHashtagRepository(dbConn)
	replyRepo := repository.NewReplyRepository(dbConn)
	interactionRepo := repository.NewInteractionRepository(dbConn)
	pollRepo := repository.NewPollRepository(dbConn)

	userClient := connectToUserService()

	// Create a user repository adapter using the user client
	userRepo := repository.NewUserRepositoryAdapter(userClient)

	// Create the user relation service for checking reply permissions
	userRelationSvc := service.NewUserRelationService(userClient)

	threadService := service.NewThreadService(threadRepo, mediaRepo, hashtagRepo, replyRepo)
	replyService := service.NewReplyService(replyRepo, threadRepo, mediaRepo, userRelationSvc)
	interactionService := service.NewInteractionService(interactionRepo, threadRepo, userRepo)
	pollService := service.NewPollService(pollRepo)

	handler := handlers.NewThreadHandler(
		threadService,
		replyService,
		interactionService,
		pollService,
		interactionRepo,
		userClient,
		hashtagRepo,
		threadRepo,
		mediaRepo,
	)

	var opts []grpc.ServerOption
	tlsEnabled := getEnv("TLS_ENABLED", "false") == "true"
	if tlsEnabled {
		log.Println("TLS is enabled but not configured. Please update the code to configure TLS.")
	}

	grpcServer := grpc.NewServer(opts...)
	thread.RegisterThreadServiceServer(grpcServer, handler)
	log.Printf("Thread service started on port %s, environment: %s", port, environment)
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

func connectToUserService() service.UserClient {
	userServiceHost := getEnv("USER_SERVICE_HOST", "user_service")
	userServicePort := getEnv("USER_SERVICE_PORT", "9091")
	userServiceAddr := userServiceHost + ":" + userServicePort

	log.Printf("Connecting to User service at %s", userServiceAddr)

	maxRetries := 5
	retryDelay := 2 * time.Second

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

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
