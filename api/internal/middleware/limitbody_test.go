package middleware

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// limit is the body cap every test in this file applies.
const limit = 10

// readBodyHandler returns a handler that reads the whole request body and
// stores the outcome in *got and *err, so a test can inspect what it saw.
func readBodyHandler(got *string, err *error) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, e := io.ReadAll(r.Body)
		*got, *err = string(b), e
	})
}

func TestLimitBodyAllowsSmallBody(t *testing.T) {
	// --- Arrange ---
	// Wrap a handler that records the body it reads.
	var got string
	var readErr error
	h := LimitBody(limit)(readBodyHandler(&got, &readErr))

	// Build a request whose body is exactly at the limit.
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("0123456789"))
	rec := httptest.NewRecorder()

	// --- Act ---
	// Send the request through the middleware.
	h.ServeHTTP(rec, req)

	// --- Assert ---
	// The handler must read the full body without error.
	if readErr != nil {
		t.Fatalf("read body: %v", readErr)
	}
	if got != "0123456789" {
		t.Errorf("body = %q, want %q", got, "0123456789")
	}
}

func TestLimitBodyRejectsDeclaredOversizeBody(t *testing.T) {
	// --- Arrange ---
	// Wrap a handler that records whether it ran.
	called := false
	h := LimitBody(limit)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))

	// Build a request whose Content-Length is one byte over the limit.
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("0123456789x"))
	rec := httptest.NewRecorder()

	// --- Act ---
	// Send the request through the middleware.
	h.ServeHTTP(rec, req)

	// --- Assert ---
	// The middleware must answer 413 itself.
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusRequestEntityTooLarge)
	}

	// The wrapped handler must never run.
	if called {
		t.Error("handler ran, want it skipped")
	}
}

func TestLimitBodyCapsUndeclaredOversizeBody(t *testing.T) {
	// --- Arrange ---
	// Wrap a handler that records the outcome of reading the body.
	var got string
	var readErr error
	h := LimitBody(limit)(readBodyHandler(&got, &readErr))

	// Build an oversized request and hide its length, as a chunked upload would,
	// so the middleware cannot reject it up front.
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("0123456789x"))
	req.ContentLength = -1
	rec := httptest.NewRecorder()

	// --- Act ---
	// Send the request through the middleware.
	h.ServeHTTP(rec, req)

	// --- Assert ---
	// Reading past the limit must fail with *http.MaxBytesError.
	var maxErr *http.MaxBytesError
	if !errors.As(readErr, &maxErr) {
		t.Fatalf("read error = %v, want *http.MaxBytesError", readErr)
	}
}
