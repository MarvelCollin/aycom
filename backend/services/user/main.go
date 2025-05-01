package main

import (
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/Acad600-Tpa/WEB-MV-242/backend/services/user/internal/handlers"
	"github.com/Acad600-Tpa/WEB-MV-242/backend/services/user/model"
	"github.com/Acad600-Tpa/WEB-MV-242/backend/services/user/pkg/db"
	pb "github.com/Acad600-Tpa/WEB-MV-242/backend/services/user/proto"
	"github.com/Acad600-Tpa/WEB-MV-242/backend/services/user/repository"
	"github.com/Acad600-Tpa/WEB-MV-242/backend/services/user/service"
	"google.golang.org/grpc"
)

func main() {
	// Setup command line flags
	skipServer := flag.Bool("skip-server", false, "Skip starting the server after running migrations/seeds")
	flag.Parse()

	// Get the command (if any)
	var command string
	if flag.NArg() > 0 {
		command = flag.Arg(0)
	}

	// Database connection
	dbConn, err := db.ConnectDatabaseWithRetry()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Println("Successfully connected to database")

	// Auto-migrate database tables
	log.Println("Running database migrations...")
	if err := dbConn.AutoMigrate(
		&model.User{},
		&model.UserProfile{},
		&model.Contact{},
	); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
	log.Println("Database migration completed")

	// Handle specific commands
	switch command {
	case "migrate":
		if *skipServer {
			log.Println("Server startup skipped due to --skip-server flag")
			os.Exit(0)
		}

	case "seed":
		log.Println("Seeding user data...")
		seeder := repository.NewUserSeeder(dbConn)
		if err := seeder.SeedUsers(); err != nil {
			log.Fatalf("Failed to seed users: %v", err)
		}
		log.Println("Seeding completed.")

		if *skipServer {
			log.Println("Server startup skipped due to --skip-server flag")
			os.Exit(0)
		}

	case "status":
		// Check if database is properly migrated and print status
		log.Println("Checking database migration status...")

		// Get all tables in the database
		var tables []string
		dbConn.Raw("SELECT tablename FROM pg_catalog.pg_tables WHERE schemaname = 'public'").Scan(&tables)

		log.Println("Database tables found:")
		for _, table := range tables {
			var count int64
			dbConn.Table(table).Count(&count)
			log.Printf("- %s (%d rows)", table, count)
		}

		log.Println("Database status check completed")
		return
	}

	// If --skip-server flag is set, don't start the server
	if *skipServer {
		log.Println("Server startup skipped due to --skip-server flag")
		os.Exit(0)
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(dbConn)

	// Initialize services
	userService := service.NewUserService(userRepo)

	// Initialize handler
	userHandler := handlers.NewUserHandler(userService)

	// Setup gRPC server
	port := os.Getenv("USER_SERVICE_PORT")
	if port == "" {
		port = "9091" // Default port
	}

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", port, err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterUserServiceServer(grpcServer, userHandler)

	log.Printf("User Service starting on :%s", port)

	// Start the gRPC server in a goroutine
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// Set up graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down user service...")
	grpcServer.GracefulStop()
	log.Println("User service stopped.")
}
