package bootstrap

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/zlatovlaska707/avito-kitchen/venue-simulator/internal/config"
	"github.com/zlatovlaska707/avito-kitchen/venue-simulator/internal/platformapi"
)

const clientTimeout = 15 * time.Second

// Platform wraps the generated oapi-codegen client for bootstrap operations.
type Platform struct {
	api *platformapi.ClientWithResponses
	cfg config.Config
}

func NewPlatform(cfg config.Config) (*Platform, error) {
	api, err := platformapi.NewPartnerClient(cfg.PlatformURL, cfg.APIKey, clientTimeout)
	if err != nil {
		return nil, err
	}
	return &Platform{api: api, cfg: cfg}, nil
}

func Run(cfg config.Config, setSecret func(string)) error {
	p, err := NewPlatform(cfg)
	if err != nil {
		return err
	}
	return p.Run(setSecret)
}

func (p *Platform) Run(setSecret func(string)) error {
	ctx := context.Background()
	if err := p.waitReady(ctx); err != nil {
		return err
	}
	if err := p.SyncMenu(ctx); err != nil {
		log.Printf("menu sync warning: %v", err)
	}
	secret, err := p.registerWebhook(ctx)
	if err != nil {
		log.Printf("webhook register warning: %v", err)
		secret = p.cfg.WebhookSecret
	}
	if setSecret != nil && secret != "" {
		setSecret(secret)
	}
	if err := p.setAcceptingOrders(ctx); err != nil {
		return err
	}
	log.Println("venue-simulator bootstrap complete")
	return nil
}

func SyncMenu(cfg config.Config) error {
	p, err := NewPlatform(cfg)
	if err != nil {
		return err
	}
	return p.SyncMenu(context.Background())
}

func SetItemAvailability(cfg config.Config, externalID string, available bool) error {
	p, err := NewPlatform(cfg)
	if err != nil {
		return err
	}
	return p.SetItemAvailability(context.Background(), externalID, available)
}

func (p *Platform) SyncMenu(ctx context.Context) error {
	content, err := os.ReadFile(p.cfg.MenuFile)
	if err != nil {
		return err
	}
	resp, err := p.api.SyncMenuWithBodyWithResponse(ctx, "application/json", bytes.NewReader(content))
	if err != nil {
		return err
	}
	if resp.JSON200 != nil {
		return nil
	}
	return fmt.Errorf("menu sync status %d: %s", resp.StatusCode(), string(resp.Body))
}

func (p *Platform) SetItemAvailability(ctx context.Context, externalID string, available bool) error {
	resp, err := p.api.UpdateItemAvailabilityWithResponse(ctx, externalID, platformapi.UpdateItemAvailabilityJSONRequestBody{
		IsAvailable: available,
	})
	if err != nil {
		return err
	}
	if resp.StatusCode() < 300 {
		return nil
	}
	return fmt.Errorf("availability status %d: %s", resp.StatusCode(), string(resp.Body))
}

func (p *Platform) waitReady(ctx context.Context) error {
	for i := 0; i < 30; i++ {
		resp, err := p.api.HealthReadyWithResponse(ctx)
		if err == nil && resp.StatusCode() == 200 {
			return nil
		}
		time.Sleep(time.Second)
	}
	return fmt.Errorf("platform not ready after 30s")
}

func (p *Platform) registerWebhook(ctx context.Context) (string, error) {
	secret := p.cfg.WebhookSecret
	if secret == "" {
		secret = "local-simulator-webhook-secret-change-me"
	}
	body, err := json.Marshal(map[string]any{
		"url":    p.cfg.WebhookPublicURL,
		"secret": secret,
		"event_types": []string{
			"order.created", "order.accepted", "order.ready", "order.completed",
			"order.cancelled", "order.rejected", "order.delivery_assigned",
		},
	})
	if err != nil {
		return "", err
	}
	resp, err := p.api.RegisterWebhookWithBodyWithResponse(ctx, "application/json", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	if resp.StatusCode() == 409 {
		return secret, nil
	}
	if resp.StatusCode() < 300 {
		return secret, nil
	}
	return "", fmt.Errorf("webhook register status %d: %s", resp.StatusCode(), string(resp.Body))
}

func (p *Platform) setAcceptingOrders(ctx context.Context) error {
	resp, err := p.api.UpdateVenueStatusWithResponse(ctx, platformapi.UpdateVenueStatusJSONRequestBody{
		AcceptsOrders: true,
	})
	if err != nil {
		return err
	}
	if resp.StatusCode() < 300 {
		return nil
	}
	return fmt.Errorf("venue status status %d: %s", resp.StatusCode(), string(resp.Body))
}
