package shared_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/repository/pgerr"
	venuerepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/venue"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase/shared"
)

type venueRepoMock struct {
	byID    *venuerepo.Venue
	bySlug  *venuerepo.Venue
	idErr   error
	slugErr error
}

func (m *venueRepoMock) Create(context.Context, venuerepo.CreateIn) (*venuerepo.Venue, error) {
	return nil, nil
}

func (m *venueRepoMock) Update(context.Context, venuerepo.UpdateIn) (*venuerepo.Venue, error) {
	return nil, nil
}

func (m *venueRepoMock) GetByID(context.Context, uuid.UUID) (*venuerepo.Venue, error) {
	return m.byID, m.idErr
}

func (m *venueRepoMock) GetBySlug(context.Context, string) (*venuerepo.Venue, error) {
	return m.bySlug, m.slugErr
}

func (m *venueRepoMock) List(context.Context, venuerepo.ListFilter) ([]venuerepo.Venue, error) {
	return nil, nil
}
func (m *venueRepoMock) UpdateStatus(context.Context, venuerepo.StatusUpdateIn) error { return nil }
func (m *venueRepoMock) GetWorkingHours(context.Context, uuid.UUID) ([]venuerepo.WorkingHour, error) {
	return nil, nil
}

func (m *venueRepoMock) ReplaceWorkingHours(context.Context, uuid.UUID, []venuerepo.WorkingHour) error {
	return nil
}

func (m *venueRepoMock) IsSlugTaken(context.Context, string, *uuid.UUID) (bool, error) {
	return false, nil
}

func TestResolveVenue_ByUUID(t *testing.T) {
	id := uuid.New()
	v := &venuerepo.Venue{ID: id, Slug: "pizza"}
	got, err := shared.ResolveVenue(context.Background(), &venueRepoMock{byID: v}, id.String())
	require.NoError(t, err)
	assert.Equal(t, id, got.ID)
}

func TestResolveVenue_BySlug(t *testing.T) {
	v := &venuerepo.Venue{Slug: "pizza"}
	got, err := shared.ResolveVenue(context.Background(), &venueRepoMock{bySlug: v}, "pizza")
	require.NoError(t, err)
	assert.Equal(t, "pizza", got.Slug)
}

func TestResolveVenue_SlugNotFound(t *testing.T) {
	_, err := shared.ResolveVenue(context.Background(), &venueRepoMock{slugErr: pgerr.ErrNotFound}, "missing")
	assert.ErrorIs(t, err, pgerr.ErrNotFound)
}
