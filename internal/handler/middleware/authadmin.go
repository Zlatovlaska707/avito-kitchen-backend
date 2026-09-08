package middleware

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"net/http"

	"github.com/zlatovlaska707/avito-kitchen/internal/handler/apierr"
	"github.com/zlatovlaska707/avito-kitchen/internal/handler/httpx"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase"
)

type adminCtxKey struct{}

func AdminAuth(expectedToken string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw, ok := bearerToken(r)
			if !ok || subtle.ConstantTimeCompare([]byte(raw), []byte(expectedToken)) != 1 {
				httpx.RespondError(w, r.Context(), apierr.MapDomainError(usecase.ErrUnauthorized), nil)
				return
			}
			ctx := context.WithValue(r.Context(), adminCtxKey{}, struct{}{})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func AdminFromContext(ctx context.Context) bool {
	_, ok := ctx.Value(adminCtxKey{}).(struct{})
	return ok
}

func hashAPIKey(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
