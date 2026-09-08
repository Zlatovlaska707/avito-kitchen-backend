package get_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/repository/pgerr"
	cartget "github.com/zlatovlaska707/avito-kitchen/internal/usecase/cart/get"
	cartrepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/cart"
	catalogrepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/catalog"
	venuerepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/venue"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase/uctest"
)

type cartRepo struct {
	cart      *cartrepo.Cart
	getErr    error
	items     []cartrepo.Item
	itemsErr  error
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
	if m.itemsErr != nil {
		return nil, m.itemsErr
	}
	return m.items, nil
}

func (m *cartRepo) AddItem(context.Context, cartrepo.AddItemIn) (*cartrepo.Item, error) {
	return nil, nil
}

func (m *cartRepo) UpdateItem(context.Context, cartrepo.UpdateItemIn) (*cartrepo.Item, error) {
	return nil, nil
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

func TestRun_EmptyWhenCartNotFound(t *testing.T) {
	now := time.Date(2026, time.September, 8, 12, 0, 0, 0, time.UTC)
	uc := cartget.New(
		&cartRepo{getErr: pgerr.ErrNotFound},
		&catalogRepo{},
		&uctest.VenueRepo{},
		uctest.FixedNower{T: now},
		time.UTC,
	)

	out, err := uc.Run(context.Background(), cartget.In{CustomerID: uuid.New()})
	require.NoError(t, err)
	assert.Nil(t, out.Result.Cart)
	assert.Empty(t, out.Result.Lines)
}

func TestRun_RepoError(t *testing.T) {
	now := time.Date(2026, time.September, 8, 12, 0, 0, 0, time.UTC)
	uc := cartget.New(
		&cartRepo{getErr: assert.AnError},
		&catalogRepo{},
		&uctest.VenueRepo{},
		uctest.FixedNower{T: now},
		time.UTC,
	)

	_, err := uc.Run(context.Background(), cartget.In{CustomerID: uuid.New()})
	assert.ErrorIs(t, err, assert.AnError)
}

func TestRun_HappyPath(t *testing.T) {
	now := time.Date(2026, time.September, 8, 12, 0, 0, 0, time.UTC)
	venueID := uuid.New()
	cartID := uuid.New()
	customerID := uuid.New()
	menuItemID := uuid.New()

	carts := &cartRepo{
		cart: &cartrepo.Cart{ID: cartID, CustomerID: customerID, VenueID: venueID},
		items: []cartrepo.Item{{
			ID: uuid.New(), CartID: cartID, MenuItemID: menuItemID, Quantity: 2,
			SelectedOptions: json.RawMessage(`[]`),
		}},
	}
	catalog := &catalogRepo{items: []catalogrepo.MenuItem{{
		ID: menuItemID, Name: "Margherita", PriceMinor: 49000, IsActive: true, IsAvailable: true,
	}}}
	venues := &uctest.VenueRepo{
		Venue: &venuerepo.Venue{
			ID: venueID, Name: "Pizza Mario", Status: "active", MinOrderMinor: 0,
		},
		Hours: openHours(now),
	}

	out, err := cartget.New(carts, catalog, venues, uctest.FixedNower{T: now}, time.UTC).
		Run(context.Background(), cartget.In{CustomerID: customerID})
	require.NoError(t, err)
	require.NotNil(t, out.Result.Cart)
	assert.Equal(t, cartID, out.Result.Cart.ID)
	assert.Len(t, out.Result.Lines, 1)
	assert.Equal(t, int64(98000), out.Result.ItemsTotalMinor)
}
