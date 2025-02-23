package main

import (
	"log"
	"net/http"

	"github.com/dangerousmonk/short-url/internal/handlers"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handlers.URLShortenerHandler)
	mux.HandleFunc(`/{id}`, handlers.GetFullURLHandler)

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatalf("Ошибка запуска: %v", err)
		panic(err)
	}
}
