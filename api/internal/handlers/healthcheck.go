package handlers

import (
	"log/slog"
	"net/http"

	"github.com/gabrielroriz/payout-orchestrator/api/internal/config"
	"github.com/gabrielroriz/payout-orchestrator/api/internal/httpjson"
)

// Healthcheck reports that the API is up, with its environment and version.
func Healthcheck(env config.Environment, version string, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Report the build and environment, so a caller can tell which version is running.
		data := map[string]string{
			"status":  "available",
			"env":     string(env),
			"version": version,
		}

		// A failed write means the client has gone, so log it and move on.
		if err := httpjson.Write(w, http.StatusOK, data); err != nil {
			logger.Error("write healthcheck response", "error", err)
		}
	}
}
