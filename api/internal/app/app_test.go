package app

import (
	"context"
	"io"
	"log/slog"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/gabrielroriz/payout-orchestrator/api/internal/config"
	"github.com/gabrielroriz/payout-orchestrator/api/internal/routes"
)

// TestRunServesTheAPI checks the whole stack over real HTTP: routes, handlers and server.
func TestRunServesTheAPI(t *testing.T) {
	// Build the config and a logger that throws its output away.
	cfg := config.Config{Env: config.Development, Version: "test"}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	// Build the app with the real routes, just as main does.
	app, err := NewApplication(cfg, logger, routes.New(cfg, logger))
	if err != nil {
		t.Fatal(err)
	}

	// Open a free local port; the test can read its address from ln.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	// Create the context that will stop the app, standing in for Ctrl+C.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the app in the background, because Run blocks until it stops.
	done := make(chan error, 1)
	go func() { done <- app.Run(ctx, ln) }()

	// Call the healthcheck over real HTTP.
	resp, err := http.Get("http://" + ln.Addr().String() + "/v1/healthcheck")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	// The route must exist and answer successfully.
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	// Ask the app to stop.
	cancel()

	// Run must return nil within 5 seconds, meaning it stopped cleanly.
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run() = %v, want nil", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not return after the context was canceled")
	}
}
