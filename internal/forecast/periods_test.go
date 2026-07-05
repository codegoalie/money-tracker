package forecast_test

import (
	"testing"
	"time"

	"github.com/codegoalie/money-tracker/internal/forecast"
)

func TestPeriods(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		anchor time.Time
		count  int
		want   []forecast.Period
	}{
		"typical 13-period 6-month case": {
			anchor: time.Date(2026, time.January, 5, 9, 30, 0, 0, time.UTC),
			count:  13,
			want: []forecast.Period{
				{Start: date(2026, 1, 5), End: date(2026, 1, 18)},
				{Start: date(2026, 1, 19), End: date(2026, 2, 1)},
				{Start: date(2026, 2, 2), End: date(2026, 2, 15)},
				{Start: date(2026, 2, 16), End: date(2026, 3, 1)},
				{Start: date(2026, 3, 2), End: date(2026, 3, 15)},
				{Start: date(2026, 3, 16), End: date(2026, 3, 29)},
				{Start: date(2026, 3, 30), End: date(2026, 4, 12)},
				{Start: date(2026, 4, 13), End: date(2026, 4, 26)},
				{Start: date(2026, 4, 27), End: date(2026, 5, 10)},
				{Start: date(2026, 5, 11), End: date(2026, 5, 24)},
				{Start: date(2026, 5, 25), End: date(2026, 6, 7)},
				{Start: date(2026, 6, 8), End: date(2026, 6, 21)},
				{Start: date(2026, 6, 22), End: date(2026, 7, 5)},
			},
		},
		"count zero returns empty slice": {
			anchor: date(2026, 1, 5),
			count:  0,
			want:   []forecast.Period{},
		},
		"month and year rollover": {
			anchor: date(2025, 12, 20),
			count:  3,
			want: []forecast.Period{
				{Start: date(2025, 12, 20), End: date(2026, 1, 2)},
				{Start: date(2026, 1, 3), End: date(2026, 1, 16)},
				{Start: date(2026, 1, 17), End: date(2026, 1, 30)},
			},
		},
		"time-of-day and timezone are normalized away": {
			anchor: time.Date(2026, time.March, 10, 23, 59, 59, 0, time.FixedZone("EST", -5*60*60)),
			count:  1,
			// 2026-03-10 23:59:59 -05:00 is 2026-03-11 04:59:59 UTC; the
			// civil date used for bucketing is the UTC calendar date.
			want: []forecast.Period{
				{Start: date(2026, 3, 11), End: date(2026, 3, 24)},
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := forecast.Periods(tt.anchor, tt.count)

			if len(got) != len(tt.want) {
				t.Fatalf("Periods() returned %d periods, want %d", len(got), len(tt.want))
			}
			for i := range got {
				if !got[i].Start.Equal(tt.want[i].Start) {
					t.Errorf("period %d Start = %v, want %v", i, got[i].Start, tt.want[i].Start)
				}
				if !got[i].End.Equal(tt.want[i].End) {
					t.Errorf("period %d End = %v, want %v", i, got[i].End, tt.want[i].End)
				}
			}
		})
	}
}

func date(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}
