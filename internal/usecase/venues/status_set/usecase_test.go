package status_set_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/repository/pgerr"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase/uctest"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase/venues/status_set"
)

func TestRun_HappyPath(t *testing.T) {
	venueID := uuid.New()
	status := "active"
	accepts := true
	venues := &uctest.VenueRepo{}
	uc := status_set.New(venues)

	out, err := uc.Run(context.Background(), status_set.In{
		VenueID:       venueID,
		Status:        &status,
		AcceptsOrders: &accepts,
	})
	require.NoError(t, err)
	require.NotNil(t, out)
}

func TestRun_VenueNotFound(t *testing.T) {
	venues := &uctest.VenueRepo{StatusErr: pgerr.ErrNotFound}
	uc := status_set.New(venues)

	_, err := uc.Run(context.Background(), status_set.In{VenueID: uuid.New()})
	assert.ErrorIs(t, err, usecase.ErrVenueNotFound)
}

func TestRun_UpdateStatusError(t *testing.T) {
	venues := &uctest.VenueRepo{StatusErr: assert.AnError}
	uc := status_set.New(venues)

	_, err := uc.Run(context.Background(), status_set.In{VenueID: uuid.New()})
	assert.ErrorIs(t, err, assert.AnError)
}
