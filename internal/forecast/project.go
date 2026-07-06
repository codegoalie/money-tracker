package forecast

import "time"

// Project buckets events into the periods that contain their date and
// computes a running balance across the periods.
//
// Each event is assigned to the period whose [Start, End] range (inclusive
// on both ends) contains the event's civil date. Events dated before the
// first period's Start or after the last period's End are ignored. Income
// is the sum of positive-amount events in a period; Expenses is the sum of
// negative-amount events (kept negative). Net is Income+Expenses, and
// EndBalance is the running balance starting from startingBalance.
//
// Project returns a new slice; it does not mutate periods.
func Project(startingBalance float64, events []CashEvent, periods []Period) []Period {
	result := make([]Period, len(periods))
	for i, p := range periods {
		result[i] = Period{Start: p.Start, End: p.End}
	}

	for _, e := range events {
		idx := periodIndexFor(e.Date, result)
		if idx == -1 {
			continue
		}
		switch {
		case e.Amount > 0:
			result[idx].Income += e.Amount
		case e.Amount < 0:
			result[idx].Expenses += e.Amount
		}
	}

	balance := startingBalance
	for i := range result {
		result[i].Net = result[i].Income + result[i].Expenses
		balance += result[i].Net
		result[i].EndBalance = balance
	}

	return result
}

// periodIndexFor returns the index of the period in periods whose [Start,
// End] range contains the civil date of d, or -1 if no period contains it.
func periodIndexFor(d time.Time, periods []Period) int {
	cd := civilDate(d)
	for i, p := range periods {
		if !cd.Before(p.Start) && !cd.After(p.End) {
			return i
		}
	}
	return -1
}
