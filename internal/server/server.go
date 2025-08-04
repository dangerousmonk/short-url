// Package server is used to describe main application entities as well as
// describes chi router and main HTTP handlers
package server

import (
	"compress/gzip"
	"context"
	"fmt"
	"net"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/dangerousmonk/short-url/cmd/config"
	_ "github.com/dangerousmonk/short-url/docs"
	"github.com/dangerousmonk/short-url/internal/auth"
	"github.com/dangerousmonk/short-url/internal/compress"
	"github.com/dangerousmonk/short-url/internal/handlers"
	"github.com/dangerousmonk/short-url/internal/logging"
	"github.com/dangerousmonk/short-url/internal/models"
	"github.com/dangerousmonk/short-url/internal/network"
	"github.com/dangerousmonk/short-url/internal/server/proto"
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
func (app *ShortURLApp) Start() error {
	r := app.initRouter()

	rootCtx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer cancel()

	server := &http.Server{
		Addr:    app.Config.ServerAddr,
		Handler: r,
	}

	gRPCServer := grpc.NewServer()
	proto.RegisterURLShortenerServiceServer(gRPCServer, proto.NewShortenerGRPCServer(app.Service, app.Config))
	reflection.Register(gRPCServer)

	go func() {
		if err := listenAndServeHTTP(app, server); err != nil && err != http.ErrServerClosed {
			app.Logger.Fatal("HTTP server failed to start")
		}
	}()

	go func() {
		if err := startgRPS(rootCtx, app, app.Service); err != nil {
			app.Logger.Fatal("gRPS server failed to start %v", err)
		}
	}()

	<-rootCtx.Done()
	app.Logger.Info("Received shutdown signal, shutting down.")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Duration(app.Service.Cfg.ShutDownTimeout)*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}
	return nil
}

func (app *ShortURLApp) initRouter() *chi.Mux {
	r := chi.NewRouter()
	compressor := middleware.NewCompressor(gzip.DefaultCompression, compress.CompressedContentTypes...)
	jwtAuthenticator, err := auth.NewJWTAuthenticator(app.Config.JWTSecret)
	if err != nil {
		logging.Log.Fatalf("Server failed initialize jwtAuthenticator | %v", err)
	}

	// middleware
	r.Use(logging.RequestLogger)
	r.Use(compress.DecompressMiddleware)
	r.Use(compressor.Handler)

	// swagger
	swaggerJSONURL := fmt.Sprintf("http://%s/swagger/doc.json", app.Config.ServerAddr)
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL(swaggerJSONURL),
	))

	// handlers
	httpHandler := handlers.NewHandler(*app.Service)
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

	r.Route("/api/internal", func(r chi.Router) {
		internalRoutesGroup := r.Group(nil)
		internalRoutesGroup.Use(network.CheckSubnet(app.Config))
		internalRoutesGroup.Get("/stats", httpHandler.GetStats)
	})

	r.Mount("/debug", middleware.Profiler())
	return r
}

// listenAndServeHTTP handles which method for serving http should be called based on config
func listenAndServeHTTP(app *ShortURLApp, s *http.Server) error {
	if app.Config.EnableHTTPS {
		app.Logger.Infof("Running HTTPS on %s...", app.Config.ServerAddr)
		return s.ListenAndServeTLS(app.Config.CertPath, app.Config.CertPrivateKeyPath)
	}

	app.Logger.Infof("Running HTTP on %s...", app.Config.ServerAddr)
	return s.ListenAndServe()
}

// startgRPS starts grps server
func startgRPS(ctx context.Context, app *ShortURLApp, s service.URLShortener) error {
	lis, err := net.Listen("tcp", app.Config.GRPCServerAddress)
	if err != nil {
		return fmt.Errorf("failed to listen on gRPC: %w", err)
	}

	gRPCServer := grpc.NewServer()
	proto.RegisterURLShortenerServiceServer(gRPCServer, proto.NewShortenerGRPCServer(s, app.Config))
	reflection.Register(gRPCServer)

	go func() {
		if err := gRPCServer.Serve(lis); err != nil {
			app.Logger.Fatal("gRPC server failed to start %v", err)
		}
	}()
	app.Logger.Infof("Running gRPC on %s...", app.Config.GRPCServerAddress)

	go func() {
		<-ctx.Done()
		gRPCServer.GracefulStop()
	}()

	return nil
}
