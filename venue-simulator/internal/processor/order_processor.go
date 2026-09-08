package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync/atomic"
	"time"

	"github.com/google/uuid"

	"github.com/zlatovlaska707/avito-kitchen/venue-simulator/internal/config"
	"github.com/zlatovlaska707/avito-kitchen/venue-simulator/internal/platformapi"
	"github.com/zlatovlaska707/avito-kitchen/venue-simulator/internal/runtime"
)

type OrderProcessor struct {
	api     *platformapi.ClientWithResponses
	cfg     config.Config
	state   *runtime.State
	counter atomic.Int64
}

func New(cfg config.Config, state *runtime.State) (*OrderProcessor, error) {
	api, err := platformapi.NewPartnerClient(cfg.PlatformURL, cfg.APIKey, 15*time.Second)
	if err != nil {
		return nil, err
	}
	return &OrderProcessor{api: api, cfg: cfg, state: state}, nil
}

func (p *OrderProcessor) ProcessOrder(orderID string, webhookBody json.RawMessage) {
	mode := p.state.Mode()
	if mode == "silent" {
		return
	}
	if mode == "broken_webhook" && webhookBody != nil {
		return
	}
	ctx := context.Background()
	if mode == "auto_reject" {
		time.Sleep(p.cfg.AcceptDelay)
		p.reject(ctx, orderID)
		return
	}

	time.Sleep(p.cfg.AcceptDelay)
	n := p.counter.Add(1)
	ext := fmt.Sprintf("SIM-%d", n)
	prep := 25
	if err := p.accept(ctx, orderID, ext, prep); err != nil {
		log.Printf("accept order %s: %v", orderID, err)
		return
	}
	time.Sleep(p.cfg.CookDelay)
	if err := p.ready(ctx, orderID); err != nil {
		log.Printf("ready order %s: %v", orderID, err)
		return
	}
	time.Sleep(p.cfg.HandoverDelay)
	if err := p.handover(ctx, orderID); err != nil {
		log.Printf("handover order %s: %v", orderID, err)
	}
}

func (p *OrderProcessor) reject(ctx context.Context, orderID string) {
	id, err := uuid.Parse(orderID)
	if err != nil {
		log.Printf("reject order %s: invalid id: %v", orderID, err)
		return
	}
	resp, err := p.api.RejectOrderWithResponse(ctx, platformapi.OrderIdPath(id), platformapi.RejectOrderJSONRequestBody{
		Reason:     "out of ingredients",
		ReasonCode: "out_of_stock",
	})
	if err != nil {
		log.Printf("reject order %s: %v", orderID, err)
		return
	}
	if resp.StatusCode() >= 300 {
		log.Printf("reject order %s: status %d: %s", orderID, resp.StatusCode(), string(resp.Body))
	}
}

func (p *OrderProcessor) accept(ctx context.Context, orderID, externalID string, prepMinutes int) error {
	id, err := uuid.Parse(orderID)
	if err != nil {
		return err
	}
	resp, err := p.api.AcceptOrderWithResponse(ctx, platformapi.OrderIdPath(id), platformapi.AcceptOrderJSONRequestBody{
		ExternalOrderId: &externalID,
		PrepMinutes:     &prepMinutes,
	})
	if err != nil {
		return err
	}
	if resp.StatusCode() >= 300 {
		return fmt.Errorf("status %d: %s", resp.StatusCode(), string(resp.Body))
	}
	return nil
}

func (p *OrderProcessor) ready(ctx context.Context, orderID string) error {
	id, err := uuid.Parse(orderID)
	if err != nil {
		return err
	}
	resp, err := p.api.MarkOrderReadyWithResponse(ctx, platformapi.OrderIdPath(id))
	if err != nil {
		return err
	}
	if resp.StatusCode() >= 300 {
		return fmt.Errorf("status %d: %s", resp.StatusCode(), string(resp.Body))
	}
	return nil
}

func (p *OrderProcessor) handover(ctx context.Context, orderID string) error {
	id, err := uuid.Parse(orderID)
	if err != nil {
		return err
	}
	resp, err := p.api.HandoverOrderWithResponse(ctx, platformapi.OrderIdPath(id))
	if err != nil {
		return err
	}
	if resp.StatusCode() >= 300 {
		return fmt.Errorf("status %d: %s", resp.StatusCode(), string(resp.Body))
	}
	return nil
}
