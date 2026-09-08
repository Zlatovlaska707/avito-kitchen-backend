package create_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	trm "github.com/avito-tech/go-transaction-manager/trm/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zlatovlaska707/avito-kitchen/internal/domain/orderstatus"
	"github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/repository/pgerr"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase"
	cartrepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/cart"
	catalogrepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/catalog"
	idempotencyrepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/idempotency"
	orderrepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/order"
	venuerepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/venue"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase/ordering/create"
)

type trmStub struct{}

func (trmStub) Do(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

func (trmStub) DoWithSettings(ctx context.Context, _ trm.Settings, fn func(context.Context) error) error {
	return fn(ctx)
}

type fixedNower struct{ t time.Time }

func (n fixedNower) Now() time.Time { return n.t }

type idempotencyMock struct {
	claimResult idempotencyrepo.ClaimResult
	record      *idempotencyrepo.Record
	deleted     bool
}

func (m *idempotencyMock) Claim(context.Context, idempotencyrepo.ClaimIn) (idempotencyrepo.ClaimResult, error) {
	return m.claimResult, nil
}

func (m *idempotencyMock) Get(context.Context, string, string) (*idempotencyrepo.Record, error) {
	return m.record, nil
}

func (m *idempotencyMock) SaveResponse(context.Context, idempotencyrepo.SaveResponseIn) error {
	return nil
}

func (m *idempotencyMock) Delete(context.Context, string, string) error {
	m.deleted = true
	return nil
}

func (m *idempotencyMock) DeleteExpired(context.Context) (int, error) { return 0, nil }

type deps struct {
	carts       *cartMock
	catalog     *catalogMock
	venues      *venueMock
	orders      *orderMock
	idempotency *idempotencyMock
	nower       fixedNower
}

type cartMock struct {
	cart  *cartrepo.Cart
	items []cartrepo.Item
	err   error
}

func (m *cartMock) GetActiveForUpdate(context.Context, uuid.UUID) (*cartrepo.Cart, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.cart, nil
}

func (m *cartMock) GetActive(context.Context, uuid.UUID) (*cartrepo.Cart, error) {
	return m.cart, m.err
}

func (m *cartMock) Create(context.Context, cartrepo.CreateIn) (*cartrepo.Cart, error) {
	return nil, nil
}
func (m *cartMock) GetItems(context.Context, uuid.UUID) ([]cartrepo.Item, error) { return m.items, nil }
func (m *cartMock) AddItem(context.Context, cartrepo.AddItemIn) (*cartrepo.Item, error) {
	return nil, nil
}

func (m *cartMock) UpdateItem(context.Context, cartrepo.UpdateItemIn) (*cartrepo.Item, error) {
	return nil, nil
}
func (m *cartMock) DeleteItem(context.Context, uuid.UUID, uuid.UUID) error { return nil }
func (m *cartMock) ClearItems(context.Context, uuid.UUID) error            { return nil }
func (m *cartMock) Abandon(context.Context, uuid.UUID) error               { return nil }
func (m *cartMock) MarkConverted(context.Context, uuid.UUID, uuid.UUID) error {
	return nil
}
func (m *cartMock) CountItems(context.Context, uuid.UUID) (int, error) { return len(m.items), nil }
func (m *cartMock) AbandonExpired(context.Context) (int, error)        { return 0, nil }

type catalogMock struct {
	items []catalogrepo.MenuItem
}

func (m *catalogMock) GetMenu(context.Context, uuid.UUID) (*catalogrepo.Menu, string, error) {
	return nil, "", nil
}

func (m *catalogMock) SyncMenu(context.Context, catalogrepo.SyncIn) (*catalogrepo.SyncResult, error) {
	return nil, nil
}

func (m *catalogMock) SetItemAvailability(context.Context, catalogrepo.AvailabilityIn) error {
	return nil
}

func (m *catalogMock) GetItemsByIDs(context.Context, []uuid.UUID) ([]catalogrepo.MenuItem, error) {
	return m.items, nil
}
func (m *catalogMock) TryMenuSyncLock(context.Context, uuid.UUID) (bool, error) { return true, nil }
func (m *catalogMock) RestoreExpiredStoplist(context.Context) (int, error)      { return 0, nil }

type venueMock struct {
	venue *venuerepo.Venue
	hours []venuerepo.WorkingHour
	err   error
}

func (m *venueMock) GetByID(context.Context, uuid.UUID) (*venuerepo.Venue, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.venue, nil
}

func (m *venueMock) GetWorkingHours(context.Context, uuid.UUID) ([]venuerepo.WorkingHour, error) {
	return m.hours, nil
}

func (m *venueMock) Create(context.Context, venuerepo.CreateIn) (*venuerepo.Venue, error) {
	return nil, nil
}

func (m *venueMock) Update(context.Context, venuerepo.UpdateIn) (*venuerepo.Venue, error) {
	return nil, nil
}
func (m *venueMock) GetBySlug(context.Context, string) (*venuerepo.Venue, error) { return nil, nil }
func (m *venueMock) List(context.Context, venuerepo.ListFilter) ([]venuerepo.Venue, error) {
	return nil, nil
}
func (m *venueMock) UpdateStatus(context.Context, venuerepo.StatusUpdateIn) error { return nil }
func (m *venueMock) ReplaceWorkingHours(context.Context, uuid.UUID, []venuerepo.WorkingHour) error {
	return nil
}
func (m *venueMock) IsSlugTaken(context.Context, string, *uuid.UUID) (bool, error) { return false, nil }

type orderMock struct {
	created *orderrepo.CreateIn
}

func (m *orderMock) Create(ctx context.Context, in orderrepo.CreateIn) (*orderrepo.Order, error) {
	m.created = &in
	return &orderrepo.Order{
		ID: uuid.New(), Status: orderstatus.StatusPlaced, TotalMinor: in.TotalMinor,
		FulfillmentType: in.FulfillmentType,
	}, nil
}

func (m *orderMock) GetByID(context.Context, uuid.UUID) (*orderrepo.Order, error) { return nil, nil }
func (m *orderMock) GetByIDForCustomer(context.Context, uuid.UUID, uuid.UUID) (*orderrepo.Order, error) {
	return nil, nil
}

func (m *orderMock) GetByIDForVenue(context.Context, uuid.UUID, uuid.UUID) (*orderrepo.Order, error) {
	return nil, nil
}

func (m *orderMock) List(context.Context, orderrepo.ListFilter) ([]orderrepo.Order, error) {
	return nil, nil
}
func (m *orderMock) GetItems(context.Context, uuid.UUID) ([]orderrepo.Item, error) { return nil, nil }
func (m *orderMock) GetStatusHistory(context.Context, uuid.UUID) ([]orderrepo.StatusEvent, error) {
	return nil, nil
}

func (m *orderMock) ApplyTransition(context.Context, orderrepo.TransitionIn) (*orderrepo.Order, error) {
	return nil, nil
}

func (m *orderMock) FindPlacedBefore(context.Context, time.Time, int) ([]uuid.UUID, error) {
	return nil, nil
}

func newUC(d deps) *create.UseCase {
	return create.New(
		trmStub{}, d.carts, d.catalog, d.venues, d.orders, d.idempotency, d.nower, time.UTC,
		time.Hour, 30, 2.0,
	)
}

func openVenueHours(now time.Time) []venuerepo.WorkingHour {
	return []venuerepo.WorkingHour{{
		Weekday: weekdayISO(now), OpensAt: 0, ClosesAt: 24 * time.Hour,
	}}
}

func weekdayISO(t time.Time) int {
	wd := int(t.Weekday())
	if wd == 0 {
		return 7
	}
	return wd
}

func baseDeps() deps {
	venueID := uuid.New()
	menuID := uuid.New()
	cartID := uuid.New()
	customerID := uuid.New()
	now := time.Date(2026, time.September, 8, 12, 0, 0, 0, time.UTC) // Monday noon UTC

	return deps{
		carts: &cartMock{
			cart: &cartrepo.Cart{ID: cartID, CustomerID: customerID, VenueID: venueID},
			items: []cartrepo.Item{{
				ID: uuid.New(), CartID: cartID, MenuItemID: menuID, Quantity: 2,
				SelectedOptions: json.RawMessage(`[]`),
			}},
		},
		catalog: &catalogMock{items: []catalogrepo.MenuItem{{
			ID: menuID, ExternalID: "margherita", Name: "Margherita", PriceMinor: 49000,
			IsActive: true, IsAvailable: true,
		}}},
		venues: &venueMock{venue: &venuerepo.Venue{
			ID: venueID, Status: "active", AcceptsOrders: true, Timezone: "UTC",
			Lat: 55.7558, Lon: 37.6173, MinOrderMinor: 50000, DeliveryFeeMinor: 19900,
			DeliveryRadiusM: 5000, AvgPrepMinutes: 30,
		}, hours: openVenueHours(now)},
		orders:      &orderMock{},
		idempotency: &idempotencyMock{claimResult: idempotencyrepo.ClaimGranted},
		nower:       fixedNower{t: now},
	}
}

func deliveryAddress(lat, lon float64) json.RawMessage {
	raw, _ := json.Marshal(map[string]any{"text": "Test", "lat": lat, "lon": lon})
	return raw
}

func TestRun_CartNotFound(t *testing.T) {
	d := baseDeps()
	d.carts.err = pgerr.ErrNotFound
	_, err := newUC(d).Run(context.Background(), create.In{
		CustomerID: uuid.New(), FulfillmentType: "delivery",
		DeliveryAddress: deliveryAddress(55.76, 37.62), IdempotencyKey: "k1", RequestHash: "h1",
	})
	assert.ErrorIs(t, err, usecase.ErrCartNotFound)
}

func TestRun_CartEmpty(t *testing.T) {
	d := baseDeps()
	d.carts.items = nil
	_, err := newUC(d).Run(context.Background(), create.In{
		CustomerID: d.carts.cart.CustomerID, FulfillmentType: "delivery",
		DeliveryAddress: deliveryAddress(55.76, 37.62), IdempotencyKey: "k1", RequestHash: "h1",
	})
	assert.ErrorIs(t, err, usecase.ErrCartEmpty)
}

func TestRun_VenueNotFound(t *testing.T) {
	d := baseDeps()
	d.venues.err = pgerr.ErrNotFound
	_, err := newUC(d).Run(context.Background(), create.In{
		CustomerID: d.carts.cart.CustomerID, FulfillmentType: "delivery",
		DeliveryAddress: deliveryAddress(55.76, 37.62), IdempotencyKey: "k1", RequestHash: "h1",
	})
	assert.ErrorIs(t, err, usecase.ErrVenueNotFound)
}

func TestRun_VenuePaused(t *testing.T) {
	d := baseDeps()
	d.venues.venue.AcceptsOrders = false
	_, err := newUC(d).Run(context.Background(), create.In{
		CustomerID: d.carts.cart.CustomerID, FulfillmentType: "delivery",
		DeliveryAddress: deliveryAddress(55.76, 37.62), IdempotencyKey: "k1", RequestHash: "h1",
	})
	assert.ErrorIs(t, err, usecase.ErrVenueClosed)
}

func TestRun_VenueClosedBySchedule(t *testing.T) {
	d := baseDeps()
	d.venues.hours = []venuerepo.WorkingHour{{
		Weekday: weekdayISO(d.nower.t), OpensAt: 14 * time.Hour, ClosesAt: 22 * time.Hour,
	}}
	_, err := newUC(d).Run(context.Background(), create.In{
		CustomerID: d.carts.cart.CustomerID, FulfillmentType: "delivery",
		DeliveryAddress: deliveryAddress(55.76, 37.62), IdempotencyKey: "k1", RequestHash: "h1",
	})
	assert.ErrorIs(t, err, usecase.ErrVenueClosed)
}

func TestRun_OutsideDeliveryZone(t *testing.T) {
	d := baseDeps()
	d.venues.venue.DeliveryRadiusM = 100
	_, err := newUC(d).Run(context.Background(), create.In{
		CustomerID: d.carts.cart.CustomerID, FulfillmentType: "delivery",
		DeliveryAddress: deliveryAddress(55.90, 37.90), IdempotencyKey: "k1", RequestHash: "h1",
	})
	assert.ErrorIs(t, err, usecase.ErrOutsideDeliveryZone)
}

func TestRun_ItemUnavailable(t *testing.T) {
	d := baseDeps()
	d.catalog.items[0].IsAvailable = false
	_, err := newUC(d).Run(context.Background(), create.In{
		CustomerID: d.carts.cart.CustomerID, FulfillmentType: "delivery",
		DeliveryAddress: deliveryAddress(55.76, 37.62), IdempotencyKey: "k1", RequestHash: "h1",
	})
	assert.ErrorIs(t, err, usecase.ErrMenuItemUnavailable)
}

func TestRun_BelowMinOrder(t *testing.T) {
	d := baseDeps()
	d.carts.items[0].Quantity = 1
	_, err := newUC(d).Run(context.Background(), create.In{
		CustomerID: d.carts.cart.CustomerID, FulfillmentType: "delivery",
		DeliveryAddress: deliveryAddress(55.76, 37.62), IdempotencyKey: "k1", RequestHash: "h1",
	})
	assert.ErrorIs(t, err, usecase.ErrBelowMinOrder)
}

func TestRun_PriceChanged(t *testing.T) {
	d := baseDeps()
	wrong := int64(1)
	_, err := newUC(d).Run(context.Background(), create.In{
		CustomerID: d.carts.cart.CustomerID, FulfillmentType: "delivery",
		DeliveryAddress: deliveryAddress(55.76, 37.62), ExpectedTotalMinor: &wrong,
		IdempotencyKey: "k1", RequestHash: "h1",
	})
	assert.ErrorIs(t, err, usecase.ErrPriceChanged)
}

func TestRun_IdempotencyReplay(t *testing.T) {
	status := 201
	body := json.RawMessage(`{"id":"replay"}`)
	d := baseDeps()
	d.idempotency.claimResult = idempotencyrepo.ClaimExists
	d.idempotency.record = &idempotencyrepo.Record{
		RequestHash: "same", ResponseStatus: &status, ResponseBody: body,
	}
	out, err := newUC(d).Run(context.Background(), create.In{
		CustomerID: uuid.New(), FulfillmentType: "delivery",
		IdempotencyKey: "k1", RequestHash: "same",
	})
	require.NoError(t, err)
	assert.True(t, out.Replayed)
	assert.Equal(t, body, out.ReplayBody)
	assert.Equal(t, 201, out.ReplayCode)
}

func TestRun_IdempotencyKeyReuse(t *testing.T) {
	d := baseDeps()
	d.idempotency.claimResult = idempotencyrepo.ClaimExists
	d.idempotency.record = &idempotencyrepo.Record{RequestHash: "other"}
	_, err := newUC(d).Run(context.Background(), create.In{
		CustomerID: uuid.New(), IdempotencyKey: "k1", RequestHash: "new",
	})
	assert.ErrorIs(t, err, usecase.ErrIdempotencyKeyReuse)
}

func TestRun_HappyPathDelivery(t *testing.T) {
	d := baseDeps()
	expected := int64(98000 + 19900)
	out, err := newUC(d).Run(context.Background(), create.In{
		CustomerID:         d.carts.cart.CustomerID,
		FulfillmentType:    "delivery",
		DeliveryAddress:    deliveryAddress(55.76, 37.62),
		ContactPhone:       "+79001234567",
		PaymentMethod:      "cash",
		ExpectedTotalMinor: &expected,
		IdempotencyKey:     "k1",
		RequestHash:        "h1",
	})
	require.NoError(t, err)
	require.NotNil(t, out.Order)
	assert.False(t, out.Replayed)
	require.NotNil(t, d.orders.created)
	assert.Equal(t, expected, d.orders.created.TotalMinor)
	assert.True(t, d.orders.created.CreateDelivery)
	assert.NotEmpty(t, d.orders.created.OutboxPayload)
	assert.Len(t, d.orders.created.Items, 1)
}

func TestRun_HappyPathPickup(t *testing.T) {
	d := baseDeps()
	expected := int64(98000)
	out, err := newUC(d).Run(context.Background(), create.In{
		CustomerID:         d.carts.cart.CustomerID,
		FulfillmentType:    "pickup",
		ContactPhone:       "+79001234567",
		PaymentMethod:      "cash",
		ExpectedTotalMinor: &expected,
		IdempotencyKey:     "k1",
		RequestHash:        "h1",
	})
	require.NoError(t, err)
	require.NotNil(t, out.Order)
	assert.False(t, d.orders.created.CreateDelivery)
	assert.Equal(t, int64(0), d.orders.created.DeliveryFeeMinor)
}

func TestRun_InvalidFulfillmentType(t *testing.T) {
	d := baseDeps()
	_, err := newUC(d).Run(context.Background(), create.In{
		CustomerID: d.carts.cart.CustomerID, FulfillmentType: "drone",
		IdempotencyKey: "k1", RequestHash: "h1",
	})
	assert.ErrorIs(t, err, usecase.ErrValidation)
}

func TestRun_BusinessErrorDeletesIdempotencyKey(t *testing.T) {
	d := baseDeps()
	d.carts.items = nil
	_, err := newUC(d).Run(context.Background(), create.In{
		CustomerID: d.carts.cart.CustomerID, FulfillmentType: "pickup",
		IdempotencyKey: "k1", RequestHash: "h1",
	})
	assert.Error(t, err)
	assert.True(t, d.idempotency.deleted)
}

func TestRun_RequestInProgress(t *testing.T) {
	d := baseDeps()
	d.idempotency.claimResult = idempotencyrepo.ClaimExists
	d.idempotency.record = &idempotencyrepo.Record{RequestHash: "h1"}
	_, err := newUC(d).Run(context.Background(), create.In{
		CustomerID: uuid.New(), IdempotencyKey: "k1", RequestHash: "h1",
	})
	assert.ErrorIs(t, err, usecase.ErrRequestInProgress)
}

func TestRun_MenuItemNotFound(t *testing.T) {
	d := baseDeps()
	d.catalog.items = nil
	_, err := newUC(d).Run(context.Background(), create.In{
		CustomerID: d.carts.cart.CustomerID, FulfillmentType: "pickup",
		IdempotencyKey: "k1", RequestHash: "h1",
	})
	assert.ErrorIs(t, err, usecase.ErrMenuItemNotFound)
}

var _ trm.Manager = trmStub{}
