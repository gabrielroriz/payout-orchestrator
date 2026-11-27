package routes

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gabrielroriz/payout-orchestrator/api/internal/config"
	"github.com/gabrielroriz/payout-orchestrator/api/internal/middleware"
)

func TestRoutes(t *testing.T) {
	// --- Arrange ---
	// Build the router with every route registered and a logger that throws its output away.
	routes := New(config.Config{}, slog.New(slog.NewTextHandler(io.Discard, nil)))

	// List each request to try and the status code the router must answer with.
	tests := []struct {
		method, path string
		want         int
	}{
		{http.MethodGet, "/v1/healthcheck", http.StatusOK},
		{http.MethodPost, "/v1/healthcheck", http.StatusMethodNotAllowed},
		{http.MethodGet, "/v1/unknown", http.StatusNotFound},
	}

	// Run each case as its own subtest, named after the request.
	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			// --- Arrange ---
			// Use a recorder: it stands in for the client and captures the response in memory.
			rec := httptest.NewRecorder()

			// --- Act ---
			// Send the request through the router, without a server or network.
			routes.ServeHTTP(rec, httptest.NewRequest(tt.method, tt.path, nil))

			// --- Assert ---
			// The router must answer with the expected status code.
			if rec.Code != tt.want {
				t.Fatalf("status = %d, want %d", rec.Code, tt.want)
			}
		})
	}
}

func TestRoutesLimitBody(t *testing.T) {
	// --- Arrange ---
	// Build the router with every route registered and a logger that throws its output away.
	routes := New(config.Config{}, slog.New(slog.NewTextHandler(io.Discard, nil)))

	// Build a request to a real route with a body one byte over the limit.
	body := strings.NewReader(strings.Repeat("a", maxRequestBodyBytes+1))
	req := httptest.NewRequest(http.MethodGet, "/v1/healthcheck", body)

	// Use a recorder: it stands in for the client and captures the response in memory.
	rec := httptest.NewRecorder()

	// --- Act ---
	// Send the request through the router, without a server or network.
	routes.ServeHTTP(rec, req)

	// --- Assert ---
	// The router must refuse the body before the route handles it.
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusRequestEntityTooLarge)
	}
}

func TestRoutesLogRequests(t *testing.T) {
	// --- Arrange ---
	// Build the router with a JSON logger that writes into a buffer the test can read.
	var buf bytes.Buffer
	routes := New(config.Config{}, slog.New(slog.NewJSONHandler(&buf, nil)))

	// Build a request with an oversized body, so the body limit answers instead of a route.
	body := strings.NewReader(strings.Repeat("a", maxRequestBodyBytes+1))
	req := httptest.NewRequest(http.MethodGet, "/v1/healthcheck", body)

	// --- Act ---
	// Send the request through the router, without a server or network.
	routes.ServeHTTP(httptest.NewRecorder(), req)

	// --- Assert ---
	// Exactly one JSON log line must have been written.
	var line map[string]any
	if err := json.Unmarshal(buf.Bytes(), &line); err != nil {
		t.Fatalf("decode log line %q: %v", buf.String(), err)
	}

	// It must record the 413, which proves logging wraps the body limit.
	if line["msg"] != "request" || line["status"] != float64(http.StatusRequestEntityTooLarge) {
		t.Errorf("log = %v, want a request line with status 413", line)
	}
}

func TestRoutesSetRequestID(t *testing.T) {
	// --- Arrange ---
	// Build the router with every route registered and a logger that throws its output away.
	routes := New(config.Config{}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	rec := httptest.NewRecorder()

	// --- Act ---
	// Send a request without a request ID.
	routes.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/v1/healthcheck", nil))

	// --- Assert ---
	// The response must carry a generated request ID.
	if got := rec.Header().Get(middleware.RequestIDHeader); got == "" {
		t.Errorf("%s header is missing", middleware.RequestIDHeader)
	}
}

func TestRouterRecoversPanic(t *testing.T) {
	// --- Arrange ---
	// Build the shared router, logging into a buffer, and add a route that panics.
	var buf bytes.Buffer
	h := newRouter(slog.New(slog.NewJSONHandler(&buf, nil)))
	h.Get("/", func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	})

	// Build a request with a fixed request ID, to find it in the logs.
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(middleware.RequestIDHeader, "req-9")
	rec := httptest.NewRecorder()

	// --- Act ---
	// Send the request through the chain.
	h.ServeHTTP(rec, req)

	// --- Assert ---
	// The client must get a 500.
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}

	// Two log lines must follow: the panic from Recover, then the request from LogRequests.
	dec := json.NewDecoder(&buf)
	var panicLine, requestLine map[string]any
	if err := dec.Decode(&panicLine); err != nil {
		t.Fatalf("decode panic line: %v", err)
	}
	if err := dec.Decode(&requestLine); err != nil {
		t.Fatalf("decode request line: %v", err)
	}

	// The panic line must carry the request ID, which proves RequestID runs before Recover.
	if panicLine["msg"] != "panic" || panicLine["request_id"] != "req-9" {
		t.Errorf("panic line = %v, want msg panic with request_id req-9", panicLine)
	}

	// The request line must record the 500, which proves LogRequests wraps Recover.
	if requestLine["msg"] != "request" || requestLine["status"] != float64(http.StatusInternalServerError) {
		t.Errorf("request line = %v, want msg request with status 500", requestLine)
	}
}
