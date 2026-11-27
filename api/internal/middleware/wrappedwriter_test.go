package middleware

import (
	"bufio"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

// plainWriter is a ResponseWriter with only the three interface methods, so it supports neither flushing nor hijacking.
type plainWriter struct{ http.ResponseWriter }

// errHijacked is what the fake writers return from Hijack, so a test can tell the call reached them.
var errHijacked = errors.New("hijack reached the original writer")

// hijackOnlyWriter is a ResponseWriter that can be hijacked but not flushed.
type hijackOnlyWriter struct{ http.ResponseWriter }

// Hijack returns errHijacked instead of a real connection.
func (hijackOnlyWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return nil, nil, errHijacked
}

// flushHijackWriter is a ResponseWriter that can flush and be hijacked, like net/http's HTTP/1 writer.
// Flush comes from the embedded recorder.
type flushHijackWriter struct{ *httptest.ResponseRecorder }

// Hijack returns errHijacked instead of a real connection.
func (flushHijackWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return nil, nil, errHijacked
}

func TestWrappedWriterStatus(t *testing.T) {
	// --- Arrange ---
	// List the calls a handler makes and the status the writer must report.
	tests := []struct {
		name  string
		write func(w http.ResponseWriter)
		want  int
	}{
		{"nothing written", func(w http.ResponseWriter) {}, http.StatusOK},
		{"body without WriteHeader", func(w http.ResponseWriter) { w.Write([]byte("hi")) }, http.StatusOK},
		{"explicit status", func(w http.ResponseWriter) { w.WriteHeader(http.StatusCreated) }, http.StatusCreated},
		{"second WriteHeader ignored", func(w http.ResponseWriter) {
			w.WriteHeader(http.StatusAccepted)
			w.WriteHeader(http.StatusTeapot)
		}, http.StatusAccepted},
		{"WriteHeader after body ignored", func(w http.ResponseWriter) {
			w.Write([]byte("hi"))
			w.WriteHeader(http.StatusInternalServerError)
		}, http.StatusOK},
		{"interim 1xx skipped", func(w http.ResponseWriter) {
			w.WriteHeader(http.StatusEarlyHints)
			w.WriteHeader(http.StatusNoContent)
		}, http.StatusNoContent},
		{"101 is final", func(w http.ResponseWriter) { w.WriteHeader(http.StatusSwitchingProtocols) }, http.StatusSwitchingProtocols},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// --- Arrange ---
			// Wrap a recorder that stands in for the client.
			ww := WrapWriter(httptest.NewRecorder())

			// --- Act ---
			// Make the handler's calls on the wrapped writer.
			tt.write(ww)

			// --- Assert ---
			// The writer must report the status the client received.
			if got := ww.Status(); got != tt.want {
				t.Errorf("Status() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestWrappedWriterBytesWritten(t *testing.T) {
	// --- Arrange ---
	// Wrap a recorder that stands in for the client.
	inner := httptest.NewRecorder()
	ww := WrapWriter(inner)

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

func TestWrappedWriterUnwrap(t *testing.T) {
	// --- Arrange ---
	// Wrap a recorder that stands in for the client.
	inner := httptest.NewRecorder()
	ww := WrapWriter(inner)

	// --- Act ---
	// Ask for the original writer.
	got := ww.Unwrap()

	// --- Assert ---
	// It must be the recorder, so http.ResponseController can reach it.
	if got != inner {
		t.Errorf("Unwrap() = %v, want the original writer", got)
	}
}

func TestWrappedWriterKeepsOptionalFeatures(t *testing.T) {
	// --- Arrange ---
	// List original writers with different optional features, and which features the wrapped writer must expose for each.
	tests := []struct {
		name                string
		inner               http.ResponseWriter
		wantFlush, wantHijk bool
	}{
		{"plain", plainWriter{httptest.NewRecorder()}, false, false},
		{"flush only", httptest.NewRecorder(), true, false},
		{"hijack only", hijackOnlyWriter{httptest.NewRecorder()}, false, true},
		{"flush and hijack", flushHijackWriter{httptest.NewRecorder()}, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// --- Act ---
			// Wrap the original writer.
			ww := WrapWriter(tt.inner)

			// --- Assert ---
			// The wrapped writer must flush exactly when the original can.
			if _, ok := ww.(http.Flusher); ok != tt.wantFlush {
				t.Errorf("implements http.Flusher = %v, want %v", ok, tt.wantFlush)
			}

			// The wrapped writer must hijack exactly when the original can.
			if _, ok := ww.(http.Hijacker); ok != tt.wantHijk {
				t.Errorf("implements http.Hijacker = %v, want %v", ok, tt.wantHijk)
			}
		})
	}
}

func TestWrappedWriterFlush(t *testing.T) {
	// --- Arrange ---
	// Wrap writers that can flush, with and without hijacking.
	tests := []struct {
		name  string
		inner *httptest.ResponseRecorder
		wrap  func(*httptest.ResponseRecorder) http.ResponseWriter
	}{
		{"flush only", httptest.NewRecorder(), func(r *httptest.ResponseRecorder) http.ResponseWriter { return r }},
		{"flush and hijack", httptest.NewRecorder(), func(r *httptest.ResponseRecorder) http.ResponseWriter { return flushHijackWriter{r} }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// --- Arrange ---
			// Wrap the original writer.
			ww := WrapWriter(tt.wrap(tt.inner))

			// --- Act ---
			// Flush, then try to change the status, which is too late.
			ww.(http.Flusher).Flush()
			ww.WriteHeader(http.StatusInternalServerError)

			// --- Assert ---
			// The flush must reach the original writer.
			if !tt.inner.Flushed {
				t.Error("original writer was not flushed")
			}

			// Flushing sent a 200 header, so that is the status to report.
			if got := ww.Status(); got != http.StatusOK {
				t.Errorf("Status() = %d, want 200", got)
			}
		})
	}
}

func TestWrappedWriterHijack(t *testing.T) {
	// --- Arrange ---
	// Wrap writers that can be hijacked, with and without flushing.
	tests := []struct {
		name  string
		inner http.ResponseWriter
	}{
		{"hijack only", hijackOnlyWriter{httptest.NewRecorder()}},
		{"flush and hijack", flushHijackWriter{httptest.NewRecorder()}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// --- Arrange ---
			// Wrap the original writer.
			ww := WrapWriter(tt.inner)

			// --- Act ---
			// Hijack through the wrapped writer.
			_, _, err := ww.(http.Hijacker).Hijack()

			// --- Assert ---
			// The call must reach the original writer.
			if !errors.Is(err, errHijacked) {
				t.Errorf("Hijack() error = %v, want errHijacked", err)
			}
		})
	}
}

func TestWrappedWriterHeaderSent(t *testing.T) {
	// --- Arrange ---
	// List the calls a handler makes and whether the header has gone out after them.
	tests := []struct {
		name  string
		write func(w http.ResponseWriter)
		want  bool
	}{
		{"nothing written", func(w http.ResponseWriter) {}, false},
		{"headers only edited", func(w http.ResponseWriter) { w.Header().Set("X-A", "1") }, false},
		{"interim 1xx only", func(w http.ResponseWriter) { w.WriteHeader(http.StatusEarlyHints) }, false},
		{"WriteHeader", func(w http.ResponseWriter) { w.WriteHeader(http.StatusOK) }, true},
		{"Write", func(w http.ResponseWriter) { w.Write([]byte("hi")) }, true},
		{"Flush", func(w http.ResponseWriter) { w.(http.Flusher).Flush() }, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// --- Arrange ---
			// Wrap a recorder, which can flush.
			ww := WrapWriter(httptest.NewRecorder())

			// --- Act ---
			// Make the handler's calls on the wrapped writer.
			tt.write(ww)

			// --- Assert ---
			// HeaderSent must match whether the status has gone out.
			if got := ww.HeaderSent(); got != tt.want {
				t.Errorf("HeaderSent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWrapWriterReusesExistingWrapper(t *testing.T) {
	// --- Arrange ---
	// Wrap a recorder once, as the first middleware would.
	first := WrapWriter(httptest.NewRecorder())

	// --- Act ---
	// Wrap the already-wrapped writer again, as a later middleware would.
	second := WrapWriter(first)

	// --- Assert ---
	// The same wrapper must come back, so both middleware share one status.
	if second != first {
		t.Error("WrapWriter returned a new wrapper, want the existing one")
	}
}
