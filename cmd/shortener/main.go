package main

import (
	"log"
	"net/http"

	"github.com/dangerousmonk/short-url/cmd/config"
	"github.com/dangerousmonk/short-url/internal/handlers"
	"github.com/dangerousmonk/short-url/internal/storage"
	"github.com/go-chi/chi/v5"
)

func main() {
	cfg := config.InitConfig()
	storage := storage.NewMapStorage()
	r := chi.NewRouter()

	shortenHandler := handlers.URLShortenerHandler{Config: cfg, MapStorage: storage}
	getFullURLHandler := handlers.GetFullURLHandler{Config: cfg, MapStorage: storage}

	r.Post("/", shortenHandler.ServeHTTP)
	r.Get("/{hash}", getFullURLHandler.ServeHTTP)
	log.Printf("Running app on %s...\n", cfg.ServerAddr)

	err := http.ListenAndServe(cfg.ServerAddr, r)
	if err != nil {
		log.Fatalf("App startup failed: %v", err)
	}
}
