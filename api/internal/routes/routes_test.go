package routes

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
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
