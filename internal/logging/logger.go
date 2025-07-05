// Package logging containts all the tools required to setup logging in the application
package logging

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

// Log is object of zap.SugaredLogger
var Log *zap.SugaredLogger

// InitLogger is a helper function to initalize logger.
func InitLogger(level string, env string) (*zap.SugaredLogger, error) {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return nil, err
	}
	var cfg zap.Config

	switch env {
	case "dev":
		cfg = zap.NewDevelopmentConfig()
	default:
		cfg = zap.NewProductionConfig()
	}

	cfg.Level = lvl
	logger, err := cfg.Build()
	if err != nil {
		return nil, err
	}
	Log = logger.Sugar()
	return Log, nil
}

// RequestLogger is helper middleware function to log every request main parameters: method,URI,response status code,
// response size and how long it took to handle request.
func RequestLogger(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)
		elapsed := time.Since(start)
		Log.Infoln(
			"Incoming HTTP request:",
			"method", r.Method,
			"URI", r.RequestURI,
			"statusCode", ww.Status(),
			"elapsedTime", elapsed,
			"responseSize", ww.BytesWritten(),
		)
	}
	return http.HandlerFunc(fn)
}
