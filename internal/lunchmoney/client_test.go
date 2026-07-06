package lunchmoney_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/codegoalie/money-tracker/internal/lunchmoney"
)

// mustReadFixture reads a testdata fixture file or fails the test.
func mustReadFixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatalf("reading fixture %q: %v", name, err)
	}
	return data
}

func TestGetMe_HappyPath(t *testing.T) {
	body := mustReadFixture(t, "me.json")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	}))
	defer srv.Close()

	client := lunchmoney.New(srv.URL, "test-token")
	me, err := client.GetMe(context.Background())
	if err != nil {
		t.Fatalf("GetMe returned error: %v", err)
	}

	want := lunchmoney.Me{
		UserID:     1001,
		UserName:   "Healthy Hannah",
		UserEmail:  "hannah@example.com",
		AccountID:  2001,
		BudgetName: "Hannah's Budget",
	}
	if *me != want {
		t.Fatalf("GetMe = %+v, want %+v", *me, want)
	}
}

func TestGetMe_SendsAuthorizationHeader(t *testing.T) {
	body := mustReadFixture(t, "me.json")
	var gotHeader string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeader = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	}))
	defer srv.Close()

	client := lunchmoney.New(srv.URL, "secret-token-123")
	if _, err := client.GetMe(context.Background()); err != nil {
		t.Fatalf("GetMe returned error: %v", err)
	}

	want := "Bearer secret-token-123"
	if gotHeader != want {
		t.Fatalf("Authorization header = %q, want %q", gotHeader, want)
	}
}

func TestGetMe_MalformedJSON(t *testing.T) {
	body := mustReadFixture(t, "malformed.json")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(body)
	}))
	defer srv.Close()

	client := lunchmoney.New(srv.URL, "test-token")
	_, err := client.GetMe(context.Background())
	if err == nil {
		t.Fatal("expected error for malformed JSON body, got nil")
	}
}
