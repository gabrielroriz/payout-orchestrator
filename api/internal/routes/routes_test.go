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
