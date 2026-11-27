package middleware

import (
	"bufio"
	"net"
	"net/http"
)

// WrappedWriter is an http.ResponseWriter that remembers what the handler sent, for middleware to read afterwards.
type WrappedWriter interface {
	http.ResponseWriter

	// Status returns the status code sent to the client, or 200 if the handler wrote a body without one.
	Status() int

	// BytesWritten returns the number of body bytes sent to the client.
	BytesWritten() int

	// HeaderSent reports whether the status has gone out and can no longer change.
	HeaderSent() bool

	// Unwrap returns the original writer, for http.ResponseController.
	Unwrap() http.ResponseWriter
}

// WrapWriter wraps w in a WrappedWriter, or returns w unchanged if it is already one.
func WrapWriter(w http.ResponseWriter) WrappedWriter {
	// Reuse an existing wrapper, so each request has exactly one and every middleware reads the same status.
	if ww, ok := w.(WrappedWriter); ok {
		return ww
	}

	// Start at 200, which net/http sends when a handler writes a body without calling WriteHeader.
	ww := &wrappedWriter{ResponseWriter: w, status: http.StatusOK}

	// Return a type with the same optional methods as w (http.Flusher, http.Hijacker), because embedding only http.ResponseWriter would hide them.
	_, canFlush := w.(http.Flusher)
	_, canHijack := w.(http.Hijacker)
	switch {
	case canFlush && canHijack:
		return &wrappedFlushHijacker{ww}
	case canFlush:
		return &wrappedFlusher{ww}
	case canHijack:
		return &wrappedHijacker{ww}
	default:
		return ww
	}
}

// wrappedWriter is the base WrappedWriter, without optional features.
type wrappedWriter struct {
	http.ResponseWriter
	status       int
	bytesWritten int
	headerSent   bool
}

// WriteHeader remembers the first final status code and passes every call on.
func (t *wrappedWriter) WriteHeader(code int) {
	// Informational codes (1xx, except 101) are interim responses sent before the final one, so they are not remembered.
	isInterim := code >= 100 && code < 200 && code != http.StatusSwitchingProtocols

	// Only the first final status reaches the client, so only that one is kept.
	if !t.headerSent && !isInterim {
		t.status = code
		t.headerSent = true
	}
	t.ResponseWriter.WriteHeader(code)
}

// Write counts the body bytes and passes them on.
func (t *wrappedWriter) Write(b []byte) (int, error) {
	// Writing sends the header too, so the status can no longer change.
	t.headerSent = true

	n, err := t.ResponseWriter.Write(b)
	t.bytesWritten += n
	return n, err
}

// Status returns the status code sent to the client.
func (t *wrappedWriter) Status() int { return t.status }

// BytesWritten returns the number of body bytes sent to the client.
func (t *wrappedWriter) BytesWritten() int { return t.bytesWritten }

// HeaderSent reports whether the status and headers have gone out.
func (t *wrappedWriter) HeaderSent() bool { return t.headerSent }

// Unwrap returns the original writer.
func (t *wrappedWriter) Unwrap() http.ResponseWriter { return t.ResponseWriter }

// flush passes a flush on to the original writer.
func (t *wrappedWriter) flush() {
	// Flushing sends the header, so the status can no longer change.
	t.headerSent = true

	t.ResponseWriter.(http.Flusher).Flush()
}

// hijack passes a hijack on to the original writer.
func (t *wrappedWriter) hijack() (net.Conn, *bufio.ReadWriter, error) {
	return t.ResponseWriter.(http.Hijacker).Hijack()
}

// wrappedFlusher is a WrappedWriter for writers that can flush.
type wrappedFlusher struct{ *wrappedWriter }

// Flush sends any buffered data to the client.
func (t *wrappedFlusher) Flush() { t.flush() }

// wrappedHijacker is a WrappedWriter for writers that can be hijacked.
type wrappedHijacker struct{ *wrappedWriter }

// Hijack lets the caller take over the connection.
func (t *wrappedHijacker) Hijack() (net.Conn, *bufio.ReadWriter, error) { return t.hijack() }

// wrappedFlushHijacker is a WrappedWriter for writers that can flush and be hijacked, like net/http's HTTP/1 writer.
type wrappedFlushHijacker struct{ *wrappedWriter }

// Flush sends any buffered data to the client.
func (t *wrappedFlushHijacker) Flush() { t.flush() }

// Hijack lets the caller take over the connection.
func (t *wrappedFlushHijacker) Hijack() (net.Conn, *bufio.ReadWriter, error) { return t.hijack() }
