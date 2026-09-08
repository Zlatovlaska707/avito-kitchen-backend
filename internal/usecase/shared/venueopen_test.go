package shared_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	venuerepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/venue"
	"github.com/zlatovlaska707/avito-kitchen/internal/usecase/shared"
)

func TestVenueIsOpen_Nil(t *testing.T) {
	open, reason := shared.VenueIsOpen(nil, nil, time.Now(), time.UTC)
	assert.False(t, open)
	assert.Equal(t, "inactive", reason)
}

func TestVenueIsOpen_Inactive(t *testing.T) {
	v := &venuerepo.Venue{Status: "draft", AcceptsOrders: true}
	open, reason := shared.VenueIsOpen(v, nil, time.Now(), time.UTC)
	assert.False(t, open)
	assert.Equal(t, "inactive", reason)
}

func TestVenueIsOpen_Paused(t *testing.T) {
	v := &venuerepo.Venue{Status: "active", AcceptsOrders: false}
	open, reason := shared.VenueIsOpen(v, nil, time.Now(), time.UTC)
	assert.False(t, open)
	assert.Equal(t, "paused_by_venue", reason)
}

func TestVenueIsOpen_OpenSchedule(t *testing.T) {
	loc, _ := time.LoadLocation("Europe/Moscow")
	now := time.Date(2026, 9, 8, 12, 0, 0, 0, loc)
	v := &venuerepo.Venue{Status: "active", AcceptsOrders: true, Timezone: "Europe/Moscow"}
	hours := []venuerepo.WorkingHour{{
		Weekday: int(now.Weekday()), OpensAt: 9 * time.Hour, ClosesAt: 22 * time.Hour,
	}}
	open, reason := shared.VenueIsOpen(v, hours, now, loc)
	assert.True(t, open)
	assert.Empty(t, reason)
}
