package forecast_test

import (
	"testing"
	"time"

	"github.com/codegoalie/money-tracker/internal/forecast"
)

func TestEvaluate(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		periods []forecast.Period
		floor   float64
		want    forecast.Verdict
	}{
		"all periods above floor": {
			periods: []forecast.Period{
				{Start: date(2026, 1, 1), End: date(2026, 1, 14), EndBalance: 500},
				{Start: date(2026, 1, 15), End: date(2026, 1, 28), EndBalance: 300},
				{Start: date(2026, 1, 29), End: date(2026, 2, 11), EndBalance: 800},
			},
			floor: 100,
			want: forecast.Verdict{
				OK:           true,
				MinBalance:   300,
				MinDate:      date(2026, 1, 28),
				BreachPeriod: nil,
			},
		},
		"breach in a middle period reports the first breach": {
			periods: []forecast.Period{
				{Start: date(2026, 1, 1), End: date(2026, 1, 14), EndBalance: 500},
				{Start: date(2026, 1, 15), End: date(2026, 1, 28), EndBalance: 50},
				{Start: date(2026, 1, 29), End: date(2026, 2, 11), EndBalance: 20},
			},
			floor: 100,
			want: forecast.Verdict{
				OK:         false,
				MinBalance: 20,
				MinDate:    date(2026, 2, 11),
				BreachPeriod: &forecast.Period{
					Start: date(2026, 1, 15), End: date(2026, 1, 28), EndBalance: 50,
				},
			},
		},
		"EndBalance exactly equal to floor is OK, not a breach": {
			periods: []forecast.Period{
				{Start: date(2026, 1, 1), End: date(2026, 1, 14), EndBalance: 100},
			},
			floor: 100,
			want: forecast.Verdict{
				OK:           true,
				MinBalance:   100,
				MinDate:      date(2026, 1, 14),
				BreachPeriod: nil,
			},
		},
		"empty periods slice is OK with zero-value minimums": {
			periods: nil,
			floor:   100,
			want: forecast.Verdict{
				OK:           true,
				MinBalance:   0,
				MinDate:      time.Time{},
				BreachPeriod: nil,
			},
		},
		"negative floor allows negative balances above it": {
			periods: []forecast.Period{
				{Start: date(2026, 1, 1), End: date(2026, 1, 14), EndBalance: -50},
				{Start: date(2026, 1, 15), End: date(2026, 1, 28), EndBalance: -200},
			},
			floor: -100,
			want: forecast.Verdict{
				OK:         false,
				MinBalance: -200,
				MinDate:    date(2026, 1, 28),
				BreachPeriod: &forecast.Period{
					Start: date(2026, 1, 15), End: date(2026, 1, 28), EndBalance: -200,
				},
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := forecast.Evaluate(tt.periods, tt.floor)

			if got.OK != tt.want.OK {
				t.Errorf("OK = %v, want %v", got.OK, tt.want.OK)
			}
			if got.MinBalance != tt.want.MinBalance {
				t.Errorf("MinBalance = %v, want %v", got.MinBalance, tt.want.MinBalance)
			}
			if !got.MinDate.Equal(tt.want.MinDate) {
				t.Errorf("MinDate = %v, want %v", got.MinDate, tt.want.MinDate)
			}
			switch {
			case tt.want.BreachPeriod == nil && got.BreachPeriod != nil:
				t.Errorf("BreachPeriod = %+v, want nil", got.BreachPeriod)
			case tt.want.BreachPeriod != nil && got.BreachPeriod == nil:
				t.Errorf("BreachPeriod = nil, want %+v", tt.want.BreachPeriod)
			case tt.want.BreachPeriod != nil && got.BreachPeriod != nil:
				if !got.BreachPeriod.Start.Equal(tt.want.BreachPeriod.Start) ||
					!got.BreachPeriod.End.Equal(tt.want.BreachPeriod.End) ||
					got.BreachPeriod.EndBalance != tt.want.BreachPeriod.EndBalance {
					t.Errorf("BreachPeriod = %+v, want %+v", got.BreachPeriod, tt.want.BreachPeriod)
				}
			}
		})
	}
}
