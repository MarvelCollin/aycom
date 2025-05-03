package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"sync"

	handlers "aycom/backend/services/thread/api"
	"aycom/backend/services/thread/db"
	"aycom/backend/services/thread/proto"
	"aycom/backend/services/thread/service"

	"google.golang.org/grpc"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		port := os.Getenv("PORT")
		if port == "" {
			port = "9092"
		}
		listener, err := net.Listen("tcp", ":"+port)
		if err != nil {
			log.Fatalf("Failed to listen: %v", err)
		}
		dsn := fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			getEnv("DB_HOST", "thread_db"),
			getEnv("DB_PORT", "5432"),
			getEnv("DB_USER", "kolin"),
			getEnv("DB_PASSWORD", "kolin"),
			getEnv("DB_NAME", "thread_db"),
		)
		dbConn, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}
		threadRepo := db.NewThreadRepository(dbConn)
		mediaRepo := db.NewMediaRepository(dbConn)
		hashtagRepo := db.NewHashtagRepository(dbConn)
		mentionRepo := db.NewMentionRepository(dbConn)
		threadService := service.NewThreadService(threadRepo, mediaRepo, hashtagRepo, mentionRepo)
		handler := handlers.NewThreadHandler(threadService, nil, nil, nil)
		grpcServer := grpc.NewServer()
		proto.RegisterThreadServiceServer(grpcServer, handler)
		log.Printf("Thread service started on port %s", port)
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	go func() {
		defer wg.Done()
		http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
			w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})
		log.Println("Thread service health endpoint started on :8082")
		http.ListenAndServe(":8082", nil)
	}()

	wg.Wait()
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
