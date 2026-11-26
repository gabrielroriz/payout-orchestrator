package app

import (
	"net/http"

	"github.com/gabrielroriz/payout-orchestrator/api/internal/handlers"
)

// Routes returns the HTTP handler with every API route registered.
func (app *Application) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /v1/healthcheck", handlers.Healthcheck(app.Config.Env, app.Config.Version, app.Logger))

	return mux
}
