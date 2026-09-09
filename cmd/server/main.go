package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/StellarYard/stellaryard-core/internal/api"
	"github.com/StellarYard/stellaryard-core/internal/docker"
	"github.com/StellarYard/stellaryard-core/internal/storage"
)

func main() {
	// Configuration
	port := os.Getenv("STELLARYARD_PORT")
	if port == "" {
		port = "8080"
	}
	dbPath := os.Getenv("STELLARYARD_DB_PATH")
	if dbPath == "" {
		dbPath = "./stellaryard.db"
	}

	log.Printf("Starting stellaryard-core on port %s", port)

	// Initialize storage
	db, err := storage.Open(dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()
	log.Println("Database initialized")

	// Initialize Docker client
	dockerClient, err := docker.NewClient()
	if err != nil {
		log.Fatalf("Failed to create Docker client: %v", err)
	}
	defer dockerClient.Close()
	log.Println("Docker client connected")

	// Create router
	router := api.NewRouter(dockerClient, db)

	// Create HTTP server
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("API server listening on http://localhost:%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	fmt.Println("Server stopped")
}
