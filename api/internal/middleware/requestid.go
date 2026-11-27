package middleware

import (
	"context"
	"crypto/rand"
	"net/http"
)

// RequestIDHeader is the header that carries the request ID, both ways.
const RequestIDHeader = "X-Request-ID"

// maxRequestIDLen caps the length of a request ID accepted from a client.
const maxRequestIDLen = 128

// requestIDKey is the context key for the request ID; its own type keeps other packages from clashing with it.
type requestIDKey struct{}

// RequestID gives every request an ID that ties its log lines together.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Reuse the client's ID when it is safe to log, so both sides share it.
		id := r.Header.Get(RequestIDHeader)
		if !validRequestID(id) {
			// Otherwise generate a random one.
			id = rand.Text()
		}

		// Send the ID back, so the client can quote it when reporting a problem.
		w.Header().Set(RequestIDHeader, id)

		// Store it in the context, where RequestIDFrom can read it.
		ctx := context.WithValue(r.Context(), requestIDKey{}, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequestIDFrom returns the request ID stored in ctx by RequestID, or "" if there is none.
func RequestIDFrom(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey{}).(string)
	return id
}

// validRequestID reports whether a client-supplied ID is safe to reuse.
func validRequestID(id string) bool {
	// Refuse an empty or overly long ID.
	if id == "" || len(id) > maxRequestIDLen {
		return false
	}

	// Allow only letters, digits, '-', '_' and '.', since anything else could break log lines or forge them.
	for _, c := range id {
		switch {
		case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c >= '0' && c <= '9':
		case c == '-', c == '_', c == '.':
		default:
			return false
		}
	}
	return true
}
