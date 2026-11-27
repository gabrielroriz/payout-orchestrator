package httpserver

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"
)

// shutdownTimeout is how long in-flight requests get to finish once shutdown starts; tests shorten it.
var shutdownTimeout = 15 * time.Second

// Run serves h on ln until ctx is canceled, then shuts down gracefully and closes ln.
func Run(ctx context.Context, ln net.Listener, logger *slog.Logger, h http.Handler) error {
	srv := &http.Server{
		Handler:           h,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       time.Minute,
		ErrorLog:          slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}

	// Serve blocks, so run it in the background and report how it ended.
	serverErr := make(chan error, 1)
	go func() {
		logger.Info("starting server", "addr", ln.Addr().String())
		serverErr <- srv.Serve(ln)
	}()

	// Wait until either the server has stopped or the context is canceled.
	select {
	case err := <-serverErr:
		// The server stopped on its own, for example because the listener failed.
		return fmt.Errorf("serve: %w", err)
	case <-ctx.Done():
		// The context was canceled, so the server should shut down gracefully.
	}

	// Gracefully shut down the server, giving in-flight requests a chance to finish.
	logger.Info("shutting down server", "reason", ctx.Err())

	// Stop accepting connections and wait for in-flight requests to finish.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown: %w", err)
	}

	logger.Info("server stopped")
	return nil
}
