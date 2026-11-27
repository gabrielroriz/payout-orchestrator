package routes

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/gabrielroriz/payout-orchestrator/api/internal/config"
	"github.com/gabrielroriz/payout-orchestrator/api/internal/handlers"
	"github.com/gabrielroriz/payout-orchestrator/api/internal/middleware"
)

// maxRequestBodyBytes caps the body of every request at 1 MB.
const maxRequestBodyBytes = 1 << 20

// New returns the HTTP handler with every API route registered.
func New(cfg config.Config, logger *slog.Logger) http.Handler {
	r := newRouter(logger)

	r.Get("/v1/healthcheck", handlers.Healthcheck(cfg.Env, cfg.Version, logger))

	return r
}

// newRouter returns a router with the middleware and fallbacks every route shares, but no routes yet.
func newRouter(logger *slog.Logger) *chi.Mux {
	r := chi.NewRouter()

	// Middleware runs in the order it is added, for every request, including 404s and 405s.

	// Assign the request ID first, so every step after it can log it.
	r.Use(middleware.RequestID)

	// Log every response from below, including the 500 that Recover sends.
	r.Use(middleware.LogRequests(logger))

	// Turn a panic anywhere below into a 500.
	r.Use(middleware.Recover(logger))

	// Cap body size closest to the handlers that read bodies.
	r.Use(middleware.LimitBody(maxRequestBodyBytes))

	// Answer requests no route matches in JSON, like every other error.
	r.NotFound(notFound)
	r.MethodNotAllowed(methodNotAllowed(r))

	return r
}
