package twin_test

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/codegoalie/money-tracker/uat/twin"
)

// TestTwin_Assets_ValidToken is a positive test of the twin infrastructure
// itself: it hits the running twin directly over HTTP (not through the SUT
// binary) and confirms the fixture is served correctly.
func TestTwin_Assets_ValidToken(t *testing.T) {
	tw := twin.New(twin.ScenarioHealthy)
	defer tw.Close()

	req, err := http.NewRequest(http.MethodGet, tw.URL()+"/v1/assets", nil)
	if err != nil {
		t.Fatalf("building request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+twin.ValidToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("performing request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("reading body: %v", err)
	}

	var parsed map[string]json.RawMessage
	if err := json.Unmarshal(body, &parsed); err != nil {
		t.Fatalf("body did not parse as JSON: %v", err)
	}
	if _, ok := parsed["assets"]; !ok {
		t.Fatalf("expected body to contain an %q key, got: %s", "assets", body)
	}
}

// TestTwin_Assets_WrongToken confirms the twin enforces bearer-token auth.
func TestTwin_Assets_WrongToken(t *testing.T) {
	tw := twin.New(twin.ScenarioHealthy)
	defer tw.Close()

	req, err := http.NewRequest(http.MethodGet, tw.URL()+"/v1/assets", nil)
	if err != nil {
		t.Fatalf("building request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer wrong-token")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("performing request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", resp.StatusCode)
	}
}
