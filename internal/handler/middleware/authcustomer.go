package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"

	"github.com/zlatovlaska707/avito-kitchen/internal/handler/apierr"
	"github.com/zlatovlaska707/avito-kitchen/internal/handler/httpx"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase"
)

type customerCtxKey struct{}

func CustomerIdentity() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw := r.Header.Get("X-Customer-Id")
			if raw == "" {
				httpx.RespondError(w, r.Context(), apierr.MapDomainError(usecase.ErrInvalidCustomerID), nil)
				return
			}
			id, err := uuid.Parse(raw)
			if err != nil {
				httpx.RespondError(w, r.Context(), apierr.MapDomainError(usecase.ErrInvalidCustomerID), nil)
				return
			}
			ctx := context.WithValue(r.Context(), customerCtxKey{}, id)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func CustomerFromContext(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(customerCtxKey{}).(uuid.UUID)
	return id, ok
}
