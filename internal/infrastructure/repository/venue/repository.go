package venue

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Masterminds/squirrel"
	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/zlatovlaska707/avito-kitchen/internal/domain/geo"
	"github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/repository/db"
	"github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/repository/pgerr"
	venuecontract "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/venue"
)

type Repository struct {
	db.RepositoryBase
}

func New(pool *pgxpool.Pool, getter *trmpgx.CtxGetter) *Repository {
	return &Repository{RepositoryBase: db.RepositoryBase{Pool: pool, Getter: getter}}
}

var psql = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

func (r *Repository) Create(ctx context.Context, in venuecontract.CreateIn) (*venuecontract.Venue, error) {
	const q = `
		INSERT INTO venues (
			slug, name, description, city_id, address_text, lat, lon, timezone,
			avg_prep_minutes, delivery_fee_minor, min_order_minor, delivery_radius_m, status, accepts_orders
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,'draft',false)
		RETURNING id, slug, name, description, status, accepts_orders, city_id, address_text,
			lat, lon, timezone, avg_prep_minutes, delivery_fee_minor, min_order_minor,
			delivery_radius_m, rating, pause_reason, created_at, updated_at`

	row := r.Q(ctx).QueryRow(ctx, q,
		in.Slug, in.Name, in.Description, in.CityID, in.AddressText, in.Lat, in.Lon, in.Timezone,
		in.AvgPrepMinutes, in.DeliveryFeeMinor, in.MinOrderMinor, in.DeliveryRadiusM,
	)
	v, err := scanVenue(row)
	if err != nil {
		return nil, fmt.Errorf("create venue: %w", err)
	}

	if len(in.WorkingHours) > 0 {
		if err := r.ReplaceWorkingHours(ctx, v.ID, in.WorkingHours); err != nil {
			return nil, err
		}
	}
	return v, nil
}

func (r *Repository) Update(ctx context.Context, in venuecontract.UpdateIn) (*venuecontract.Venue, error) {
	b := psql.Update("venues").Where(squirrel.Eq{"id": in.ID})
	if in.Name != nil {
		b = b.Set("name", *in.Name)
	}
	if in.Description != nil {
		b = b.Set("description", *in.Description)
	}
	if in.AddressText != nil {
		b = b.Set("address_text", *in.AddressText)
	}
	if in.Lat != nil {
		b = b.Set("lat", *in.Lat)
	}
	if in.Lon != nil {
		b = b.Set("lon", *in.Lon)
	}
	if in.Timezone != nil {
		b = b.Set("timezone", *in.Timezone)
	}
	if in.AvgPrepMinutes != nil {
		b = b.Set("avg_prep_minutes", *in.AvgPrepMinutes)
	}
	if in.DeliveryFeeMinor != nil {
		b = b.Set("delivery_fee_minor", *in.DeliveryFeeMinor)
	}
	if in.MinOrderMinor != nil {
		b = b.Set("min_order_minor", *in.MinOrderMinor)
	}
	if in.DeliveryRadiusM != nil {
		b = b.Set("delivery_radius_m", *in.DeliveryRadiusM)
	}
	b = b.Set("updated_at", squirrel.Expr("now()"))
	b = b.Suffix(`RETURNING id, slug, name, description, status, accepts_orders, city_id, address_text,
		lat, lon, timezone, avg_prep_minutes, delivery_fee_minor, min_order_minor,
		delivery_radius_m, rating, pause_reason, created_at, updated_at`)

	sql, args, err := b.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build update venue: %w", err)
	}
	row := r.Q(ctx).QueryRow(ctx, sql, args...)
	v, err := scanVenue(row)
	if err != nil {
		return nil, pgerr.WrapNotFound(fmt.Errorf("update venue: %w", err))
	}
	return v, nil
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*venuecontract.Venue, error) {
	const q = `SELECT id, slug, name, description, status, accepts_orders, city_id, address_text,
		lat, lon, timezone, avg_prep_minutes, delivery_fee_minor, min_order_minor,
		delivery_radius_m, rating, pause_reason, created_at, updated_at
		FROM venues WHERE id = $1`
	row := r.Q(ctx).QueryRow(ctx, q, id)
	v, err := scanVenue(row)
	if err != nil {
		return nil, pgerr.WrapNotFound(fmt.Errorf("get venue: %w", err))
	}
	return v, nil
}

func (r *Repository) GetBySlug(ctx context.Context, slug string) (*venuecontract.Venue, error) {
	const q = `SELECT id, slug, name, description, status, accepts_orders, city_id, address_text,
		lat, lon, timezone, avg_prep_minutes, delivery_fee_minor, min_order_minor,
		delivery_radius_m, rating, pause_reason, created_at, updated_at
		FROM venues WHERE slug = $1`
	row := r.Q(ctx).QueryRow(ctx, q, slug)
	v, err := scanVenue(row)
	if err != nil {
		return nil, pgerr.WrapNotFound(fmt.Errorf("get venue by slug: %w", err))
	}
	return v, nil
}

const defaultListRadiusM = 50000

func (r *Repository) List(ctx context.Context, filter venuecontract.ListFilter) ([]venuecontract.Venue, error) {
	cols := `id, slug, name, description, status, accepts_orders, city_id, address_text,
		lat, lon, timezone, avg_prep_minutes, delivery_fee_minor, min_order_minor,
		delivery_radius_m, rating, pause_reason, created_at, updated_at`

	b := psql.Select(cols).From("venues")

	if filter.CityID != nil {
		b = b.Where(squirrel.Eq{"city_id": *filter.CityID})
	}
	if filter.Status != nil {
		b = b.Where(squirrel.Eq{"status": *filter.Status})
	}
	if filter.Query != nil && strings.TrimSpace(*filter.Query) != "" {
		q := "%" + escapeLike(strings.TrimSpace(*filter.Query)) + "%"
		b = b.Where(squirrel.Or{
			squirrel.ILike{"name": q},
			squirrel.ILike{"slug": q},
		})
	}
	if filter.Lat != nil && filter.Lon != nil {
		radiusM := float64(defaultListRadiusM)
		if filter.RadiusM != nil && *filter.RadiusM > 0 {
			radiusM = float64(*filter.RadiusM)
		}
		if box, ok := geo.ComputeBoundingBox(*filter.Lat, *filter.Lon, radiusM); ok {
			b = b.Where(squirrel.And{
				squirrel.GtOrEq{"lat": box.MinLat},
				squirrel.LtOrEq{"lat": box.MaxLat},
				squirrel.GtOrEq{"lon": box.MinLon},
				squirrel.LtOrEq{"lon": box.MaxLon},
			})
		}
	}

	if filter.CursorTS != nil && filter.CursorID != nil {
		b = b.Where(squirrel.Expr("(created_at, id) < (?, ?)", *filter.CursorTS, *filter.CursorID))
	}

	if filter.Sort == "distance" && filter.Lat != nil && filter.Lon != nil {
		b = b.OrderByClause(squirrel.Expr(
			`(6371000 * acos(LEAST(1.0, GREATEST(-1.0,
				cos(radians(?)) * cos(radians(lat)) * cos(radians(lon) - radians(?)) +
				sin(radians(?)) * sin(radians(lat))
			)))) ASC`, *filter.Lat, *filter.Lon, *filter.Lat))
		b = b.OrderBy("id ASC")
	} else {
		b = b.OrderBy("created_at DESC", "id DESC")
	}

	if filter.Limit > 0 {
		b = b.Limit(uint64(filter.Limit))
	}

	sql, args, err := b.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build list venues: %w", err)
	}

	rows, err := r.Q(ctx).Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("list venues: %w", err)
	}
	defer rows.Close()

	var result []venuecontract.Venue
	for rows.Next() {
		v, err := scanVenue(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, *v)
	}
	return result, rows.Err()
}

func escapeLike(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `%`, `\%`)
	s = strings.ReplaceAll(s, `_`, `\_`)
	return s
}

func (r *Repository) UpdateStatus(ctx context.Context, in venuecontract.StatusUpdateIn) error {
	b := psql.Update("venues").Where(squirrel.Eq{"id": in.VenueID})
	if in.Status != nil {
		b = b.Set("status", *in.Status)
	}
	if in.AcceptsOrders != nil {
		b = b.Set("accepts_orders", *in.AcceptsOrders)
	}
	if in.PauseReason != nil {
		b = b.Set("pause_reason", *in.PauseReason)
	}
	b = b.Set("updated_at", squirrel.Expr("now()"))
	sql, args, err := b.ToSql()
	if err != nil {
		return fmt.Errorf("build update status: %w", err)
	}
	tag, err := r.Q(ctx).Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("update venue status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return pgerr.ErrNotFound
	}
	return nil
}

func (r *Repository) GetWorkingHours(ctx context.Context, venueID uuid.UUID) ([]venuecontract.WorkingHour, error) {
	const q = `
		SELECT id, weekday, opens_at, closes_at, crosses_midnight
		FROM venue_working_hours
		WHERE venue_id = $1
		ORDER BY weekday, opens_at`

	rows, err := r.Q(ctx).Query(ctx, q, venueID)
	if err != nil {
		return nil, fmt.Errorf("get working hours: %w", err)
	}
	defer rows.Close()

	var result []venuecontract.WorkingHour
	for rows.Next() {
		var wh venuecontract.WorkingHour
		var opensAt, closesAt time.Time
		if err := rows.Scan(&wh.ID, &wh.Weekday, &opensAt, &closesAt, &wh.CrossesMidnight); err != nil {
			return nil, fmt.Errorf("scan working hour: %w", err)
		}
		wh.OpensAt = timeDurationSinceMidnight(opensAt)
		wh.ClosesAt = timeDurationSinceMidnight(closesAt)
		result = append(result, wh)
	}
	return result, rows.Err()
}

func (r *Repository) ReplaceWorkingHours(ctx context.Context, venueID uuid.UUID, hours []venuecontract.WorkingHour) error {
	if _, err := r.Q(ctx).Exec(ctx, `DELETE FROM venue_working_hours WHERE venue_id = $1`, venueID); err != nil {
		return fmt.Errorf("delete working hours: %w", err)
	}
	for _, h := range hours {
		opens := durationToTime(h.OpensAt)
		closes := durationToTime(h.ClosesAt)
		const q = `
			INSERT INTO venue_working_hours (venue_id, weekday, opens_at, closes_at, crosses_midnight)
			VALUES ($1, $2, $3, $4, $5)`
		if _, err := r.Q(ctx).Exec(ctx, q, venueID, h.Weekday, opens, closes, h.CrossesMidnight); err != nil {
			return fmt.Errorf("insert working hour: %w", err)
		}
	}
	return nil
}

func (r *Repository) IsSlugTaken(ctx context.Context, slug string, excludeID *uuid.UUID) (bool, error) {
	b := psql.Select("1").From("venues").Where(squirrel.Eq{"slug": slug}).Limit(1)
	if excludeID != nil {
		b = b.Where(squirrel.NotEq{"id": *excludeID})
	}
	sql, args, err := b.ToSql()
	if err != nil {
		return false, err
	}
	var one int
	err = r.Q(ctx).QueryRow(ctx, sql, args...).Scan(&one)
	if pgerr.IsNotFound(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("is slug taken: %w", err)
	}
	return true, nil
}

func durationToTime(d time.Duration) time.Time {
	return time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC).Add(d)
}

func timeDurationSinceMidnight(t time.Time) time.Duration {
	return time.Duration(t.Hour())*time.Hour + time.Duration(t.Minute())*time.Minute + time.Duration(t.Second())*time.Second
}

type scannable interface {
	Scan(dest ...any) error
}

func scanVenue(row scannable) (*venuecontract.Venue, error) {
	var v venuecontract.Venue
	if err := row.Scan(
		&v.ID, &v.Slug, &v.Name, &v.Description, &v.Status, &v.AcceptsOrders, &v.CityID, &v.AddressText,
		&v.Lat, &v.Lon, &v.Timezone, &v.AvgPrepMinutes, &v.DeliveryFeeMinor, &v.MinOrderMinor,
		&v.DeliveryRadiusM, &v.Rating, &v.PauseReason, &v.CreatedAt, &v.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &v, nil
}
