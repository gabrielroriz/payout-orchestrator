package httpserver

import (
	"context"
	"io"
	"log/slog"
	"net"
	"net/http"
	"testing"
)

var discard = slog.New(slog.NewTextHandler(io.Discard, nil))

func TestRunStopsWhenContextIsCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := Run(ctx, 0, discard, http.NotFoundHandler()); err != nil {
		t.Fatalf("Run() = %v, want nil", err)
	}
}

func TestRunFailsWhenPortIsInUse(t *testing.T) {
	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	port := ln.Addr().(*net.TCPAddr).Port

	if err := Run(context.Background(), port, discard, http.NotFoundHandler()); err == nil {
		t.Fatal("Run() = nil, want an error")
	}
}
