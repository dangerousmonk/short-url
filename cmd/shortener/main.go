package main

import (
	"log"
	"net/http"

	"github.com/dangerousmonk/short-url/cmd/config"
	"github.com/dangerousmonk/short-url/internal/handlers"
	"github.com/dangerousmonk/short-url/internal/logging"
	"github.com/dangerousmonk/short-url/internal/storage"
	"github.com/go-chi/chi/v5"
)

func main() {
	cfg := config.InitConfig()
	storage := storage.NewMapStorage()
	logger, err := logging.InitLogger(cfg.LogLevel, cfg.Env)
	if err != nil {
		log.Fatalf("Failed init log: %v", err)
	}
	r := chi.NewRouter()
	r.Use(logging.RequestLogger)

	shortenHandler := handlers.URLShortenerHandler{Config: cfg, MapStorage: storage}
	getFullURLHandler := handlers.GetFullURLHandler{Config: cfg, MapStorage: storage}

	r.Post("/", shortenHandler.ServeHTTP)
	r.Get("/{hash}", getFullURLHandler.ServeHTTP)
	logger.Infof("Running app on %s...", cfg.ServerAddr)

	err = http.ListenAndServe(cfg.ServerAddr, r)
	if err != nil {
		logger.Fatalf("App startup failed: %v", err)
	}
}
