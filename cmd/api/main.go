package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/zlatovlaska707/avito-kitchen/internal/app"
	"github.com/zlatovlaska707/avito-kitchen/internal/config"
	"github.com/zlatovlaska707/avito-kitchen/internal/logging"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "config/config.yaml", "path to config yaml")
	flag.Parse()

	cfg := config.MustLoad(configPath)
	logger := logging.New(cfg.Log.Level, cfg.Log.Format)
	slog.SetDefault(logger)

	application, err := app.New(context.Background(), cfg, logger)
	if err != nil {
		fmt.Fprintf(os.Stderr, "app init: %v\n", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := application.Run(ctx); err != nil {
		logger.Error("application stopped with error", "error", err)
		os.Exit(1)
	}
}
