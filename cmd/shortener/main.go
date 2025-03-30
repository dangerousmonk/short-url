package main

import (
	"compress/gzip"
	"context"
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
	shortenHandler := handlers.URLShortenerHandler{Config: cfg, Storage: appStorage}
	apiShortenerHandler := handlers.APIShortenerHandler{Config: cfg, Storage: appStorage}
	getFullURLHandler := handlers.GetFullURLHandler{Config: cfg, Storage: appStorage}
	pingHandler := handlers.PingHandler{Config: cfg, Storage: appStorage}

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
