package middleware

import (
	"context"
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/google/uuid"

	apikeyrepo "github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/repository/apikey"
	"github.com/zlatovlaska707/avito-kitchen/internal/handler/apierr"
	"github.com/zlatovlaska707/avito-kitchen/internal/handler/httpx"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase"
)

type partnerCtxKey struct{}

type PartnerContext struct {
	VenueID uuid.UUID
	KeyID   uuid.UUID
	Scopes  []string
	Prefix  string
}

func PartnerAuth(_ string, repo *apikeyrepo.CachedRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw, ok := bearerToken(r)
			if !ok {
				writeAuthError(w, r)
				return
			}
			prefix, secret, ok := parseAPIKey(raw)
			if !ok {
				writeAuthError(w, r)
				return
			}
			rec, err := repo.LookupByPrefix(r.Context(), prefix)
			if err != nil {
				writeAuthError(w, r)
				return
			}
			if subtle.ConstantTimeCompare([]byte(rec.KeyHash), []byte(hashAPIKey(raw))) != 1 {
				writeAuthError(w, r)
				return
			}
			_ = secret
			ctx := context.WithValue(r.Context(), partnerCtxKey{}, PartnerContext{
				VenueID: rec.VenueID,
				KeyID:   rec.ID,
				Scopes:  rec.Scopes,
				Prefix:  prefix,
			})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func PartnerFromContext(ctx context.Context) (PartnerContext, bool) {
	v, ok := ctx.Value(partnerCtxKey{}).(PartnerContext)
	return v, ok
}

func bearerToken(r *http.Request) (string, bool) {
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		return "", false
	}
	token := strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))
	return token, token != ""
}

func parseAPIKey(raw string) (prefix, secret string, ok bool) {
	parts := strings.Split(raw, "_")
	if len(parts) != 3 || parts[0] != "ak" || len(parts[1]) != 8 {
		return "", "", false
	}
	return parts[1], parts[2], true
}

func writeAuthError(w http.ResponseWriter, r *http.Request) {
	httpx.RespondError(w, r.Context(), apierr.MapDomainError(usecase.ErrUnauthorized), nil)
}
