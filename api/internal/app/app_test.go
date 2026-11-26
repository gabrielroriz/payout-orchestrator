package app

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"testing"

	"github.com/gabrielroriz/payout-orchestrator/api/internal/config"
)

func TestRunStopsWhenContextIsCanceled(t *testing.T) {
	app, err := NewApplication(
		config.Config{Port: 0}, // any free port
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		http.NotFoundHandler(),
	)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := app.Run(ctx); err != nil {
		t.Fatalf("Run() = %v, want nil", err)
	}
}
