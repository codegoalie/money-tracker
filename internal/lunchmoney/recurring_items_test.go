package lunchmoney_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/codegoalie/money-tracker/internal/lunchmoney"
)

func mustParseDate(t *testing.T, s string) time.Time {
	t.Helper()
	d, err := time.Parse("2006-01-02", s)
	if err != nil {
		t.Fatalf("parsing date %q: %v", s, err)
	}
	return d
}

func TestGetRecurringItems_HappyPathAndSortedOccurrences(t *testing.T) {
	body := mustReadFixture(t, "recurring_items.json")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	}))
	defer srv.Close()

	client := lunchmoney.New(srv.URL, "test-token")
	start := mustParseDate(t, "2026-07-01")
	end := mustParseDate(t, "2026-08-31")
	items, err := client.GetRecurringItems(context.Background(), start, end)
	if err != nil {
		t.Fatalf("GetRecurringItems returned error: %v", err)
	}

	if len(items) != 2 {
		t.Fatalf("GetRecurringItems returned %d items, want 2: %+v", len(items), items)
	}

	payroll := items[0]
	if payroll.ID != 10 || payroll.Payee != "Acme Employer Payroll" || payroll.Amount != 3100.00 ||
		payroll.ToBase != 3100.00 || !payroll.IsIncome || payroll.Granularity != "weekly" || payroll.Quantity != 2 {
		t.Fatalf("payroll item = %+v, unexpected field values", payroll)
	}
	wantOccurrences := []lunchmoney.Occurrence{
		{Date: mustParseDate(t, "2026-07-10"), Amount: 3100.00},
		{Date: mustParseDate(t, "2026-07-24"), Amount: 3100.00},
		{Date: mustParseDate(t, "2026-08-07"), Amount: 3100.00},
		{Date: mustParseDate(t, "2026-08-21"), Amount: 3100.00},
	}
	if len(payroll.Occurrences) != len(wantOccurrences) {
		t.Fatalf("payroll occurrences = %+v, want %+v", payroll.Occurrences, wantOccurrences)
	}
	for i, want := range wantOccurrences {
		got := payroll.Occurrences[i]
		if !got.Date.Equal(want.Date) || got.Amount != want.Amount {
			t.Fatalf("payroll occurrence[%d] = %+v, want %+v (occurrences must be sorted ascending by date)", i, got, want)
		}
	}

	rent := items[1]
	if rent.ID != 11 || rent.Payee != "Maple Street Apartments" || rent.Amount != 1450.00 ||
		rent.ToBase != -1450.00 || rent.IsIncome {
		t.Fatalf("rent item = %+v, unexpected field values", rent)
	}
	if len(rent.Occurrences) != 2 {
		t.Fatalf("rent occurrences = %+v, want 2 entries", rent.Occurrences)
	}
	if !rent.Occurrences[0].Date.Equal(mustParseDate(t, "2026-07-01")) || rent.Occurrences[0].Amount != -1450.00 {
		t.Fatalf("rent occurrence[0] = %+v, want 2026-07-01/-1450.00 (occurrences must be sorted ascending by date)", rent.Occurrences[0])
	}
	if !rent.Occurrences[1].Date.Equal(mustParseDate(t, "2026-08-01")) || rent.Occurrences[1].Amount != -1450.00 {
		t.Fatalf("rent occurrence[1] = %+v, want 2026-08-01/-1450.00", rent.Occurrences[1])
	}
}

func TestGetRecurringItems_SendsAuthorizationHeaderAndQueryParams(t *testing.T) {
	body := mustReadFixture(t, "recurring_items.json")
	var gotHeader string
	var gotQuery = map[string]string{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeader = r.Header.Get("Authorization")
		gotQuery["debit_as_negative"] = r.URL.Query().Get("debit_as_negative")
		gotQuery["start_date"] = r.URL.Query().Get("start_date")
		gotQuery["end_date"] = r.URL.Query().Get("end_date")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	}))
	defer srv.Close()

	client := lunchmoney.New(srv.URL, "secret-token-123")
	start := mustParseDate(t, "2026-07-01")
	end := mustParseDate(t, "2026-08-31")
	if _, err := client.GetRecurringItems(context.Background(), start, end); err != nil {
		t.Fatalf("GetRecurringItems returned error: %v", err)
	}

	if want := "Bearer secret-token-123"; gotHeader != want {
		t.Fatalf("Authorization header = %q, want %q", gotHeader, want)
	}
	if gotQuery["debit_as_negative"] != "true" {
		t.Fatalf("debit_as_negative query param = %q, want %q", gotQuery["debit_as_negative"], "true")
	}
	if gotQuery["start_date"] != "2026-07-01" {
		t.Fatalf("start_date query param = %q, want %q", gotQuery["start_date"], "2026-07-01")
	}
	if gotQuery["end_date"] != "2026-08-31" {
		t.Fatalf("end_date query param = %q, want %q", gotQuery["end_date"], "2026-08-31")
	}
}

func TestGetRecurringItems_MalformedAmount(t *testing.T) {
	body := mustReadFixture(t, "recurring_items_malformed_amount.json")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	}))
	defer srv.Close()

	client := lunchmoney.New(srv.URL, "test-token")
	start := mustParseDate(t, "2026-07-01")
	end := mustParseDate(t, "2026-08-31")
	_, err := client.GetRecurringItems(context.Background(), start, end)
	if err == nil {
		t.Fatal("expected error for malformed amount string, got nil")
	}
}

func TestGetRecurringItems_MalformedOccurrenceAmount(t *testing.T) {
	body := mustReadFixture(t, "recurring_items_malformed_occurrence.json")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	}))
	defer srv.Close()

	client := lunchmoney.New(srv.URL, "test-token")
	start := mustParseDate(t, "2026-07-01")
	end := mustParseDate(t, "2026-08-31")
	_, err := client.GetRecurringItems(context.Background(), start, end)
	if err == nil {
		t.Fatal("expected error for malformed occurrence amount string, got nil")
	}
}
