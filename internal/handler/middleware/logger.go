package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/zlatovlaska707/avito-kitchen/internal/handler/httpx"
	"github.com/zlatovlaska707/avito-kitchen/internal/metrics"
)

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func Logger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(sw, r)
			duration := time.Since(start)
			path := r.URL.Path
			metrics.ObserveHTTP(r.Method, path, sw.status, duration)
			logger.Info("request completed",
				"method", r.Method,
				"path", path,
				"status", sw.status,
				"duration_ms", duration.Milliseconds(),
				"request_id", httpx.RequestIDFromContext(r.Context()),
			)
		})
	}
}
