package logger

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewWritesJSONWithServiceAndApplication(t *testing.T) {
	// --- Arrange ---
	// Capture the output in memory instead of stdout.
	var buf bytes.Buffer

	// Build a JSON logger, as production uses.
	log := New(Options{Application: "api", JSON: true, Output: &buf})

	// --- Act ---
	// Write one line with an extra attribute.
	log.Info("hello", "payee", "p_1")

	// --- Assert ---
	// The line must be valid JSON.
	var line map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &line), "decode log line")

	// It must carry the message, the level, the attribute and the tags that say where it came from.
	assert.Equal(t, "hello", line["msg"])
	assert.Equal(t, "INFO", line["level"])
	assert.Equal(t, "p_1", line["payee"])
	assert.Equal(t, "payout-orchestrator", line["service"])
	assert.Equal(t, "api", line["application"])
}

func TestNewWritesTextByDefault(t *testing.T) {
	// --- Arrange ---
	// Capture the output in memory instead of stdout.
	var buf bytes.Buffer

	// Build a logger without asking for JSON, as local development does.
	log := New(Options{Application: "api", Output: &buf})

	// --- Act ---
	// Write one line.
	log.Info("hello")

	// --- Assert ---
	// The line must be in key=value text form, with the tags that say where it came from.
	out := buf.String()
	assert.Contains(t, out, "msg=hello")
	assert.Contains(t, out, "service=payout-orchestrator")
	assert.Contains(t, out, "application=api")

	// It must not be JSON.
	assert.False(t, strings.HasPrefix(out, "{"), "output should be text, got %q", out)
}

func TestNewLevel(t *testing.T) {
	// --- Arrange ---
	// List configured levels with the lowest level each one must let through.
	tests := []struct {
		name  string
		level string
		want  slog.Level
	}{
		{"debug", "debug", slog.LevelDebug},
		{"warn", "warn", slog.LevelWarn},
		{"error", "error", slog.LevelError},
		{"empty falls back to info", "", slog.LevelInfo},
		{"invalid falls back to info", "loud", slog.LevelInfo},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// --- Arrange ---
			// Build a logger with the configured level, writing to memory.
			var buf bytes.Buffer
			log := New(Options{Level: tt.level, Output: &buf})

			// --- Act ---
			// Write one line at the expected level and one just below it.
			log.Log(t.Context(), tt.want, "at level")
			log.Log(t.Context(), tt.want-1, "below level")

			// --- Assert ---
			// Only the line at the expected level must be written.
			out := buf.String()
			assert.Contains(t, out, "at level")
			assert.NotContains(t, out, "below level")
		})
	}
}
