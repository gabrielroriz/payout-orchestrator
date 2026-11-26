package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gabrielroriz/payout-orchestrator/api/internal/config"
)

func TestHealthcheck(t *testing.T) {
	// --- Arrange ---
	// Build the handler with known values and a logger that throws its output away.
	h := Healthcheck(config.Development, "test", discardLogger())

	// Use a recorder: it stands in for the client and captures the response in memory.
	rec := httptest.NewRecorder()

	// --- Act ---
	// Call the handler directly, without a server or network.
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	// --- Assert ---
	// The handler must answer 200 with a JSON body.
	body := decodeJSONResponse[map[string]string](t, rec, http.StatusOK)

	// Each field must hold the value the handler was built with.
	want := map[string]string{"status": "available", "env": "development", "version": "test"}
	for k, v := range want {
		if body[k] != v {
			t.Errorf("body[%q] = %q, want %q", k, body[k], v)
		}
	}
}
