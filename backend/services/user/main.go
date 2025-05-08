package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/status"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"aycom/backend/proto/user"
	handlers "aycom/backend/services/user/api"
	"aycom/backend/services/user/db"
	"aycom/backend/services/user/model"
	"aycom/backend/services/user/service"
)

func main() {
	if err := loadEnv(); err != nil {
		log.Printf("Warning: %v", err)
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()

		port := os.Getenv("PORT")
		if port == "" {
			port = "9091"
		}

		listener, err := net.Listen("tcp", ":"+port)
		if err != nil {
			log.Fatalf("Failed to listen: %v", err)
		}

		dbConn, err := setupDatabase()
		if err != nil {
			log.Fatalf("Failed to set up database: %v", err)
		}

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

		// Create adapter instance
		adapter := &userServiceAdapter{h: handler}

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

		// Register our adapter as the service implementation
		user.RegisterUserServiceServer(grpcServer, adapter)

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

	// Try loading from project root (direct parent directories method)
	if err := godotenv.Load("../../.env"); err == nil {
		log.Println("Loaded .env file from project root (../../.env)")
		return nil
	}

	// Try another path to root (going up three levels)
	if err := godotenv.Load("../../../.env"); err == nil {
		log.Println("Loaded .env file from project root (../../../.env)")
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

// Create an adapter that implements only the methods we need
// and uses the UnimplementedUserServiceServer for the rest
type userServiceAdapter struct {
	user.UnimplementedUserServiceServer
	h *handlers.UserHandler
}

// Forward implemented methods to our actual handler
func (a *userServiceAdapter) GetUser(ctx context.Context, req *user.GetUserRequest) (*user.GetUserResponse, error) {
	return a.h.GetUser(ctx, req)
}

func (a *userServiceAdapter) CreateUser(ctx context.Context, req *user.CreateUserRequest) (*user.CreateUserResponse, error) {
	return a.h.CreateUser(ctx, req)
}

func (a *userServiceAdapter) UpdateUser(ctx context.Context, req *user.UpdateUserRequest) (*user.UpdateUserResponse, error) {
	return a.h.UpdateUser(ctx, req)
}

func (a *userServiceAdapter) DeleteUser(ctx context.Context, req *user.DeleteUserRequest) (*user.DeleteUserResponse, error) {
	return a.h.DeleteUser(ctx, req)
}

func (a *userServiceAdapter) UpdateUserVerificationStatus(ctx context.Context, req *user.UpdateUserVerificationStatusRequest) (*user.UpdateUserVerificationStatusResponse, error) {
	return a.h.UpdateUserVerificationStatus(ctx, req)
}

func (a *userServiceAdapter) LoginUser(ctx context.Context, req *user.LoginUserRequest) (*user.LoginUserResponse, error) {
	return a.h.LoginUser(ctx, req)
}

func (a *userServiceAdapter) GetUserByEmail(ctx context.Context, req *user.GetUserByEmailRequest) (*user.GetUserByEmailResponse, error) {
	return a.h.GetUserByEmail(ctx, req)
}

// SearchUsers - Implementation matching the UserServiceServer interface
func (a *userServiceAdapter) SearchUsers(ctx context.Context, req *user.SearchUsersRequest) (*user.SearchUsersResponse, error) {
	if req.Query == "" {
		return nil, status.Error(codes.InvalidArgument, "Search query is required")
	}

	// Set default pagination values
	page := int(req.Page)
	if page <= 0 {
		page = 1
	}

	limit := int(req.Limit)
	if limit <= 0 {
		limit = 10
	}

	// Create model request
	searchReq := &model.SearchUsersRequest{
		Query:  req.Query,
		Filter: req.Filter,
		Page:   page,
		Limit:  limit,
	}

	// Get service from handler
	svc := a.h.GetService()

	// Call service layer
	users, totalCount, err := svc.SearchUsers(ctx, searchReq)
	if err != nil {
		return nil, err
	}

	// Map results
	protoUsers := make([]*user.User, 0, len(users))
	for _, u := range users {
		protoUser := &user.User{
			Id:                u.ID.String(),
			Name:              u.Name,
			Username:          u.Username,
			Email:             u.Email,
			Gender:            u.Gender,
			ProfilePictureUrl: u.ProfilePictureURL,
			BannerUrl:         u.BannerURL,
		}

		// Handle time fields
		if u.DateOfBirth != nil {
			protoUser.DateOfBirth = u.DateOfBirth.Format("2006-01-02")
		}
		if !u.CreatedAt.IsZero() {
			protoUser.CreatedAt = u.CreatedAt.Format(time.RFC3339)
		}
		if !u.UpdatedAt.IsZero() {
			protoUser.UpdatedAt = u.UpdatedAt.Format(time.RFC3339)
		}

		protoUsers = append(protoUsers, protoUser)
	}

	// Return the proper response type as required by the interface
	return &user.SearchUsersResponse{
		Users:      protoUsers,
		TotalCount: int32(totalCount),
	}, nil
}

// Create adapter instance
