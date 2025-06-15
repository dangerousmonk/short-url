package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/dangerousmonk/short-url/cmd/config"
	"github.com/dangerousmonk/short-url/internal/logging"
	"github.com/dangerousmonk/short-url/internal/models"
	"github.com/dangerousmonk/short-url/internal/repository/memory"
	"github.com/dangerousmonk/short-url/internal/service"
)

func Example() {
	// Init dependencies
	cfg := config.InitConfig()
	_, err := logging.InitLogger(cfg.LogLevel, cfg.Env)
	repo := memory.NewMemoryRepository(cfg)
	delCh := make(chan models.DeleteURLChannelMessage)
	defer close(delCh)

	s := service.NewShortenerService(repo, cfg, delCh)
	httpHandler := NewHandler(*s)

	// Init router and attach http handlers
	r := chi.NewRouter()
	r.Post("/api/shorten", httpHandler.APIShorten)
	r.Post("/api/shorten/batch", httpHandler.APIShortenBatch)
	r.Get("/api/user/urls", httpHandler.GetUserURLs)
	r.Delete("/api/user/urls", httpHandler.APIDeleteBatch)
	r.Get("/{hash}", httpHandler.GetURL)
	r.Post("/", httpHandler.Shorten)

	// Start the server
	err = http.ListenAndServe("localhost:8080", r)
	if err != nil {
		panic(err)
	}

}
