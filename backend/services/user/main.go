package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Acad600-Tpa/WEB-MV-242/backend/services/user/repository"
	"github.com/Acad600-Tpa/WEB-MV-242/backend/services/user/service"
)

func main() {
	// Check if a command is specified
	if len(os.Args) > 1 {
		cmd := os.Args[1]

		switch cmd {
		case "migrate":
			log.Println("Running migrations...")
			svc, err := service.NewUserService()
			if err != nil {
				log.Fatalf("Failed to initialize service: %v", err)
			}
			// Use svc to avoid unused variable error
			if err := svc.GetMigrationStatus(); err != nil {
				log.Fatalf("Failed to get migration status: %v", err)
			}
			log.Println("Migrations completed successfully")

			// Seed default users after migration
			if err := seedDefaultUsers(svc); err != nil {
				log.Fatalf("Failed to seed default users: %v", err)
			}

			return

		case "status":
			log.Println("Getting migration status...")
			svc, err := service.NewUserService()
			if err != nil {
				log.Fatalf("Failed to initialize service: %v", err)
			}
			if err := svc.GetMigrationStatus(); err != nil {
				log.Fatalf("Failed to get migration status: %v", err)
			}
			return
		case "seed":
			log.Println("Seeding default users...")
			svc, err := service.NewUserService()
			if err != nil {
				log.Fatalf("Failed to initialize service: %v", err)
			}
			if err := seedDefaultUsers(svc); err != nil {
				log.Fatalf("Failed to seed default users: %v", err)
			}
			log.Println("User seeding completed successfully")
			return
		}
	}

	// Normal application startup
	userService, err := service.NewUserService()
	if err != nil {
		log.Fatalf("Failed to initialize service: %v", err)
	}

	// Seed default users on startup
	if err := seedDefaultUsers(userService); err != nil {
		log.Printf("Warning: Failed to seed default users: %v", err)
	}

	port := os.Getenv("USER_SERVICE_PORT")
	if port == "" {
		port = "9091"
	}

	// Create a simple HTTP server
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("User Service is running"))
	})

	// Health check endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Example endpoint using the user service (to avoid unused variable warning)
	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if userService.DB != nil {
			w.Write([]byte("DB Connected"))
		} else {
			w.Write([]byte("DB Not Connected"))
		}
	})

	// Start server in a goroutine
	server := &http.Server{
		Addr:    ":" + port,
		Handler: http.DefaultServeMux,
	}

	go func() {
		fmt.Printf("User service started on port: %s\n", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Set up graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down user service...")

	// Give the server 5 seconds to finish ongoing requests
	time.Sleep(5 * time.Second)

	log.Println("User service stopped")
}

// seedDefaultUsers initializes the user seeder and seeds default users
func seedDefaultUsers(svc *service.UserServiceImpl) error {
	if svc.DB == nil {
		return fmt.Errorf("database connection not available")
	}

	// Initialize the user seeder
	seeder := repository.NewUserSeeder(svc.DB)

	// Seed default users
	return seeder.SeedUsers()
}
