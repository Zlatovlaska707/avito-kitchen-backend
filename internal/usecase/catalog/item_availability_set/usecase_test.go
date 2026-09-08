package item_availability_set_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/repository/pgerr"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase/catalog/item_availability_set"
	catalogrepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/catalog"
)

type catalogMock struct {
	setErr error
	setIn  *catalogrepo.AvailabilityIn
}

func (m *catalogMock) GetMenu(context.Context, uuid.UUID) (*catalogrepo.Menu, string, error) {
	return nil, "", nil
}

func (m *catalogMock) SyncMenu(context.Context, catalogrepo.SyncIn) (*catalogrepo.SyncResult, error) {
	return nil, nil
}

func (m *catalogMock) SetItemAvailability(_ context.Context, in catalogrepo.AvailabilityIn) error {
	m.setIn = &in
	return m.setErr
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

func TestRun_HappyPath(t *testing.T) {
	venueID := uuid.New()
	until := time.Now().Add(time.Hour)
	mock := &catalogMock{}
	uc := item_availability_set.New(mock)

	out, err := uc.Run(context.Background(), item_availability_set.In{
		VenueID:          venueID,
		ExternalID:       "item-1",
		IsAvailable:      false,
		UnavailableUntil: &until,
	})
	require.NoError(t, err)
	require.NotNil(t, out)
	require.NotNil(t, mock.setIn)
	assert.Equal(t, venueID, mock.setIn.VenueID)
	assert.Equal(t, "item-1", mock.setIn.ExternalID)
	assert.False(t, mock.setIn.IsAvailable)
	assert.Equal(t, &until, mock.setIn.UnavailableUntil)
}

func TestRun_ItemNotFound(t *testing.T) {
	mock := &catalogMock{setErr: pgerr.ErrNotFound}
	uc := item_availability_set.New(mock)

	_, err := uc.Run(context.Background(), item_availability_set.In{
		VenueID:    uuid.New(),
		ExternalID: "missing",
	})
	assert.ErrorIs(t, err, usecase.ErrMenuItemNotFound)
}

func TestRun_SetAvailabilityError(t *testing.T) {
	mock := &catalogMock{setErr: assert.AnError}
	uc := item_availability_set.New(mock)

	_, err := uc.Run(context.Background(), item_availability_set.In{
		VenueID:    uuid.New(),
		ExternalID: "item-1",
	})
	assert.ErrorIs(t, err, assert.AnError)
}
