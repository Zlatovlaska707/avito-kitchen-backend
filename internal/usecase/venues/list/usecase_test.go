package list_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	venuerepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/venue"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase/uctest"
	venuelist "github.com/zlatovlaska707/avito-kitchen/internal/usecase/venues/list"
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

func closedHours(now time.Time) []venuerepo.WorkingHour {
	return []venuerepo.WorkingHour{{
		Weekday: weekdayISO(now), OpensAt: 14 * time.Hour, ClosesAt: 22 * time.Hour,
	}}
}

func TestRun_DefaultLimit(t *testing.T) {
	now := time.Date(2026, time.September, 8, 12, 0, 0, 0, time.UTC)
	venues := make([]venuerepo.Venue, 21)
	for i := range venues {
		venues[i] = venuerepo.Venue{
			ID: uuid.New(), Name: "Venue", Status: "active", AcceptsOrders: true, Timezone: "UTC",
			AvgPrepMinutes: 20, CreatedAt: now.Add(time.Duration(i) * time.Minute),
		}
	}
	uc := venuelist.New(
		&uctest.VenueRepo{Venues: venues, Hours: openHours(now)},
		uctest.FixedNower{T: now},
		time.UTC,
	)

	out, err := uc.Run(context.Background(), venuelist.In{})
	require.NoError(t, err)
	assert.Len(t, out.Items, 20)
	assert.True(t, out.HasNext)
}

type hoursByVenueRepo struct {
	uctest.VenueRepo
	openHours   []venuerepo.WorkingHour
	closedHours []venuerepo.WorkingHour
	openVenueID uuid.UUID
}

func (m *hoursByVenueRepo) GetWorkingHours(_ context.Context, venueID uuid.UUID) ([]venuerepo.WorkingHour, error) {
	if venueID == m.openVenueID {
		return m.openHours, nil
	}
	return m.closedHours, nil
}

func TestRun_OnlyOpenFilter(t *testing.T) {
	now := time.Date(2026, time.September, 8, 12, 0, 0, 0, time.UTC)
	openID := uuid.New()
	closedID := uuid.New()
	repo := &hoursByVenueRepo{
		VenueRepo: uctest.VenueRepo{
			Venues: []venuerepo.Venue{
				{ID: openID, Name: "Open", Status: "active", AcceptsOrders: true, Timezone: "UTC", AvgPrepMinutes: 15, CreatedAt: now},
				{ID: closedID, Name: "Closed", Status: "active", AcceptsOrders: true, Timezone: "UTC", AvgPrepMinutes: 15, CreatedAt: now.Add(time.Minute)},
			},
		},
		openVenueID: openID,
		openHours:   openHours(now),
		closedHours: closedHours(now),
	}
	uc := venuelist.New(repo, uctest.FixedNower{T: now}, time.UTC)

	out, err := uc.Run(context.Background(), venuelist.In{OnlyOpen: true, Limit: 10})
	require.NoError(t, err)
	require.Len(t, out.Items, 1)
	assert.Equal(t, openID, out.Items[0].Venue.ID)
}

func TestRun_DistanceAndETA(t *testing.T) {
	now := time.Date(2026, time.September, 8, 12, 0, 0, 0, time.UTC)
	venueID := uuid.New()
	lat := 55.7558
	lon := 37.6173
	repo := &uctest.VenueRepo{
		Venues: []venuerepo.Venue{{
			ID: venueID, Name: "Near", Status: "active", AcceptsOrders: true, Timezone: "UTC",
			AvgPrepMinutes: 25, Lat: lat, Lon: lon, CreatedAt: now,
		}},
		Hours: openHours(now),
	}
	uc := venuelist.New(repo, uctest.FixedNower{T: now}, time.UTC)

	out, err := uc.Run(context.Background(), venuelist.In{
		Lat: &lat, Lon: &lon, Limit: 10,
	})
	require.NoError(t, err)
	require.Len(t, out.Items, 1)
	require.NotNil(t, out.Items[0].ETAMinutes)
	assert.Equal(t, 35, *out.Items[0].ETAMinutes)
	require.NotNil(t, out.Items[0].DistanceM)
	assert.Equal(t, 0, *out.Items[0].DistanceM)
}

func TestRun_PaginationCursor(t *testing.T) {
	now := time.Date(2026, time.September, 8, 12, 0, 0, 0, time.UTC)
	v1 := venuerepo.Venue{ID: uuid.New(), Name: "V1", Status: "active", AcceptsOrders: true, Timezone: "UTC", CreatedAt: now}
	v2 := venuerepo.Venue{ID: uuid.New(), Name: "V2", Status: "active", AcceptsOrders: true, Timezone: "UTC", CreatedAt: now.Add(time.Minute)}
	uc := venuelist.New(
		&uctest.VenueRepo{Venues: []venuerepo.Venue{v1, v2}, Hours: closedHours(now)},
		uctest.FixedNower{T: now},
		time.UTC,
	)

	out, err := uc.Run(context.Background(), venuelist.In{Limit: 1})
	require.NoError(t, err)
	require.Len(t, out.Items, 1)
	assert.True(t, out.HasNext)
	require.NotNil(t, out.LastTS)
	require.NotNil(t, out.LastID)
	assert.Equal(t, v1.ID, *out.LastID)
}

func TestRun_ListError(t *testing.T) {
	now := time.Date(2026, time.September, 8, 12, 0, 0, 0, time.UTC)
	uc := venuelist.New(
		&uctest.VenueRepo{ListErr: assert.AnError},
		uctest.FixedNower{T: now},
		time.UTC,
	)

	_, err := uc.Run(context.Background(), venuelist.In{})
	assert.ErrorIs(t, err, assert.AnError)
}
