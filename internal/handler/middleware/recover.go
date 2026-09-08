package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/zlatovlaska707/avito-kitchen/internal/handler/httpx"
)

func Recover(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqCtx := r.Context()
			defer func() {
				if rec := recover(); rec != nil {
					logger.Error("panic recovered",
						"panic", rec,
						"stack", string(debug.Stack()),
						"request_id", httpx.RequestIDFromContext(reqCtx),
					)
					apiErr := httpx.APIError{Status: http.StatusInternalServerError, Code: "internal_error", Message: "Internal server error"}
					httpx.RespondError(w, reqCtx, apiErr, nil)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
