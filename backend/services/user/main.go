package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	// "gorm.io/gorm/logger" // Commented out logger import

	"aycom/backend/proto/user"
	handlers "aycom/backend/services/user/api"
	"aycom/backend/services/user/db"
	"aycom/backend/services/user/model"
	"aycom/backend/services/user/service"
)

func main() {
	// Load environment variables
	if err := loadEnv(); err != nil {
		log.Printf("Warning: %v", err)
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()

		// Get service port from environment with fallback
		port := os.Getenv("PORT")
		if port == "" {
			port = "9091"
		}

		// Set up gRPC listener
		listener, err := net.Listen("tcp", ":"+port)
		if err != nil {
			log.Fatalf("Failed to listen: %v", err)
		}

		// Configure database connection
		dbConn, err := setupDatabase()
		if err != nil {
			log.Fatalf("Failed to set up database: %v", err)
		}

		// Get SQL DB connection for connection pooling settings
		sqlDB, err := dbConn.DB()
		if err != nil {
			log.Fatalf("Failed to get database connection: %v", err)
		}

		// Configure connection pool
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetMaxOpenConns(100)
		sqlDB.SetConnMaxLifetime(time.Hour)

		// Run migrations
		if err := runMigrations(dbConn); err != nil {
			log.Fatalf("Failed to run migrations: %v", err)
		}

		// Initialize repository, service, and handler
		repo := db.NewPostgresUserRepository(dbConn)
		svc := service.NewUserService(repo)
		handler := handlers.NewUserHandler(svc)

		// Configure gRPC server with keep-alive settings
		grpcServer := grpc.NewServer(
			grpc.KeepaliveParams(keepalive.ServerParameters{
				MaxConnectionIdle:     15 * time.Minute,
				MaxConnectionAge:      30 * time.Minute,
				MaxConnectionAgeGrace: 5 * time.Minute,
				Time:                  5 * time.Minute,
				Timeout:               20 * time.Second,
			}),
		)

		// Register services
		user.RegisterUserServiceServer(grpcServer, handler)

		// Set up health check
		healthServer := health.NewServer()
		healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
		grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)

		log.Printf("User service started on port %s", port)

		// Start gRPC server
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	wg.Wait()
}

// loadEnv loads environment variables from .env file
func loadEnv() error {
	// Try loading from current directory first
	if err := godotenv.Load(); err == nil {
		log.Println("Loaded .env file from current directory")
		return nil
	}

	// Try loading from project root
	if err := godotenv.Load("../../.env"); err == nil {
		log.Println("Loaded .env file from project root")
		return nil
	}

	return fmt.Errorf("no .env file found, using environment variables")
}

// setupDatabase establishes a connection to the database
func setupDatabase() (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		getEnv("DATABASE_HOST", "user_db"),
		getEnv("DATABASE_PORT", "5432"),
		getEnv("DATABASE_USER", "kolin"),
		getEnv("DATABASE_PASSWORD", "kolin"),
		getEnv("DATABASE_NAME", "user_db"),
	)

	// Configure GORM logger
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	// Try to connect with retry mechanism
	var dbConn *gorm.DB
	var err error

	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		dbConn, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: newLogger,
		})

		if err == nil {
			log.Println("Successfully connected to database")
			return dbConn, nil
		}

		retryDelay := time.Duration(i+1) * time.Second
		log.Printf("Failed to connect to database (attempt %d/%d): %v. Retrying in %v...",
			i+1, maxRetries, err, retryDelay)
		time.Sleep(retryDelay)
	}

	return nil, fmt.Errorf("failed to connect to database after %d attempts: %w", maxRetries, err)
}

// runMigrations applies database migrations
func runMigrations(dbConn *gorm.DB) error {
	log.Println("Running database migrations")

	// Run auto-migrate for core models
	if err := dbConn.AutoMigrate(&model.User{}, &model.Session{}); err != nil {
		return fmt.Errorf("failed to run auto-migrate: %w", err)
	}

	log.Println("Database migrations completed successfully")
	return nil
}

// getEnv gets an environment variable with a fallback value
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
