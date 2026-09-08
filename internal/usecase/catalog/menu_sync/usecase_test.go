package menu_sync_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zlatovlaska707/avito-kitchen/internal/usecase"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase/catalog/menu_sync"
	catalogrepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/catalog"
)

type catalogMock struct {
	lockOK  bool
	lockErr error
	syncRes *catalogrepo.SyncResult
	syncErr error
	syncIn  *catalogrepo.SyncIn
}

func (m *catalogMock) GetMenu(context.Context, uuid.UUID) (*catalogrepo.Menu, string, error) {
	return nil, "", nil
}

func (m *catalogMock) SyncMenu(ctx context.Context, in catalogrepo.SyncIn) (*catalogrepo.SyncResult, error) {
	m.syncIn = &in
	if m.syncErr != nil {
		return nil, m.syncErr
	}
	return m.syncRes, nil
}

func (m *catalogMock) SetItemAvailability(context.Context, catalogrepo.AvailabilityIn) error {
	return nil
}

func (m *catalogMock) GetItemsByIDs(context.Context, []uuid.UUID) ([]catalogrepo.MenuItem, error) {
	return nil, nil
}

func (m *catalogMock) TryMenuSyncLock(context.Context, uuid.UUID) (bool, error) {
	if m.lockErr != nil {
		return false, m.lockErr
	}
	return m.lockOK, nil
}

func (m *catalogMock) RestoreExpiredStoplist(context.Context) (int, error) {
	return 0, nil
}

func TestRun_HappyPath(t *testing.T) {
	mock := &catalogMock{
		lockOK:  true,
		syncRes: &catalogrepo.SyncResult{Created: 2, Updated: 1},
	}
	uc := menu_sync.New(mock, 10, 100)
	venueID := uuid.New()
	out, err := uc.Run(context.Background(), menu_sync.In{
		VenueID: venueID,
		Categories: []catalogrepo.SyncCategory{{
			ExternalID: "cat-1", Name: "Pizza",
			Items: []catalogrepo.SyncItem{{ExternalID: "item-1", Name: "Margherita", PriceMinor: 49000}},
		}},
	})
	require.NoError(t, err)
	assert.Equal(t, 2, out.Result.Created)
	require.NotNil(t, mock.syncIn)
	assert.Equal(t, venueID, mock.syncIn.VenueID)
	assert.Equal(t, "cat-1", mock.syncIn.Categories[0].Items[0].CategoryExt)
}

func TestRun_TooManyCategories(t *testing.T) {
	uc := menu_sync.New(&catalogMock{}, 1, 100)
	cats := []catalogrepo.SyncCategory{
		{ExternalID: "c1", Items: nil},
		{ExternalID: "c2", Items: nil},
	}
	_, err := uc.Run(context.Background(), menu_sync.In{VenueID: uuid.New(), Categories: cats})
	assert.ErrorIs(t, err, usecase.ErrValidation)
}

func TestRun_TooManyItems(t *testing.T) {
	items := make([]catalogrepo.SyncItem, 3)
	for i := range items {
		items[i] = catalogrepo.SyncItem{ExternalID: "item-" + string(rune('a'+i)), PriceMinor: 100}
	}
	uc := menu_sync.New(&catalogMock{}, 10, 2)
	_, err := uc.Run(context.Background(), menu_sync.In{
		VenueID:    uuid.New(),
		Categories: []catalogrepo.SyncCategory{{ExternalID: "cat", Items: items}},
	})
	assert.ErrorIs(t, err, usecase.ErrValidation)
}

func TestRun_DuplicateCategoryExternalID(t *testing.T) {
	uc := menu_sync.New(&catalogMock{}, 10, 100)
	_, err := uc.Run(context.Background(), menu_sync.In{
		VenueID: uuid.New(),
		Categories: []catalogrepo.SyncCategory{
			{ExternalID: "dup", Items: nil},
			{ExternalID: "dup", Items: nil},
		},
	})
	assert.ErrorIs(t, err, usecase.ErrDuplicateExternalID)
}

func TestRun_DuplicateItemExternalID(t *testing.T) {
	uc := menu_sync.New(&catalogMock{}, 10, 100)
	_, err := uc.Run(context.Background(), menu_sync.In{
		VenueID: uuid.New(),
		Categories: []catalogrepo.SyncCategory{{
			ExternalID: "cat",
			Items: []catalogrepo.SyncItem{
				{ExternalID: "same", PriceMinor: 100},
				{ExternalID: "same", PriceMinor: 200},
			},
		}},
	})
	assert.ErrorIs(t, err, usecase.ErrDuplicateExternalID)
}

func TestRun_MenuSyncInProgress(t *testing.T) {
	mock := &catalogMock{lockOK: false}
	uc := menu_sync.New(mock, 10, 100)
	_, err := uc.Run(context.Background(), menu_sync.In{
		VenueID: uuid.New(),
		Categories: []catalogrepo.SyncCategory{{
			ExternalID: "cat", Items: []catalogrepo.SyncItem{{ExternalID: "i1", PriceMinor: 1}},
		}},
	})
	assert.ErrorIs(t, err, usecase.ErrMenuSyncInProgress)
}

func TestRun_LockError(t *testing.T) {
	mock := &catalogMock{lockErr: assert.AnError}
	uc := menu_sync.New(mock, 10, 100)
	_, err := uc.Run(context.Background(), menu_sync.In{
		VenueID:    uuid.New(),
		Categories: []catalogrepo.SyncCategory{{ExternalID: "cat"}},
	})
	assert.ErrorIs(t, err, assert.AnError)
}
