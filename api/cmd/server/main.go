package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/gabrielroriz/payout-orchestrator/api/internal/app"
	"github.com/gabrielroriz/payout-orchestrator/api/internal/config"
	"github.com/gabrielroriz/payout-orchestrator/api/internal/routes"
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

	log := logger.New(logger.Options{
		Application: "api",
		Level:       cfg.LogLevel,
		JSON:        cfg.Env == config.Production,
	})

	app, err := app.NewApplication(
		cfg,
		log,
		routes.New(cfg, log),
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// Open the port here so a failure, such as the port being in use, stops startup at once.
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))
	if err != nil {
		app.Logger.Error("cannot listen on port", "port", cfg.Port, "error", err)
		os.Exit(1)
	}

	// After the first signal, stop listening so a second Ctrl+C kills the process at once.
	// A second Ctrl+C usually means "force quit", including on a production box.
	context.AfterFunc(ctx, stop)

	// Blocks until a signal arrives and the application has shut down.
	if err := app.Run(ctx, ln); err != nil {
		app.Logger.Error("app stopped with error", "error", err)
		os.Exit(1)
	}
}
