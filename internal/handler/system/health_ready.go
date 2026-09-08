package system

import (
	"context"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/zlatovlaska707/avito-kitchen/internal/handler/httpx"
)

func Ready(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), time.Second)
		defer cancel()
		if err := pool.Ping(ctx); err != nil {
			httpx.RespondError(w, r.Context(), httpx.APIError{
				Status: http.StatusServiceUnavailable, Code: "service_unavailable", Message: "Database unavailable",
			}, err)
			return
		}
		httpx.RespondJSON(w, http.StatusOK, map[string]string{"status": "ready"})
	}
}
