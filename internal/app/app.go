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

type App struct {
	cfg    config.Config
	logger *slog.Logger
	pool   *pgxpool.Pool
	server *http.Server
	deps   *Dependencies
}

func New(ctx context.Context, cfg config.Config, logger *slog.Logger) (*App, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.PostgresDSN)
	if err != nil {
		return nil, fmt.Errorf("parse postgres dsn: %w", err)
	}
	poolCfg.MaxConns = cfg.Postgres.MaxConns
	poolCfg.MinConns = cfg.Postgres.MinConns
	poolCfg.MaxConnLifetime = cfg.Postgres.MaxConnLifetime
	poolCfg.MaxConnIdleTime = cfg.Postgres.MaxConnIdleTime
	if poolCfg.ConnConfig.RuntimeParams == nil {
		poolCfg.ConnConfig.RuntimeParams = map[string]string{}
	}
	poolCfg.ConnConfig.RuntimeParams["application_name"] = "avito-kitchen"

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("create postgres pool: %w", err)
	}

	deps, err := setup(ctx, cfg, pool, logger)
	if err != nil {
		pool.Close()
		return nil, err
	}

	router := chi.NewRouter()
	registerRoutes(router, deps)

	if deps.Workers != nil {
		deps.Workers.Start(ctx)
	}

	server := &http.Server{
		Addr:         cfg.HTTP.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		IdleTimeout:  cfg.HTTP.IdleTimeout,
	}

	return &App{cfg: cfg, logger: logger, pool: pool, server: server, deps: deps}, nil
}

func (a *App) Run(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		a.logger.Info("starting http server", "addr", a.cfg.HTTP.Address)
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), a.cfg.HTTP.ShutdownTimeout)
		defer cancel()
		if a.deps.Workers != nil {
			a.deps.Workers.Stop()
		}
		if shutdownErr := a.server.Shutdown(shutdownCtx); shutdownErr != nil { //nolint:contextcheck // intentional detached shutdown context
			return fmt.Errorf("shutdown server: %w", shutdownErr)
		}
		a.pool.Close()
		a.logger.Info("application stopped")
		return nil
	case err := <-errCh:
		if a.deps.Workers != nil {
			a.deps.Workers.Stop()
		}
		a.pool.Close()
		return err
	}
}

func (a *App) Stop(ctx context.Context) error {
	if a.deps.Workers != nil {
		a.deps.Workers.Stop()
	}
	return a.server.Shutdown(ctx)
}
