package item_add_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/repository/pgerr"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase"
	cartitemadd "github.com/zlatovlaska707/avito-kitchen/internal/usecase/cart/item_add"
	cartrepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/cart"
	catalogrepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/catalog"
	venuerepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/venue"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase/uctest"
)

type cartRepo struct {
	cart       *cartrepo.Cart
	getErr     error
	items      []cartrepo.Item
	itemCount  int
	created    *cartrepo.CreateIn
	added      *cartrepo.AddItemIn
	createCart *cartrepo.Cart
}

func (m *cartRepo) GetActiveForUpdate(context.Context, uuid.UUID) (*cartrepo.Cart, error) {
	return m.cart, m.getErr
}

func (m *cartRepo) GetActive(context.Context, uuid.UUID) (*cartrepo.Cart, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.cart, nil
}

func (m *cartRepo) Create(_ context.Context, in cartrepo.CreateIn) (*cartrepo.Cart, error) {
	m.created = &in
	if m.createCart != nil {
		return m.createCart, nil
	}
	return &cartrepo.Cart{
		ID: uuid.New(), CustomerID: in.CustomerID, VenueID: in.VenueID, ExpiresAt: in.ExpiresAt,
	}, nil
}

func (m *cartRepo) GetItems(context.Context, uuid.UUID) ([]cartrepo.Item, error) {
	return m.items, nil
}

func (m *cartRepo) AddItem(_ context.Context, in cartrepo.AddItemIn) (*cartrepo.Item, error) {
	m.added = &in
	return &cartrepo.Item{
		ID: uuid.New(), CartID: in.CartID, MenuItemID: in.MenuItemID, Quantity: in.Quantity,
		SelectedOptions: in.SelectedOptions, OptionsHash: in.OptionsHash,
	}, nil
}

func (m *cartRepo) UpdateItem(context.Context, cartrepo.UpdateItemIn) (*cartrepo.Item, error) {
	return nil, nil
}

func (m *cartRepo) DeleteItem(context.Context, uuid.UUID, uuid.UUID) error { return nil }

func (m *cartRepo) ClearItems(context.Context, uuid.UUID) error { return nil }

func (m *cartRepo) Abandon(context.Context, uuid.UUID) error { return nil }

func (m *cartRepo) MarkConverted(context.Context, uuid.UUID, uuid.UUID) error { return nil }

func (m *cartRepo) CountItems(context.Context, uuid.UUID) (int, error) {
	if m.itemCount > 0 {
		return m.itemCount, nil
	}
	return len(m.items), nil
}

func (m *cartRepo) AbandonExpired(context.Context) (int, error) { return 0, nil }

type catalogRepo struct {
	items []catalogrepo.MenuItem
}

func (m *catalogRepo) GetMenu(context.Context, uuid.UUID) (*catalogrepo.Menu, string, error) {
	return nil, "", nil
}

func (m *catalogRepo) SyncMenu(context.Context, catalogrepo.SyncIn) (*catalogrepo.SyncResult, error) {
	return nil, nil
}

func (m *catalogRepo) SetItemAvailability(context.Context, catalogrepo.AvailabilityIn) error {
	return nil
}

func (m *catalogRepo) GetItemsByIDs(_ context.Context, ids []uuid.UUID) ([]catalogrepo.MenuItem, error) {
	if len(m.items) == 0 {
		return nil, nil
	}
	out := make([]catalogrepo.MenuItem, 0, len(ids))
	for _, id := range ids {
		for _, item := range m.items {
			if item.ID == id {
				out = append(out, item)
				break
			}
		}
	}
	return out, nil
}

func (m *catalogRepo) TryMenuSyncLock(context.Context, uuid.UUID) (bool, error) { return true, nil }

func (m *catalogRepo) RestoreExpiredStoplist(context.Context) (int, error) { return 0, nil }

func weekdayISO(t time.Time) int {
	wd := int(t.Weekday())
	if wd == 0 {
		return 7
	}
	return wd
}

func openHours(now time.Time) []venuerepo.WorkingHour {
	return []venuerepo.WorkingHour{{
		Weekday: weekdayISO(now), OpensAt: 0, ClosesAt: 24 * time.Hour,
	}}
}

func menuItem(venueID, menuItemID uuid.UUID) catalogrepo.MenuItem {
	return catalogrepo.MenuItem{
		ID: menuItemID, VenueID: venueID, Name: "Margherita", PriceMinor: 49000,
		IsActive: true, IsAvailable: true,
	}
}

func TestRun_QuantityLimit(t *testing.T) {
	now := time.Date(2026, time.September, 8, 12, 0, 0, 0, time.UTC)
	venueID := uuid.New()
	menuItemID := uuid.New()
	uc := cartitemadd.New(
		&cartRepo{getErr: pgerr.ErrNotFound},
		&catalogRepo{items: []catalogrepo.MenuItem{menuItem(venueID, menuItemID)}},
		&uctest.VenueRepo{},
		uctest.FixedNower{T: now},
		time.UTC,
		time.Hour,
		10,
	)

	for _, qty := range []int{0, 100} {
		_, err := uc.Run(context.Background(), cartitemadd.In{
			CustomerID: uuid.New(), VenueID: venueID, MenuItemID: menuItemID, Quantity: qty,
		})
		assert.ErrorIs(t, err, usecase.ErrQuantityLimit)
	}
}

func TestRun_VenueMismatch(t *testing.T) {
	now := time.Date(2026, time.September, 8, 12, 0, 0, 0, time.UTC)
	cartVenueID := uuid.New()
	requestVenueID := uuid.New()
	menuItemID := uuid.New()
	customerID := uuid.New()

	carts := &cartRepo{
		cart: &cartrepo.Cart{ID: uuid.New(), CustomerID: customerID, VenueID: cartVenueID},
	}
	uc := cartitemadd.New(
		carts,
		&catalogRepo{items: []catalogrepo.MenuItem{menuItem(requestVenueID, menuItemID)}},
		&uctest.VenueRepo{},
		uctest.FixedNower{T: now},
		time.UTC,
		time.Hour,
		10,
	)

	_, err := uc.Run(context.Background(), cartitemadd.In{
		CustomerID: customerID, VenueID: requestVenueID, MenuItemID: menuItemID, Quantity: 1,
	})
	assert.ErrorIs(t, err, usecase.ErrCartVenueMismatch)
}

func TestRun_CartItemLimit(t *testing.T) {
	now := time.Date(2026, time.September, 8, 12, 0, 0, 0, time.UTC)
	venueID := uuid.New()
	menuItemID := uuid.New()
	customerID := uuid.New()

	carts := &cartRepo{
		cart:      &cartrepo.Cart{ID: uuid.New(), CustomerID: customerID, VenueID: venueID},
		itemCount: 10,
	}
	uc := cartitemadd.New(
		carts,
		&catalogRepo{items: []catalogrepo.MenuItem{menuItem(venueID, menuItemID)}},
		&uctest.VenueRepo{},
		uctest.FixedNower{T: now},
		time.UTC,
		time.Hour,
		10,
	)

	_, err := uc.Run(context.Background(), cartitemadd.In{
		CustomerID: customerID, VenueID: venueID, MenuItemID: menuItemID, Quantity: 1,
	})
	assert.ErrorIs(t, err, usecase.ErrCartItemLimit)
}

func TestRun_HappyPathNewCart(t *testing.T) {
	now := time.Date(2026, time.September, 8, 12, 0, 0, 0, time.UTC)
	venueID := uuid.New()
	menuItemID := uuid.New()
	customerID := uuid.New()
	cartID := uuid.New()

	carts := &cartRepo{
		getErr: pgerr.ErrNotFound,
		createCart: &cartrepo.Cart{
			ID: cartID, CustomerID: customerID, VenueID: venueID,
		},
	}
	venues := &uctest.VenueRepo{
		Venue: &venuerepo.Venue{ID: venueID, Name: "Pizza Mario", MinOrderMinor: 0},
		Hours: openHours(now),
	}
	uc := cartitemadd.New(
		carts,
		&catalogRepo{items: []catalogrepo.MenuItem{menuItem(venueID, menuItemID)}},
		venues,
		uctest.FixedNower{T: now},
		time.UTC,
		time.Hour,
		10,
	)

	out, err := uc.Run(context.Background(), cartitemadd.In{
		CustomerID: customerID, VenueID: venueID, MenuItemID: menuItemID, Quantity: 2,
	})
	require.NoError(t, err)
	require.NotNil(t, carts.created)
	assert.Equal(t, customerID, carts.created.CustomerID)
	assert.Equal(t, venueID, carts.created.VenueID)
	require.NotNil(t, carts.added)
	assert.Equal(t, cartID, carts.added.CartID)
	assert.Equal(t, menuItemID, carts.added.MenuItemID)
	assert.Equal(t, 2, carts.added.Quantity)
	require.NotNil(t, out.Result.Cart)
}

func TestRun_HappyPathExistingCart(t *testing.T) {
	now := time.Date(2026, time.September, 8, 12, 0, 0, 0, time.UTC)
	venueID := uuid.New()
	menuItemID := uuid.New()
	customerID := uuid.New()
	cartID := uuid.New()

	carts := &cartRepo{
		cart: &cartrepo.Cart{ID: cartID, CustomerID: customerID, VenueID: venueID},
		items: []cartrepo.Item{{
			ID: uuid.New(), CartID: cartID, MenuItemID: menuItemID, Quantity: 1,
			SelectedOptions: json.RawMessage(`[]`),
		}},
	}
	venues := &uctest.VenueRepo{
		Venue: &venuerepo.Venue{ID: venueID, Name: "Pizza Mario", MinOrderMinor: 0},
		Hours: openHours(now),
	}
	uc := cartitemadd.New(
		carts,
		&catalogRepo{items: []catalogrepo.MenuItem{menuItem(venueID, menuItemID)}},
		venues,
		uctest.FixedNower{T: now},
		time.UTC,
		time.Hour,
		10,
	)

	out, err := uc.Run(context.Background(), cartitemadd.In{
		CustomerID: customerID, VenueID: venueID, MenuItemID: menuItemID, Quantity: 1,
	})
	require.NoError(t, err)
	require.NotNil(t, carts.added)
	assert.Nil(t, carts.created)
	require.NotNil(t, out.Result.Cart)
}

func TestRun_MenuItemNotFound(t *testing.T) {
	now := time.Date(2026, time.September, 8, 12, 0, 0, 0, time.UTC)
	uc := cartitemadd.New(
		&cartRepo{getErr: pgerr.ErrNotFound},
		&catalogRepo{},
		&uctest.VenueRepo{},
		uctest.FixedNower{T: now},
		time.UTC,
		time.Hour,
		10,
	)

	_, err := uc.Run(context.Background(), cartitemadd.In{
		CustomerID: uuid.New(), VenueID: uuid.New(), MenuItemID: uuid.New(), Quantity: 1,
	})
	assert.ErrorIs(t, err, usecase.ErrMenuItemNotFound)
}

func TestRun_MenuItemUnavailable(t *testing.T) {
	now := time.Date(2026, time.September, 8, 12, 0, 0, 0, time.UTC)
	venueID := uuid.New()
	menuItemID := uuid.New()
	item := menuItem(venueID, menuItemID)
	item.IsAvailable = false

	uc := cartitemadd.New(
		&cartRepo{getErr: pgerr.ErrNotFound},
		&catalogRepo{items: []catalogrepo.MenuItem{item}},
		&uctest.VenueRepo{},
		uctest.FixedNower{T: now},
		time.UTC,
		time.Hour,
		10,
	)

	_, err := uc.Run(context.Background(), cartitemadd.In{
		CustomerID: uuid.New(), VenueID: venueID, MenuItemID: menuItemID, Quantity: 1,
	})
	assert.ErrorIs(t, err, usecase.ErrMenuItemUnavailable)
}
