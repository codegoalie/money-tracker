package lunchmoney

import (
	"context"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"time"
)

// Occurrence is a single dated occurrence of a RecurringItem within the
// requested date range.
//
// Amount follows the "debit_as_negative" convention described on
// GetRecurringItems: negative for expenses, positive for income.
type Occurrence struct {
	Date   time.Time
	Amount float64
}

// RecurringItem is a recurring income or expense definition, as returned by
// GET /v1/recurring_items, together with its dated Occurrences within the
// requested date range.
//
// Amount is the item's magnitude as configured in LunchMoney (its sign is
// not guaranteed by the API); ToBase and each Occurrence's Amount carry the
// signed, "debit_as_negative" convention. Granularity (e.g. "weekly",
// "monthly") and Quantity are cadence metadata describing how often the item
// recurs; callers that need dated cash events should use Occurrences rather
// than reimplementing cadence math from Granularity/Quantity.
type RecurringItem struct {
	ID          int
	Payee       string
	Amount      float64
	ToBase      float64
	IsIncome    bool
	Granularity string
	Quantity    int
	Occurrences []Occurrence
}

// recurringItemWire mirrors the wire format of a single element of the
// GET /v1/recurring_items response array. Amount is a JSON string; each
// value in Occurrences is also a JSON string.
type recurringItemWire struct {
	ID          int               `json:"id"`
	Payee       string            `json:"payee"`
	Amount      string            `json:"amount"`
	ToBase      float64           `json:"to_base"`
	IsIncome    bool              `json:"is_income"`
	Granularity string            `json:"granularity"`
	Quantity    int               `json:"quantity"`
	Occurrences map[string]string `json:"occurrences"`
}

// dateQueryLayout is the YYYY-MM-DD layout the LunchMoney API expects for
// start_date/end_date query parameters and uses for recurring_items
// occurrence keys.
const dateQueryLayout = "2006-01-02"

// GetRecurringItems fetches recurring income/expense items with their dated
// occurrences between startDate and endDate (inclusive, per the LunchMoney
// API), from GET /v1/recurring_items. Dates are sent as YYYY-MM-DD.
//
// The request always sets debit_as_negative=true, so the API returns
// expenses as negative amounts and income as positive amounts; this client
// passes that sign straight through into RecurringItem.ToBase and
// Occurrence.Amount without any additional sign-flipping.
//
// Each RecurringItem's Occurrences are sorted ascending by Date before
// being returned, since the API represents them as a JSON object (map) and
// Go's JSON decoding of a map does not preserve or guarantee any particular
// key order.
func (c *Client) GetRecurringItems(ctx context.Context, startDate, endDate time.Time) ([]RecurringItem, error) {
	q := url.Values{}
	q.Set("debit_as_negative", "true")
	q.Set("start_date", startDate.Format(dateQueryLayout))
	q.Set("end_date", endDate.Format(dateQueryLayout))

	var wireItems []recurringItemWire
	if err := c.get("/v1/recurring_items", q.Encode(), &wireItems); err != nil {
		return nil, err
	}

	items := make([]RecurringItem, 0, len(wireItems))
	for _, w := range wireItems {
		amount, err := strconv.ParseFloat(w.Amount, 64)
		if err != nil {
			return nil, fmt.Errorf("lunchmoney: parsing amount %q for recurring item %d (%s): %w", w.Amount, w.ID, w.Payee, err)
		}

		occurrences := make([]Occurrence, 0, len(w.Occurrences))
		for dateStr, amountStr := range w.Occurrences {
			date, err := time.Parse(dateQueryLayout, dateStr)
			if err != nil {
				return nil, fmt.Errorf("lunchmoney: parsing occurrence date %q for recurring item %d (%s): %w", dateStr, w.ID, w.Payee, err)
			}
			occAmount, err := strconv.ParseFloat(amountStr, 64)
			if err != nil {
				return nil, fmt.Errorf("lunchmoney: parsing occurrence amount %q for recurring item %d (%s) on %s: %w", amountStr, w.ID, w.Payee, dateStr, err)
			}
			occurrences = append(occurrences, Occurrence{Date: date, Amount: occAmount})
		}
		sort.Slice(occurrences, func(i, j int) bool {
			return occurrences[i].Date.Before(occurrences[j].Date)
		})

		items = append(items, RecurringItem{
			ID:          w.ID,
			Payee:       w.Payee,
			Amount:      amount,
			ToBase:      w.ToBase,
			IsIncome:    w.IsIncome,
			Granularity: w.Granularity,
			Quantity:    w.Quantity,
			Occurrences: occurrences,
		})
	}
	return items, nil
}
