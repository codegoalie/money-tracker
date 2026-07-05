package uat

import (
	"strings"
	"testing"

	"github.com/codegoalie/money-tracker/uat/twin"
)

// TestCheck_ValidToken_NotYetImplemented encodes the target behavior of a
// hypothetical future `moneytracker check` subcommand that validates the
// configured token/connection. It doesn't exist yet: the Phase-0 stub exits
// 2 and prints an "unknown command" message for any invocation.
func TestCheck_ValidToken_NotYetImplemented(t *testing.T) {
	tw := twin.New(twin.ScenarioHealthy)
	defer tw.Close()

	exitCode, stdout, _ := runSUT(t, tw.URL(), twin.ValidToken, "check")

	// REQUIREMENT: check exits 0 when the token is valid. Flip to assert
	// exitCode == 0 when implemented.
	if exitCode == 0 {
		t.Errorf("expected check to NOT exit 0 yet (Phase 0 stub), got exit code 0")
	}

	// REQUIREMENT: check prints confirmation of a valid token/account, e.g.
	// contains "token OK". Flip to assert strings.Contains(stdout, "token
	// OK") when implemented.
	if strings.Contains(stdout, "token OK") {
		t.Errorf("expected stdout to NOT contain %q yet, got: %s", "token OK", stdout)
	}
}

// TestCheck_InvalidToken_NotYetImplemented encodes the target behavior of
// `moneytracker check` when the configured token is rejected by the API.
// The target exit code (2, operational error) is the SAME as the Phase-0
// stub's default exit code for everything, so exit code is NOT a usable
// signal here — asserting exitCode == 2 today would pass for the wrong
// reason (the stub, not real token validation) and would not flip
// meaningfully when implemented. Instead we assert on stderr content that
// only real token validation would produce.
func TestCheck_InvalidToken_NotYetImplemented(t *testing.T) {
	tw := twin.New(twin.ScenarioHealthy)
	defer tw.Close()

	_, _, stderr := runSUT(t, tw.URL(), "wrong-token", "check")

	// REQUIREMENT: check exits 2 with a clear error message mentioning
	// invalid token when the API rejects the configured token. Exit code is
	// not assertable today (see comment above). Flip to assert
	// strings.Contains(strings.ToLower(stderr), "invalid token") when
	// implemented (in addition to asserting exitCode == 2).
	if strings.Contains(strings.ToLower(stderr), "invalid token") {
		t.Errorf("expected stderr to NOT contain %q yet, got: %s", "invalid token", stderr)
	}
}
