package polling

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/zlatovlaska707/avito-kitchen/venue-simulator/internal/config"
	"github.com/zlatovlaska707/avito-kitchen/venue-simulator/internal/platformapi"
	"github.com/zlatovlaska707/avito-kitchen/venue-simulator/internal/webhook"
)

type Poller struct {
	api *platformapi.ClientWithResponses
	cfg config.Config
	wh  *webhook.Receiver
}

func New(cfg config.Config, wh *webhook.Receiver) (*Poller, error) {
	api, err := platformapi.NewPartnerClient(cfg.PlatformURL, cfg.APIKey, 10*time.Second)
	if err != nil {
		return nil, err
	}
	return &Poller{api: api, cfg: cfg, wh: wh}, nil
}

func (p *Poller) Run(ctx context.Context) {
	ticker := time.NewTicker(p.cfg.PollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.pollOnce(ctx)
		}
	}
}

func (p *Poller) pollOnce(ctx context.Context) {
	status := platformapi.OrderStatusMultiQuery{"placed"}
	resp, err := p.api.ListPartnerOrdersWithResponse(ctx, &platformapi.ListPartnerOrdersParams{
		Status: &status,
	})
	if err != nil {
		log.Printf("poll orders: %v", err)
		return
	}
	if resp.StatusCode() != http.StatusOK || resp.JSON200 == nil {
		return
	}
	for _, item := range resp.JSON200.Items {
		log.Printf("poll found placed order %s", item.Id)
		go p.wh.ProcessPolledOrder(item.Id.String())
	}
}
