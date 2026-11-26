package app

import (
	"log/slog"

	"github.com/gabrielroriz/payout-orchestrator/api/internal/config"
)

type Application struct {
	Config config.Config
	Logger *slog.Logger
}

func NewApplication(cfg config.Config, logger *slog.Logger) (*Application, error) {
	return &Application{Config: cfg, Logger: logger}, nil
}
