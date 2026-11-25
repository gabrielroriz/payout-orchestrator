package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gabrielroriz/payout-orchestrator/internal/api/app"
	"github.com/gabrielroriz/payout-orchestrator/internal/api/config"
	"github.com/gabrielroriz/payout-orchestrator/internal/api/logger"
)

// version is set at build time: go build -ldflags "-X main.version=<git sha>"
var version = "dev"

func main() {

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	cfg.Version = version

	app, err := app.NewApplication(
		logger.NewLogger(cfg),
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	app.Logger.Info("app running...")

	<-ctx.Done() // wait for a signal
	app.Logger.Info("shutting down")

	// gracefully shutting down...
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// then close the DB pool, etc., in reverse order of creation
}
