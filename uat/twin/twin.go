// Package twin implements an in-process digital twin of the LunchMoney API
// for use by the black-box UAT suite. It serves canned fixtures over HTTP so
// the suite can exec the real moneytracker binary against a realistic (but
// entirely local and deterministic) backend.
package twin

import (
	"embed"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
)

// Scenario selects which set of canned fixtures a Twin serves.
type Scenario string

const (
	// ScenarioHealthy models an account with comfortable cash flow: income
	// well exceeds recurring expenses, so the projected balance stays
	// healthy over a multi-month horizon.
	ScenarioHealthy Scenario = "healthy"
	// ScenarioFloorBreach models an account where large recurring expenses
	// outpace income, driving the projected balance below zero partway
	// through the forecast horizon.
	ScenarioFloorBreach Scenario = "floor-breach"
)

// ValidToken is the bearer token every Twin accepts, regardless of scenario.
// Any request presenting a different token receives a 401.
const ValidToken = "test-valid-token"

//go:embed testdata
var fixturesFS embed.FS

// Twin is a running digital twin of the LunchMoney API, backed by an
// httptest.Server. Create one with New and shut it down with Close.
type Twin struct {
	server *httptest.Server
	me     json.RawMessage
	assets json.RawMessage
	plaid  json.RawMessage
	recur  json.RawMessage
}

// New starts a digital twin serving the fixtures for the given scenario.
// It panics if the scenario's fixtures cannot be loaded or parsed, since
// this is test infrastructure: a bad fixture should fail loudly and
// immediately rather than surface as a confusing test failure downstream.
func New(scenario Scenario) *Twin {
	tw := &Twin{
		me:     loadFixture(scenario, "me.json"),
		assets: loadFixture(scenario, "assets.json"),
		plaid:  loadFixture(scenario, "plaid_accounts.json"),
		recur:  loadFixture(scenario, "recurring_items.json"),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/me", tw.withAuth(tw.handleMe))
	mux.HandleFunc("GET /v1/assets", tw.withAuth(tw.handleAssets))
	mux.HandleFunc("GET /v1/plaid_accounts", tw.withAuth(tw.handlePlaidAccounts))
	mux.HandleFunc("GET /v1/recurring_items", tw.withAuth(tw.handleRecurringItems))

	tw.server = httptest.NewServer(mux)
	return tw
}

// loadFixture reads and validates (as JSON) the fixture file for the given
// scenario, panicking on any failure. Fixtures are embedded at build time so
// loading never depends on the test binary's working directory.
func loadFixture(scenario Scenario, name string) json.RawMessage {
	path := fmt.Sprintf("testdata/%s/%s", scenario, name)
	data, err := fixturesFS.ReadFile(path)
	if err != nil {
		panic(fmt.Sprintf("twin: failed to load fixture %q: %v", path, err))
	}
	if !json.Valid(data) {
		panic(fmt.Sprintf("twin: fixture %q is not valid JSON", path))
	}
	return json.RawMessage(data)
}

// URL returns the base URL of the running twin server, e.g.
// "http://127.0.0.1:54321".
func (tw *Twin) URL() string {
	return tw.server.URL
}

// Token returns the bearer token this twin accepts. It is always ValidToken
// today, but exposed as a method so callers don't need to depend on the
// package constant directly.
func (tw *Twin) Token() string {
	return ValidToken
}

// Close shuts down the twin's underlying httptest.Server.
func (tw *Twin) Close() {
	tw.server.Close()
}

// withAuth wraps a handler so it requires a valid bearer token, matching the
// real LunchMoney API's behavior of returning 401 with a JSON error body for
// missing/incorrect credentials.
func (tw *Twin) withAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		want := "Bearer " + ValidToken
		if r.Header.Get("Authorization") != want {
			writeUnauthorized(w)
			return
		}
		next(w, r)
	}
}

func writeUnauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write([]byte(`{"name": "Unauthorized", "message": "Access token does not exist."}`))
}

func writeJSON(w http.ResponseWriter, body json.RawMessage) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(body)
}

func (tw *Twin) handleMe(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, tw.me)
}

func (tw *Twin) handleAssets(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, tw.assets)
}

func (tw *Twin) handlePlaidAccounts(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, tw.plaid)
}

// handleRecurringItems accepts start_date/end_date query params for
// compatibility with the real API's request shape, but currently ignores
// their values and returns the scenario's fixture as-is.
func (tw *Twin) handleRecurringItems(w http.ResponseWriter, r *http.Request) {
	_ = r.URL.Query().Get("start_date")
	_ = r.URL.Query().Get("end_date")
	writeJSON(w, tw.recur)
}
