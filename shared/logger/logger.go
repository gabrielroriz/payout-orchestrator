package logger

import (
	"log/slog"
	"os"
)

// Options configures a logger; each service fills it from its own config.
type Options struct {
	Application string // api, rules-engine, …
	Level       string // debug, info, warn or error; info when empty or invalid
	JSON        bool   // JSON for production; text is easier to read locally
}

// New returns a structured logger tagged with the service and application names.
func New(opts Options) *slog.Logger {

	// Parse the level, falling back to info when it is empty or invalid.
	var level slog.Level

	if err := level.UnmarshalText([]byte(opts.Level)); err != nil {
		level = slog.LevelInfo
	}
	handlerOpts := &slog.HandlerOptions{Level: level}

	// Pick the output format: JSON for machines in production, text for people locally.
	var h slog.Handler

	if opts.JSON {
		h = slog.NewJSONHandler(os.Stdout, handlerOpts)
	} else {
		h = slog.NewTextHandler(os.Stdout, handlerOpts) // readable locally
	}

	// Tag every line with where it came from.
	logger := slog.New(h).
		With("service", "payout-orchestrator").
		With("application", opts.Application)

	return logger
}
