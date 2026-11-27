package middleware

import (
	"net/http"

	chimw "github.com/go-chi/chi/v5/middleware"
)

// wrap returns w as chi's WrapResponseWriter, which remembers what the handler sent.
func wrap(w http.ResponseWriter, r *http.Request) chimw.WrapResponseWriter {
	// Reuse an existing wrapper, so each request has exactly one and every middleware reads the same status.
	if ww, ok := w.(chimw.WrapResponseWriter); ok {
		return ww
	}

	// The protocol version decides which optional features (Flush, Hijack) the wrapper keeps.
	return chimw.NewWrapResponseWriter(w, r.ProtoMajor)
}

// statusOf returns the status the client received from ww.
func statusOf(ww chimw.WrapResponseWriter) int {
	// chi reports 0 when the handler wrote nothing, but net/http then sends 200.
	if status := ww.Status(); status != 0 {
		return status
	}
	return http.StatusOK
}

// headerSent reports whether ww's status has gone out and can no longer change.
func headerSent(ww chimw.WrapResponseWriter) bool {
	// chi sets the status on the first WriteHeader, Write or Flush, which are exactly the calls that send it.
	return ww.Status() != 0
}
