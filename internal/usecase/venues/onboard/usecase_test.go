package onboard_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zlatovlaska707/avito-kitchen/internal/usecase"
	venuerepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/venue"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase/uctest"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase/venues/onboard"
)

type venueMock struct {
	uctest.VenueRepo
	slugTakenErr error
}

func (m *venueMock) IsSlugTaken(ctx context.Context, slug string, excludeID *uuid.UUID) (bool, error) {
	if m.slugTakenErr != nil {
		return false, m.slugTakenErr
	}
	return m.VenueRepo.IsSlugTaken(ctx, slug, excludeID)
}

func validIn() onboard.In {
	return onboard.In{
		Slug:             "new-venue",
		Name:             "New Venue",
		CityID:           1,
		AddressText:      "123 Main St",
		Lat:              55.75,
		Lon:              37.62,
		Timezone:         "Europe/Moscow",
		AvgPrepMinutes:   30,
		DeliveryFeeMinor: 19900,
		MinOrderMinor:    50000,
		DeliveryRadiusM:  5000,
	}
}

func TestRun_HappyPath(t *testing.T) {
	venueID := uuid.New()
	venues := &uctest.VenueRepo{
		Created: &venuerepo.Venue{
			ID:   venueID,
			Slug: "new-venue",
			Name: "New Venue",
		},
	}
	uc := onboard.New(venues)

	out, err := uc.Run(context.Background(), validIn())
	require.NoError(t, err)
	require.NotNil(t, out)
	assert.Equal(t, venueID, out.Venue.ID)
	assert.Equal(t, "new-venue", out.Venue.Slug)
}

func TestRun_SlugTaken(t *testing.T) {
	venues := &uctest.VenueRepo{SlugTaken: true}
	uc := onboard.New(venues)

	_, err := uc.Run(context.Background(), validIn())
	assert.ErrorIs(t, err, usecase.ErrVenueSlugTaken)
}

func TestRun_IsSlugTakenError(t *testing.T) {
	venues := &venueMock{slugTakenErr: assert.AnError}
	uc := onboard.New(venues)

	_, err := uc.Run(context.Background(), validIn())
	assert.ErrorIs(t, err, assert.AnError)
}

func TestRun_CreateUniqueViolation(t *testing.T) {
	venues := &uctest.VenueRepo{
		CreateErr: &pgconn.PgError{Code: "23505"},
	}
	uc := onboard.New(venues)

	_, err := uc.Run(context.Background(), validIn())
	assert.ErrorIs(t, err, usecase.ErrVenueSlugTaken)
}

func TestRun_CreateError(t *testing.T) {
	venues := &uctest.VenueRepo{CreateErr: assert.AnError}
	uc := onboard.New(venues)

	_, err := uc.Run(context.Background(), validIn())
	assert.ErrorIs(t, err, assert.AnError)
}
