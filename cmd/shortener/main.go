package main

import (
	"compress/gzip"
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/dangerousmonk/short-url/cmd/config"
	"github.com/dangerousmonk/short-url/internal/auth"
	"github.com/dangerousmonk/short-url/internal/compress"
	"github.com/dangerousmonk/short-url/internal/handlers"
	"github.com/dangerousmonk/short-url/internal/logging"
	"github.com/dangerousmonk/short-url/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
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

	r := chi.NewRouter()
	compressor := middleware.NewCompressor(gzip.DefaultCompression, compress.CompressedContentTypes...)

	// middleware
	r.Use(logging.RequestLogger)
	r.Use(compress.DecompressMiddleware)
	r.Use(compressor.Handler)

	// handlers
	pingHandler := handlers.PingHandler{Config: cfg, Storage: appStorage}
	shortenHandler := handlers.URLShortenerHandler{Config: cfg, Storage: appStorage}
	getFullURLHandler := handlers.GetFullURLHandler{Config: cfg, Storage: appStorage}
	apiShortenerHandler := handlers.APIShortenerHandler{Config: cfg, Storage: appStorage}
	apiBatchHandler := handlers.APIShortenBatchHandler{Config: cfg, Storage: appStorage}
	apiGetUserURLsHandler := handlers.APIGetUserURLsHandler{Config: cfg, Storage: appStorage}

	r.Get("/ping", pingHandler.ServeHTTP)

	r.Group(func(r chi.Router) {
		r.Use(auth.AuthMiddleware)
		r.Post("/api/shorten", apiShortenerHandler.ServeHTTP)
		r.Post("/api/shorten/batch", apiBatchHandler.ServeHTTP)
		r.Get("/{hash}", getFullURLHandler.ServeHTTP)
		r.Post("/", shortenHandler.ServeHTTP)
	})

	r.Group(func(r chi.Router) {
		r.Use(auth.AuthMiddleware)
		r.Get("/api/user/urls", apiGetUserURLsHandler.ServeHTTP)
	})

	logger.Infof("Running app on %s...", cfg.ServerAddr)

	err = http.ListenAndServe(cfg.ServerAddr, r)
	if err != nil {
		logger.Fatalf("App startup failed: %v", err)
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
