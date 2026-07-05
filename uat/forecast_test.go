package uat

import (
	"strings"
	"testing"

	"github.com/codegoalie/money-tracker/uat/twin"
)

// TestForecast_HealthyAccount_NotYetImplemented encodes the target behavior
// of `moneytracker forecast` against a healthy account. None of this exists
// yet: the Phase-0 stub exits 2 and prints an "unknown command" message for
// any invocation, so every assertion here is negative today.
func TestForecast_HealthyAccount_NotYetImplemented(t *testing.T) {
	tw := twin.New(twin.ScenarioHealthy)
	defer tw.Close()

	exitCode, stdout, stderr := runSUT(t, tw.URL(), twin.ValidToken, "forecast")

	// REQUIREMENT: forecast against a healthy account exits 0. Flip to
	// assert exitCode == 0 when implemented.
	if exitCode == 0 {
		t.Errorf("expected forecast to NOT exit 0 yet (Phase 0 stub), got exit code 0")
	}

	// REQUIREMENT: stdout starts with a cash position summary containing
	// "Cash position:". Flip to assert strings.Contains(stdout, "Cash
	// position:") (or strings.HasPrefix, per the final format) when
	// implemented.
	if strings.Contains(stdout, "Cash position:") {
		t.Errorf("expected stdout to NOT contain %q yet, got: %s", "Cash position:", stdout)
	}

	// REQUIREMENT: the period table has a "Balance" column header. Flip to
	// assert strings.Contains(stdout, "Balance") when implemented.
	if strings.Contains(stdout, "Balance") {
		t.Errorf("expected stdout to NOT contain %q yet, got: %s", "Balance", stdout)
	}

	// REQUIREMENT: on success the verdict line contains "OK —" (em dash).
	// Flip to assert strings.Contains(combined, "OK —") when implemented.
	combined := stdout + stderr
	if strings.Contains(combined, "OK —") {
		t.Errorf("expected combined output to NOT contain %q yet, got: %s", "OK —", combined)
	}
}

// TestForecast_FloorBreachAccount_NotYetImplemented encodes the target
// behavior of `moneytracker forecast` against an account whose projected
// balance breaches the floor. Not implemented yet, so both assertions are
// negative.
func TestForecast_FloorBreachAccount_NotYetImplemented(t *testing.T) {
	tw := twin.New(twin.ScenarioFloorBreach)
	defer tw.Close()

	exitCode, stdout, stderr := runSUT(t, tw.URL(), twin.ValidToken, "forecast")

	// REQUIREMENT: forecast exits 1 when the floor is breached. Flip to
	// assert exitCode == 1 when implemented.
	if exitCode == 1 {
		t.Errorf("expected forecast to NOT exit 1 yet (Phase 0 stub), got exit code 1")
	}

	// REQUIREMENT: the verdict line contains "NOT OK" when the floor is
	// breached. Flip to assert strings.Contains(combined, "NOT OK") when
	// implemented.
	combined := stdout + stderr
	if strings.Contains(combined, "NOT OK") {
		t.Errorf("expected combined output to NOT contain %q yet, got: %s", "NOT OK", combined)
	}
}

// TestForecast_PeriodTable_NotYetImplemented encodes the target shape of the
// forecast's period table, which will have "Income" and "Expenses" column
// headers regardless of scenario. Not implemented yet.
func TestForecast_PeriodTable_NotYetImplemented(t *testing.T) {
	tw := twin.New(twin.ScenarioHealthy)
	defer tw.Close()

	_, stdout, _ := runSUT(t, tw.URL(), twin.ValidToken, "forecast")

	// REQUIREMENT: the period table has an "Income" column header. Flip to
	// assert strings.Contains(stdout, "Income") when implemented.
	if strings.Contains(stdout, "Income") {
		t.Errorf("expected stdout to NOT contain %q yet, got: %s", "Income", stdout)
	}

	// REQUIREMENT: the period table has an "Expenses" column header. Flip to
	// assert strings.Contains(stdout, "Expenses") when implemented.
	if strings.Contains(stdout, "Expenses") {
		t.Errorf("expected stdout to NOT contain %q yet, got: %s", "Expenses", stdout)
	}
}
