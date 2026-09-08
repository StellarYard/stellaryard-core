package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/StellarYard/stellaryard-core/internal/api"
	"github.com/StellarYard/stellaryard-core/internal/docker"
	"github.com/StellarYard/stellaryard-core/internal/signer"
	"github.com/StellarYard/stellaryard-core/internal/storage"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	db, err := storage.New("stellaryard.db")
	if err != nil {
		log.Fatalf("failed to initialize storage: %v", err)
	}
	defer db.Close()

	s := signer.NewLocalTestSigner(db)

	dockerClient, err := docker.NewClient()
	if err != nil {
		log.Fatalf("failed to initialize docker client: %v", err)
	}

	router := api.NewRouter(db, s, dockerClient)

	addr := fmt.Sprintf(":%s", port)
	log.Printf("stellaryard-core listening on %s", addr)

	go func() {
		if err := http.ListenAndServe(addr, router); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down...")
}
