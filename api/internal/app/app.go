package app

import (
	"log/slog"
	"net/http"

	"github.com/gabrielroriz/payout-orchestrator/api/internal/config"
)

type Application struct {
	Config config.Config
	Logger *slog.Logger
	Routes http.Handler
}

func NewApplication(cfg config.Config, logger *slog.Logger, routes http.Handler) (*Application, error) {
	return &Application{Config: cfg, Logger: logger, Routes: routes}, nil
}
