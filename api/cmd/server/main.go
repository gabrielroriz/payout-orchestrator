package main

import (
	"fmt"
	"os"

	"github.com/gabrielroriz/payout-orchestrator/internal/api/app"
	"github.com/gabrielroriz/payout-orchestrator/internal/api/config"
	"github.com/gabrielroriz/payout-orchestrator/internal/api/logger"
)

// version is set at build time: go build -ldflags "-X main.version=<git sha>"
var version = "dev"

func main() {
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
}
