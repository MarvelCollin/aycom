package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"sync"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	// "gorm.io/gorm/logger" // Commented out logger import

	handlers "aycom/backend/services/user/api"
	"aycom/backend/services/user/db"
	"aycom/backend/services/user/model"
	"aycom/backend/services/user/proto"
	"aycom/backend/services/user/service"
)

func main() {
	_ = godotenv.Load(".env")
	var wg sync.WaitGroup
	wg.Add(2)

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
		dsn := fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			getEnv("DATABASE_HOST", "user_db"),
			getEnv("DATABASE_PORT", "5432"),
			getEnv("DATABASE_USER", "kolin"),
			getEnv("DATABASE_PASSWORD", "kolin"),
			getEnv("DATABASE_NAME", "user_db"),
		)

		// Configure GORM logger - COMMENTED OUT
		// newLogger := logger.New(
		// 	log.New(os.Stdout, "\r\n", log.LstdFlags),
		// 	logger.Config{
		// 		SlowThreshold:             time.Second,
		// 		LogLevel:                  logger.Info,
		// 		IgnoreRecordNotFoundError: true,
		// 		Colorful:                  true,
		// 	},
		// )

		// Restore original GORM connection
		dbConn, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}

		// Run direct AutoMigrate for core models
		log.Println("Running direct AutoMigrate for User and Session models...")
		if err := dbConn.AutoMigrate(&model.User{}, &model.Session{}); err != nil {
			log.Fatalf("Failed to run direct AutoMigrate: %v", err)
		}
		log.Println("Direct AutoMigrate completed.")

		// Custom Migrations & Seeder - COMMENTED OUT
		// log.Println("Running database migrations...")
		// if err := db.Migrate(dbConn); err != nil { ... }
		// log.Println("Database migrations completed.")
		// if getEnv("RUN_SEEDER", "false") == "true" { ... }
		// log.Println("Skipping database seeding (Direct migration active).")

		repo := db.NewPostgresUserRepository(dbConn)
		svc := service.NewUserService(repo)
		handler := handlers.NewUserHandler(svc)
		grpcServer := grpc.NewServer()
		proto.RegisterUserServiceServer(grpcServer, handler)
		log.Printf("User service started on port %s", port)
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	go func() {
		defer wg.Done()
		healthPort := getEnv("HEALTH_PORT", "8081")
		http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			// Add CORS headers
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type")

			// Handle preflight OPTIONS request
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			// Ensure proper HTTP response
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			// Return JSON response
			_, err := w.Write([]byte(`{"status":"OK","service":"user_service"}`))
			if err != nil {
				log.Printf("Error writing health check response: %v", err)
			}
		})
		log.Printf("User service health endpoint started on port %s", healthPort)

		// Use a properly configured HTTP server
		server := &http.Server{
			Addr:    ":" + healthPort,
			Handler: nil, // Use default ServeMux
		}

		if err := server.ListenAndServe(); err != nil {
			log.Fatalf("Failed to start health endpoint: %v", err)
		}
	}()

	wg.Wait()
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
