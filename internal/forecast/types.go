// Package forecast is the pure domain core for cash-flow forecasting.
//
// It performs no I/O, makes no HTTP calls, and never reads the system clock;
// all dates and times must be supplied by the caller. All functions operate
// on plain structs in, plain structs out.
package forecast

import "time"

// CashEvent represents a single dated cash movement.
//
// Amount uses a signed convention: positive values are income and negative
// values are expenses. IsIncome mirrors the sign of Amount and is kept as an
// explicit field so callers can classify an event without inspecting the
// sign of Amount.
type CashEvent struct {
	Date     time.Time
	Amount   float64
	Payee    string
	IsIncome bool
}

// Period represents one bi-weekly (14-day) forecast window.
//
// End is the last day INCLUDED in the period (inclusive), not exclusive. For
// a 14-day period, End = Start + 13 days.
type Period struct {
	Start      time.Time
	End        time.Time
	Income     float64
	Expenses   float64
	Net        float64
	EndBalance float64
}

// Verdict is the result of evaluating a projected series of Periods against
// a minimum balance floor.
//
// OK is true iff every period's EndBalance is greater than or equal to the
// floor. BreachPeriod points to the FIRST period whose EndBalance fell below
// the floor, and is nil when OK is true.
type Verdict struct {
	OK           bool
	MinBalance   float64
	MinDate      time.Time
	BreachPeriod *Period
}
