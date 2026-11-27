package middleware

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gabrielroriz/payout-orchestrator/api/internal/httpjson"
)

// limit is the body cap every test in this file applies.
const limit = 10

// readBodyHandler returns a handler that reads the whole request body and stores the outcome in *got and *err, so a test can inspect what it saw.
func readBodyHandler(got *string, err *error) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, e := io.ReadAll(r.Body)
		*got, *err = string(b), e
	})
}

func TestLimitBodyAllowsSmallBody(t *testing.T) {
	// --- Arrange ---
	// Wrap a handler that records the body it reads.
	var got string
	var readErr error
	h := LimitBody(limit)(readBodyHandler(&got, &readErr))

	// Build a request whose body is exactly at the limit.
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("0123456789"))
	rec := httptest.NewRecorder()

	// --- Act ---
	// Send the request through the middleware.
	h.ServeHTTP(rec, req)

	// --- Assert ---
	// The handler must read the full body without error.
	require.NoError(t, readErr, "read body")
	assert.Equal(t, "0123456789", got)
}

func TestLimitBodyRejectsDeclaredOversizeBody(t *testing.T) {
	// --- Arrange ---
	// Wrap a handler that records whether it ran.
	called := false
	h := LimitBody(limit)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))

	// Build a request whose Content-Length is one byte over the limit.
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("0123456789x"))
	rec := httptest.NewRecorder()

	// --- Act ---
	// Send the request through the middleware.
	h.ServeHTTP(rec, req)

	// --- Assert ---
	// The middleware must answer 413 itself, as a JSON error.
	assert.Equal(t, http.StatusRequestEntityTooLarge, rec.Code)
	var body httpjson.ErrorResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&body), "decode body")
	assert.Equal(t, httpjson.ErrorDetail{Code: httpjson.CodeBodyTooLarge, Message: "request body too large"}, body.Error)

	// The wrapped handler must never run.
	assert.False(t, called, "handler ran, want it skipped")
}

func TestLimitBodyCapsUndeclaredOversizeBody(t *testing.T) {
	// --- Arrange ---
	// Wrap a handler that records the outcome of reading the body.
	var got string
	var readErr error
	h := LimitBody(limit)(readBodyHandler(&got, &readErr))

	// Build an oversized request and hide its length, as a chunked upload would, so the middleware cannot reject it up front.
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("0123456789x"))
	req.ContentLength = -1
	rec := httptest.NewRecorder()

	// --- Act ---
	// Send the request through the middleware.
	h.ServeHTTP(rec, req)

	// --- Assert ---
	// Reading past the limit must fail with *http.MaxBytesError.
	var maxErr *http.MaxBytesError
	assert.ErrorAs(t, readErr, &maxErr)
}
