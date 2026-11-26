package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gabrielroriz/payout-orchestrator/api/internal/app"
	"github.com/gabrielroriz/payout-orchestrator/api/internal/config"
	"github.com/gabrielroriz/payout-orchestrator/api/internal/httpserver"
	"github.com/gabrielroriz/payout-orchestrator/shared/logger"
)

// version is set at build time: go build -ldflags "-X main.version=<git sha>"
var version = "dev"

func main() {

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	cfg.Version = version

	app, err := app.NewApplication(
		cfg,
		logger.New(logger.Options{
			Application: "api",
			Level:       cfg.LogLevel,
			JSON:        cfg.Env == config.Production,
		}),
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	app.Logger.Info("app running", "env", cfg.Env, "version", cfg.Version)

	// After the first signal, stop listening so a second Ctrl+C kills the process at once.
	context.AfterFunc(ctx, stop)

	// Blocks until a signal arrives and the server has finished shutting down.
	if err := httpserver.Run(ctx, cfg.Port, app.Logger, app.Routes()); err != nil {
		app.Logger.Error("server stopped with error", "error", err)
		os.Exit(1)
	}
	// then close the DB pool, etc., in reverse order of creation
}
