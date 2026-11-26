package app

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gabrielroriz/payout-orchestrator/api/internal/config"
	"github.com/gabrielroriz/payout-orchestrator/api/internal/httpserver"
)

type Application struct {
	Config config.Config
	Logger *slog.Logger
	Routes http.Handler
}

func NewApplication(cfg config.Config, logger *slog.Logger, routes http.Handler) (*Application, error) {
	return &Application{Config: cfg, Logger: logger, Routes: routes}, nil
}

// Run serves the API until ctx is canceled, then shuts everything down.
// It blocks until the server has stopped and every resource is released.
func (app *Application) Run(ctx context.Context) error {
	app.Logger.Info("app running", "env", app.Config.Env, "version", app.Config.Version)

	if err := httpserver.Run(ctx, app.Config.Port, app.Logger, app.Routes); err != nil {
		return err
	}

	// then close the DB pool, etc., in reverse order of creation

	return nil
}
