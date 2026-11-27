// Package middleware holds HTTP middleware shared by every route.
package middleware

import (
	"net/http"

	"github.com/gabrielroriz/payout-orchestrator/api/internal/httpjson"
)

// LimitBody caps every request body at maxBytes.
func LimitBody(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Reject a body that declares it is too large up front, without reading it.
			if r.ContentLength > maxBytes {
				httpjson.Error(w, http.StatusRequestEntityTooLarge, httpjson.CodeBodyTooLarge, "request body too large")
				return
			}

			// Cap any other body, so reading past the limit fails with *http.MaxBytesError.
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}
