package lunchmoney_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/codegoalie/money-tracker/internal/lunchmoney"
)

// TestGetMe_ContextCancellation verifies that a request aborts promptly when
// its context is cancelled/times out, rather than blocking until the (here,
// deliberately unresponsive) server responds.
func TestGetMe_ContextCancellation(t *testing.T) {
	block := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-block // never respond before the test's own timeout
	}))
	// Deferred LIFO: close(block) must run BEFORE srv.Close(), otherwise
	// srv.Close() deadlocks waiting for the still-blocked handler to return.
	defer srv.Close()
	defer close(block)

	client := lunchmoney.New(srv.URL, "test-token")
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		_, err := client.GetMe(ctx)
		done <- err
	}()

	select {
	case err := <-done:
		if err == nil {
			t.Fatal("expected error due to context deadline, got nil")
		}
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("expected error wrapping context.DeadlineExceeded, got: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("GetMe did not return promptly after the context deadline was exceeded (request is not honoring context cancellation)")
	}
}
