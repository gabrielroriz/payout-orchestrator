package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gabrielroriz/payout-orchestrator/api/internal/httpjson"
)

// panicAfter returns a handler that runs write and then panics with "boom".
func panicAfter(write func(w http.ResponseWriter)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		write(w)
		panic("boom")
	})
}

// recoveredValue runs fn and returns the value of any panic that escapes it, or nil if none does.
func recoveredValue(fn func()) (v any) {
	defer func() { v = recover() }()
	fn()
	return nil
}

func TestRecoverPassesThrough(t *testing.T) {
	// --- Arrange ---
	// Wrap a handler that answers normally.
	logger, buf := captureLogger()
	h := Recover(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	rec := httptest.NewRecorder()

	// --- Act ---
	// Send a request through the middleware.
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	// --- Assert ---
	// The handler's response must reach the client unchanged.
	assert.Equal(t, http.StatusNoContent, rec.Code)

	// Nothing went wrong, so nothing must be logged.
	assert.Empty(t, buf.String())
}

func TestRecoverBeforeResponse(t *testing.T) {
	// --- Arrange ---
	// Wrap a handler that sets headers for its response and then panics before sending it.
	logger, buf := captureLogger()
	h := Recover(logger)(panicAfter(func(w http.ResponseWriter) {
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Length", "999")
	}))

	// Build a request that already carries a request ID header in its response.
	req := httptest.NewRequest(http.MethodPost, "/v1/payouts?token=secret", nil)
	rec := httptest.NewRecorder()
	rec.Header().Set(RequestIDHeader, "req-1")

	// --- Act ---
	// Send the request through the middleware.
	h.ServeHTTP(rec, req)

	// --- Assert ---
	// The client must get a 500 with a JSON error body.
	require.Equal(t, http.StatusInternalServerError, rec.Code)
	var body httpjson.ErrorResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&body), "decode body")
	assert.Equal(t, httpjson.ErrorDetail{Code: httpjson.CodeInternal, Message: "internal server error"}, body.Error)

	// The headers meant for the unfinished response must be gone.
	assert.Empty(t, rec.Header().Get("Content-Length"))
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	// The request ID must survive, so the client can still quote it.
	assert.Equal(t, "req-1", rec.Header().Get(RequestIDHeader))

	// The panic must be logged at error level with its value, a stack trace and the path.
	line := decodeLogLine(t, buf)
	assert.Subset(t, line, map[string]any{"level": "ERROR", "msg": "panic", "panic": "boom", "path": "/v1/payouts"})
	assert.NotEmpty(t, line["stack"], "log has no stack trace")
}

func TestRecoverLogsRequestID(t *testing.T) {
	// --- Arrange ---
	// Put RequestID in front, as the router does, with a fixed client ID.
	logger, buf := captureLogger()
	h := RequestID(Recover(logger)(panicAfter(func(http.ResponseWriter) {})))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(RequestIDHeader, "req-42")

	// --- Act ---
	// Send the request through both middleware.
	h.ServeHTTP(httptest.NewRecorder(), req)

	// --- Assert ---
	// The panic log must carry the request ID, so it can be matched to the client's report.
	assert.Equal(t, "req-42", decodeLogLine(t, buf)["request_id"])
}

func TestRecoverAfterResponseStarted(t *testing.T) {
	// --- Arrange ---
	// Wrap a handler that sends part of its body and then panics.
	logger, buf := captureLogger()
	h := Recover(logger)(panicAfter(func(w http.ResponseWriter) {
		w.Write([]byte("partial"))
	}))

	// --- Act ---
	// Send a request through the middleware and catch whatever panic escapes.
	v := recoveredValue(func() {
		h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
	})

	// --- Assert ---
	// The status was already sent, so the response must be aborted instead.
	assert.Equal(t, http.ErrAbortHandler, v)

	// The original panic must still be logged.
	assert.Equal(t, "boom", decodeLogLine(t, buf)["panic"])
}

func TestRecoverLetsAbortThrough(t *testing.T) {
	// --- Arrange ---
	// Wrap a handler that aborts on purpose.
	logger, buf := captureLogger()
	h := Recover(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic(http.ErrAbortHandler)
	}))

	// --- Act ---
	// Send a request through the middleware and catch whatever panic escapes.
	v := recoveredValue(func() {
		h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
	})

	// --- Assert ---
	// The deliberate abort must reach net/http unchanged.
	assert.Equal(t, http.ErrAbortHandler, v)

	// A deliberate abort is not a bug, so nothing must be logged.
	assert.Empty(t, buf.String())
}

func TestRecoverReusesWrappedWriter(t *testing.T) {
	// --- Arrange ---
	// Wrap the writer up front, as LogRequests does.
	logger, _ := captureLogger()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ww := wrap(httptest.NewRecorder(), req)

	// Wrap a handler that records the writer it receives.
	var got http.ResponseWriter
	h := Recover(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = w
	}))

	// --- Act ---
	// Send a request through the middleware with the already-wrapped writer.
	h.ServeHTTP(ww, req)

	// --- Assert ---
	// The handler must receive that same writer, not a second wrapper.
	assert.Same(t, ww, got)
}
