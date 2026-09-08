package openhours

import "time"

type ClosedReason string

const (
	ReasonOpen          ClosedReason = "open"
	ReasonOutOfSchedule ClosedReason = "out_of_schedule"
	ReasonNoSchedule    ClosedReason = "no_schedule"
)

type Interval struct {
	Weekday         int
	OpensAt         time.Duration
	ClosesAt        time.Duration
	CrossesMidnight bool
}

func IsOpenNow(intervals []Interval, tzName string, now time.Time, defaultTZ *time.Location) (bool, ClosedReason) {
	if len(intervals) == 0 {
		return false, ReasonNoSchedule
	}

	loc := defaultTZ
	if tzName != "" {
		if l, err := time.LoadLocation(tzName); err == nil {
			loc = l
		}
	}
	if loc == nil {
		loc = time.UTC
	}

	local := now.In(loc)
	todayWeekday := isoWeekday(local)
	offsetToday := local.Sub(startOfDay(local))

	for _, iv := range intervals {
		if iv.Weekday != todayWeekday {
			continue
		}
		if iv.CrossesMidnight {
			if offsetToday >= iv.OpensAt {
				return true, ReasonOpen
			}
			continue
		}
		if offsetToday >= iv.OpensAt && offsetToday < iv.ClosesAt {
			return true, ReasonOpen
		}
	}

	yesterday := isoWeekday(local.AddDate(0, 0, -1))
	for _, iv := range intervals {
		if iv.Weekday != yesterday || !iv.CrossesMidnight {
			continue
		}
		if offsetToday < iv.ClosesAt {
			return true, ReasonOpen
		}
	}

	return false, ReasonOutOfSchedule
}

func isoWeekday(t time.Time) int {
	wd := int(t.Weekday())
	if wd == 0 {
		return 7
	}
	return wd
}

func startOfDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}
