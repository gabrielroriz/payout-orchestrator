package logger

import (
	"log/slog"
	"os"
)

// Options configures a logger. Each service fills it from its own config.
type Options struct {
	Application string // api, rules-engine, …
	Level       string // debug, info, warn or error; info when empty or invalid
	JSON        bool   // JSON for production; text is easier to read locally
}

// New returns a structured logger tagged with the service and application names.
func New(opts Options) *slog.Logger {

	var level slog.Level

	if err := level.UnmarshalText([]byte(opts.Level)); err != nil {
		// level info by default (error, warn, info)
		level = slog.LevelInfo
	}
	handlerOpts := &slog.HandlerOptions{Level: level}

	var h slog.Handler

	if opts.JSON {
		h = slog.NewJSONHandler(os.Stdout, handlerOpts)
	} else {
		h = slog.NewTextHandler(os.Stdout, handlerOpts) // readable locally
	}

	logger := slog.New(h).
		With("service", "payout-orchestrator").
		With("application", opts.Application)

	return logger
}
