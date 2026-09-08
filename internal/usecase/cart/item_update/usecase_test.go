package item_update_test

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
	cartitemupdate "github.com/zlatovlaska707/avito-kitchen/internal/usecase/cart/item_update"
	cartrepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/cart"
	catalogrepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/catalog"
	venuerepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/venue"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase/uctest"
)

type cartRepo struct {
	cart       *cartrepo.Cart
	getErr     error
	items      []cartrepo.Item
	updateErr  error
	updatedQty int
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

func (m *cartRepo) Create(context.Context, cartrepo.CreateIn) (*cartrepo.Cart, error) {
	return nil, nil
}

func (m *cartRepo) GetItems(context.Context, uuid.UUID) ([]cartrepo.Item, error) {
	return m.items, nil
}

func (m *cartRepo) AddItem(context.Context, cartrepo.AddItemIn) (*cartrepo.Item, error) {
	return nil, nil
}

func (m *cartRepo) UpdateItem(_ context.Context, in cartrepo.UpdateItemIn) (*cartrepo.Item, error) {
	if m.updateErr != nil {
		return nil, m.updateErr
	}
	m.updatedQty = in.Quantity
	return &cartrepo.Item{ID: in.ItemID, CartID: in.CartID, Quantity: in.Quantity}, nil
}

func (m *cartRepo) DeleteItem(context.Context, uuid.UUID, uuid.UUID) error { return nil }

func (m *cartRepo) ClearItems(context.Context, uuid.UUID) error { return nil }

func (m *cartRepo) Abandon(context.Context, uuid.UUID) error { return nil }

func (m *cartRepo) MarkConverted(context.Context, uuid.UUID, uuid.UUID) error { return nil }

func (m *cartRepo) CountItems(context.Context, uuid.UUID) (int, error) { return len(m.items), nil }

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

func (m *catalogRepo) GetItemsByIDs(context.Context, []uuid.UUID) ([]catalogrepo.MenuItem, error) {
	return m.items, nil
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

func baseDeps(now time.Time) (*cartRepo, *catalogRepo, *uctest.VenueRepo) {
	venueID := uuid.New()
	cartID := uuid.New()
	customerID := uuid.New()
	itemID := uuid.New()
	menuItemID := uuid.New()

	carts := &cartRepo{
		cart: &cartrepo.Cart{ID: cartID, CustomerID: customerID, VenueID: venueID},
		items: []cartrepo.Item{{
			ID: itemID, CartID: cartID, MenuItemID: menuItemID, Quantity: 2,
			SelectedOptions: json.RawMessage(`[]`),
		}},
	}
	catalog := &catalogRepo{items: []catalogrepo.MenuItem{{
		ID: menuItemID, PriceMinor: 49000, IsActive: true, IsAvailable: true,
	}}}
	venues := &uctest.VenueRepo{
		Venue: &venuerepo.Venue{ID: venueID, Name: "Pizza Mario", MinOrderMinor: 0},
		Hours: openHours(now),
	}
	return carts, catalog, venues
}

func TestRun_QuantityLimit(t *testing.T) {
	now := time.Date(2026, time.September, 8, 12, 0, 0, 0, time.UTC)
	carts, catalog, venues := baseDeps(now)
	uc := cartitemupdate.New(carts, catalog, venues, uctest.FixedNower{T: now}, time.UTC)

	for _, qty := range []int{0, 100} {
		_, err := uc.Run(context.Background(), cartitemupdate.In{
			CustomerID: carts.cart.CustomerID, ItemID: carts.items[0].ID, Quantity: qty,
		})
		assert.ErrorIs(t, err, usecase.ErrQuantityLimit)
	}
}

func TestRun_CartNotFound(t *testing.T) {
	now := time.Date(2026, time.September, 8, 12, 0, 0, 0, time.UTC)
	carts, catalog, venues := baseDeps(now)
	carts.getErr = pgerr.ErrNotFound
	uc := cartitemupdate.New(carts, catalog, venues, uctest.FixedNower{T: now}, time.UTC)

	_, err := uc.Run(context.Background(), cartitemupdate.In{
		CustomerID: carts.cart.CustomerID, ItemID: uuid.New(), Quantity: 1,
	})
	assert.ErrorIs(t, err, usecase.ErrCartNotFound)
}

func TestRun_ItemNotFound(t *testing.T) {
	now := time.Date(2026, time.September, 8, 12, 0, 0, 0, time.UTC)
	carts, catalog, venues := baseDeps(now)
	carts.updateErr = pgerr.ErrNotFound
	uc := cartitemupdate.New(carts, catalog, venues, uctest.FixedNower{T: now}, time.UTC)

	_, err := uc.Run(context.Background(), cartitemupdate.In{
		CustomerID: carts.cart.CustomerID, ItemID: uuid.New(), Quantity: 3,
	})
	assert.ErrorIs(t, err, usecase.ErrCartNotFound)
}

func TestRun_HappyPath(t *testing.T) {
	now := time.Date(2026, time.September, 8, 12, 0, 0, 0, time.UTC)
	carts, catalog, venues := baseDeps(now)
	uc := cartitemupdate.New(carts, catalog, venues, uctest.FixedNower{T: now}, time.UTC)

	out, err := uc.Run(context.Background(), cartitemupdate.In{
		CustomerID: carts.cart.CustomerID, ItemID: carts.items[0].ID, Quantity: 5,
	})
	require.NoError(t, err)
	assert.Equal(t, 5, carts.updatedQty)
	require.NotNil(t, out.Result.Cart)
}
