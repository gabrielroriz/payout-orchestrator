package handlers

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

// discardLogger returns a logger that throws its output away, for handlers whose logs a test ignores.
func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// decodeJSONResponse checks that rec holds a JSON response with wantStatus and returns its body decoded into a T.
func decodeJSONResponse[T any](t *testing.T, rec *httptest.ResponseRecorder, wantStatus int) T {
	t.Helper()

	// The handler must answer with the expected status.
	require.Equal(t, wantStatus, rec.Code, "status")

	// The response must be declared as JSON.
	require.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	// The body must be valid JSON that fits T.
	var body T
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&body), "decode body")
	return body
}
