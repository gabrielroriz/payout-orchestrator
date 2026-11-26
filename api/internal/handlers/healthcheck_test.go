package handlers

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gabrielroriz/payout-orchestrator/api/internal/config"
)

func TestHealthcheck(t *testing.T) {
	// --- Arrange ---
	// Build the handler with known values and a logger that throws its output away.
	h := Healthcheck(config.Development, "test", slog.New(slog.NewTextHandler(io.Discard, nil)))

	// Use a recorder: it stands in for the client and captures the response in memory.
	rec := httptest.NewRecorder()

	// --- Act ---
	// Call the handler directly, without a server or network.
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	// --- Assert ---
	// The handler must answer successfully.
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	// The response must be declared as JSON.
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}

	// The body must be valid JSON.
	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}

	// Each field must hold the value the handler was built with.
	want := map[string]string{"status": "available", "env": "development", "version": "test"}
	for k, v := range want {
		if body[k] != v {
			t.Errorf("body[%q] = %q, want %q", k, body[k], v)
		}
	}
}
