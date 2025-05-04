package main

import (
	"log"
	"net"
	"net/http"
	"os"
	"sync"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"

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
	wg.Add(2)

	go func() {
		defer wg.Done()
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
		userServiceHost := getEnv("USER_SERVICE_HOST", "user_service")
		userServicePort := getEnv("USER_SERVICE_PORT", "9091")
		userServiceAddr := userServiceHost + ":" + userServicePort

		log.Printf("User service is at %s, but using mock client for now", userServiceAddr)
		// TODO: Fix proto imports to establish real connection
		userClient := service.NewUserClient(nil)

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
	}()

	// Health check endpoint
	go func() {
		defer wg.Done()
		healthPort := getEnv("HEALTH_PORT", "8082")
		http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", getEnv("CORS_ORIGIN", "http://localhost:3000"))
			w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})
		log.Printf("Thread service health endpoint started on port %s", healthPort)
		if err := http.ListenAndServe(":"+healthPort, nil); err != nil {
			log.Fatalf("Failed to start health server: %v", err)
		}
	}()

	wg.Wait()
}

// getEnv retrieves an environment variable value or returns a default if not set
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
