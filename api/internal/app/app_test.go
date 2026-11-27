package app

import (
	"context"
	"io"
	"log/slog"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gabrielroriz/payout-orchestrator/api/internal/config"
	"github.com/gabrielroriz/payout-orchestrator/api/internal/routes"
)

// TestRunServesTheAPI checks the whole stack over real HTTP: routes, handlers and server.
func TestRunServesTheAPI(t *testing.T) {
	// --- Arrange ---
	// Build the config and a logger that throws its output away.
	cfg := config.Config{Env: config.Development, Version: "test"}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	// Build the app with the real routes, just as main does.
	app, err := NewApplication(cfg, logger, routes.New(cfg, logger))
	require.NoError(t, err)

	// Open a free local port; the test can read its address from ln.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	// Create the context that will stop the app, standing in for Ctrl+C.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the app in the background, because Run blocks until it stops.
	done := make(chan error, 1)
	go func() { done <- app.Run(ctx, ln) }()

	// --- Act ---
	// Call the healthcheck over real HTTP.
	resp, err := http.Get("http://" + ln.Addr().String() + "/v1/healthcheck")
	require.NoError(t, err)
	resp.Body.Close()

	// --- Assert ---
	// The route must exist and answer successfully.
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// --- Act ---
	// Ask the app to stop.
	cancel()

	// --- Assert ---
	// Run must return nil within 5 seconds, meaning it stopped cleanly.
	select {
	case err := <-done:
		require.NoError(t, err, "Run")
	case <-time.After(5 * time.Second):
		require.Fail(t, "Run did not return after the context was canceled")
	}
}
