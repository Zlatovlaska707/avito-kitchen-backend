package update_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/repository/pgerr"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase"
	venuerepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/venue"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase/uctest"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase/venues/update"
)

func TestRun_HappyPath(t *testing.T) {
	venueID := uuid.New()
	name := "Updated Name"
	venues := &uctest.VenueRepo{
		Updated: &venuerepo.Venue{ID: venueID, Name: name},
	}
	uc := update.New(venues)

	out, err := uc.Run(context.Background(), update.In{ID: venueID, Name: &name})
	require.NoError(t, err)
	require.NotNil(t, out)
	assert.Equal(t, venueID, out.Venue.ID)
	assert.Equal(t, name, out.Venue.Name)
}

func TestRun_VenueNotFound(t *testing.T) {
	venues := &uctest.VenueRepo{UpdateErr: pgerr.ErrNotFound}
	uc := update.New(venues)

	_, err := uc.Run(context.Background(), update.In{ID: uuid.New()})
	assert.ErrorIs(t, err, usecase.ErrVenueNotFound)
}

func TestRun_UpdateError(t *testing.T) {
	venues := &uctest.VenueRepo{UpdateErr: assert.AnError}
	uc := update.New(venues)

	_, err := uc.Run(context.Background(), update.In{ID: uuid.New()})
	assert.ErrorIs(t, err, assert.AnError)
}
