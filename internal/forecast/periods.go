package forecast

import "time"

// periodDays is the length, in days, of a single forecast period.
const periodDays = 14

// Periods returns count consecutive 14-day periods starting at anchor's
// civil date (time-of-day and timezone are normalized away; the anchor's
// date is taken in UTC).
//
// If count is 0 or negative, Periods returns an empty, non-nil slice.
func Periods(anchor time.Time, count int) []Period {
	if count < 0 {
		count = 0
	}

	start := civilDate(anchor)
	periods := make([]Period, 0, count)
	for i := range count {
		periodStart := start.AddDate(0, 0, i*periodDays)
		periodEnd := periodStart.AddDate(0, 0, periodDays-1)
		periods = append(periods, Period{Start: periodStart, End: periodEnd})
	}
	return periods
}

// civilDate strips the time-of-day and timezone from t, returning the
// midnight-UTC instant for t's UTC calendar date.
func civilDate(t time.Time) time.Time {
	t = t.UTC()
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}
