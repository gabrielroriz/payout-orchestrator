package routes

import (
	"log/slog"
	"net/http"

	"github.com/gabrielroriz/payout-orchestrator/api/internal/config"
	"github.com/gabrielroriz/payout-orchestrator/api/internal/handlers"
)

// New returns the HTTP handler with every API route registered.
func New(cfg config.Config, logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /v1/healthcheck", handlers.Healthcheck(cfg.Env, cfg.Version, logger))

	return mux
}
