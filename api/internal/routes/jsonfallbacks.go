package routes

import (
	"net/http"
	"strings"

	"github.com/gabrielroriz/payout-orchestrator/api/internal/httpjson"
)

// jsonFallbacks answers requests no route matches (404, 405) with a JSON error instead of plain text.
func jsonFallbacks(mux *http.ServeMux) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// mux reports an empty pattern exactly when no route matches.
		if _, pattern := mux.Handler(r); pattern == "" {
			// Let mux pick the status and headers, such as Allow on a 405, but swap its body for JSON.
			w = &jsonErrorWriter{ResponseWriter: w}
		}
		mux.ServeHTTP(w, r)
	})
}

// jsonErrorWriter turns a plain-text error from net/http into a JSON one.
type jsonErrorWriter struct {
	http.ResponseWriter
	headerSent bool
}

// WriteHeader sends a JSON error for the status, such as code "not_found" with message "not found".
func (j *jsonErrorWriter) WriteHeader(status int) {
	// Only the first status counts, as with any ResponseWriter.
	if j.headerSent {
		return
	}
	j.headerSent = true

	httpjson.Error(j.ResponseWriter, status, codeFor(status), strings.ToLower(http.StatusText(status)))
}

// Write throws away net/http's plain-text body, which the JSON error replaces.
func (j *jsonErrorWriter) Write(b []byte) (int, error) {
	// A body without a status means 200, as with any ResponseWriter.
	if !j.headerSent {
		j.WriteHeader(http.StatusOK)
	}

	// Report the bytes as written, so net/http sees no error.
	return len(b), nil
}

// codeFor returns the error code for a status net/http answers with on its own.
func codeFor(status int) httpjson.Code {
	switch status {
	case http.StatusNotFound:
		return httpjson.CodeNotFound
	case http.StatusMethodNotAllowed:
		return httpjson.CodeMethodNotAllowed
	default:
		// Build any other code from the status text, such as "bad_request".
		return httpjson.Code(strings.ReplaceAll(strings.ToLower(http.StatusText(status)), " ", "_"))
	}
}
