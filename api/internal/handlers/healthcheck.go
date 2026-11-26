package handlers

import (
	"log/slog"
	"net/http"

	"github.com/gabrielroriz/payout-orchestrator/api/internal/config"
)

// Healthcheck reports that the API is up, with its environment and version.
func Healthcheck(env config.Environment, version string, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := map[string]string{
			"status":  "available",
			"env":     string(env),
			"version": version,
		}

		if err := writeJSON(w, http.StatusOK, data); err != nil {
			logger.Error("write healthcheck response", "error", err)
		}
	}
}
