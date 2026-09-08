package get_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zlatovlaska707/avito-kitchen/internal/infrastructure/repository/pgerr"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase"
	venuerepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/venue"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase/uctest"
	venueget "github.com/zlatovlaska707/avito-kitchen/internal/usecase/venues/get"
)

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

func TestRun_VenueNotFoundByID(t *testing.T) {
	now := time.Date(2026, time.September, 8, 12, 0, 0, 0, time.UTC)
	uc := venueget.New(
		&uctest.VenueRepo{GetByIDErr: pgerr.ErrNotFound},
		uctest.FixedNower{T: now},
		time.UTC,
	)

	_, err := uc.Run(context.Background(), venueget.In{VenueRef: uuid.New().String()})
	assert.ErrorIs(t, err, usecase.ErrVenueNotFound)
}

func TestRun_VenueNotFoundBySlug(t *testing.T) {
	now := time.Date(2026, time.September, 8, 12, 0, 0, 0, time.UTC)
	uc := venueget.New(
		&uctest.VenueRepo{GetBySlugErr: pgerr.ErrNotFound},
		uctest.FixedNower{T: now},
		time.UTC,
	)

	_, err := uc.Run(context.Background(), venueget.In{VenueRef: "missing-slug"})
	assert.ErrorIs(t, err, usecase.ErrVenueNotFound)
}

func TestRun_InactiveVenue(t *testing.T) {
	now := time.Date(2026, time.September, 8, 12, 0, 0, 0, time.UTC)
	venueID := uuid.New()
	uc := venueget.New(
		&uctest.VenueRepo{Venue: &venuerepo.Venue{ID: venueID, Status: "paused", AvgPrepMinutes: 25}},
		uctest.FixedNower{T: now},
		time.UTC,
	)

	_, err := uc.Run(context.Background(), venueget.In{VenueRef: venueID.String()})
	assert.ErrorIs(t, err, usecase.ErrVenueNotFound)
}

func TestRun_HappyPath(t *testing.T) {
	now := time.Date(2026, time.September, 8, 12, 0, 0, 0, time.UTC)
	venueID := uuid.New()
	venue := &venuerepo.Venue{
		ID: venueID, Slug: "pizza-mario", Name: "Pizza Mario", Status: "active",
		AcceptsOrders: true, Timezone: "UTC", AvgPrepMinutes: 30,
	}
	uc := venueget.New(
		&uctest.VenueRepo{Venue: venue, Hours: openHours(now)},
		uctest.FixedNower{T: now},
		time.UTC,
	)

	out, err := uc.Run(context.Background(), venueget.In{VenueRef: venueID.String()})
	require.NoError(t, err)
	assert.Equal(t, venueID, out.Venue.ID)
	assert.True(t, out.IsOpen)
	assert.Nil(t, out.ClosedReason)
	assert.Equal(t, 40, out.ETAMinutes)
	assert.Len(t, out.Hours, 1)
}

func TestRun_ClosedBySchedule(t *testing.T) {
	now := time.Date(2026, time.September, 8, 12, 0, 0, 0, time.UTC)
	venueID := uuid.New()
	uc := venueget.New(
		&uctest.VenueRepo{
			Venue: &venuerepo.Venue{
				ID: venueID, Status: "active", AcceptsOrders: true, Timezone: "UTC", AvgPrepMinutes: 20,
			},
			Hours: []venuerepo.WorkingHour{{
				Weekday: weekdayISO(now), OpensAt: 14 * time.Hour, ClosesAt: 22 * time.Hour,
			}},
		},
		uctest.FixedNower{T: now},
		time.UTC,
	)

	out, err := uc.Run(context.Background(), venueget.In{VenueRef: venueID.String()})
	require.NoError(t, err)
	assert.False(t, out.IsOpen)
	require.NotNil(t, out.ClosedReason)
	assert.Equal(t, "schedule", *out.ClosedReason)
}
