package config

import (
	"fmt"
	"os"
	"strconv"
)

type Environment string

const (
	Production  Environment = "production"
	Development Environment = "development"
)

const defaultPort = 8089

type Config struct {
	Env      Environment
	Port     int
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
		Port:     defaultPort,
		LogLevel: os.Getenv("LOG_LEVEL"),
	}

	switch cfg.Env {
	case Production, Development:
	default:
		return Config{}, fmt.Errorf("invalid APP_ENV %q: must be %q or %q", cfg.Env, Production, Development)
	}

	if p := os.Getenv("PORT"); p != "" {
		port, err := strconv.Atoi(p)
		if err != nil || port < 1 || port > 65535 {
			return Config{}, fmt.Errorf("invalid PORT %q: must be a number between 1 and 65535", p)
		}
		cfg.Port = port
	}

	return cfg, nil
}
