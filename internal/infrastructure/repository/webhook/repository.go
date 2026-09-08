package webhook

import (
	"context"
	"fmt"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/repository/db"
	"github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/repository/pgerr"
	webhookcontract "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/webhook"
)

type Repository struct {
	db.RepositoryBase
}

func New(pool *pgxpool.Pool, getter *trmpgx.CtxGetter) *Repository {
	return &Repository{RepositoryBase: db.RepositoryBase{Pool: pool, Getter: getter}}
}

func (r *Repository) Create(ctx context.Context, in webhookcontract.CreateIn) (*webhookcontract.Webhook, error) {
	const q = `
		INSERT INTO venue_webhooks (venue_id, url, secret, event_types)
		VALUES ($1, $2, $3, $4)
		RETURNING id, venue_id, url, secret, event_types, is_active, created_at, updated_at`
	row := r.Q(ctx).QueryRow(ctx, q, in.VenueID, in.URL, in.Secret, in.EventTypes)
	return scanWebhook(row)
}

func (r *Repository) Delete(ctx context.Context, venueID, webhookID uuid.UUID) error {
	tag, err := r.Q(ctx).Exec(ctx, `DELETE FROM venue_webhooks WHERE id = $1 AND venue_id = $2`, webhookID, venueID)
	if err != nil {
		return fmt.Errorf("delete webhook: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return pgerr.ErrNotFound
	}
	return nil
}

func (r *Repository) GetByID(ctx context.Context, venueID, webhookID uuid.UUID) (*webhookcontract.Webhook, error) {
	const q = `
		SELECT id, venue_id, url, secret, event_types, is_active, created_at, updated_at
		FROM venue_webhooks WHERE id = $1 AND venue_id = $2`
	return scanWebhook(r.Q(ctx).QueryRow(ctx, q, webhookID, venueID))
}

func (r *Repository) ListByVenue(ctx context.Context, venueID uuid.UUID) ([]webhookcontract.Webhook, error) {
	const q = `
		SELECT id, venue_id, url, secret, event_types, is_active, created_at, updated_at
		FROM venue_webhooks WHERE venue_id = $1 ORDER BY created_at DESC`
	rows, err := r.Q(ctx).Query(ctx, q, venueID)
	if err != nil {
		return nil, fmt.Errorf("list webhooks: %w", err)
	}
	defer rows.Close()

	var result []webhookcontract.Webhook
	for rows.Next() {
		w, err := scanWebhook(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, *w)
	}
	return result, rows.Err()
}

func (r *Repository) FindActiveByVenueAndEvent(ctx context.Context, venueID uuid.UUID, eventType string) ([]webhookcontract.Webhook, error) {
	const q = `
		SELECT id, venue_id, url, secret, event_types, is_active, created_at, updated_at
		FROM venue_webhooks
		WHERE venue_id = $1 AND is_active = true AND $2 = ANY(event_types)`
	rows, err := r.Q(ctx).Query(ctx, q, venueID, eventType)
	if err != nil {
		return nil, fmt.Errorf("find active webhooks: %w", err)
	}
	defer rows.Close()

	var result []webhookcontract.Webhook
	for rows.Next() {
		w, err := scanWebhook(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, *w)
	}
	return result, rows.Err()
}

type scannable interface {
	Scan(dest ...any) error
}

func scanWebhook(row scannable) (*webhookcontract.Webhook, error) {
	var w webhookcontract.Webhook
	if err := row.Scan(&w.ID, &w.VenueID, &w.URL, &w.Secret, &w.EventTypes, &w.IsActive, &w.CreatedAt, &w.UpdatedAt); err != nil {
		return nil, pgerr.WrapNotFound(fmt.Errorf("scan webhook: %w", err))
	}
	return &w, nil
}
