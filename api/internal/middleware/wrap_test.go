package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// writeCase is a sequence of calls a handler makes, with the status and header state they must leave behind.
type writeCase struct {
	name       string
	write      func(w http.ResponseWriter)
	wantStatus int
	wantSent   bool
}

// writeCases lists the handler behaviors statusOf and headerSent must handle.
var writeCases = []writeCase{
	{"nothing written", func(w http.ResponseWriter) {}, http.StatusOK, false},
	{"headers only edited", func(w http.ResponseWriter) { w.Header().Set("X-A", "1") }, http.StatusOK, false},
	{"body without WriteHeader", func(w http.ResponseWriter) { w.Write([]byte("hi")) }, http.StatusOK, true},
	{"explicit status", func(w http.ResponseWriter) { w.WriteHeader(http.StatusCreated) }, http.StatusCreated, true},
	{"second WriteHeader ignored", func(w http.ResponseWriter) {
		w.WriteHeader(http.StatusAccepted)
		w.WriteHeader(http.StatusTeapot)
	}, http.StatusAccepted, true},
	{"WriteHeader after body ignored", func(w http.ResponseWriter) {
		w.Write([]byte("hi"))
		w.WriteHeader(http.StatusInternalServerError)
	}, http.StatusOK, true},
	{"interim 1xx only", func(w http.ResponseWriter) { w.WriteHeader(http.StatusEarlyHints) }, http.StatusOK, false},
	{"interim 1xx skipped", func(w http.ResponseWriter) {
		w.WriteHeader(http.StatusEarlyHints)
		w.WriteHeader(http.StatusNoContent)
	}, http.StatusNoContent, true},
	{"flush sends the header", func(w http.ResponseWriter) { w.(http.Flusher).Flush() }, http.StatusOK, true},
}

func TestWrapTracksResponse(t *testing.T) {
	for _, tt := range writeCases {
		t.Run(tt.name, func(t *testing.T) {
			// --- Arrange ---
			// Wrap a recorder, which can flush, as the middleware does.
			ww := wrap(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

			// --- Act ---
			// Make the handler's calls on the wrapped writer.
			tt.write(ww)

			// --- Assert ---
			// statusOf must report the status the client received.
			assert.Equal(t, tt.wantStatus, statusOf(ww), "statusOf")

			// headerSent must match whether that status has gone out.
			assert.Equal(t, tt.wantSent, headerSent(ww), "headerSent")
		})
	}
}

func TestWrapCountsBytes(t *testing.T) {
	// --- Arrange ---
	// Wrap a recorder that stands in for the client.
	inner := httptest.NewRecorder()
	ww := wrap(inner, httptest.NewRequest(http.MethodGet, "/", nil))

	// --- Act ---
	// Write the body in two parts.
	ww.Write([]byte("hello "))
	ww.Write([]byte("world"))

	// --- Assert ---
	// The count must cover both parts.
	assert.Equal(t, 11, ww.BytesWritten())

	// The client must receive the body unchanged.
	assert.Equal(t, "hello world", inner.Body.String())
}

func TestWrapReusesExistingWrapper(t *testing.T) {
	// --- Arrange ---
	// Wrap a recorder once, as the first middleware would.
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	first := wrap(httptest.NewRecorder(), req)

	// --- Act ---
	// Wrap the already-wrapped writer again, as a later middleware would.
	second := wrap(first, req)

	// --- Assert ---
	// The same wrapper must come back, so both middleware share one status.
	assert.Same(t, first, second)
}

func TestWrapKeepsFlush(t *testing.T) {
	// --- Arrange ---
	// Wrap a recorder, which can flush.
	inner := httptest.NewRecorder()
	ww := wrap(inner, httptest.NewRequest(http.MethodGet, "/", nil))

	// --- Act ---
	// Flush through the wrapper, as a streaming handler would.
	flusher, ok := ww.(http.Flusher)
	require.True(t, ok, "wrapper does not implement http.Flusher")
	flusher.Flush()

	// --- Assert ---
	// The flush must reach the original writer.
	assert.True(t, inner.Flushed, "original writer was not flushed")
}

func TestWrapUnwrap(t *testing.T) {
	// --- Arrange ---
	// Wrap a recorder that stands in for the client.
	inner := httptest.NewRecorder()
	ww := wrap(inner, httptest.NewRequest(http.MethodGet, "/", nil))

	// --- Act ---
	// Ask for the original writer.
	got := ww.Unwrap()

	// --- Assert ---
	// It must be the recorder, so http.ResponseController can reach it.
	assert.Same(t, inner, got)
}
