package main

import (
	"log"
	"net/http"

	"github.com/dangerousmonk/short-url/internal/handlers"
	"github.com/go-chi/chi/v5"
)

func main() {
	r := chi.NewRouter()
	r.Post("/", handlers.URLShortenerHandler)
	r.Get("/{hash}", handlers.GetFullURLHandler)

	err := http.ListenAndServe(":8080", r)
	if err != nil {
		log.Fatalf("Ошибка запуска: %v", err)
		panic(err)
	}
}
