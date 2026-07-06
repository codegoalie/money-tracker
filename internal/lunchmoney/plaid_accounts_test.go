package lunchmoney_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/codegoalie/money-tracker/internal/lunchmoney"
)

func TestGetPlaidAccounts_HappyPath(t *testing.T) {
	body := mustReadFixture(t, "plaid_accounts.json")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	}))
	defer srv.Close()

	client := lunchmoney.New(srv.URL, "test-token")
	accounts, err := client.GetPlaidAccounts(context.Background())
	if err != nil {
		t.Fatalf("GetPlaidAccounts returned error: %v", err)
	}

	want := []lunchmoney.PlaidAccount{
		{ID: 3, Name: "Rewards Credit Card", Type: "credit", Balance: 1230.00, ToBase: -1230.00, Currency: "usd"},
	}
	if len(accounts) != len(want) {
		t.Fatalf("GetPlaidAccounts returned %d accounts, want %d: %+v", len(accounts), len(want), accounts)
	}
	for i := range want {
		if accounts[i] != want[i] {
			t.Fatalf("account[%d] = %+v, want %+v", i, accounts[i], want[i])
		}
	}
}

func TestGetPlaidAccounts_SendsAuthorizationHeader(t *testing.T) {
	body := mustReadFixture(t, "plaid_accounts.json")
	var gotHeader string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeader = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	}))
	defer srv.Close()

	client := lunchmoney.New(srv.URL, "secret-token-123")
	if _, err := client.GetPlaidAccounts(context.Background()); err != nil {
		t.Fatalf("GetPlaidAccounts returned error: %v", err)
	}

	want := "Bearer secret-token-123"
	if gotHeader != want {
		t.Fatalf("Authorization header = %q, want %q", gotHeader, want)
	}
}

func TestGetPlaidAccounts_MalformedBalance(t *testing.T) {
	body := mustReadFixture(t, "plaid_accounts_malformed_balance.json")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	}))
	defer srv.Close()

	client := lunchmoney.New(srv.URL, "test-token")
	_, err := client.GetPlaidAccounts(context.Background())
	if err == nil {
		t.Fatal("expected error for malformed balance string, got nil")
	}
}
