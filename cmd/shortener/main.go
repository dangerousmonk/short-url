package main

import (
	"context"
	"log"

	"github.com/dangerousmonk/short-url/cmd/config"
	"github.com/dangerousmonk/short-url/internal/database"
	"github.com/dangerousmonk/short-url/internal/logging"
	"github.com/dangerousmonk/short-url/internal/models"
	"github.com/dangerousmonk/short-url/internal/repository"
	"github.com/dangerousmonk/short-url/internal/repository/memory"
	"github.com/dangerousmonk/short-url/internal/server"
	"github.com/dangerousmonk/short-url/internal/service"
)

func main() {
	cfg := config.InitConfig()
	logger, err := logging.InitLogger(cfg.LogLevel, cfg.Env)
	if err != nil {
		log.Fatalf("Failed init log: %v", err)
	}
	defer func() {
		if err := logger.Sync(); err != nil {
			logger.Warnf("Failed to sync logger: %v", err)
		}
	}()

	ctx := context.Background()
	var appRepo repository.Repository
	if cfg.DatabaseDSN != "" {
		database.ApplyMigrations(cfg)
		db, err := database.InitDB(ctx, cfg.DatabaseDSN)
		if err != nil {
			logger.Fatalf("Failed init postgresql: %v", err)
		}
		defer db.Close()
		appRepo = repository.NewPostgresRepo(db)
	} else {
		repo := memory.NewMemoryRepository(cfg)
		err = memory.LoadFromFile(repo, cfg)
		if err != nil {
			logger.Fatalf("Failed init file storage: %v", err)
		}
		appRepo = repo
	}

	delCh := make(chan models.DeleteURLChannelMessage)
	defer close(delCh)

	s := service.NewShortenerService(appRepo, cfg, delCh)
	go s.FlushDeleteMessages()

	server := server.NewApp(cfg, logger, delCh, s)
	err = server.Start()
	if err != nil {
		logger.Fatalf("Failed init server: %v", err)
	}
}
