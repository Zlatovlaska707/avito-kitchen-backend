package menu_get_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/repository/pgerr"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase/catalog/menu_get"
	catalogrepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/catalog"
	venuerepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/venue"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase/uctest"
)

type catalogMock struct {
	menu    *catalogrepo.Menu
	etag    string
	menuErr error
}

func (m *catalogMock) GetMenu(_ context.Context, venueID uuid.UUID) (*catalogrepo.Menu, string, error) {
	if m.menuErr != nil {
		return nil, "", m.menuErr
	}
	if m.menu != nil {
		return m.menu, m.etag, nil
	}
	return &catalogrepo.Menu{VenueID: venueID}, m.etag, nil
}

func (m *catalogMock) SyncMenu(context.Context, catalogrepo.SyncIn) (*catalogrepo.SyncResult, error) {
	return nil, nil
}

func (m *catalogMock) SetItemAvailability(context.Context, catalogrepo.AvailabilityIn) error {
	return nil
}

func (m *catalogMock) GetItemsByIDs(context.Context, []uuid.UUID) ([]catalogrepo.MenuItem, error) {
	return nil, nil
}

func (m *catalogMock) TryMenuSyncLock(context.Context, uuid.UUID) (bool, error) {
	return false, nil
}

func (m *catalogMock) RestoreExpiredStoplist(context.Context) (int, error) {
	return 0, nil
}

func activeVenue(id uuid.UUID) *venuerepo.Venue {
	return &venuerepo.Venue{ID: id, Slug: "pizza-place", Status: "active"}
}

func TestRun_HappyPathByID(t *testing.T) {
	venueID := uuid.New()
	menu := &catalogrepo.Menu{VenueID: venueID, UpdatedAt: time.Now()}
	catalog := &catalogMock{menu: menu, etag: `"v1"`}
	venues := &uctest.VenueRepo{Venue: activeVenue(venueID)}
	uc := menu_get.New(catalog, venues)

	out, err := uc.Run(context.Background(), menu_get.In{VenueRef: venueID.String()})
	require.NoError(t, err)
	require.NotNil(t, out)
	assert.Equal(t, menu, out.Menu)
	assert.Equal(t, `"v1"`, out.ETag)
}

func TestRun_HappyPathBySlug(t *testing.T) {
	venueID := uuid.New()
	catalog := &catalogMock{etag: `"v2"`}
	venues := &uctest.VenueRepo{Venue: activeVenue(venueID)}
	uc := menu_get.New(catalog, venues)

	out, err := uc.Run(context.Background(), menu_get.In{VenueRef: "pizza-place"})
	require.NoError(t, err)
	require.NotNil(t, out)
	assert.Equal(t, venueID, out.Menu.VenueID)
	assert.Equal(t, `"v2"`, out.ETag)
}

func TestRun_VenueNotFound(t *testing.T) {
	uc := menu_get.New(&catalogMock{}, &uctest.VenueRepo{GetBySlugErr: pgerr.ErrNotFound})

	_, err := uc.Run(context.Background(), menu_get.In{VenueRef: "missing"})
	assert.ErrorIs(t, err, usecase.ErrVenueNotFound)
}

func TestRun_VenueNotActive(t *testing.T) {
	venueID := uuid.New()
	venues := &uctest.VenueRepo{Venue: &venuerepo.Venue{ID: venueID, Status: "inactive"}}
	uc := menu_get.New(&catalogMock{}, venues)

	_, err := uc.Run(context.Background(), menu_get.In{VenueRef: venueID.String()})
	assert.ErrorIs(t, err, usecase.ErrVenueNotFound)
}

func TestRun_ResolveVenueError(t *testing.T) {
	uc := menu_get.New(&catalogMock{}, &uctest.VenueRepo{GetByIDErr: assert.AnError})

	_, err := uc.Run(context.Background(), menu_get.In{VenueRef: uuid.New().String()})
	assert.ErrorIs(t, err, assert.AnError)
}

func TestRun_GetMenuError(t *testing.T) {
	venueID := uuid.New()
	catalog := &catalogMock{menuErr: assert.AnError}
	venues := &uctest.VenueRepo{Venue: activeVenue(venueID)}
	uc := menu_get.New(catalog, venues)

	_, err := uc.Run(context.Background(), menu_get.In{VenueRef: venueID.String()})
	assert.ErrorIs(t, err, assert.AnError)
}
