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

	// Logging wraps everything else, so responses written by other middleware are logged too.
	var h http.Handler = mux
	h = middleware.LimitBody(maxRequestBodyBytes)(h)
	h = middleware.LogRequests(logger)(h)
	return h
}
