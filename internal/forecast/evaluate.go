package forecast

// Evaluate checks a projected series of Periods against a minimum balance
// floor.
//
// OK is true iff every period's EndBalance is >= floor (a balance exactly
// equal to the floor is OK, not a breach). MinBalance and MinDate track the
// smallest EndBalance across all periods and the End date of the period it
// occurred in. BreachPeriod points to a copy of the FIRST period whose
// EndBalance fell below floor, and is nil when OK is true.
//
// For an empty periods slice, Evaluate returns OK: true with MinBalance and
// MinDate left at their zero values, since there is nothing to evaluate.
func Evaluate(periods []Period, floor float64) Verdict {
	v := Verdict{OK: true}

	for i, p := range periods {
		if i == 0 || p.EndBalance < v.MinBalance {
			v.MinBalance = p.EndBalance
			v.MinDate = p.End
		}
		if p.EndBalance < floor {
			v.OK = false
			if v.BreachPeriod == nil {
				breach := p
				v.BreachPeriod = &breach
			}
		}
	}

	return v
}
