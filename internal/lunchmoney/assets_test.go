package lunchmoney_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/codegoalie/money-tracker/internal/lunchmoney"
)

func TestGetAssets_HappyPath(t *testing.T) {
	body := mustReadFixture(t, "assets.json")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	}))
	defer srv.Close()

	client := lunchmoney.New(srv.URL, "test-token")
	assets, err := client.GetAssets(context.Background())
	if err != nil {
		t.Fatalf("GetAssets returned error: %v", err)
	}

	want := []lunchmoney.Asset{
		{ID: 1, Name: "Everyday Checking", TypeName: "checking", Balance: 7000.00, ToBase: 7000.00, Currency: "usd"},
		{ID: 2, Name: "Rainy Day Savings", TypeName: "savings", Balance: 2650.00, ToBase: 2650.00, Currency: "usd"},
	}
	if len(assets) != len(want) {
		t.Fatalf("GetAssets returned %d assets, want %d: %+v", len(assets), len(want), assets)
	}
	for i := range want {
		if assets[i] != want[i] {
			t.Fatalf("asset[%d] = %+v, want %+v", i, assets[i], want[i])
		}
	}
}

func TestGetAssets_SendsAuthorizationHeader(t *testing.T) {
	body := mustReadFixture(t, "assets.json")
	var gotHeader string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeader = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	}))
	defer srv.Close()

	client := lunchmoney.New(srv.URL, "secret-token-123")
	if _, err := client.GetAssets(context.Background()); err != nil {
		t.Fatalf("GetAssets returned error: %v", err)
	}

	want := "Bearer secret-token-123"
	if gotHeader != want {
		t.Fatalf("Authorization header = %q, want %q", gotHeader, want)
	}
}

func TestGetAssets_MalformedBalance(t *testing.T) {
	body := mustReadFixture(t, "assets_malformed_balance.json")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	}))
	defer srv.Close()

	client := lunchmoney.New(srv.URL, "test-token")
	_, err := client.GetAssets(context.Background())
	if err == nil {
		t.Fatal("expected error for malformed balance string, got nil")
	}
}
