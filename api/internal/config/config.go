package config

import (
	"fmt"
	"os"
	"strconv"
)

// Environment is where the API runs, which decides defaults such as the log format.
type Environment string

// The environments the API accepts in APP_ENV.
const (
	Production  Environment = "production"
	Development Environment = "development"
)

// defaultPort is the port the API listens on when PORT is not set.
const defaultPort = 8089

// Config is the API configuration, read once at startup.
type Config struct {
	Env      Environment
	Port     int
	LogLevel string
	Version  string
}

// Load reads the configuration from environment variables and fails on invalid values.
func Load() (Config, error) {

	// Default to development, so running locally needs no setup.
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = string(Development)
	}

	// Start from the defaults and the values that need no checking.
	cfg := Config{
		Env:      Environment(env),
		Port:     defaultPort,
		LogLevel: os.Getenv("LOG_LEVEL"),
	}

	// Refuse an unknown environment, rather than run with surprising defaults.
	switch cfg.Env {
	case Production, Development:
	default:
		return Config{}, fmt.Errorf("invalid APP_ENV %q: must be %q or %q", cfg.Env, Production, Development)
	}

	// Use PORT when set, as long as it is a valid port number.
	if p := os.Getenv("PORT"); p != "" {
		port, err := strconv.Atoi(p)
		if err != nil || port < 1 || port > 65535 {
			return Config{}, fmt.Errorf("invalid PORT %q: must be a number between 1 and 65535", p)
		}
		cfg.Port = port
	}

	return cfg, nil
}
