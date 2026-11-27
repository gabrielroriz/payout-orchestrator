package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
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
			if got := statusOf(ww); got != tt.wantStatus {
				t.Errorf("statusOf() = %d, want %d", got, tt.wantStatus)
			}

			// headerSent must match whether that status has gone out.
			if got := headerSent(ww); got != tt.wantSent {
				t.Errorf("headerSent() = %v, want %v", got, tt.wantSent)
			}
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
	if got := ww.BytesWritten(); got != 11 {
		t.Errorf("BytesWritten() = %d, want 11", got)
	}

	// The client must receive the body unchanged.
	if got := inner.Body.String(); got != "hello world" {
		t.Errorf("body = %q, want %q", got, "hello world")
	}
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
	if second != first {
		t.Error("wrap returned a new wrapper, want the existing one")
	}
}

func TestWrapKeepsFlush(t *testing.T) {
	// --- Arrange ---
	// Wrap a recorder, which can flush.
	inner := httptest.NewRecorder()
	ww := wrap(inner, httptest.NewRequest(http.MethodGet, "/", nil))

	// --- Act ---
	// Flush through the wrapper, as a streaming handler would.
	flusher, ok := ww.(http.Flusher)
	if !ok {
		t.Fatal("wrapper does not implement http.Flusher")
	}
	flusher.Flush()

	// --- Assert ---
	// The flush must reach the original writer.
	if !inner.Flushed {
		t.Error("original writer was not flushed")
	}
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
	if got != inner {
		t.Errorf("Unwrap() = %v, want the original writer", got)
	}
}
