package mapper

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"github.com/zlatovlaska707/avito-kitchen/internal/domain/orderstatus"
	"github.com/zlatovlaska707/avito-kitchen/internal/handler/httpx"
	cartrecalc "github.com/zlatovlaska707/avito-kitchen/internal/usecase/cart/recalc"
	catalogrepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/catalog"
	deliveryrepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/delivery"
	orderrepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/order"
	venuerepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/venue"
)

type Venue struct {
	ID               uuid.UUID `json:"id"`
	Slug             string    `json:"slug"`
	Name             string    `json:"name"`
	Rating           *float64  `json:"rating"`
	DeliveryFeeMinor int64     `json:"delivery_fee_minor"`
	MinOrderMinor    int64     `json:"min_order_minor"`
	ETAMinutes       *int      `json:"eta_minutes,omitempty"`
	IsOpenNow        bool      `json:"is_open_now"`
	DistanceM        *int      `json:"distance_m,omitempty"`
	Currency         string    `json:"currency"`
}

type VenueDetails struct {
	Venue
	Description   *string       `json:"description,omitempty"`
	Status        string        `json:"status"`
	AcceptsOrders bool          `json:"accepts_orders"`
	CityID        int           `json:"city_id"`
	AddressText   string        `json:"address_text"`
	Lat           float64       `json:"lat"`
	Lon           float64       `json:"lon"`
	Timezone      string        `json:"timezone"`
	WorkingHours  []WorkingHour `json:"working_hours"`
	ClosedReason  *string       `json:"closed_reason,omitempty"`
}

type WorkingHour struct {
	Weekday         int    `json:"weekday"`
	OpensAt         string `json:"opens_at"`
	ClosesAt        string `json:"closes_at"`
	CrossesMidnight bool   `json:"crosses_midnight"`
}

type Menu struct {
	VenueID    uuid.UUID      `json:"venue_id"`
	Categories []MenuCategory `json:"categories"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

type MenuCategory struct {
	ID         uuid.UUID  `json:"id"`
	ExternalID string     `json:"external_id"`
	Name       string     `json:"name"`
	SortOrder  int        `json:"sort_order"`
	IsActive   bool       `json:"is_active"`
	Items      []MenuItem `json:"items"`
}

type MenuItem struct {
	ID               uuid.UUID     `json:"id"`
	ExternalID       string        `json:"external_id"`
	Name             string        `json:"name"`
	Description      *string       `json:"description,omitempty"`
	PriceMinor       int64         `json:"price_minor"`
	WeightGrams      *int          `json:"weight_grams,omitempty"`
	ImageURL         *string       `json:"image_url,omitempty"`
	SortOrder        int           `json:"sort_order"`
	IsActive         bool          `json:"is_active"`
	IsAvailable      bool          `json:"is_available"`
	UnavailableUntil *time.Time    `json:"unavailable_until,omitempty"`
	OptionGroups     []OptionGroup `json:"option_groups"`
}

type OptionGroup struct {
	ID         uuid.UUID `json:"id"`
	ExternalID string    `json:"external_id"`
	Name       string    `json:"name"`
	MinChoices int       `json:"min_choices"`
	MaxChoices int       `json:"max_choices"`
	SortOrder  int       `json:"sort_order"`
	Options    []Option  `json:"options"`
}

type Option struct {
	ID              uuid.UUID `json:"id"`
	ExternalID      string    `json:"external_id"`
	Name            string    `json:"name"`
	PriceDeltaMinor int64     `json:"price_delta_minor"`
	IsAvailable     bool      `json:"is_available"`
	SortOrder       int       `json:"sort_order"`
}

type Cart struct {
	ID               uuid.UUID   `json:"id"`
	VenueID          uuid.UUID   `json:"venue_id"`
	VenueName        string      `json:"venue_name,omitempty"`
	Lines            []CartLine  `json:"lines"`
	ItemsTotalMinor  int64       `json:"items_total_minor"`
	DeliveryFeeMinor int64       `json:"delivery_fee_minor"`
	TotalMinor       int64       `json:"total_minor"`
	Currency         string      `json:"currency"`
	Issues           []CartIssue `json:"issues"`
	ExpiresAt        time.Time   `json:"expires_at"`
}

type CartLine struct {
	ID              uuid.UUID            `json:"id"`
	MenuItemID      uuid.UUID            `json:"menu_item_id"`
	Name            string               `json:"name"`
	Quantity        int                  `json:"quantity"`
	UnitPriceMinor  int64                `json:"unit_price_minor"`
	LineTotalMinor  int64                `json:"line_total_minor"`
	SelectedOptions []CartSelectedOption `json:"selected_options"`
}

type CartSelectedOption struct {
	ID              uuid.UUID `json:"id"`
	ExternalID      string    `json:"external_id"`
	Name            string    `json:"name"`
	PriceDeltaMinor int64     `json:"price_delta_minor"`
}

type CartIssue struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}

type Order struct {
	ID               uuid.UUID                   `json:"id"`
	PublicNumber     string                      `json:"public_number"`
	CustomerID       uuid.UUID                   `json:"customer_id"`
	VenueID          uuid.UUID                   `json:"venue_id"`
	Status           orderstatus.Status          `json:"status"`
	FulfillmentType  orderstatus.FulfillmentType `json:"fulfillment_type"`
	DeliveryAddress  json.RawMessage             `json:"delivery_address,omitempty"`
	ContactPhone     string                      `json:"contact_phone"`
	ItemsTotalMinor  int64                       `json:"items_total_minor"`
	DeliveryFeeMinor int64                       `json:"delivery_fee_minor"`
	DiscountMinor    int64                       `json:"discount_minor"`
	TotalMinor       int64                       `json:"total_minor"`
	Currency         string                      `json:"currency"`
	PaymentMethod    string                      `json:"payment_method"`
	PaymentStatus    string                      `json:"payment_status"`
	Comment          *string                     `json:"comment,omitempty"`
	ETAAt            *time.Time                  `json:"eta_at,omitempty"`
	PlacedAt         *time.Time                  `json:"placed_at,omitempty"`
	ExternalOrderID  *string                     `json:"external_order_id,omitempty"`
	CancelReason     *string                     `json:"cancel_reason,omitempty"`
	Lines            []OrderLine                 `json:"lines"`
	StatusHistory    []OrderStatusEvent          `json:"status_history"`
	Delivery         *Delivery                   `json:"delivery,omitempty"`
}

type OrderLine struct {
	MenuItemID      *uuid.UUID      `json:"menu_item_id,omitempty"`
	ItemExternalID  string          `json:"item_external_id"`
	NameSnapshot    string          `json:"name_snapshot"`
	UnitPriceMinor  int64           `json:"unit_price_minor"`
	Quantity        int             `json:"quantity"`
	OptionsSnapshot json.RawMessage `json:"options_snapshot"`
	LineTotalMinor  int64           `json:"line_total_minor"`
}

type OrderStatusEvent struct {
	FromStatus *orderstatus.Status   `json:"from_status,omitempty"`
	ToStatus   orderstatus.Status    `json:"to_status"`
	ActorType  orderstatus.ActorType `json:"actor_type"`
	ActorID    *string               `json:"actor_id,omitempty"`
	Reason     *string               `json:"reason,omitempty"`
	CreatedAt  time.Time             `json:"created_at"`
}

type Delivery struct {
	ID           uuid.UUID `json:"id"`
	Status       string    `json:"status"`
	CourierName  *string   `json:"courier_name,omitempty"`
	CourierPhone *string   `json:"courier_phone,omitempty"`
	ETAMinutes   *int      `json:"eta_minutes,omitempty"`
}

func VenueFrom(v venuerepo.Venue, isOpen bool, eta *int, distance *int) Venue {
	return Venue{
		ID: v.ID, Slug: v.Slug, Name: v.Name, Rating: v.Rating,
		DeliveryFeeMinor: v.DeliveryFeeMinor, MinOrderMinor: v.MinOrderMinor,
		ETAMinutes: eta, IsOpenNow: isOpen, DistanceM: distance, Currency: "RUB",
	}
}

func WorkingHoursFrom(hours []venuerepo.WorkingHour) []WorkingHour {
	out := make([]WorkingHour, 0, len(hours))
	for _, h := range hours {
		out = append(out, WorkingHour{
			Weekday: h.Weekday,
			OpensAt: formatDuration(h.OpensAt), ClosesAt: formatDuration(h.ClosesAt),
			CrossesMidnight: h.CrossesMidnight,
		})
	}
	return out
}

func formatDuration(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	return time.Date(0, 1, 1, h, m, 0, 0, time.UTC).Format("15:04")
}

func MenuFrom(m *catalogrepo.Menu) Menu {
	cats := make([]MenuCategory, 0, len(m.Categories))
	for _, c := range m.Categories {
		items := make([]MenuItem, 0, len(c.Items))
		for _, it := range c.Items {
			items = append(items, menuItemFrom(it))
		}
		cats = append(cats, MenuCategory{
			ID: c.ID, ExternalID: c.ExternalID, Name: c.Name,
			SortOrder: c.SortOrder, IsActive: c.IsActive, Items: items,
		})
	}
	return Menu{VenueID: m.VenueID, Categories: cats, UpdatedAt: m.UpdatedAt}
}

func menuItemFrom(it catalogrepo.MenuItem) MenuItem {
	groups := make([]OptionGroup, 0, len(it.OptionGroups))
	for _, g := range it.OptionGroups {
		opts := make([]Option, 0, len(g.Options))
		for _, o := range g.Options {
			opts = append(opts, Option{
				ID: o.ID, ExternalID: o.ExternalID, Name: o.Name,
				PriceDeltaMinor: o.PriceDeltaMinor, IsAvailable: o.IsAvailable, SortOrder: o.SortOrder,
			})
		}
		groups = append(groups, OptionGroup{
			ID: g.ID, ExternalID: g.ExternalID, Name: g.Name,
			MinChoices: g.MinChoices, MaxChoices: g.MaxChoices, SortOrder: g.SortOrder, Options: opts,
		})
	}
	return MenuItem{
		ID: it.ID, ExternalID: it.ExternalID, Name: it.Name, Description: it.Description,
		PriceMinor: it.PriceMinor, WeightGrams: it.WeightGrams, ImageURL: it.ImageURL,
		SortOrder: it.SortOrder, IsActive: it.IsActive, IsAvailable: it.IsAvailable,
		UnavailableUntil: it.UnavailableUntil, OptionGroups: groups,
	}
}

func MenuItemFrom(it catalogrepo.MenuItem) MenuItem { return menuItemFrom(it) }

func CartFrom(c cartrecalc.Result) *Cart {
	if c.Cart == nil {
		return nil
	}
	lines := make([]CartLine, 0, len(c.Lines))
	for _, l := range c.Lines {
		cid, _ := uuid.Parse(l.ID)
		mid, _ := uuid.Parse(l.MenuItemID)
		opts := make([]CartSelectedOption, 0, len(l.SelectedOptions))
		for _, o := range l.SelectedOptions {
			oid, _ := uuid.Parse(o.ID)
			opts = append(opts, CartSelectedOption{ID: oid, ExternalID: o.ExternalID, Name: o.Name, PriceDeltaMinor: o.PriceDeltaMinor})
		}
		lines = append(lines, CartLine{
			ID: cid, MenuItemID: mid, Name: l.Name, Quantity: l.Quantity,
			UnitPriceMinor: l.UnitPriceMinor, LineTotalMinor: l.LineTotalMinor, SelectedOptions: opts,
		})
	}
	issues := make([]CartIssue, 0, len(c.Issues))
	for _, iss := range c.Issues {
		issues = append(issues, CartIssue{Code: iss.Code, Message: iss.Message, Details: iss.Details})
	}
	return &Cart{
		ID: c.Cart.ID, VenueID: c.Cart.VenueID, VenueName: c.VenueName, Lines: lines,
		ItemsTotalMinor: c.ItemsTotalMinor, DeliveryFeeMinor: c.DeliveryFeeMinor,
		TotalMinor: c.TotalMinor, Currency: "RUB", Issues: issues, ExpiresAt: c.Cart.ExpiresAt,
	}
}

func OrderFrom(o *orderrepo.Order, items []orderrepo.Item, del *deliveryrepo.Delivery) Order {
	return OrderFromWithHistory(o, items, del, nil)
}

func OrderFromWithHistory(o *orderrepo.Order, items []orderrepo.Item, del *deliveryrepo.Delivery, history []orderrepo.StatusEvent) Order {
	lines := make([]OrderLine, 0, len(items))
	for _, it := range items {
		lines = append(lines, OrderLine{
			MenuItemID: it.MenuItemID, ItemExternalID: it.ItemExternalID,
			NameSnapshot: it.NameSnapshot, UnitPriceMinor: it.UnitPriceMinor, Quantity: it.Quantity,
			OptionsSnapshot: it.OptionsSnapshot, LineTotalMinor: it.LineTotalMinor,
		})
	}
	events := make([]OrderStatusEvent, 0, len(history))
	for _, ev := range history {
		events = append(events, OrderStatusEvent{
			FromStatus: ev.FromStatus, ToStatus: ev.ToStatus, ActorType: ev.ActorType,
			ActorID: ev.ActorID, Reason: ev.Reason, CreatedAt: ev.CreatedAt,
		})
	}
	var placed *time.Time
	if !o.PlacedAt.IsZero() {
		t := o.PlacedAt
		placed = &t
	}
	var delivery *Delivery
	if del != nil {
		delivery = &Delivery{
			ID: del.ID, Status: del.Status, CourierName: del.CourierName,
			CourierPhone: del.CourierPhone, ETAMinutes: del.ETAMinutes,
		}
	}
	return Order{
		ID: o.ID, PublicNumber: o.PublicNumber, CustomerID: o.CustomerID, VenueID: o.VenueID,
		Status: o.Status, FulfillmentType: o.FulfillmentType, DeliveryAddress: o.DeliveryAddress,
		ContactPhone: o.ContactPhone, ItemsTotalMinor: o.ItemsTotalMinor, DeliveryFeeMinor: o.DeliveryFeeMinor,
		DiscountMinor: o.DiscountMinor, TotalMinor: o.TotalMinor, Currency: o.Currency,
		PaymentMethod: o.PaymentMethod, PaymentStatus: o.PaymentStatus, Comment: o.Comment,
		ETAAt: o.ETAAt, PlacedAt: placed, ExternalOrderID: o.ExternalOrderID, CancelReason: o.CancelReason,
		Lines: lines, StatusHistory: events, Delivery: delivery,
	}
}

func PageFrom(limit int, hasNext bool, lastTS *time.Time, lastID *uuid.UUID) httpx.PageInfo {
	return httpx.BuildPageInfo(limit, hasNext, lastTS, lastID)
}
