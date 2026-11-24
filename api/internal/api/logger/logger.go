package logger

import (
	"log/slog"
	"os"

	"github.com/gabrielroriz/payout-orchestrator/internal/api/config"
)

func NewLogger(cfg config.Config) *slog.Logger {

	var level slog.Level

	if err := level.UnmarshalText([]byte(cfg.LogLevel)); err != nil {
		// level info by default (error, warn, info)
		level = slog.LevelInfo
	}
	opts := &slog.HandlerOptions{Level: level}

	var h slog.Handler

	switch cfg.Env {
	case config.Development:
		h = slog.NewTextHandler(os.Stdout, opts) // readable locally
	default:
		// json handler for production (and any other environment)
		h = slog.NewJSONHandler(os.Stdout, opts)
	}

	logger := slog.New(h).
		With("service", "payout-orchestrator").
		With("application", "api")

	return logger
}
