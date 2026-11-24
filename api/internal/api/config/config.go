package config

import (
	"fmt"
	"os"
)

type Environment string

const (
	Production  Environment = "production"
	Development Environment = "development"
)

type Config struct {
	Env      Environment
	LogLevel string
	Version  string
}

// Load reads the configuration from environment variables and fails on invalid values.
func Load() (Config, error) {

	env := os.Getenv("APP_ENV")
	if env == "" {
		env = string(Development)
	}

	cfg := Config{
		Env:      Environment(env),
		LogLevel: os.Getenv("LOG_LEVEL"),
	}

	switch cfg.Env {
	case Production, Development:
	default:
		return Config{}, fmt.Errorf("invalid APP_ENV %q: must be %q or %q", cfg.Env, Production, Development)
	}

	return cfg, nil
}
