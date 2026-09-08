package shared

import (
	"time"

	"github.com/zlatovlaska707/avito-kitchen/internal/domain/openhours"
	venuerepo "github.com/zlatovlaska707/avito-kitchen/internal/usecase/contract/repository/venue"
)

func VenueIsOpen(v *venuerepo.Venue, hours []venuerepo.WorkingHour, now time.Time, defaultTZ *time.Location) (bool, string) {
	if v == nil {
		return false, "inactive"
	}
	if v.Status != "active" {
		return false, "inactive"
	}
	if !v.AcceptsOrders {
		return false, "paused_by_venue"
	}
	intervals := make([]openhours.Interval, 0, len(hours))
	for _, h := range hours {
		intervals = append(intervals, openhours.Interval{
			Weekday:         h.Weekday,
			OpensAt:         h.OpensAt,
			ClosesAt:        h.ClosesAt,
			CrossesMidnight: h.CrossesMidnight,
		})
	}
	open, reason := openhours.IsOpenNow(intervals, v.Timezone, now, defaultTZ)
	if !open {
		if reason == openhours.ReasonNoSchedule || reason == openhours.ReasonOutOfSchedule {
			return false, "schedule"
		}
		return false, "schedule"
	}
	return true, ""
}
