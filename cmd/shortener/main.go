package main

import (
	"compress/gzip"
	"context"
	"database/sql"
	"log"
	"net/http"

	"github.com/dangerousmonk/short-url/cmd/config"
	"github.com/dangerousmonk/short-url/internal/compress"
	"github.com/dangerousmonk/short-url/internal/handlers"
	"github.com/dangerousmonk/short-url/internal/logging"
	"github.com/dangerousmonk/short-url/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	cfg := config.InitConfig()
	storage := storage.NewMapStorage()
	logger, err := logging.InitLogger(cfg.LogLevel, cfg.Env)
	if err != nil {
		log.Fatalf("Failed init log: %v", err)
	}
	defer func() {
		if err := logger.Sync(); err != nil {
			logger.Warnf("Failed to sync logger: %v", err)
		}
	}()

	err = storage.LoadFromFile(cfg)
	if err != nil {
		logger.Fatalf("Failed init storage: %v", err)
	}

	db, err := sql.Open("pgx", cfg.DatabaseDSN)
	if err != nil {
		logger.Warnf("could not connect to database: %v", err)
	}
	defer db.Close()

	if err := db.PingContext(context.Background()); err != nil {
		logger.Warnf("unable to reach database: %v", err)
	}

	logger.Info("Database setup complete")

	r := chi.NewRouter()
	compressor := middleware.NewCompressor(gzip.DefaultCompression, compress.CompressedContentTypes...)

	// middleware
	r.Use(logging.RequestLogger)
	r.Use(compress.DecompressMiddleware)
	r.Use(compressor.Handler)

	// handlers
	shortenHandler := handlers.URLShortenerHandler{Config: cfg, MapStorage: storage}
	apiShortenerHandler := handlers.APIShortenerHandler{Config: cfg, MapStorage: storage}
	getFullURLHandler := handlers.GetFullURLHandler{Config: cfg, MapStorage: storage}
	pingHandler := handlers.PingHandler{Config: cfg, DB: db}

	r.Post("/", shortenHandler.ServeHTTP)
	r.Post("/api/shorten", apiShortenerHandler.ServeHTTP)
	r.Get("/{hash}", getFullURLHandler.ServeHTTP)
	r.Get("/ping", pingHandler.ServeHTTP)

	logger.Infof("Running app on %s...", cfg.ServerAddr)

	err = http.ListenAndServe(cfg.ServerAddr, r)
	if err != nil {
		logger.Fatalf("App startup failed: %v", err)
	}
}
