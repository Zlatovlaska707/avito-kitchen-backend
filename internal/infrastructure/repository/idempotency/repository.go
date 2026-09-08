package idempotency

import (
	"context"
	"fmt"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/repository/db"
	"github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/repository/pgerr"
	idempotencycontract "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/idempotency"
)

type Repository struct {
	db.RepositoryBase
}

func New(pool *pgxpool.Pool, getter *trmpgx.CtxGetter) *Repository {
	return &Repository{RepositoryBase: db.RepositoryBase{Pool: pool, Getter: getter}}
}

func (r *Repository) Claim(ctx context.Context, in idempotencycontract.ClaimIn) (idempotencycontract.ClaimResult, error) {
	const q = `
		INSERT INTO idempotency_keys (scope, key, request_hash, expires_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (scope, key) DO NOTHING`
	tag, err := r.Q(ctx).Exec(ctx, q, in.Scope, in.Key, in.RequestHash, in.ExpiresAt)
	if err != nil {
		return idempotencycontract.ClaimExists, fmt.Errorf("claim idempotency key: %w", err)
	}
	if tag.RowsAffected() == 1 {
		return idempotencycontract.ClaimGranted, nil
	}
	return idempotencycontract.ClaimExists, nil
}

func (r *Repository) Get(ctx context.Context, scope, key string) (*idempotencycontract.Record, error) {
	const q = `
		SELECT scope, key, request_hash, response_status, response_body, resource_id, created_at, expires_at
		FROM idempotency_keys WHERE scope = $1 AND key = $2`
	row := r.Q(ctx).QueryRow(ctx, q, scope, key)
	var rec idempotencycontract.Record
	if err := row.Scan(&rec.Scope, &rec.Key, &rec.RequestHash, &rec.ResponseStatus, &rec.ResponseBody, &rec.ResourceID, &rec.CreatedAt, &rec.ExpiresAt); err != nil {
		return nil, pgerr.WrapNotFound(fmt.Errorf("get idempotency key: %w", err))
	}
	return &rec, nil
}

func (r *Repository) SaveResponse(ctx context.Context, in idempotencycontract.SaveResponseIn) error {
	const q = `
		UPDATE idempotency_keys
		SET response_status = $3, response_body = $4, resource_id = $5
		WHERE scope = $1 AND key = $2`
	tag, err := r.Q(ctx).Exec(ctx, q, in.Scope, in.Key, in.ResponseStatus, in.ResponseBody, in.ResourceID)
	if err != nil {
		return fmt.Errorf("save idempotency response: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return pgerr.ErrNotFound
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, scope, key string) error {
	_, err := r.Q(ctx).Exec(ctx, `DELETE FROM idempotency_keys WHERE scope = $1 AND key = $2`, scope, key)
	if err != nil {
		return fmt.Errorf("delete idempotency key: %w", err)
	}
	return nil
}

func (r *Repository) DeleteExpired(ctx context.Context) (int, error) {
	tag, err := r.Q(ctx).Exec(ctx, `DELETE FROM idempotency_keys WHERE expires_at < now()`)
	if err != nil {
		return 0, fmt.Errorf("delete expired idempotency keys: %w", err)
	}
	return int(tag.RowsAffected()), nil
}
