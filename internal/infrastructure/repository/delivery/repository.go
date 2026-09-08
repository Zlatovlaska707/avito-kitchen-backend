package delivery

import (
	"context"
	"fmt"
	"time"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/repository/db"
	"github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/repository/pgerr"
	deliverycontract "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/delivery"
)

type Repository struct {
	db.RepositoryBase
}

func New(pool *pgxpool.Pool, getter *trmpgx.CtxGetter) *Repository {
	return &Repository{RepositoryBase: db.RepositoryBase{Pool: pool, Getter: getter}}
}

func (r *Repository) Create(ctx context.Context, in deliverycontract.CreateIn) (*deliverycontract.Delivery, error) {
	const q = `
		INSERT INTO deliveries (order_id, status, eta_minutes)
		VALUES ($1, 'pending', $2)
		RETURNING id, order_id, status, courier_name, courier_phone, eta_minutes,
			assigned_at, picked_up_at, delivered_at, created_at, updated_at`
	return scanDelivery(r.Q(ctx).QueryRow(ctx, q, in.OrderID, in.ETAMinutes))
}

func (r *Repository) UpdateStatus(ctx context.Context, in deliverycontract.UpdateStatusIn) (*deliverycontract.Delivery, error) {
	const q = `
		UPDATE deliveries
		SET status = $2::varchar(16),
			courier_name = COALESCE($3, courier_name),
			courier_phone = COALESCE($4, courier_phone),
			picked_up_at = COALESCE($5, picked_up_at),
			delivered_at = COALESCE($6, delivered_at),
			assigned_at = CASE WHEN $2::varchar(16) = 'assigned' AND assigned_at IS NULL THEN now() ELSE assigned_at END,
			updated_at = now()
		WHERE order_id = $1
		RETURNING id, order_id, status, courier_name, courier_phone, eta_minutes,
			assigned_at, picked_up_at, delivered_at, created_at, updated_at`
	return scanDelivery(r.Q(ctx).QueryRow(ctx, q, in.OrderID, in.Status, in.CourierName, in.CourierPhone, in.PickedUpAt, in.DeliveredAt))
}

func (r *Repository) GetByOrderID(ctx context.Context, orderID uuid.UUID) (*deliverycontract.Delivery, error) {
	const q = `
		SELECT id, order_id, status, courier_name, courier_phone, eta_minutes,
			assigned_at, picked_up_at, delivered_at, created_at, updated_at
		FROM deliveries WHERE order_id = $1`
	return scanDelivery(r.Q(ctx).QueryRow(ctx, q, orderID))
}

func (r *Repository) FindPendingForAssign(ctx context.Context, before time.Time, limit int) ([]deliverycontract.Delivery, error) {
	const q = `
		SELECT d.id, d.order_id, d.status, d.courier_name, d.courier_phone, d.eta_minutes,
			d.assigned_at, d.picked_up_at, d.delivered_at, d.created_at, d.updated_at
		FROM deliveries d
		JOIN orders o ON o.id = d.order_id
		WHERE d.status = 'pending' AND o.ready_at IS NOT NULL AND o.ready_at <= $1
		LIMIT $2`
	return r.scanDeliveries(ctx, q, before, limit)
}

func (r *Repository) FindAssignedForTransit(ctx context.Context, before time.Time, limit int) ([]deliverycontract.Delivery, error) {
	const q = `
		SELECT id, order_id, status, courier_name, courier_phone, eta_minutes,
			assigned_at, picked_up_at, delivered_at, created_at, updated_at
		FROM deliveries
		WHERE status = 'assigned' AND assigned_at IS NOT NULL AND assigned_at <= $1
		LIMIT $2`
	return r.scanDeliveries(ctx, q, before, limit)
}

func (r *Repository) scanDeliveries(ctx context.Context, q string, before time.Time, limit int) ([]deliverycontract.Delivery, error) {
	rows, err := r.Q(ctx).Query(ctx, q, before, limit)
	if err != nil {
		return nil, fmt.Errorf("scan deliveries: %w", err)
	}
	defer rows.Close()

	var result []deliverycontract.Delivery
	for rows.Next() {
		d, err := scanDelivery(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, *d)
	}
	return result, rows.Err()
}

type scannable interface {
	Scan(dest ...any) error
}

func scanDelivery(row scannable) (*deliverycontract.Delivery, error) {
	var d deliverycontract.Delivery
	if err := row.Scan(&d.ID, &d.OrderID, &d.Status, &d.CourierName, &d.CourierPhone, &d.ETAMinutes, &d.AssignedAt, &d.PickedUpAt, &d.DeliveredAt, &d.CreatedAt, &d.UpdatedAt); err != nil {
		return nil, pgerr.WrapNotFound(fmt.Errorf("scan delivery: %w", err))
	}
	return &d, nil
}
