package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Acad600-Tpa/WEB-MV-242/services/user/service"
)

func main() {
	// Check if migration command is specified
	if len(os.Args) > 1 && os.Args[1] == "migrate" {
		log.Println("Running migrations...")
		_, err := service.NewUserService()
		if err != nil {
			log.Fatalf("Failed to initialize service: %v", err)
		}
		log.Println("Migrations completed successfully")
		return
	}

	// Normal application startup
	userService, err := service.NewUserService()
	if err != nil {
		log.Fatalf("Failed to initialize service: %v", err)
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
