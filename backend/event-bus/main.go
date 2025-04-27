package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// Create a simple HTTP server for health checks
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Event Bus is running"))
	})

	// Health check endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Start server in a goroutine
	server := &http.Server{
		Addr:    ":8000", // Just for health checks
		Handler: http.DefaultServeMux,
	}

	go func() {
		fmt.Println("Event bus service started")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Set up graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down event bus...")
	
	// Give the server 5 seconds to finish ongoing requests
	time.Sleep(5 * time.Second)
	
	log.Println("Event bus stopped")
}
