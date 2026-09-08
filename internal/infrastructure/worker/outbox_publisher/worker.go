package outbox_publisher

import (
	"context"
	"fmt"
	"time"

	"github.com/zlatovlaska707/avito-kitchen/internal/config"
	"github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/webhook"
	outboxrepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/outbox"
	webhookrepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/webhook"
)

var backoff = []time.Duration{time.Second, 5 * time.Second, 30 * time.Second, 2 * time.Minute, 10 * time.Minute}

type Worker struct {
	outbox   outboxrepo.Repository
	webhooks webhookrepo.Repository
	sender   *webhook.Sender
	cfg      config.WebhookConfig
	batch    int
	lease    time.Duration
	interval time.Duration
}

func New(outbox outboxrepo.Repository, webhooks webhookrepo.Repository, sender *webhook.Sender, cfg config.Config) *Worker {
	return &Worker{
		outbox: outbox, webhooks: webhooks, sender: sender, cfg: cfg.Webhook,
		batch: cfg.Outbox.BatchSize, lease: 30 * time.Second, interval: cfg.Outbox.PollInterval,
	}
}

func (w *Worker) Name() string { return "outbox_publisher" }

func (w *Worker) Interval() time.Duration { return w.interval }

func (w *Worker) RunOnce(ctx context.Context) (int, error) {
	events, err := w.outbox.FetchPendingWithLease(ctx, w.batch, w.lease)
	if err != nil {
		return 0, err
	}
	processed := 0
	for _, e := range events {
		if err := w.publishOne(ctx, e); err != nil {
			continue
		}
		processed++
	}
	return processed, nil
}

func (w *Worker) publishOne(ctx context.Context, e outboxrepo.Event) error {
	if e.VenueID == nil {
		return w.outbox.MarkSkipped(ctx, e.ID)
	}
	hooks, err := w.webhooks.FindActiveByVenueAndEvent(ctx, *e.VenueID, e.EventType)
	if err != nil {
		return err
	}
	if len(hooks) == 0 {
		return w.outbox.MarkSkipped(ctx, e.ID)
	}
	var lastErr error
	for _, h := range hooks {
		res, err := w.sender.Send(ctx, webhook.SendRequest{
			URL: h.URL, Secret: h.Secret, EventType: e.EventType,
			DeliveryID: fmt.Sprintf("%d", e.ID), Payload: e.Payload, Timeout: w.cfg.Timeout,
		})
		status := 0
		if res != nil {
			status = res.StatusCode
		}
		dur := 0
		if res != nil {
			dur = int(res.Duration.Milliseconds())
		}
		var errStr *string
		if err != nil {
			s := err.Error()
			errStr = &s
		}
		whID := h.ID
		_ = w.outbox.RecordDelivery(ctx, outboxrepo.DeliveryIn{
			OutboxEventID: e.ID, WebhookID: &whID, Attempt: e.Attempts + 1,
			HTTPStatus: &status, DurationMS: &dur, Error: errStr,
		})
		if err != nil {
			lastErr = err
			continue
		}
		if status >= 200 && status < 300 {
			return w.outbox.MarkPublished(ctx, e.ID)
		}
		lastErr = fmt.Errorf("webhook status %d", status)
	}
	attempts := e.Attempts + 1
	next := time.Now().Add(backoffAt(attempts))
	if attempts >= w.cfg.MaxAttempts {
		return w.outbox.MarkFailed(ctx, e.ID, attempts, next, fmt.Sprintf("%v", lastErr))
	}
	return w.outbox.MarkFailed(ctx, e.ID, attempts, next, fmt.Sprintf("%v", lastErr))
}

func backoffAt(attempt int) time.Duration {
	if attempt <= 0 {
		return backoff[0]
	}
	if attempt > len(backoff) {
		return backoff[len(backoff)-1]
	}
	return backoff[attempt-1]
}
