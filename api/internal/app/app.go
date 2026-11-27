package app

import (
	"context"
	"log/slog"
	"net"
	"net/http"

	"github.com/gabrielroriz/payout-orchestrator/api/internal/config"
	"github.com/gabrielroriz/payout-orchestrator/api/internal/httpserver"
)

// Application holds everything the API needs to run.
type Application struct {
	Config config.Config
	Logger *slog.Logger
	Routes http.Handler
}

// NewApplication builds an Application from its dependencies.
func NewApplication(cfg config.Config, logger *slog.Logger, routes http.Handler) (*Application, error) {
	return &Application{Config: cfg, Logger: logger, Routes: routes}, nil
}

// Run serves the API on ln until ctx is canceled, then releases every resource before returning.
func (app *Application) Run(ctx context.Context, ln net.Listener) error {
	app.Logger.Info("app running", "env", app.Config.Env, "version", app.Config.Version)

	if err := httpserver.Run(ctx, ln, app.Logger, app.Routes); err != nil {
		return err
	}

	// Then close the DB pool and other resources, in reverse order of creation.

	return nil
}
