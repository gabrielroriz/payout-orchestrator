package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setEnv sets each variable for the duration of the test, so tests do not depend on the shell that runs them.
func setEnv(t *testing.T, vars map[string]string) {
	t.Helper()

	// t.Setenv restores the previous value when the test ends.
	for k, v := range vars {
		t.Setenv(k, v)
	}
}

func TestLoadDefaults(t *testing.T) {
	// --- Arrange ---
	// Leave every variable empty, as a fresh local checkout would.
	setEnv(t, map[string]string{"APP_ENV": "", "PORT": "", "LOG_LEVEL": ""})

	// --- Act ---
	// Read the configuration.
	cfg, err := Load()

	// --- Assert ---
	// Empty variables must not be an error.
	require.NoError(t, err)

	// The defaults must let the API run locally with no setup.
	assert.Equal(t, Config{Env: Development, Port: defaultPort}, cfg)
}

func TestLoadReadsEnvironment(t *testing.T) {
	// --- Arrange ---
	// Set every variable to a valid, non-default value.
	setEnv(t, map[string]string{"APP_ENV": "production", "PORT": "9000", "LOG_LEVEL": "debug"})

	// --- Act ---
	// Read the configuration.
	cfg, err := Load()

	// --- Assert ---
	// Valid values must be accepted.
	require.NoError(t, err)

	// Each field must hold the value from its variable.
	assert.Equal(t, Config{Env: Production, Port: 9000, LogLevel: "debug"}, cfg)
}

func TestLoadRejectsInvalidValues(t *testing.T) {
	// --- Arrange ---
	// List invalid variables, each with the text its error must mention.
	tests := []struct {
		name    string
		env     map[string]string
		wantErr string
	}{
		{"unknown environment", map[string]string{"APP_ENV": "staging", "PORT": ""}, "invalid APP_ENV"},
		{"port not a number", map[string]string{"APP_ENV": "", "PORT": "http"}, "invalid PORT"},
		{"port zero", map[string]string{"APP_ENV": "", "PORT": "0"}, "invalid PORT"},
		{"port too high", map[string]string{"APP_ENV": "", "PORT": "65536"}, "invalid PORT"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// --- Arrange ---
			// Set the invalid variables.
			setEnv(t, tt.env)

			// --- Act ---
			// Read the configuration.
			cfg, err := Load()

			// --- Assert ---
			// Load must fail and name the bad variable.
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)

			// It must not return a partly filled configuration.
			assert.Equal(t, Config{}, cfg)
		})
	}
}

func TestLoadAcceptsPortBounds(t *testing.T) {
	// --- Arrange ---
	// List the lowest and highest valid ports.
	tests := []struct {
		port string
		want int
	}{
		{"1", 1},
		{"65535", 65535},
	}

	for _, tt := range tests {
		t.Run(tt.port, func(t *testing.T) {
			// --- Arrange ---
			// Set the port at the edge of the valid range.
			setEnv(t, map[string]string{"APP_ENV": "", "PORT": tt.port})

			// --- Act ---
			// Read the configuration.
			cfg, err := Load()

			// --- Assert ---
			// The edge value must be accepted as is.
			require.NoError(t, err)
			assert.Equal(t, tt.want, cfg.Port)
		})
	}
}
