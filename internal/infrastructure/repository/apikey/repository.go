package apikey

import (
	"context"
	"fmt"
	"time"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/repository/db"
	"github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/repository/pgerr"
)

type Repository struct {
	db.RepositoryBase
}

func New(pool *pgxpool.Pool, getter *trmpgx.CtxGetter) *Repository {
	return &Repository{RepositoryBase: db.RepositoryBase{Pool: pool, Getter: getter}}
}

type Record struct {
	ID         uuid.UUID
	VenueID    uuid.UUID
	KeyPrefix  string
	KeyHash    string
	Scopes     []string
	IsActive   bool
	LastUsedAt *time.Time
	RevokedAt  *time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (r *Repository) LookupByPrefix(ctx context.Context, prefix string) (*Record, error) {
	const q = `
		SELECT id, venue_id, key_prefix, key_hash, scopes, is_active, last_used_at, revoked_at, created_at, updated_at
		FROM venue_api_keys
		WHERE key_prefix = $1 AND is_active = true AND revoked_at IS NULL
		LIMIT 1`

	row := r.Q(ctx).QueryRow(ctx, q, prefix)
	var rec Record
	if err := row.Scan(
		&rec.ID, &rec.VenueID, &rec.KeyPrefix, &rec.KeyHash, &rec.Scopes,
		&rec.IsActive, &rec.LastUsedAt, &rec.RevokedAt, &rec.CreatedAt, &rec.UpdatedAt,
	); err != nil {
		return nil, pgerr.WrapNotFound(fmt.Errorf("lookup api key: %w", err))
	}
	return &rec, nil
}

func (r *Repository) Issue(ctx context.Context, venueID uuid.UUID, prefix, hash string, scopes []string) (*Record, error) {
	const q = `
		INSERT INTO venue_api_keys (venue_id, key_prefix, key_hash, scopes)
		VALUES ($1, $2, $3, $4)
		RETURNING id, venue_id, key_prefix, key_hash, scopes, is_active, last_used_at, revoked_at, created_at, updated_at`

	row := r.Q(ctx).QueryRow(ctx, q, venueID, prefix, hash, scopes)
	var rec Record
	if err := row.Scan(
		&rec.ID, &rec.VenueID, &rec.KeyPrefix, &rec.KeyHash, &rec.Scopes,
		&rec.IsActive, &rec.LastUsedAt, &rec.RevokedAt, &rec.CreatedAt, &rec.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("issue api key: %w", err)
	}
	return &rec, nil
}

func (r *Repository) Revoke(ctx context.Context, venueID, keyID uuid.UUID) (string, error) {
	const q = `
		UPDATE venue_api_keys
		SET is_active = false, revoked_at = now(), updated_at = now()
		WHERE id = $1 AND venue_id = $2 AND revoked_at IS NULL
		RETURNING key_prefix`

	var prefix string
	err := r.Q(ctx).QueryRow(ctx, q, keyID, venueID).Scan(&prefix)
	if err != nil {
		return "", pgerr.WrapNotFound(fmt.Errorf("revoke api key: %w", err))
	}
	return prefix, nil
}

func (r *Repository) ListByVenue(ctx context.Context, venueID uuid.UUID) ([]Record, error) {
	const q = `
		SELECT id, venue_id, key_prefix, key_hash, scopes, is_active, last_used_at, revoked_at, created_at, updated_at
		FROM venue_api_keys
		WHERE venue_id = $1
		ORDER BY created_at DESC`

	rows, err := r.Q(ctx).Query(ctx, q, venueID)
	if err != nil {
		return nil, fmt.Errorf("list api keys: %w", err)
	}
	defer rows.Close()

	var result []Record
	for rows.Next() {
		var rec Record
		if err := rows.Scan(
			&rec.ID, &rec.VenueID, &rec.KeyPrefix, &rec.KeyHash, &rec.Scopes,
			&rec.IsActive, &rec.LastUsedAt, &rec.RevokedAt, &rec.CreatedAt, &rec.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan api key: %w", err)
		}
		result = append(result, rec)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate api keys: %w", err)
	}
	return result, nil
}
