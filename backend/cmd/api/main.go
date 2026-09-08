package main

import (
	"log"
	"time"

	"eventman/backend/internal/config"
	"eventman/backend/internal/database"
	"eventman/backend/internal/jobs"
	"eventman/backend/internal/server"
	"eventman/backend/internal/services"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	cfg := config.Load()

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	database.Seed(db)

	midtransSvc := services.NewMidtransService(cfg)
	txSvc := services.NewTransactionService(db, midtransSvc, cfg)
	jobs.StartTransactionExpirySweeper(txSvc, 5*time.Minute)

	router := server.NewRouter(db, cfg)

	log.Printf("event-management API listening on :%s (env=%s)", cfg.Port, cfg.AppEnv)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
