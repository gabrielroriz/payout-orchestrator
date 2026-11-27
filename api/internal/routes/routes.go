package routes

import (
	"log/slog"
	"net/http"

	"github.com/gabrielroriz/payout-orchestrator/api/internal/config"
	"github.com/gabrielroriz/payout-orchestrator/api/internal/handlers"
	"github.com/gabrielroriz/payout-orchestrator/api/internal/middleware"
)

// maxRequestBodyBytes caps the body of every request at 1 MB.
const maxRequestBodyBytes = 1 << 20

// New returns the HTTP handler with every API route registered.
func New(cfg config.Config, logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /v1/healthcheck", handlers.Healthcheck(cfg.Env, cfg.Version, logger))

	return withMiddleware(jsonFallbacks(mux), logger)
}

// withMiddleware wraps h in the middleware every request goes through.
func withMiddleware(h http.Handler, logger *slog.Logger) http.Handler {
	// Each wrap goes around the previous one, so a request meets them from the bottom up.

	// Cap body size closest to the handlers that read bodies.
	h = middleware.LimitBody(maxRequestBodyBytes)(h)

	// Turn a panic anywhere below into a 500.
	h = middleware.Recover(logger)(h)

	// Log every response from below, including the 500 that Recover sends.
	h = middleware.LogRequests(logger)(h)

	// Assign the request ID first, so every step below can log it.
	h = middleware.RequestID(h)

	return h
}
