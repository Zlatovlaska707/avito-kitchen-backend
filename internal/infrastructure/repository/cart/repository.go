package cart

import (
	"context"
	"fmt"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/repository/db"
	"github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/repository/pgerr"
	cartcontract "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/cart"
)

type Repository struct {
	db.RepositoryBase
}

func New(pool *pgxpool.Pool, getter *trmpgx.CtxGetter) *Repository {
	return &Repository{RepositoryBase: db.RepositoryBase{Pool: pool, Getter: getter}}
}

func (r *Repository) GetActiveForUpdate(ctx context.Context, customerID uuid.UUID) (*cartcontract.Cart, error) {
	const q = `
		SELECT id, customer_id, venue_id, status, order_id, expires_at, created_at, updated_at
		FROM carts
		WHERE customer_id = $1 AND status = 'active'
		FOR UPDATE`
	return r.scanCart(r.Q(ctx).QueryRow(ctx, q, customerID))
}

func (r *Repository) GetActive(ctx context.Context, customerID uuid.UUID) (*cartcontract.Cart, error) {
	const q = `
		SELECT id, customer_id, venue_id, status, order_id, expires_at, created_at, updated_at
		FROM carts
		WHERE customer_id = $1 AND status = 'active'`
	return r.scanCart(r.Q(ctx).QueryRow(ctx, q, customerID))
}

func (r *Repository) Create(ctx context.Context, in cartcontract.CreateIn) (*cartcontract.Cart, error) {
	const q = `
		INSERT INTO carts (customer_id, venue_id, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id, customer_id, venue_id, status, order_id, expires_at, created_at, updated_at`
	return r.scanCart(r.Q(ctx).QueryRow(ctx, q, in.CustomerID, in.VenueID, in.ExpiresAt))
}

func (r *Repository) GetItems(ctx context.Context, cartID uuid.UUID) ([]cartcontract.Item, error) {
	const q = `
		SELECT id, cart_id, menu_item_id, quantity, selected_options, options_hash, created_at, updated_at
		FROM cart_items WHERE cart_id = $1 ORDER BY created_at`
	rows, err := r.Q(ctx).Query(ctx, q, cartID)
	if err != nil {
		return nil, fmt.Errorf("get cart items: %w", err)
	}
	defer rows.Close()

	var items []cartcontract.Item
	for rows.Next() {
		var it cartcontract.Item
		if err := rows.Scan(&it.ID, &it.CartID, &it.MenuItemID, &it.Quantity, &it.SelectedOptions, &it.OptionsHash, &it.CreatedAt, &it.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan cart item: %w", err)
		}
		items = append(items, it)
	}
	return items, rows.Err()
}

func (r *Repository) AddItem(ctx context.Context, in cartcontract.AddItemIn) (*cartcontract.Item, error) {
	const q = `
		INSERT INTO cart_items (cart_id, menu_item_id, quantity, selected_options, options_hash)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (cart_id, menu_item_id, options_hash)
		DO UPDATE SET quantity = cart_items.quantity + EXCLUDED.quantity, updated_at = now()
		RETURNING id, cart_id, menu_item_id, quantity, selected_options, options_hash, created_at, updated_at`
	row := r.Q(ctx).QueryRow(ctx, q, in.CartID, in.MenuItemID, in.Quantity, in.SelectedOptions, in.OptionsHash)
	var it cartcontract.Item
	if err := row.Scan(&it.ID, &it.CartID, &it.MenuItemID, &it.Quantity, &it.SelectedOptions, &it.OptionsHash, &it.CreatedAt, &it.UpdatedAt); err != nil {
		return nil, fmt.Errorf("add cart item: %w", err)
	}
	return &it, nil
}

func (r *Repository) UpdateItem(ctx context.Context, in cartcontract.UpdateItemIn) (*cartcontract.Item, error) {
	const q = `
		UPDATE cart_items SET quantity = $3, updated_at = now()
		WHERE cart_id = $1 AND id = $2
		RETURNING id, cart_id, menu_item_id, quantity, selected_options, options_hash, created_at, updated_at`
	row := r.Q(ctx).QueryRow(ctx, q, in.CartID, in.ItemID, in.Quantity)
	var it cartcontract.Item
	if err := row.Scan(&it.ID, &it.CartID, &it.MenuItemID, &it.Quantity, &it.SelectedOptions, &it.OptionsHash, &it.CreatedAt, &it.UpdatedAt); err != nil {
		return nil, pgerr.WrapNotFound(fmt.Errorf("update cart item: %w", err))
	}
	return &it, nil
}

func (r *Repository) DeleteItem(ctx context.Context, cartID, itemID uuid.UUID) error {
	tag, err := r.Q(ctx).Exec(ctx, `DELETE FROM cart_items WHERE cart_id = $1 AND id = $2`, cartID, itemID)
	if err != nil {
		return fmt.Errorf("delete cart item: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return pgerr.ErrNotFound
	}
	return nil
}

func (r *Repository) ClearItems(ctx context.Context, cartID uuid.UUID) error {
	_, err := r.Q(ctx).Exec(ctx, `DELETE FROM cart_items WHERE cart_id = $1`, cartID)
	if err != nil {
		return fmt.Errorf("clear cart items: %w", err)
	}
	return nil
}

func (r *Repository) Abandon(ctx context.Context, cartID uuid.UUID) error {
	tag, err := r.Q(ctx).Exec(ctx,
		`UPDATE carts SET status = 'abandoned', updated_at = now() WHERE id = $1 AND status = 'active'`, cartID)
	if err != nil {
		return fmt.Errorf("abandon cart: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return pgerr.ErrNotFound
	}
	return nil
}

func (r *Repository) MarkConverted(ctx context.Context, cartID, orderID uuid.UUID) error {
	tag, err := r.Q(ctx).Exec(ctx,
		`UPDATE carts SET status = 'converted', order_id = $2, updated_at = now() WHERE id = $1`, cartID, orderID)
	if err != nil {
		return fmt.Errorf("mark cart converted: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return pgerr.ErrNotFound
	}
	return nil
}

func (r *Repository) CountItems(ctx context.Context, cartID uuid.UUID) (int, error) {
	var count int
	err := r.Q(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM cart_items WHERE cart_id = $1`, cartID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count cart items: %w", err)
	}
	return count, nil
}

func (r *Repository) AbandonExpired(ctx context.Context) (int, error) {
	tag, err := r.Q(ctx).Exec(ctx,
		`UPDATE carts SET status = 'abandoned', updated_at = now() WHERE status = 'active' AND expires_at < now()`)
	if err != nil {
		return 0, fmt.Errorf("abandon expired carts: %w", err)
	}
	return int(tag.RowsAffected()), nil
}

func (r *Repository) scanCart(row interface{ Scan(dest ...any) error }) (*cartcontract.Cart, error) {
	var c cartcontract.Cart
	if err := row.Scan(&c.ID, &c.CustomerID, &c.VenueID, &c.Status, &c.OrderID, &c.ExpiresAt, &c.CreatedAt, &c.UpdatedAt); err != nil {
		return nil, pgerr.WrapNotFound(fmt.Errorf("scan cart: %w", err))
	}
	return &c, nil
}
