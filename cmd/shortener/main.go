package main

import (
	"log"
	"net/http"

	"github.com/dangerousmonk/short-url/cmd/config"
	"github.com/dangerousmonk/short-url/internal/handlers"
	"github.com/go-chi/chi/v5"
)

func main() {
	cfg := config.InitConfig()
	r := chi.NewRouter()

	r.Post("/", handlers.URLShortenerHandler)
	r.Get("/{hash}", handlers.GetFullURLHandler)

	log.Printf("Running app on %s...\n", cfg.ServerAddr)
	err := http.ListenAndServe(cfg.ServerAddr, r)
	if err != nil {
		log.Fatalf("App startup failed: %v", err)
		panic(err)
	}
}
