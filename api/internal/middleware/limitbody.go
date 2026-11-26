// Package middleware holds HTTP middleware shared by every route.
package middleware

import "net/http"

// LimitBody caps every request body at maxBytes. A request whose declared
// Content-Length is already over the limit gets 413 without its body being read.
// Any other body is wrapped so reading past the limit fails with *http.MaxBytesError.
func LimitBody(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.ContentLength > maxBytes {
				http.Error(w, http.StatusText(http.StatusRequestEntityTooLarge), http.StatusRequestEntityTooLarge)
				return
			}

			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}
