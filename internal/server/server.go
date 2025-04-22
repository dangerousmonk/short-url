package server

import (
	"compress/gzip"
	"context"
	"net/http"
	"time"

	"github.com/dangerousmonk/short-url/cmd/config"
	"github.com/dangerousmonk/short-url/internal/auth"
	"github.com/dangerousmonk/short-url/internal/compress"
	"github.com/dangerousmonk/short-url/internal/handlers"
	"github.com/dangerousmonk/short-url/internal/logging"
	"github.com/dangerousmonk/short-url/internal/models"
	"github.com/dangerousmonk/short-url/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

type ShortURLApp struct {
	Storage storage.Storage
	Config  *config.Config
	Logger  *zap.SugaredLogger
	DelCh   chan models.DeleteURLChannelMessage
}

func NewApp(config *config.Config, storage storage.Storage, logger *zap.SugaredLogger, delCh chan models.DeleteURLChannelMessage) *ShortURLApp {
	return &ShortURLApp{
		Config:  config,
		Storage: storage,
		Logger:  logger,
		DelCh:   delCh,
	}
}

func (server *ShortURLApp) Start() error {
	r := server.initRouter()
	server.Logger.Infof("Running app on %s...", server.Config.ServerAddr)

	err := http.ListenAndServe(server.Config.ServerAddr, r)
	if err != nil {
		return err
	}
	return nil
}

func (server *ShortURLApp) initRouter() *chi.Mux {
	r := chi.NewRouter()
	compressor := middleware.NewCompressor(gzip.DefaultCompression, compress.CompressedContentTypes...)

	// middleware
	r.Use(logging.RequestLogger)
	r.Use(compress.DecompressMiddleware)
	r.Use(compressor.Handler)

	// handlers
	pingHandler := handlers.PingHandler{Config: server.Config, Storage: server.Storage}
	shortenHandler := handlers.URLShortenerHandler{Config: server.Config, Storage: server.Storage}
	getFullURLHandler := handlers.GetFullURLHandler{Config: server.Config, Storage: server.Storage}
	apiShortenerHandler := handlers.APIShortenerHandler{Config: server.Config, Storage: server.Storage}
	apiBatchHandler := handlers.APIShortenBatchHandler{Config: server.Config, Storage: server.Storage}
	apiGetUserURLsHandler := handlers.APIGetUserURLsHandler{Config: server.Config, Storage: server.Storage}
	apiDeleteUserURLsHandler := handlers.APIDeleteUserURLsHandler{Config: server.Config, Storage: server.Storage, DelCh: server.DelCh}

	r.Get("/ping", pingHandler.ServeHTTP)

	r.Group(func(r chi.Router) {
		r.Use(auth.AuthMiddleware)
		r.Post("/api/shorten", apiShortenerHandler.ServeHTTP)
		r.Post("/api/shorten/batch", apiBatchHandler.ServeHTTP)
		r.Get("/api/user/urls", apiGetUserURLsHandler.ServeHTTP)
		r.Delete("/api/user/urls", apiDeleteUserURLsHandler.ServeHTTP)
		r.Get("/{hash}", getFullURLHandler.ServeHTTP)
		r.Post("/", shortenHandler.ServeHTTP)
	})
	return r
}

func (server *ShortURLApp) FlushDeleteMessages() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	var messages []models.DeleteURLChannelMessage

	for {
		select {
		case msg := <-server.DelCh:
			messages = append(messages, msg)
		case <-ticker.C:
			if len(messages) == 0 {
				continue
			}
			for _, msg := range messages {
				err := server.Storage.DeleteBatch(context.TODO(), msg.URLs, msg.UserID)
				if err != nil {
					server.Logger.Warnf("FlushDeleteMessages error=%v", err)
					continue
				}
			}
			messages = nil
		}
	}
}
