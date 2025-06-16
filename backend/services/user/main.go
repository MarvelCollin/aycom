package main

import (
	"aycom/backend/proto/user"
	"context"
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

	handlers "aycom/backend/services/user/api"
	"aycom/backend/services/user/db"
	"aycom/backend/services/user/repository"
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

		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetMaxOpenConns(100)
		sqlDB.SetConnMaxLifetime(time.Hour)

		if err := runMigrations(dbConn); err != nil {
			log.Fatalf("Failed to run migrations: %v", err)
		}

		userRepo := repository.NewPostgresUserRepository(dbConn)
		adminRepo := repository.NewAdminRepository(dbConn)
		blockRepo := repository.NewBlockRepository(dbConn)
		reportRepo := repository.NewReportRepository(dbConn)

		baseUserSvc := service.NewUserService(userRepo)
		adminSvc := service.NewAdminService(adminRepo, userRepo)
		blockSvc := service.NewUserBlockService(blockRepo, reportRepo, userRepo)

		userSvc := service.NewCombinedService(baseUserSvc, blockSvc)

		handler := handlers.NewUserHandler(userSvc, adminSvc)

		adapter := &userServiceAdapter{h: handler}

		grpcServer := grpc.NewServer(
			grpc.KeepaliveParams(keepalive.ServerParameters{
				MaxConnectionIdle:     15 * time.Minute,
				MaxConnectionAge:      30 * time.Minute,
				MaxConnectionAgeGrace: 5 * time.Minute,
				Time:                  5 * time.Minute,
				Timeout:               20 * time.Second,
			}),
		)

		user.RegisterUserServiceServer(grpcServer, adapter)

		healthServer := health.NewServer()
		healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
		grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)

		log.Printf("User service started on port %s", port)

		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	wg.Wait()
}

func loadEnv() error {
	if err := godotenv.Load(); err == nil {
		log.Println("Loaded .env file from current directory")
		return nil
	}

	if err := godotenv.Load("../../.env"); err == nil {
		log.Println("Loaded .env file from project root (../../.env)")
		return nil
	}

	if err := godotenv.Load("../../../.env"); err == nil {
		log.Println("Loaded .env file from project root (../../../.env)")
		return nil
	}

	return fmt.Errorf("no .env file found, using environment variables")
}

func setupDatabase() (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		getEnv("DATABASE_HOST", "user_db"),
		getEnv("DATABASE_PORT", "5432"),
		getEnv("DATABASE_USER", "kolin"),
		getEnv("DATABASE_PASSWORD", "kolin"),
		getEnv("DATABASE_NAME", "user_db"),
	)

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

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

func runMigrations(dbConn *gorm.DB) error {
	log.Println("Running database migrations")

	if err := db.Migrate(dbConn); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Database migrations completed successfully")
	return nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

type userServiceAdapter struct {
	user.UnimplementedUserServiceServer
	h *handlers.UserHandler
}

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

func (a *userServiceAdapter) GetUserByUsername(ctx context.Context, req *user.GetUserByUsernameRequest) (*user.GetUserByUsernameResponse, error) {
	return a.h.GetUserByUsername(ctx, req)
}

func (a *userServiceAdapter) IsUserBlocked(ctx context.Context, req *user.IsUserBlockedRequest) (*user.IsUserBlockedResponse, error) {
	return a.h.IsUserBlocked(ctx, req)
}

func (a *userServiceAdapter) IsFollowing(ctx context.Context, req *user.IsFollowingRequest) (*user.IsFollowingResponse, error) {
	return a.h.IsFollowing(ctx, req)
}

func (a *userServiceAdapter) FollowUser(ctx context.Context, req *user.FollowUserRequest) (*user.FollowUserResponse, error) {
	return a.h.FollowUser(ctx, req)
}

func (a *userServiceAdapter) UnfollowUser(ctx context.Context, req *user.UnfollowUserRequest) (*user.UnfollowUserResponse, error) {
	return a.h.UnfollowUser(ctx, req)
}

func (a *userServiceAdapter) GetFollowers(ctx context.Context, req *user.GetFollowersRequest) (*user.GetFollowersResponse, error) {
	return a.h.GetFollowers(ctx, req)
}

func (a *userServiceAdapter) GetFollowing(ctx context.Context, req *user.GetFollowingRequest) (*user.GetFollowingResponse, error) {
	return a.h.GetFollowing(ctx, req)
}

func (a *userServiceAdapter) SearchUsers(ctx context.Context, req *user.SearchUsersRequest) (*user.SearchUsersResponse, error) {
	return a.h.SearchUsers(ctx, req)
}

func (a *userServiceAdapter) GetAllUsers(ctx context.Context, req *user.GetAllUsersRequest) (*user.GetAllUsersResponse, error) {
	return a.h.GetAllUsers(ctx, req)
}

func (a *userServiceAdapter) GetRecommendedUsers(ctx context.Context, req *user.GetRecommendedUsersRequest) (*user.GetRecommendedUsersResponse, error) {
	return a.h.GetRecommendedUsers(ctx, req)
}

func (a *userServiceAdapter) BanUser(ctx context.Context, req *user.BanUserRequest) (*user.BanUserResponse, error) {
	return a.h.BanUser(ctx, req)
}

func (a *userServiceAdapter) SendNewsletter(ctx context.Context, req *user.SendNewsletterRequest) (*user.SendNewsletterResponse, error) {
	return a.h.SendNewsletter(ctx, req)
}

func (a *userServiceAdapter) GetCommunityRequests(ctx context.Context, req *user.GetCommunityRequestsRequest) (*user.GetCommunityRequestsResponse, error) {
	return a.h.GetCommunityRequests(ctx, req)
}

func (a *userServiceAdapter) ProcessCommunityRequest(ctx context.Context, req *user.ProcessCommunityRequestRequest) (*user.ProcessCommunityRequestResponse, error) {
	return a.h.ProcessCommunityRequest(ctx, req)
}

func (a *userServiceAdapter) GetPremiumRequests(ctx context.Context, req *user.GetPremiumRequestsRequest) (*user.GetPremiumRequestsResponse, error) {
	return a.h.GetPremiumRequests(ctx, req)
}

func (a *userServiceAdapter) ProcessPremiumRequest(ctx context.Context, req *user.ProcessPremiumRequestRequest) (*user.ProcessPremiumRequestResponse, error) {
	return a.h.ProcessPremiumRequest(ctx, req)
}

func (a *userServiceAdapter) GetReportRequests(ctx context.Context, req *user.GetReportRequestsRequest) (*user.GetReportRequestsResponse, error) {
	return a.h.GetReportRequests(ctx, req)
}

func (a *userServiceAdapter) ProcessReportRequest(ctx context.Context, req *user.ProcessReportRequestRequest) (*user.ProcessReportRequestResponse, error) {
	return a.h.ProcessReportRequest(ctx, req)
}

func (a *userServiceAdapter) GetThreadCategories(ctx context.Context, req *user.GetThreadCategoriesRequest) (*user.GetThreadCategoriesResponse, error) {
	return a.h.GetThreadCategories(ctx, req)
}

func (a *userServiceAdapter) CreateThreadCategory(ctx context.Context, req *user.CreateThreadCategoryRequest) (*user.CreateThreadCategoryResponse, error) {
	return a.h.CreateThreadCategory(ctx, req)
}

func (a *userServiceAdapter) UpdateThreadCategory(ctx context.Context, req *user.UpdateThreadCategoryRequest) (*user.UpdateThreadCategoryResponse, error) {
	return a.h.UpdateThreadCategory(ctx, req)
}

func (a *userServiceAdapter) DeleteThreadCategory(ctx context.Context, req *user.DeleteThreadCategoryRequest) (*user.DeleteThreadCategoryResponse, error) {
	return a.h.DeleteThreadCategory(ctx, req)
}

func (a *userServiceAdapter) GetCommunityCategories(ctx context.Context, req *user.GetCommunityCategoriesRequest) (*user.GetCommunityCategoriesResponse, error) {
	return a.h.GetCommunityCategories(ctx, req)
}

func (a *userServiceAdapter) CreateCommunityCategory(ctx context.Context, req *user.CreateCommunityCategoryRequest) (*user.CreateCommunityCategoryResponse, error) {
	return a.h.CreateCommunityCategory(ctx, req)
}

func (a *userServiceAdapter) UpdateCommunityCategory(ctx context.Context, req *user.UpdateCommunityCategoryRequest) (*user.UpdateCommunityCategoryResponse, error) {
	return a.h.UpdateCommunityCategory(ctx, req)
}

func (a *userServiceAdapter) DeleteCommunityCategory(ctx context.Context, req *user.DeleteCommunityCategoryRequest) (*user.DeleteCommunityCategoryResponse, error) {
	return a.h.DeleteCommunityCategory(ctx, req)
}

func (a *userServiceAdapter) CreatePremiumRequest(ctx context.Context, req *user.CreatePremiumRequestRequest) (*user.CreatePremiumRequestResponse, error) {
	return a.h.CreatePremiumRequest(ctx, req)
}
