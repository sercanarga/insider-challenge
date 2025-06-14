package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	_ "insider-challenge/docs" // swagger
	"insider-challenge/internal/handler"
	"insider-challenge/internal/repository"
	"insider-challenge/internal/service"
	"insider-challenge/pkg/config"
)

// @title           Insider Challenge API
// @version         1.0
// @description     A message processing service API
// @BasePath        /
func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	db, err := repository.InitDB(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize Redis
	if err := config.InitRedis(); err != nil {
		log.Fatalf("Failed to initialize Redis: %v", err)
	}

	repo := repository.New(db)
	svc := service.New(repo, cfg)
	h := handler.New(svc, cfg)

	go svc.StartMessageSender()

	go func() {
		if err := h.Start(cfg.ServerPort); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	svc.StopMessageSender()
}
