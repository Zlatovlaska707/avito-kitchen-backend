package catalog

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/repository/db"
	"github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/repository/pgerr"
	catalogcontract "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/catalog"
)

type Repository struct {
	db.RepositoryBase
}

func New(pool *pgxpool.Pool, getter *trmpgx.CtxGetter) *Repository {
	return &Repository{RepositoryBase: db.RepositoryBase{Pool: pool, Getter: getter}}
}

func (r *Repository) GetMenu(ctx context.Context, venueID uuid.UUID) (*catalogcontract.Menu, string, error) {
	cats, err := r.loadCategories(ctx, venueID)
	if err != nil {
		return nil, "", err
	}
	items, err := r.loadItems(ctx, venueID)
	if err != nil {
		return nil, "", err
	}

	itemMap := make(map[uuid.UUID][]catalogcontract.MenuItem)
	var maxUpdated time.Time
	for _, it := range items {
		groups, err := r.loadOptionGroups(ctx, it.ID)
		if err != nil {
			return nil, "", err
		}
		it.OptionGroups = groups
		itemMap[it.CategoryID] = append(itemMap[it.CategoryID], it)
		if it.UpdatedAt.After(maxUpdated) {
			maxUpdated = it.UpdatedAt
		}
	}

	menu := &catalogcontract.Menu{VenueID: venueID, UpdatedAt: maxUpdated}
	for _, c := range cats {
		c.Items = itemMap[c.ID]
		menu.Categories = append(menu.Categories, c)
	}

	etag := computeETag(menu)
	return menu, etag, nil
}

func (r *Repository) TryMenuSyncLock(ctx context.Context, venueID uuid.UUID) (bool, error) {
	var ok bool
	err := r.Q(ctx).QueryRow(ctx, `SELECT pg_try_advisory_xact_lock(hashtext('menu_sync:' || $1::text))`, venueID).Scan(&ok)
	if err != nil {
		return false, fmt.Errorf("menu sync lock: %w", err)
	}
	return ok, nil
}

func (r *Repository) SyncMenu(ctx context.Context, in catalogcontract.SyncIn) (*catalogcontract.SyncResult, error) {
	result := &catalogcontract.SyncResult{}
	catExtToID := make(map[string]uuid.UUID)
	itemExtSeen := make(map[string]struct{})

	for _, cat := range in.Categories {
		var catID uuid.UUID
		const catQ = `
			INSERT INTO menu_categories (venue_id, external_id, name, sort_order, is_active)
			VALUES ($1, $2, $3, $4, true)
			ON CONFLICT (venue_id, external_id) DO UPDATE SET
				name = EXCLUDED.name,
				sort_order = EXCLUDED.sort_order,
				is_active = true,
				updated_at = CASE
					WHEN (menu_categories.name, menu_categories.sort_order) IS DISTINCT FROM (EXCLUDED.name, EXCLUDED.sort_order)
					THEN now()
					ELSE menu_categories.updated_at
				END
			RETURNING id, (xmax = 0) AS inserted`
		var inserted bool
		if err := r.Q(ctx).QueryRow(ctx, catQ, in.VenueID, cat.ExternalID, cat.Name, cat.SortOrder).Scan(&catID, &inserted); err != nil {
			return nil, fmt.Errorf("upsert category: %w", err)
		}
		if inserted {
			result.Created++
		} else {
			result.Updated++
		}
		catExtToID[cat.ExternalID] = catID

		for _, item := range cat.Items {
			itemExtSeen[item.ExternalID] = struct{}{}
			categoryID := catExtToID[item.CategoryExt]
			if categoryID == uuid.Nil {
				categoryID = catID
			}
			var itemID uuid.UUID
			const itemQ = `
				INSERT INTO menu_items (
					venue_id, category_id, external_id, name, description, price_minor,
					weight_grams, image_url, sort_order, is_active, is_available
				) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,true,true)
				ON CONFLICT (venue_id, external_id) DO UPDATE SET
					category_id = EXCLUDED.category_id,
					name = EXCLUDED.name,
					description = EXCLUDED.description,
					price_minor = EXCLUDED.price_minor,
					weight_grams = EXCLUDED.weight_grams,
					image_url = EXCLUDED.image_url,
					sort_order = EXCLUDED.sort_order,
					is_active = true,
					revision = menu_items.revision + CASE
						WHEN (menu_items.name, menu_items.price_minor, menu_items.description, menu_items.sort_order)
							IS DISTINCT FROM (EXCLUDED.name, EXCLUDED.price_minor, EXCLUDED.description, EXCLUDED.sort_order)
						THEN 1 ELSE 0 END,
					updated_at = now()
				RETURNING id, (xmax = 0) AS inserted`
			var itemInserted bool
			if err := r.Q(ctx).QueryRow(ctx, itemQ,
				in.VenueID, categoryID, item.ExternalID, item.Name, item.Description, item.PriceMinor,
				item.WeightGrams, item.ImageURL, item.SortOrder,
			).Scan(&itemID, &itemInserted); err != nil {
				return nil, fmt.Errorf("upsert item: %w", err)
			}
			if itemInserted {
				result.Created++
			} else {
				result.Updated++
			}

			if err := r.syncOptionGroups(ctx, itemID, item.Groups); err != nil {
				return nil, err
			}
		}
	}

	extIDs := make([]string, 0, len(itemExtSeen))
	for ext := range itemExtSeen {
		extIDs = append(extIDs, ext)
	}
	tag, err := r.Q(ctx).Exec(ctx,
		`UPDATE menu_items SET is_active = false, updated_at = now()
		 WHERE venue_id = $1 AND ($2::text[] IS NULL OR external_id <> ALL($2)) AND is_active = true`,
		in.VenueID, extIDs)
	if err != nil {
		return nil, fmt.Errorf("deactivate items: %w", err)
	}
	result.Deactivated += int(tag.RowsAffected())

	const outboxQ = `
		INSERT INTO outbox_events (aggregate_type, aggregate_id, venue_id, event_type, payload)
		VALUES ('menu', $1, $2, 'menu.synced', '{}'::jsonb)`
	if _, err := r.Q(ctx).Exec(ctx, outboxQ, in.VenueID, in.VenueID); err != nil {
		return nil, fmt.Errorf("insert menu synced outbox: %w", err)
	}

	return result, nil
}

func (r *Repository) syncOptionGroups(ctx context.Context, itemID uuid.UUID, groups []catalogcontract.SyncOptionGroup) error {
	groupExtSeen := make([]string, 0, len(groups))
	for _, g := range groups {
		groupExtSeen = append(groupExtSeen, g.ExternalID)
		var groupID uuid.UUID
		const groupQ = `
			INSERT INTO menu_item_option_groups (item_id, external_id, name, min_choices, max_choices, sort_order)
			VALUES ($1,$2,$3,$4,$5,$6)
			ON CONFLICT (item_id, external_id) DO UPDATE SET
				name = EXCLUDED.name, min_choices = EXCLUDED.min_choices,
				max_choices = EXCLUDED.max_choices, sort_order = EXCLUDED.sort_order,
				updated_at = now()
			RETURNING id`
		if err := r.Q(ctx).QueryRow(ctx, groupQ, itemID, g.ExternalID, g.Name, g.MinChoices, g.MaxChoices, g.SortOrder).Scan(&groupID); err != nil {
			return fmt.Errorf("upsert option group: %w", err)
		}

		optExtSeen := make([]string, 0, len(g.Options))
		for _, o := range g.Options {
			optExtSeen = append(optExtSeen, o.ExternalID)
			const optQ = `
				INSERT INTO menu_item_options (group_id, external_id, name, price_delta_minor, sort_order)
				VALUES ($1,$2,$3,$4,$5)
				ON CONFLICT (group_id, external_id) DO UPDATE SET
					name = EXCLUDED.name, price_delta_minor = EXCLUDED.price_delta_minor,
					sort_order = EXCLUDED.sort_order, updated_at = now()`
			if _, err := r.Q(ctx).Exec(ctx, optQ, groupID, o.ExternalID, o.Name, o.PriceDeltaMinor, o.SortOrder); err != nil {
				return fmt.Errorf("upsert option: %w", err)
			}
		}
		if _, err := r.Q(ctx).Exec(ctx,
			`DELETE FROM menu_item_options WHERE group_id = $1 AND external_id <> ALL($2)`, groupID, optExtSeen); err != nil {
			return fmt.Errorf("delete stale options: %w", err)
		}
	}
	if _, err := r.Q(ctx).Exec(ctx,
		`DELETE FROM menu_item_option_groups WHERE item_id = $1 AND external_id <> ALL($2)`, itemID, groupExtSeen); err != nil {
		return fmt.Errorf("delete stale groups: %w", err)
	}
	return nil
}

func (r *Repository) SetItemAvailability(ctx context.Context, in catalogcontract.AvailabilityIn) error {
	tag, err := r.Q(ctx).Exec(ctx, `
		UPDATE menu_items
		SET is_available = $3, unavailable_until = $4, updated_at = now()
		WHERE venue_id = $1 AND external_id = $2`,
		in.VenueID, in.ExternalID, in.IsAvailable, in.UnavailableUntil)
	if err != nil {
		return fmt.Errorf("set item availability: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return pgerr.ErrNotFound
	}
	return nil
}

func (r *Repository) RestoreExpiredStoplist(ctx context.Context) (int, error) {
	tag, err := r.Q(ctx).Exec(ctx, `
		UPDATE menu_items SET is_available = true, unavailable_until = NULL, updated_at = now()
		WHERE unavailable_until IS NOT NULL AND unavailable_until <= now()`)
	if err != nil {
		return 0, fmt.Errorf("restore stoplist: %w", err)
	}
	return int(tag.RowsAffected()), nil
}

func (r *Repository) GetItemsByIDs(ctx context.Context, ids []uuid.UUID) ([]catalogcontract.MenuItem, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	const q = `
		SELECT id, venue_id, category_id, external_id, name, description, price_minor,
			weight_grams, image_url, sort_order, is_active, is_available, unavailable_until,
			revision, created_at, updated_at
		FROM menu_items WHERE id = ANY($1) ORDER BY id`
	rows, err := r.Q(ctx).Query(ctx, q, ids)
	if err != nil {
		return nil, fmt.Errorf("get items by ids: %w", err)
	}
	defer rows.Close()

	var items []catalogcontract.MenuItem
	for rows.Next() {
		var it catalogcontract.MenuItem
		if err := rows.Scan(
			&it.ID, &it.VenueID, &it.CategoryID, &it.ExternalID, &it.Name, &it.Description, &it.PriceMinor,
			&it.WeightGrams, &it.ImageURL, &it.SortOrder, &it.IsActive, &it.IsAvailable, &it.UnavailableUntil,
			&it.Revision, &it.CreatedAt, &it.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan menu item: %w", err)
		}
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	rows.Close()

	for i := range items {
		groups, err := r.loadOptionGroups(ctx, items[i].ID)
		if err != nil {
			return nil, err
		}
		items[i].OptionGroups = groups
	}
	return items, nil
}

func (r *Repository) loadCategories(ctx context.Context, venueID uuid.UUID) ([]catalogcontract.Category, error) {
	const q = `
		SELECT id, external_id, name, sort_order, is_active, created_at, updated_at
		FROM menu_categories WHERE venue_id = $1 AND is_active = true ORDER BY sort_order`
	rows, err := r.Q(ctx).Query(ctx, q, venueID)
	if err != nil {
		return nil, fmt.Errorf("load categories: %w", err)
	}
	defer rows.Close()

	var cats []catalogcontract.Category
	for rows.Next() {
		var c catalogcontract.Category
		var createdAt, updatedAt time.Time
		if err := rows.Scan(&c.ID, &c.ExternalID, &c.Name, &c.SortOrder, &c.IsActive, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		cats = append(cats, c)
	}
	return cats, rows.Err()
}

func (r *Repository) loadItems(ctx context.Context, venueID uuid.UUID) ([]catalogcontract.MenuItem, error) {
	const q = `
		SELECT id, venue_id, category_id, external_id, name, description, price_minor,
			weight_grams, image_url, sort_order, is_active, is_available, unavailable_until,
			revision, created_at, updated_at
		FROM menu_items WHERE venue_id = $1 AND is_active = true ORDER BY sort_order`
	rows, err := r.Q(ctx).Query(ctx, q, venueID)
	if err != nil {
		return nil, fmt.Errorf("load items: %w", err)
	}
	defer rows.Close()

	var items []catalogcontract.MenuItem
	for rows.Next() {
		var it catalogcontract.MenuItem
		if err := rows.Scan(
			&it.ID, &it.VenueID, &it.CategoryID, &it.ExternalID, &it.Name, &it.Description, &it.PriceMinor,
			&it.WeightGrams, &it.ImageURL, &it.SortOrder, &it.IsActive, &it.IsAvailable, &it.UnavailableUntil,
			&it.Revision, &it.CreatedAt, &it.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	return items, rows.Err()
}

func (r *Repository) loadOptionGroups(ctx context.Context, itemID uuid.UUID) ([]catalogcontract.OptionGroup, error) {
	const q = `
		SELECT id, external_id, name, min_choices, max_choices, sort_order
		FROM menu_item_option_groups WHERE item_id = $1 ORDER BY sort_order`
	rows, err := r.Q(ctx).Query(ctx, q, itemID)
	if err != nil {
		return nil, fmt.Errorf("load option groups: %w", err)
	}
	defer rows.Close()

	var groups []catalogcontract.OptionGroup
	for rows.Next() {
		var g catalogcontract.OptionGroup
		if err := rows.Scan(&g.ID, &g.ExternalID, &g.Name, &g.MinChoices, &g.MaxChoices, &g.SortOrder); err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	rows.Close()

	for i := range groups {
		opts, err := r.loadOptions(ctx, groups[i].ID)
		if err != nil {
			return nil, err
		}
		groups[i].Options = opts
	}
	return groups, nil
}

func (r *Repository) loadOptions(ctx context.Context, groupID uuid.UUID) ([]catalogcontract.Option, error) {
	const q = `
		SELECT id, external_id, name, price_delta_minor, is_available, sort_order
		FROM menu_item_options WHERE group_id = $1 ORDER BY sort_order`
	rows, err := r.Q(ctx).Query(ctx, q, groupID)
	if err != nil {
		return nil, fmt.Errorf("load options: %w", err)
	}
	defer rows.Close()

	var opts []catalogcontract.Option
	for rows.Next() {
		var o catalogcontract.Option
		if err := rows.Scan(&o.ID, &o.ExternalID, &o.Name, &o.PriceDeltaMinor, &o.IsAvailable, &o.SortOrder); err != nil {
			return nil, err
		}
		opts = append(opts, o)
	}
	return opts, rows.Err()
}

func computeETag(menu *catalogcontract.Menu) string {
	h := sha256.Sum256([]byte(menu.UpdatedAt.UTC().Format(time.RFC3339Nano) + menu.VenueID.String()))
	return `"` + hex.EncodeToString(h[:16]) + `"`
}
