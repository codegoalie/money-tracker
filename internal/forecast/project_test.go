package forecast_test

import (
	"testing"

	"github.com/codegoalie/money-tracker/internal/forecast"
)

func TestProject(t *testing.T) {
	t.Parallel()

	onePeriod := []forecast.Period{
		{Start: date(2026, 1, 1), End: date(2026, 1, 14)},
	}
	twoPeriods := []forecast.Period{
		{Start: date(2026, 1, 1), End: date(2026, 1, 14)},
		{Start: date(2026, 1, 15), End: date(2026, 1, 28)},
	}

	tests := map[string]struct {
		startingBalance float64
		events          []forecast.CashEvent
		periods         []forecast.Period
		want            []forecast.Period
	}{
		"empty events": {
			startingBalance: 100,
			events:          nil,
			periods:         twoPeriods,
			want: []forecast.Period{
				{Start: date(2026, 1, 1), End: date(2026, 1, 14), Income: 0, Expenses: 0, Net: 0, EndBalance: 100},
				{Start: date(2026, 1, 15), End: date(2026, 1, 28), Income: 0, Expenses: 0, Net: 0, EndBalance: 100},
			},
		},
		"event exactly on period Start": {
			startingBalance: 100,
			events: []forecast.CashEvent{
				{Date: date(2026, 1, 1), Amount: 50, Payee: "Employer", IsIncome: true},
			},
			periods: onePeriod,
			want: []forecast.Period{
				{Start: date(2026, 1, 1), End: date(2026, 1, 14), Income: 50, Expenses: 0, Net: 50, EndBalance: 150},
			},
		},
		"event exactly on period End is inclusive": {
			startingBalance: 100,
			events: []forecast.CashEvent{
				{Date: date(2026, 1, 14), Amount: -30, Payee: "Rent", IsIncome: false},
			},
			periods: onePeriod,
			want: []forecast.Period{
				{Start: date(2026, 1, 1), End: date(2026, 1, 14), Income: 0, Expenses: -30, Net: -30, EndBalance: 70},
			},
		},
		"event one day past the last period is ignored": {
			startingBalance: 100,
			events: []forecast.CashEvent{
				{Date: date(2026, 1, 15), Amount: 1000, Payee: "Windfall", IsIncome: true},
			},
			periods: onePeriod,
			want: []forecast.Period{
				{Start: date(2026, 1, 1), End: date(2026, 1, 14), Income: 0, Expenses: 0, Net: 0, EndBalance: 100},
			},
		},
		"event one day before the first period is ignored": {
			startingBalance: 100,
			events: []forecast.CashEvent{
				{Date: date(2025, 12, 31), Amount: 1000, Payee: "Windfall", IsIncome: true},
			},
			periods: onePeriod,
			want: []forecast.Period{
				{Start: date(2026, 1, 1), End: date(2026, 1, 14), Income: 0, Expenses: 0, Net: 0, EndBalance: 100},
			},
		},
		"income only across periods": {
			startingBalance: 0,
			events: []forecast.CashEvent{
				{Date: date(2026, 1, 2), Amount: 200, Payee: "Employer", IsIncome: true},
				{Date: date(2026, 1, 16), Amount: 300, Payee: "Employer", IsIncome: true},
			},
			periods: twoPeriods,
			want: []forecast.Period{
				{Start: date(2026, 1, 1), End: date(2026, 1, 14), Income: 200, Expenses: 0, Net: 200, EndBalance: 200},
				{Start: date(2026, 1, 15), End: date(2026, 1, 28), Income: 300, Expenses: 0, Net: 300, EndBalance: 500},
			},
		},
		"expense only across periods": {
			startingBalance: 500,
			events: []forecast.CashEvent{
				{Date: date(2026, 1, 5), Amount: -100, Payee: "Rent", IsIncome: false},
				{Date: date(2026, 1, 20), Amount: -50, Payee: "Groceries", IsIncome: false},
			},
			periods: twoPeriods,
			want: []forecast.Period{
				{Start: date(2026, 1, 1), End: date(2026, 1, 14), Income: 0, Expenses: -100, Net: -100, EndBalance: 400},
				{Start: date(2026, 1, 15), End: date(2026, 1, 28), Income: 0, Expenses: -50, Net: -50, EndBalance: 350},
			},
		},
		"multiple events same day": {
			startingBalance: 100,
			events: []forecast.CashEvent{
				{Date: date(2026, 1, 5), Amount: 200, Payee: "Employer", IsIncome: true},
				{Date: date(2026, 1, 5), Amount: -40, Payee: "Groceries", IsIncome: false},
				{Date: date(2026, 1, 5), Amount: -10, Payee: "Coffee", IsIncome: false},
			},
			periods: onePeriod,
			want: []forecast.Period{
				{Start: date(2026, 1, 1), End: date(2026, 1, 14), Income: 200, Expenses: -50, Net: 150, EndBalance: 250},
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := forecast.Project(tt.startingBalance, tt.events, tt.periods)

			if len(got) != len(tt.want) {
				t.Fatalf("Project() returned %d periods, want %d", len(got), len(tt.want))
			}
			for i := range got {
				want := tt.want[i]
				g := got[i]
				if !g.Start.Equal(want.Start) || !g.End.Equal(want.End) {
					t.Errorf("period %d Start/End = %v/%v, want %v/%v", i, g.Start, g.End, want.Start, want.End)
				}
				if g.Income != want.Income {
					t.Errorf("period %d Income = %v, want %v", i, g.Income, want.Income)
				}
				if g.Expenses != want.Expenses {
					t.Errorf("period %d Expenses = %v, want %v", i, g.Expenses, want.Expenses)
				}
				if g.Net != want.Net {
					t.Errorf("period %d Net = %v, want %v", i, g.Net, want.Net)
				}
				if g.EndBalance != want.EndBalance {
					t.Errorf("period %d EndBalance = %v, want %v", i, g.EndBalance, want.EndBalance)
				}
			}
		})
	}
}
