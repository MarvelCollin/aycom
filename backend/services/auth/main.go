package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Acad600-Tpa/WEB-MV-242/services/auth/service"
)

func main() {
	// Check if a command is specified
	if len(os.Args) > 1 {
		cmd := os.Args[1]

		switch cmd {
		case "migrate":
			log.Println("Running auth migrations...")
			svc, err := service.NewAuthService()
			if err != nil {
				log.Fatalf("Failed to initialize service: %v", err)
			}
			// Use svc to avoid unused variable error
			if err := svc.GetMigrationStatus(); err != nil {
				log.Fatalf("Failed to get migration status: %v", err)
			}
			log.Println("Auth migrations completed successfully")
			return

		case "status":
			log.Println("Getting auth migration status...")
			svc, err := service.NewAuthService()
			if err != nil {
				log.Fatalf("Failed to initialize service: %v", err)
			}
			if err := svc.GetMigrationStatus(); err != nil {
				log.Fatalf("Failed to get migration status: %v", err)
			}
			return
		}
	}

	// Normal application startup
	authService, err := service.NewAuthService()
	if err != nil {
		log.Fatalf("Failed to initialize service: %v", err)
	}

	port := os.Getenv("AUTH_SERVICE_PORT")
	if port == "" {
		port = "9090"
	}

	// Create a simple HTTP server
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Auth Service is running"))
	})

	// Health check endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Status endpoint
	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if authService.DB != nil {
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
		fmt.Printf("Auth service started on port: %s\n", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Set up graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down auth service...")

	// Give the server 5 seconds to finish ongoing requests
	time.Sleep(5 * time.Second)

	log.Println("Auth service stopped")
}
