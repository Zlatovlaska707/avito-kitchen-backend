package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/zlatovlaska707/avito-kitchen/internal/config"
)

// NewHTTPHandler builds the API router (no background workers). Used by contract tests.
func NewHTTPHandler(ctx context.Context, cfg config.Config, pool *pgxpool.Pool, logger *slog.Logger) (http.Handler, error) {
	deps, err := setup(ctx, cfg, pool, logger)
	if err != nil {
		return nil, fmt.Errorf("setup dependencies: %w", err)
	}
	r := chi.NewRouter()
	registerRoutes(r, deps)
	return r, nil
}
