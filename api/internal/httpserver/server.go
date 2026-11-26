package httpserver

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

// shutdownTimeout is how long in-flight requests get to finish once shutdown starts.
const shutdownTimeout = 10 * time.Second

// Run serves h on the given port until ctx is canceled, then shuts down gracefully.
// It blocks until the server has fully stopped.
func Run(ctx context.Context, port int, logger *slog.Logger, h http.Handler) error {
	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           h,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       time.Minute,
		ErrorLog:          slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}

	// ListenAndServe blocks, so run it in the background and report how it ended.
	serverErr := make(chan error, 1)
	go func() {
		logger.Info("starting server", "addr", srv.Addr)
		serverErr <- srv.ListenAndServe()
	}()

	select {
	case err := <-serverErr:
		// The server never started, for example because the port is in use.
		return fmt.Errorf("start server: %w", err)
	case <-ctx.Done():
	}

	logger.Info("shutting down server")

	// Stop accepting connections and wait for in-flight requests to finish.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown: %w", err)
	}

	logger.Info("server stopped")
	return nil
}
