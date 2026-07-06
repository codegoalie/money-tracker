package lunchmoney_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/codegoalie/money-tracker/internal/lunchmoney"
)

func TestGetMe_Unauthorized(t *testing.T) {
	body := mustReadFixture(t, "unauthorized.json")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write(body)
	}))
	defer srv.Close()

	client := lunchmoney.New(srv.URL, "bad-token")
	_, err := client.GetMe(context.Background())
	if err == nil {
		t.Fatal("expected error for 401 response, got nil")
	}
	if !errors.Is(err, lunchmoney.ErrUnauthorized) {
		t.Fatalf("errors.Is(err, ErrUnauthorized) = false, err = %v", err)
	}
	if !strings.Contains(err.Error(), "Access token does not exist.") {
		t.Fatalf("error text = %q, want it to contain %q", err.Error(), "Access token does not exist.")
	}

	var apiErr *lunchmoney.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("errors.As(err, &APIError) = false, err = %v", err)
	}
	if apiErr.StatusCode != http.StatusUnauthorized {
		t.Fatalf("apiErr.StatusCode = %d, want %d", apiErr.StatusCode, http.StatusUnauthorized)
	}
	if apiErr.Message != "Access token does not exist." {
		t.Fatalf("apiErr.Message = %q, want %q", apiErr.Message, "Access token does not exist.")
	}
}

func TestGetMe_NonOKStatus(t *testing.T) {
	body := mustReadFixture(t, "server_error.txt")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write(body)
	}))
	defer srv.Close()

	client := lunchmoney.New(srv.URL, "test-token")
	_, err := client.GetMe(context.Background())
	if err == nil {
		t.Fatal("expected error for 500 response, got nil")
	}
	if errors.Is(err, lunchmoney.ErrUnauthorized) {
		t.Fatalf("errors.Is(err, ErrUnauthorized) = true for a 500 response, want false")
	}

	var apiErr *lunchmoney.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("errors.As(err, &APIError) = false, err = %v", err)
	}
	if apiErr.StatusCode != http.StatusInternalServerError {
		t.Fatalf("apiErr.StatusCode = %d, want %d", apiErr.StatusCode, http.StatusInternalServerError)
	}
	if !strings.Contains(err.Error(), "500") {
		t.Fatalf("error text = %q, want it to contain the status code 500", err.Error())
	}
}
