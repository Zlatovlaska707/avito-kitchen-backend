package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/zlatovlaska707/avito-kitchen/venue-simulator/internal/bootstrap"
	"github.com/zlatovlaska707/avito-kitchen/venue-simulator/internal/config"
	"github.com/zlatovlaska707/avito-kitchen/venue-simulator/internal/control"
	"github.com/zlatovlaska707/avito-kitchen/venue-simulator/internal/polling"
	"github.com/zlatovlaska707/avito-kitchen/venue-simulator/internal/runtime"
	"github.com/zlatovlaska707/avito-kitchen/venue-simulator/internal/webhook"
)

func main() {
	cfg := config.Load()
	if cfg.APIKey == "" {
		log.Fatal("PARTNER_API_KEY is required")
	}

	state := runtime.NewState(cfg)
	if err := bootstrap.Run(cfg, state.SetWebhookSecret); err != nil {
		log.Fatalf("bootstrap: %v", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	wh, err := webhook.NewReceiver(cfg, state)
	if err != nil {
		log.Fatalf("webhook receiver: %v", err)
	}
	ctrl := control.New(cfg, state)
	mux := http.NewServeMux()
	mux.HandleFunc("/webhooks/orders", wh.HandleOrderCreated)
	mux.HandleFunc("/health/live", wh.HandleHealth)
	mux.HandleFunc("/health/ready", wh.HandleHealth)
	mux.HandleFunc("/orders", wh.HandleOrders)
	mux.HandleFunc("/control/mode", ctrl.HandleMode)
	mux.HandleFunc("/control/stoplist", ctrl.HandleStoplist)
	mux.HandleFunc("/control/menu/reload", ctrl.HandleMenuReload)

	srv := &http.Server{Addr: cfg.HTTPAddress, Handler: mux}
	go func() {
		log.Printf("venue-simulator listening on %s mode=%s", cfg.HTTPAddress, state.Mode())
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http server: %v", err)
		}
	}()

	if cfg.PollingEnabled {
		poller, err := polling.New(cfg, wh)
		if err != nil {
			log.Fatalf("poller: %v", err)
		}
		go poller.Run(ctx)
	}

	<-ctx.Done()
	_ = srv.Shutdown(context.Background())
	log.Println("venue-simulator stopped")
	os.Exit(0)
}
