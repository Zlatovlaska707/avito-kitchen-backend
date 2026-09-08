package stoplist_restore_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	catalogrepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/catalog"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase/catalog/stoplist_restore"
)

type catalogMock struct {
	restored   int
	restoreErr error
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
	return nil, nil
}

func (m *catalogMock) TryMenuSyncLock(context.Context, uuid.UUID) (bool, error) {
	return false, nil
}

func (m *catalogMock) RestoreExpiredStoplist(context.Context) (int, error) {
	if m.restoreErr != nil {
		return 0, m.restoreErr
	}
	return m.restored, nil
}

func TestRun_HappyPath(t *testing.T) {
	mock := &catalogMock{restored: 5}
	uc := stoplist_restore.New(mock)

	out, err := uc.Run(context.Background(), stoplist_restore.In{})
	require.NoError(t, err)
	require.NotNil(t, out)
	assert.Equal(t, 5, out.Restored)
}

func TestRun_RestoreError(t *testing.T) {
	mock := &catalogMock{restoreErr: assert.AnError}
	uc := stoplist_restore.New(mock)

	_, err := uc.Run(context.Background(), stoplist_restore.In{})
	assert.ErrorIs(t, err, assert.AnError)
}

func TestRun_NoItemsRestored(t *testing.T) {
	mock := &catalogMock{restored: 0}
	uc := stoplist_restore.New(mock)

	out, err := uc.Run(context.Background(), stoplist_restore.In{})
	require.NoError(t, err)
	assert.Equal(t, 0, out.Restored)
}
