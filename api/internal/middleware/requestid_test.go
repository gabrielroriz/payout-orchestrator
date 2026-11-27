package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// serveWithRequestID sends a request through RequestID and returns the ID the next handler saw, plus the response.
func serveWithRequestID(incoming string) (seen string, rec *httptest.ResponseRecorder) {
	// Wrap a handler that records the ID it finds in the context.
	h := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = RequestIDFrom(r.Context())
	}))

	// Send the client's ID only when the test gives one.
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if incoming != "" {
		req.Header.Set(RequestIDHeader, incoming)
	}
	// Run the request and capture the response.
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return seen, rec
}

func TestRequestIDGeneratesID(t *testing.T) {
	// --- Act ---
	// Send two requests without an X-Request-ID header.
	first, rec := serveWithRequestID("")
	second, _ := serveWithRequestID("")

	// --- Assert ---
	// The handler must find a generated ID in the context.
	if first == "" {
		t.Fatal("no request ID in context")
	}

	// The same ID must be sent back to the client.
	if got := rec.Header().Get(RequestIDHeader); got != first {
		t.Errorf("response header = %q, want %q", got, first)
	}

	// Each request must get its own ID.
	if first == second {
		t.Errorf("two requests got the same ID %q", first)
	}
}

func TestRequestIDReusesValidClientID(t *testing.T) {
	// --- Act ---
	// Send a request carrying a well-formed ID from the client.
	seen, rec := serveWithRequestID("tenant-abc_123.4")

	// --- Assert ---
	// The handler must see the client's ID.
	if seen != "tenant-abc_123.4" {
		t.Errorf("context ID = %q, want the client's ID", seen)
	}

	// The client must get its own ID back.
	if got := rec.Header().Get(RequestIDHeader); got != "tenant-abc_123.4" {
		t.Errorf("response header = %q, want the client's ID", got)
	}
}

func TestRequestIDReplacesUnsafeClientID(t *testing.T) {
	// --- Arrange ---
	// List client IDs that must not be reused, each named for the reason.
	tests := []struct {
		name string
		id   string
	}{
		{"too long", strings.Repeat("a", maxRequestIDLen+1)},
		{"space", "abc def"},
		{"quote", `abc"def`},
		{"non-ASCII", "abcé"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// --- Act ---
			// Send a request carrying the unsafe ID.
			seen, rec := serveWithRequestID(tt.id)

			// --- Assert ---
			// The handler must get a fresh, non-empty ID instead.
			if seen == "" || seen == tt.id {
				t.Errorf("context ID = %q, want a generated ID", seen)
			}

			// The client must be told the ID that was actually used.
			if got := rec.Header().Get(RequestIDHeader); got != seen {
				t.Errorf("response header = %q, want %q", got, seen)
			}
		})
	}
}

func TestRequestIDAcceptsMaxLength(t *testing.T) {
	// --- Arrange ---
	// Build a client ID exactly at the length limit.
	id := strings.Repeat("a", maxRequestIDLen)

	// --- Act ---
	// Send a request carrying it.
	seen, _ := serveWithRequestID(id)

	// --- Assert ---
	// An ID at the limit must still be reused.
	if seen != id {
		t.Errorf("context ID was replaced, want the client's ID kept")
	}
}

func TestRequestIDFromEmptyContext(t *testing.T) {
	// --- Act ---
	// Read the ID from a context that RequestID never touched.
	got := RequestIDFrom(context.Background())

	// --- Assert ---
	// There is no ID, so the result must be empty.
	if got != "" {
		t.Errorf("RequestIDFrom() = %q, want empty", got)
	}
}
