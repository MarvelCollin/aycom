package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	fmt.Println("Event bus service started")

	// Set up graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down event bus...")

	// Give the server 5 seconds to finish ongoing requests
	time.Sleep(5 * time.Second)

	log.Println("Event bus stopped")
}
