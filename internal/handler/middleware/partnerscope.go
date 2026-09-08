package middleware

import (
	"net/http"

	"github.com/zlatovlaska707/avito-kitchen/internal/handler/apierr"
	"github.com/zlatovlaska707/avito-kitchen/internal/handler/httpx"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase"
)

const (
	ScopeMenuWrite   = "menu:write"
	ScopeOrdersWrite = "orders:write"
)

// RequirePartnerScope checks that the authenticated partner key includes scope.
func RequirePartnerScope(scope string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			p, ok := PartnerFromContext(r.Context())
			if !ok {
				httpx.RespondError(w, r.Context(), apierr.MapDomainError(usecase.ErrUnauthorized), nil)
				return
			}
			if !hasScope(p.Scopes, scope) {
				httpx.RespondError(w, r.Context(), apierr.MapDomainError(usecase.ErrForbidden), nil)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func hasScope(scopes []string, want string) bool {
	for _, s := range scopes {
		if s == want {
			return true
		}
	}
	return false
}
