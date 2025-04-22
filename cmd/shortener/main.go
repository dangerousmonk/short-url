package main

import (
	"context"
	"errors"
	"log"

	"github.com/dangerousmonk/short-url/cmd/config"
	"github.com/dangerousmonk/short-url/internal/logging"
	"github.com/dangerousmonk/short-url/internal/models"
	"github.com/dangerousmonk/short-url/internal/server"
	"github.com/dangerousmonk/short-url/internal/storage"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
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
	var appStorage storage.Storage
	if cfg.DatabaseDSN != "" {
		applyMigrations(cfg)
		db, err := storage.InitDB(ctx, cfg.DatabaseDSN)
		if err != nil {
			logger.Fatalf("Failed init postgresql: %v", err)
		}
		defer db.Close()
		appStorage = &storage.PostgreSQLStorage{DB: db}
	} else {
		mapStorage := storage.InitMapStorage(cfg)
		err = storage.LoadFromFile(mapStorage, cfg)
		if err != nil {
			logger.Fatalf("Failed init file storage: %v", err)
		}
		appStorage = mapStorage
	}

	delCh := make(chan models.DeleteURLChannelMessage)
	defer close(delCh)

	server := server.NewApp(cfg, appStorage, logger, delCh)
	go server.FlushDeleteMessages()
	err = server.Start()
	if err != nil {
		logger.Fatalf("Failed init server: %v", err)
	}
}

func applyMigrations(cfg *config.Config) {
	logging.Log.Infof("DB DSN=%s", cfg.DatabaseDSN)
	migrations, err := migrate.New("file://internal/storage/migrations", cfg.DatabaseDSN)
	if err != nil {
		logging.Log.Fatalf("Failed to apply migrations: %v", err)
	}
	migrations.Up()

	if err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			logging.Log.Info("Migrations no change")
			return
		}
		log.Fatalf("Migrations failed: %v ", err)
		return
	}
	logging.Log.Info(" Migrations: success")
}
