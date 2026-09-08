package openhours_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zlatovlaska707/avito-kitchen/internal/domain/openhours"
)

func h(hour, min int) time.Duration {
	return time.Duration(hour)*time.Hour + time.Duration(min)*time.Minute
}

func at(tz *time.Location, month time.Month, day, hour, min int) time.Time {
	return time.Date(2026, month, day, hour, min, 0, 0, tz)
}

func isoWeekday(t time.Time) int {
	wd := int(t.Weekday())
	if wd == 0 {
		return 7
	}
	return wd
}

func intervalFor(t time.Time, opens, closes int) openhours.Interval {
	return openhours.Interval{
		Weekday: isoWeekday(t), OpensAt: h(opens, 0), ClosesAt: h(closes, 0),
	}
}

func TestIsOpenNow_NoSchedule(t *testing.T) {
	open, reason := openhours.IsOpenNow(nil, "UTC", time.Now(), time.UTC)
	assert.False(t, open)
	assert.Equal(t, openhours.ReasonNoSchedule, reason)
}

func TestIsOpenNow_InsideInterval(t *testing.T) {
	now := at(time.UTC, time.September, 8, 15, 0) // Tuesday
	intervals := []openhours.Interval{intervalFor(now, 10, 22)}
	open, reason := openhours.IsOpenNow(intervals, "UTC", now, time.UTC)
	assert.True(t, open)
	assert.Equal(t, openhours.ReasonOpen, reason)
}

func TestIsOpenNow_ExactlyOpensAt(t *testing.T) {
	now := at(time.UTC, time.September, 8, 10, 0)
	intervals := []openhours.Interval{intervalFor(now, 10, 22)}
	open, _ := openhours.IsOpenNow(intervals, "UTC", now, time.UTC)
	assert.True(t, open)
}

func TestIsOpenNow_ExactlyClosesAt(t *testing.T) {
	now := at(time.UTC, time.September, 8, 22, 0)
	intervals := []openhours.Interval{intervalFor(now, 10, 22)}
	open, reason := openhours.IsOpenNow(intervals, "UTC", now, time.UTC)
	assert.False(t, open)
	assert.Equal(t, openhours.ReasonOutOfSchedule, reason)
}

func TestIsOpenNow_OutsideInterval(t *testing.T) {
	now := at(time.UTC, time.September, 8, 9, 59)
	intervals := []openhours.Interval{intervalFor(now, 10, 22)}
	open, reason := openhours.IsOpenNow(intervals, "UTC", now, time.UTC)
	assert.False(t, open)
	assert.Equal(t, openhours.ReasonOutOfSchedule, reason)
}

func TestIsOpenNow_CrossesMidnight_FromYesterday(t *testing.T) {
	now := at(time.UTC, time.September, 6, 1, 30) // Sunday morning
	prev := now.AddDate(0, 0, -1)
	intervals := []openhours.Interval{{
		Weekday: isoWeekday(prev), OpensAt: h(22, 0), ClosesAt: h(4, 0), CrossesMidnight: true,
	}}
	open, reason := openhours.IsOpenNow(intervals, "UTC", now, time.UTC)
	assert.True(t, open)
	assert.Equal(t, openhours.ReasonOpen, reason)
}

func TestIsOpenNow_CrossesMidnight_SameDayAfterOpen(t *testing.T) {
	now := at(time.UTC, time.September, 5, 23, 0) // Friday
	intervals := []openhours.Interval{{
		Weekday: isoWeekday(now), OpensAt: h(22, 0), ClosesAt: h(4, 0), CrossesMidnight: true,
	}}
	open, _ := openhours.IsOpenNow(intervals, "UTC", now, time.UTC)
	assert.True(t, open)
}

func TestIsOpenNow_TwoShiftsSameDay(t *testing.T) {
	now := at(time.UTC, time.September, 3, 13, 0) // Wednesday lunch
	wd := isoWeekday(now)
	intervals := []openhours.Interval{
		{Weekday: wd, OpensAt: h(8, 0), ClosesAt: h(12, 0)},
		{Weekday: wd, OpensAt: h(14, 0), ClosesAt: h(20, 0)},
	}
	open, reason := openhours.IsOpenNow(intervals, "UTC", now, time.UTC)
	assert.False(t, open)
	assert.Equal(t, openhours.ReasonOutOfSchedule, reason)

	afternoon := at(time.UTC, time.September, 3, 15, 0)
	open, reason = openhours.IsOpenNow(intervals, "UTC", afternoon, time.UTC)
	assert.True(t, open)
	assert.Equal(t, openhours.ReasonOpen, reason)
}

func TestIsOpenNow_ForeignTimezone(t *testing.T) {
	now := at(time.UTC, time.September, 8, 6, 0) // 09:00 Moscow
	local := now.In(mustLocation(t, "Europe/Moscow"))
	intervals := []openhours.Interval{intervalFor(local, 9, 18)}
	open, _ := openhours.IsOpenNow(intervals, "Europe/Moscow", now, time.UTC)
	assert.True(t, open)
}

func TestIsOpenNow_AsiaKolkata_HalfHourOffset(t *testing.T) {
	now := at(time.UTC, time.September, 9, 5, 0) // 10:30 IST
	local := now.In(mustLocation(t, "Asia/Kolkata"))
	intervals := []openhours.Interval{{
		Weekday: isoWeekday(local), OpensAt: h(10, 30), ClosesAt: h(14, 30),
	}}
	open, reason := openhours.IsOpenNow(intervals, "Asia/Kolkata", now, time.UTC)
	assert.True(t, open)
	assert.Equal(t, openhours.ReasonOpen, reason)
}

func TestIsOpenNow_InvalidTZFallsBackToDefault(t *testing.T) {
	defaultTZ := mustLocation(t, "Europe/Moscow")
	now := at(defaultTZ, time.September, 8, 15, 0)
	intervals := []openhours.Interval{intervalFor(now, 10, 22)}
	open, _ := openhours.IsOpenNow(intervals, "Invalid/Timezone", now, defaultTZ)
	assert.True(t, open)
}

func TestIsOpenNow_SundayISOWeekday(t *testing.T) {
	now := at(time.UTC, time.September, 7, 15, 0) // Sunday
	intervals := []openhours.Interval{intervalFor(now, 12, 20)}
	open, reason := openhours.IsOpenNow(intervals, "UTC", now, time.UTC)
	assert.True(t, open)
	assert.Equal(t, openhours.ReasonOpen, reason)
}

func mustLocation(t *testing.T, name string) *time.Location {
	t.Helper()
	loc, err := time.LoadLocation(name)
	require.NoError(t, err)
	return loc
}
