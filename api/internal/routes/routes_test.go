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
	routes := New(config.Config{}, slog.New(slog.NewTextHandler(io.Discard, nil)))

	tests := []struct {
		method, path string
		want         int
	}{
		{http.MethodGet, "/v1/healthcheck", http.StatusOK},
		{http.MethodPost, "/v1/healthcheck", http.StatusMethodNotAllowed},
		{http.MethodGet, "/v1/unknown", http.StatusNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			rec := httptest.NewRecorder()
			routes.ServeHTTP(rec, httptest.NewRequest(tt.method, tt.path, nil))

			if rec.Code != tt.want {
				t.Fatalf("status = %d, want %d", rec.Code, tt.want)
			}
		})
	}
}
