package middleware

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

// captureLogger returns a JSON logger that writes into the returned buffer,
// so a test can decode the line the middleware logged.
func captureLogger() (*slog.Logger, *bytes.Buffer) {
	var buf bytes.Buffer
	return slog.New(slog.NewJSONHandler(&buf, nil)), &buf
}

// decodeLogLine decodes the single JSON log line in buf, stopping the test if
// it is missing or malformed.
func decodeLogLine(t *testing.T, buf *bytes.Buffer) map[string]any {
	t.Helper()

	var line map[string]any
	if err := json.Unmarshal(buf.Bytes(), &line); err != nil {
		t.Fatalf("decode log line %q: %v", buf.String(), err)
	}
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
	if rec.Code != http.StatusCreated || rec.Body.String() != "hello" {
		t.Fatalf("response = %d %q, want 201 \"hello\"", rec.Code, rec.Body.String())
	}

	// The log line must describe the request at info level, without the query string.
	line := decodeLogLine(t, buf)
	want := map[string]any{
		"level":  "INFO",
		"msg":    "request",
		"method": "POST",
		"path":   "/v1/payouts",
		"status": float64(201),
		"bytes":  float64(5),
	}
	for k, v := range want {
		if line[k] != v {
			t.Errorf("log[%q] = %v, want %v", k, line[k], v)
		}
	}

	// The duration and remote address must be present.
	for _, k := range []string{"duration", "remote_addr"} {
		if _, ok := line[k]; !ok {
			t.Errorf("log is missing %q", k)
		}
	}
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
	if got := decodeLogLine(t, buf)["status"]; got != float64(http.StatusOK) {
		t.Errorf("status = %v, want 200", got)
	}
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
	if got := decodeLogLine(t, buf)["status"]; got != float64(http.StatusAccepted) {
		t.Errorf("status = %v, want 202", got)
	}
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
	if got := decodeLogLine(t, buf)["level"]; got != "ERROR" {
		t.Errorf("level = %v, want ERROR", got)
	}
}
