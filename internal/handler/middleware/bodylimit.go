package middleware

import (
	"context"
	"net/http"

	"github.com/zlatovlaska707/avito-kitchen/internal/handler/httpx"
)

func BodyLimit(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}

func PayloadTooLarge() httpx.APIError {
	return httpx.APIError{Status: http.StatusRequestEntityTooLarge, Code: "payload_too_large", Message: "Payload too large"}
}

func WritePayloadTooLarge(w http.ResponseWriter, ctx context.Context) {
	httpx.RespondError(w, ctx, PayloadTooLarge(), nil)
}
