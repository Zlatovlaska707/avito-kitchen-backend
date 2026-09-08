package order

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Masterminds/squirrel"
	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/zlatovlaska707/avito-kitchen/internal/domain/orderstatus"
	"github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/repository/db"
	"github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/repository/pgerr"
	ordercontract "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/order"
)

type Repository struct {
	db.RepositoryBase
}

func New(pool *pgxpool.Pool, getter *trmpgx.CtxGetter) *Repository {
	return &Repository{RepositoryBase: db.RepositoryBase{Pool: pool, Getter: getter}}
}

var psql = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

func (r *Repository) Create(ctx context.Context, in ordercontract.CreateIn) (*ordercontract.Order, error) {
	publicNumber, err := r.nextPublicNumber(ctx)
	if err != nil {
		return nil, err
	}

	const orderQ = `
		INSERT INTO orders (
			public_number, customer_id, venue_id, cart_id, status, fulfillment_type,
			delivery_address, contact_phone, items_total_minor, delivery_fee_minor, discount_minor,
			total_minor, payment_method, payment_status, comment, prep_minutes, eta_at, placed_at
		) VALUES (
			$1,$2,$3,$4,'placed',$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,now()
		) RETURNING id, public_number, customer_id, venue_id, cart_id, status, fulfillment_type,
			delivery_address, contact_phone, items_total_minor, delivery_fee_minor, discount_minor,
			total_minor, currency, payment_method, payment_status, comment, external_order_id,
			prep_minutes, eta_at, placed_at, accepted_at, ready_at, delivered_at, closed_at,
			cancel_reason, cancelled_by, version, created_at, updated_at`

	row := r.Q(ctx).QueryRow(ctx, orderQ,
		publicNumber, in.CustomerID, in.VenueID, in.CartID, in.FulfillmentType,
		nullJSON(in.DeliveryAddress), in.ContactPhone, in.ItemsTotalMinor, in.DeliveryFeeMinor,
		in.DiscountMinor, in.TotalMinor, in.PaymentMethod, in.PaymentStatus, in.Comment,
		in.PrepMinutes, in.ETAAt,
	)
	o, err := scanOrder(row)
	if err != nil {
		return nil, fmt.Errorf("insert order: %w", err)
	}

	if err := r.insertOrderItems(ctx, o.ID, in.Items); err != nil {
		return nil, err
	}

	const histQ = `
		INSERT INTO order_status_history (order_id, from_status, to_status, actor_type, actor_id)
		VALUES ($1, NULL, 'placed', 'customer', $2)`
	if _, err := r.Q(ctx).Exec(ctx, histQ, o.ID, in.CustomerID.String()); err != nil {
		return nil, fmt.Errorf("insert order history: %w", err)
	}

	const outboxQ = `
		INSERT INTO outbox_events (aggregate_type, aggregate_id, venue_id, event_type, payload)
		VALUES ('order', $1, $2, 'order.created', $3)`
	if _, err := r.Q(ctx).Exec(ctx, outboxQ, o.ID, in.VenueID, in.OutboxPayload); err != nil {
		return nil, fmt.Errorf("insert outbox: %w", err)
	}

	if _, err := r.Q(ctx).Exec(ctx,
		`UPDATE carts SET status = 'converted', order_id = $2, updated_at = now() WHERE id = $1`,
		in.CartID, o.ID); err != nil {
		return nil, fmt.Errorf("convert cart: %w", err)
	}

	if in.CreateDelivery {
		const delQ = `INSERT INTO deliveries (order_id, status, eta_minutes) VALUES ($1, 'pending', $2)`
		if _, err := r.Q(ctx).Exec(ctx, delQ, o.ID, in.DeliveryETAMin); err != nil {
			return nil, fmt.Errorf("insert delivery: %w", err)
		}
	}

	return o, nil
}

func (r *Repository) insertOrderItems(ctx context.Context, orderID uuid.UUID, items []ordercontract.CreateItemIn) error {
	if len(items) == 0 {
		return nil
	}
	var b strings.Builder
	b.WriteString(`INSERT INTO order_items (
		order_id, menu_item_id, item_external_id, name_snapshot, unit_price_minor,
		quantity, options_snapshot, line_total_minor) VALUES `)
	args := make([]any, 0, len(items)*8)
	for i, it := range items {
		if i > 0 {
			b.WriteString(",")
		}
		base := i*8 + 1
		fmt.Fprintf(&b, "($%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d)", base, base+1, base+2, base+3, base+4, base+5, base+6, base+7)
		args = append(args, orderID, it.MenuItemID, it.ItemExternalID, it.NameSnapshot, it.UnitPriceMinor, it.Quantity, it.OptionsSnapshot, it.LineTotalMinor)
	}
	_, err := r.Q(ctx).Exec(ctx, b.String(), args...)
	if err != nil {
		return fmt.Errorf("insert order items: %w", err)
	}
	return nil
}

func (r *Repository) nextPublicNumber(ctx context.Context) (string, error) {
	var seq int64
	if err := r.Q(ctx).QueryRow(ctx, `SELECT nextval('order_public_number_seq')`).Scan(&seq); err != nil {
		return "", fmt.Errorf("next public number: %w", err)
	}
	now := time.Now().UTC()
	return fmt.Sprintf("AK-%s-%d", now.Format("060102"), seq), nil
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*ordercontract.Order, error) {
	return r.getOne(ctx, psql.Select(orderColumns...).From("orders").Where(squirrel.Eq{"id": id}))
}

func (r *Repository) GetByIDForCustomer(ctx context.Context, id, customerID uuid.UUID) (*ordercontract.Order, error) {
	return r.getOne(ctx, psql.Select(orderColumns...).From("orders").Where(squirrel.Eq{"id": id, "customer_id": customerID}))
}

func (r *Repository) GetByIDForVenue(ctx context.Context, id, venueID uuid.UUID) (*ordercontract.Order, error) {
	return r.getOne(ctx, psql.Select(orderColumns...).From("orders").Where(squirrel.Eq{"id": id, "venue_id": venueID}))
}

func (r *Repository) List(ctx context.Context, filter ordercontract.ListFilter) ([]ordercontract.Order, error) {
	b := psql.Select(orderColumns...).From("orders").OrderBy("created_at DESC", "id DESC")
	if filter.CustomerID != nil {
		b = b.Where(squirrel.Eq{"customer_id": *filter.CustomerID})
	}
	if filter.VenueID != nil {
		b = b.Where(squirrel.Eq{"venue_id": *filter.VenueID})
	}
	if filter.Status != nil {
		b = b.Where(squirrel.Eq{"status": *filter.Status})
	}
	if filter.CursorTS != nil && filter.CursorID != nil {
		b = b.Where(squirrel.Expr("(created_at, id) < (?, ?)", *filter.CursorTS, *filter.CursorID))
	}
	if filter.Limit > 0 {
		b = b.Limit(uint64(filter.Limit))
	}
	if filter.Offset > 0 {
		b = b.Offset(uint64(filter.Offset))
	}

	sql, args, err := b.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build list orders: %w", err)
	}
	rows, err := r.Q(ctx).Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("list orders: %w", err)
	}
	defer rows.Close()

	var result []ordercontract.Order
	for rows.Next() {
		o, err := scanOrder(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, *o)
	}
	return result, rows.Err()
}

func (r *Repository) GetItems(ctx context.Context, orderID uuid.UUID) ([]ordercontract.Item, error) {
	const q = `
		SELECT id, order_id, menu_item_id, item_external_id, name_snapshot, unit_price_minor,
			quantity, options_snapshot, line_total_minor, created_at
		FROM order_items WHERE order_id = $1 ORDER BY created_at`
	rows, err := r.Q(ctx).Query(ctx, q, orderID)
	if err != nil {
		return nil, fmt.Errorf("get order items: %w", err)
	}
	defer rows.Close()

	var items []ordercontract.Item
	for rows.Next() {
		var it ordercontract.Item
		if err := rows.Scan(&it.ID, &it.OrderID, &it.MenuItemID, &it.ItemExternalID, &it.NameSnapshot, &it.UnitPriceMinor, &it.Quantity, &it.OptionsSnapshot, &it.LineTotalMinor, &it.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan order item: %w", err)
		}
		items = append(items, it)
	}
	return items, rows.Err()
}

func (r *Repository) GetStatusHistory(ctx context.Context, orderID uuid.UUID) ([]ordercontract.StatusEvent, error) {
	const q = `
		SELECT from_status, to_status, actor_type, actor_id, reason, created_at
		FROM order_status_history WHERE order_id = $1 ORDER BY created_at, id`
	rows, err := r.Q(ctx).Query(ctx, q, orderID)
	if err != nil {
		return nil, fmt.Errorf("get order status history: %w", err)
	}
	defer rows.Close()

	events := make([]ordercontract.StatusEvent, 0)
	for rows.Next() {
		var (
			fromRaw, actorID, reason *string
			toRaw, actorType         string
			createdAt                time.Time
		)
		if err := rows.Scan(&fromRaw, &toRaw, &actorType, &actorID, &reason, &createdAt); err != nil {
			return nil, fmt.Errorf("scan status history: %w", err)
		}
		ev := ordercontract.StatusEvent{
			ToStatus:  orderstatus.Status(toRaw),
			ActorType: orderstatus.ActorType(actorType),
			ActorID:   actorID,
			Reason:    reason,
			CreatedAt: createdAt,
		}
		if fromRaw != nil {
			st := orderstatus.Status(*fromRaw)
			ev.FromStatus = &st
		}
		events = append(events, ev)
	}
	return events, rows.Err()
}

func (r *Repository) ApplyTransition(ctx context.Context, in ordercontract.TransitionIn) (*ordercontract.Order, error) {
	setParts := []string{"status = $3", "version = version + 1", "updated_at = now()"}
	args := []any{in.OrderID, in.FromStatus, in.ToStatus}
	argN := 4

	switch in.ToStatus {
	case orderstatus.StatusAccepted:
		setParts = append(setParts, "accepted_at = now()")
		if in.PrepMinutes != nil {
			setParts = append(setParts, fmt.Sprintf("prep_minutes = $%d", argN))
			args = append(args, *in.PrepMinutes)
			argN++
		}
		if in.ETAAt != nil {
			setParts = append(setParts, fmt.Sprintf("eta_at = $%d", argN))
			args = append(args, *in.ETAAt)
		}
	case orderstatus.StatusReady:
		setParts = append(setParts, "ready_at = now()")
	case orderstatus.StatusInDelivery:
		setParts = append(setParts, "ready_at = COALESCE(ready_at, now())")
	case orderstatus.StatusCompleted:
		setParts = append(setParts, "delivered_at = now()", "closed_at = now()")
	case orderstatus.StatusCancelled, orderstatus.StatusRejected:
		setParts = append(setParts, "closed_at = now()")
		if in.Reason != nil {
			setParts = append(setParts, fmt.Sprintf("cancel_reason = $%d", argN))
			args = append(args, *in.Reason)
			argN++
		}
		if in.CancelledBy != nil {
			setParts = append(setParts, fmt.Sprintf("cancelled_by = $%d", argN))
			args = append(args, *in.CancelledBy)
		}
	}

	q := fmt.Sprintf(`UPDATE orders SET %s WHERE id = $1 AND status = $2
		RETURNING %s`, strings.Join(setParts, ", "), strings.Join(orderColumns, ", "))
	row := r.Q(ctx).QueryRow(ctx, q, args...)
	o, err := scanOrder(row)
	if err != nil {
		return nil, pgerr.WrapNotFound(fmt.Errorf("apply transition: %w", err))
	}

	const histQ = `
		INSERT INTO order_status_history (order_id, from_status, to_status, actor_type, actor_id, reason)
		VALUES ($1, $2, $3, $4, $5, $6)`
	if _, err := r.Q(ctx).Exec(ctx, histQ, in.OrderID, in.FromStatus, in.ToStatus, in.ActorType, in.ActorID, in.Reason); err != nil {
		return nil, fmt.Errorf("insert status history: %w", err)
	}

	if in.OutboxEventType != "" {
		const outboxQ = `
			INSERT INTO outbox_events (aggregate_type, aggregate_id, venue_id, event_type, payload)
			VALUES ('order', $1, $2, $3, $4)`
		venueID := o.VenueID
		if in.VenueID != nil {
			venueID = *in.VenueID
		}
		if _, err := r.Q(ctx).Exec(ctx, outboxQ, o.ID, venueID, in.OutboxEventType, in.OutboxPayload); err != nil {
			return nil, fmt.Errorf("insert transition outbox: %w", err)
		}
	}

	if in.ToStatus == orderstatus.StatusReady && in.FulfillmentType == orderstatus.FulfillmentDelivery {
		_, _ = r.Q(ctx).Exec(ctx, `UPDATE deliveries SET status = 'pending', updated_at = now() WHERE order_id = $1`, o.ID)
	}

	return o, nil
}

func (r *Repository) FindPlacedBefore(ctx context.Context, before time.Time, limit int) ([]uuid.UUID, error) {
	const q = `SELECT id FROM orders WHERE status = 'placed' AND placed_at < $1 LIMIT $2`
	rows, err := r.Q(ctx).Query(ctx, q, before, limit)
	if err != nil {
		return nil, fmt.Errorf("find placed before: %w", err)
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (r *Repository) getOne(ctx context.Context, b squirrel.SelectBuilder) (*ordercontract.Order, error) {
	sql, args, err := b.ToSql()
	if err != nil {
		return nil, err
	}
	return scanOrder(r.Q(ctx).QueryRow(ctx, sql, args...))
}

var orderColumns = []string{
	"id", "public_number", "customer_id", "venue_id", "cart_id", "status", "fulfillment_type",
	"delivery_address", "contact_phone", "items_total_minor", "delivery_fee_minor", "discount_minor",
	"total_minor", "currency", "payment_method", "payment_status", "comment", "external_order_id",
	"prep_minutes", "eta_at", "placed_at", "accepted_at", "ready_at", "delivered_at", "closed_at",
	"cancel_reason", "cancelled_by", "version", "created_at", "updated_at",
}

type scannable interface {
	Scan(dest ...any) error
}

func scanOrder(row scannable) (*ordercontract.Order, error) {
	var o ordercontract.Order
	var status, fulfillment string
	if err := row.Scan(
		&o.ID, &o.PublicNumber, &o.CustomerID, &o.VenueID, &o.CartID, &status, &fulfillment,
		&o.DeliveryAddress, &o.ContactPhone, &o.ItemsTotalMinor, &o.DeliveryFeeMinor, &o.DiscountMinor,
		&o.TotalMinor, &o.Currency, &o.PaymentMethod, &o.PaymentStatus, &o.Comment, &o.ExternalOrderID,
		&o.PrepMinutes, &o.ETAAt, &o.PlacedAt, &o.AcceptedAt, &o.ReadyAt, &o.DeliveredAt, &o.ClosedAt,
		&o.CancelReason, &o.CancelledBy, &o.Version, &o.CreatedAt, &o.UpdatedAt,
	); err != nil {
		return nil, err
	}
	o.Status = orderstatus.Status(status)
	o.FulfillmentType = orderstatus.FulfillmentType(fulfillment)
	return &o, nil
}

func nullJSON(b []byte) any {
	if len(b) == 0 {
		return nil
	}
	return b
}
