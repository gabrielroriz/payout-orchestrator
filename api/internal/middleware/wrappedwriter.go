package middleware

import (
	"bufio"
	"net"
	"net/http"
)

// WrappedWriter is an http.ResponseWriter that keeps track of what the handler
// sent, so middleware can read it once the handler returns. net/http writers
// send the response straight to the client and cannot report it afterwards.
type WrappedWriter interface {
	http.ResponseWriter

	// Status returns the status code sent to the client: 200 if the handler
	// wrote a body without calling WriteHeader.
	Status() int

	// BytesWritten returns the number of body bytes sent to the client.
	BytesWritten() int

	// Unwrap returns the original writer, for http.ResponseController.
	Unwrap() http.ResponseWriter
}

// WrapWriter wraps w in a WrappedWriter. The result keeps the optional
// features w supports (http.Flusher, http.Hijacker), so code that checks for
// them with a type assertion still finds them.
func WrapWriter(w http.ResponseWriter) WrappedWriter {
	ww := &wrappedWriter{ResponseWriter: w, status: http.StatusOK}

	// Return a type with the same optional methods as w. Embedding only
	// http.ResponseWriter would hide them.
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
// Informational codes (1xx, except 101) are interim responses sent before the
// final one, so they are passed on without being remembered.
func (t *wrappedWriter) WriteHeader(code int) {
	isInterim := code >= 100 && code < 200 && code != http.StatusSwitchingProtocols
	if !t.headerSent && !isInterim {
		t.status = code
		t.headerSent = true
	}
	t.ResponseWriter.WriteHeader(code)
}

// Write counts the body bytes and passes them on. Writing sends the header too,
// so the status can no longer change.
func (t *wrappedWriter) Write(b []byte) (int, error) {
	t.headerSent = true
	n, err := t.ResponseWriter.Write(b)
	t.bytesWritten += n
	return n, err
}

// Status returns the status code sent to the client.
func (t *wrappedWriter) Status() int { return t.status }

// BytesWritten returns the number of body bytes sent to the client.
func (t *wrappedWriter) BytesWritten() int { return t.bytesWritten }

// Unwrap returns the original writer.
func (t *wrappedWriter) Unwrap() http.ResponseWriter { return t.ResponseWriter }

// flush passes a flush on to the original writer. Flushing sends the header,
// so the status can no longer change.
func (t *wrappedWriter) flush() {
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

// wrappedFlushHijacker is a WrappedWriter for writers that can flush and be
// hijacked, as net/http's HTTP/1 writer can.
type wrappedFlushHijacker struct{ *wrappedWriter }

// Flush sends any buffered data to the client.
func (t *wrappedFlushHijacker) Flush() { t.flush() }

// Hijack lets the caller take over the connection.
func (t *wrappedFlushHijacker) Hijack() (net.Conn, *bufio.ReadWriter, error) { return t.hijack() }
