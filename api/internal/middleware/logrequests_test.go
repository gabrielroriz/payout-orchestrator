package middleware

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// captureLogger returns a JSON logger that writes into the returned buffer, so a test can decode the line the middleware logged.
func captureLogger() (*slog.Logger, *bytes.Buffer) {
	var buf bytes.Buffer
	return slog.New(slog.NewJSONHandler(&buf, nil)), &buf
}

// decodeLogLine decodes the single JSON log line in buf, stopping the test if it is missing or malformed.
func decodeLogLine(t *testing.T, buf *bytes.Buffer) map[string]any {
	t.Helper()

	var line map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &line), "decode log line %q", buf.String())
	return line
}

func TestLogRequests(t *testing.T) {
	// --- Arrange ---
	// Wrap a handler that answers 201 with a five-byte body.
	logger, buf := captureLogger()
	h := LogRequests(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("hello"))
	}))

	// Build a request with a query string, which must not reach the log.
	req := httptest.NewRequest(http.MethodPost, "/v1/payouts?token=secret", nil)
	rec := httptest.NewRecorder()

	// --- Act ---
	// Send the request through the middleware.
	h.ServeHTTP(rec, req)

	// --- Assert ---
	// The response must reach the client unchanged.
	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Equal(t, "hello", rec.Body.String())

	// The log line must describe the request at info level, without the query string.
	line := decodeLogLine(t, buf)
	assert.Subset(t, line, map[string]any{
		"level":  "INFO",
		"msg":    "request",
		"method": "POST",
		"path":   "/v1/payouts",
		"status": float64(201),
		"bytes":  float64(5),
	})

	// The duration and remote address must be present.
	assert.Contains(t, line, "duration")
	assert.Contains(t, line, "remote_addr")
}

func TestLogRequestsDefaultsToOK(t *testing.T) {
	// --- Arrange ---
	// Wrap a handler that writes a body without calling WriteHeader.
	logger, buf := captureLogger()
	h := LogRequests(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}))

	// --- Act ---
	// Send a request through the middleware.
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	// --- Assert ---
	// The log must report 200, which is what net/http sends in that case.
	assert.Equal(t, float64(http.StatusOK), decodeLogLine(t, buf)["status"])
}

func TestLogRequestsKeepsFirstStatus(t *testing.T) {
	// --- Arrange ---
	// Wrap a handler that calls WriteHeader twice; net/http ignores the second call.
	logger, buf := captureLogger()
	h := LogRequests(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		w.WriteHeader(http.StatusTeapot)
	}))

	// --- Act ---
	// Send a request through the middleware.
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	// --- Assert ---
	// The log must report the status that was actually sent.
	assert.Equal(t, float64(http.StatusAccepted), decodeLogLine(t, buf)["status"])
}

func TestLogRequestsServerErrorLevel(t *testing.T) {
	// --- Arrange ---
	// Wrap a handler that fails with 500.
	logger, buf := captureLogger()
	h := LogRequests(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))

	// --- Act ---
	// Send a request through the middleware.
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	// --- Assert ---
	// A server error must be logged at error level.
	assert.Equal(t, "ERROR", decodeLogLine(t, buf)["level"])
}

func TestLogRequestsIncludesRequestID(t *testing.T) {
	// --- Arrange ---
	// Put RequestID in front, as the router does, with a fixed client ID.
	logger, buf := captureLogger()
	h := RequestID(LogRequests(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(RequestIDHeader, "req-7")

	// --- Act ---
	// Send the request through both middleware.
	h.ServeHTTP(httptest.NewRecorder(), req)

	// --- Assert ---
	// The log line must carry the request ID.
	assert.Equal(t, "req-7", decodeLogLine(t, buf)["request_id"])
}
