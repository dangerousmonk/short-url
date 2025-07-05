// Package server is used to describe main application entities as well as
// describes chi router and main HTTP handlers
package server

import (
	"compress/gzip"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"

	"github.com/dangerousmonk/short-url/cmd/config"
	_ "github.com/dangerousmonk/short-url/docs"
	"github.com/dangerousmonk/short-url/internal/auth"
	"github.com/dangerousmonk/short-url/internal/compress"
	"github.com/dangerousmonk/short-url/internal/handlers"
	"github.com/dangerousmonk/short-url/internal/logging"
	"github.com/dangerousmonk/short-url/internal/models"
	"github.com/dangerousmonk/short-url/internal/service"
)

// ShortURLApp is a structure to represent short-url app and its main components
type ShortURLApp struct {
	Config  *config.Config
	Logger  *zap.SugaredLogger
	DelCh   chan models.DeleteURLChannelMessage
	Service *service.URLShortenerService
}

// NewApp is a helper function that returns a pointer to the new app struct.
func NewApp(config *config.Config, logger *zap.SugaredLogger, delCh chan models.DeleteURLChannelMessage, s *service.URLShortenerService) *ShortURLApp {
	return &ShortURLApp{
		Config:  config,
		Logger:  logger,
		DelCh:   delCh,
		Service: s,
	}
}

// Start godoc
//
//	@title						short-url app
//	@version					1.0
//	@description				API Server
//	@BasePath					/
//
//	@securityDefinitions.apikey	ApiKeyAuth
//	@in							Cookie
//	@name						auth
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
	jwtAuthenticator, err := auth.NewJWTAuthenticator(server.Config.JWTSecret)
	if err != nil {
		logging.Log.Fatalf("Server failed initialize jwtAuthenticator | %v", err)
	}

	// middleware
	r.Use(logging.RequestLogger)
	r.Use(compress.DecompressMiddleware)
	r.Use(compressor.Handler)

	// swagger
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8099/swagger/doc.json"),
	))

	// handlers
	httpHandler := handlers.NewHandler(*server.Service)
	r.Get("/ping", httpHandler.Ping)

	r.Group(func(r chi.Router) {
		r.Use(auth.AuthMiddleware(jwtAuthenticator))
		r.Post("/api/shorten", httpHandler.APIShorten)
		r.Post("/api/shorten/batch", httpHandler.APIShortenBatch)
		r.Get("/api/user/urls", httpHandler.GetUserURLs)
		r.Delete("/api/user/urls", httpHandler.APIDeleteBatch)
		r.Get("/{hash}", httpHandler.GetURL)
		r.Post("/", httpHandler.Shorten)
	})

	r.Mount("/debug", middleware.Profiler())
	return r
}
