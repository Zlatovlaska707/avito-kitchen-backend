package outbox

import (
	"context"
	"fmt"
	"time"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/repository/db"
	"github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/repository/pgerr"
	outboxcontract "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/outbox"
)

type Repository struct {
	db.RepositoryBase
}

func New(pool *pgxpool.Pool, getter *trmpgx.CtxGetter) *Repository {
	return &Repository{RepositoryBase: db.RepositoryBase{Pool: pool, Getter: getter}}
}

func (r *Repository) Publish(ctx context.Context, in outboxcontract.PublishIn) (int64, error) {
	const q = `
		INSERT INTO outbox_events (aggregate_type, aggregate_id, venue_id, event_type, payload)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`
	var id int64
	if err := r.Q(ctx).QueryRow(ctx, q, in.AggregateType, in.AggregateID, in.VenueID, in.EventType, in.Payload).Scan(&id); err != nil {
		return 0, fmt.Errorf("publish outbox event: %w", err)
	}
	return id, nil
}

func (r *Repository) FetchPendingWithLease(ctx context.Context, batchSize int, lease time.Duration) ([]outboxcontract.Event, error) {
	const q = `
		WITH picked AS (
			SELECT id FROM outbox_events
			WHERE status = 'pending' AND next_attempt_at <= now()
			ORDER BY next_attempt_at, id
			LIMIT $1
			FOR UPDATE SKIP LOCKED
		)
		UPDATE outbox_events e
		SET next_attempt_at = now() + $2::interval
		FROM picked
		WHERE e.id = picked.id
		RETURNING e.id, e.aggregate_type, e.aggregate_id, e.venue_id, e.event_type,
			e.payload, e.status, e.attempts, e.next_attempt_at, e.created_at`

	rows, err := r.Q(ctx).Query(ctx, q, batchSize, fmt.Sprintf("%f seconds", lease.Seconds()))
	if err != nil {
		return nil, fmt.Errorf("fetch pending outbox: %w", err)
	}
	defer rows.Close()

	var events []outboxcontract.Event
	for rows.Next() {
		var e outboxcontract.Event
		if err := rows.Scan(&e.ID, &e.AggregateType, &e.AggregateID, &e.VenueID, &e.EventType, &e.Payload, &e.Status, &e.Attempts, &e.NextAttemptAt, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan outbox event: %w", err)
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

func (r *Repository) MarkPublished(ctx context.Context, id int64) error {
	_, err := r.Q(ctx).Exec(ctx,
		`UPDATE outbox_events SET status = 'published', published_at = now() WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("mark outbox published: %w", err)
	}
	return nil
}

func (r *Repository) MarkFailed(ctx context.Context, id int64, attempts int, nextAttemptAt time.Time, lastError string) error {
	status := "pending"
	if attempts >= 5 {
		status = "failed"
	}
	_, err := r.Q(ctx).Exec(ctx,
		`UPDATE outbox_events SET status = $2, attempts = $3, next_attempt_at = $4, last_error = $5 WHERE id = $1`,
		id, status, attempts, nextAttemptAt, lastError)
	if err != nil {
		return fmt.Errorf("mark outbox failed: %w", err)
	}
	return nil
}

func (r *Repository) MarkSkipped(ctx context.Context, id int64) error {
	_, err := r.Q(ctx).Exec(ctx, `UPDATE outbox_events SET status = 'skipped' WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("mark outbox skipped: %w", err)
	}
	return nil
}

func (r *Repository) RecordDelivery(ctx context.Context, in outboxcontract.DeliveryIn) error {
	const q = `
		INSERT INTO webhook_deliveries (outbox_event_id, webhook_id, attempt, http_status, duration_ms, error)
		VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.Q(ctx).Exec(ctx, q, in.OutboxEventID, in.WebhookID, in.Attempt, in.HTTPStatus, in.DurationMS, in.Error)
	if err != nil {
		return fmt.Errorf("record webhook delivery: %w", err)
	}
	return nil
}

func (r *Repository) GetByID(ctx context.Context, id int64) (*outboxcontract.Event, error) {
	const q = `
		SELECT id, aggregate_type, aggregate_id, venue_id, event_type, payload, status, attempts, next_attempt_at, created_at
		FROM outbox_events WHERE id = $1`
	row := r.Q(ctx).QueryRow(ctx, q, id)
	var e outboxcontract.Event
	if err := row.Scan(&e.ID, &e.AggregateType, &e.AggregateID, &e.VenueID, &e.EventType, &e.Payload, &e.Status, &e.Attempts, &e.NextAttemptAt, &e.CreatedAt); err != nil {
		return nil, pgerr.WrapNotFound(fmt.Errorf("get outbox event: %w", err))
	}
	return &e, nil
}
