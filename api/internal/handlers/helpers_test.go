package handlers

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http/httptest"
	"testing"
)

// discardLogger returns a logger that throws its output away, for handlers whose logs a test ignores.
func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// decodeJSONResponse checks that rec holds a JSON response with wantStatus and returns its body decoded into a T.
// It stops the test on any mismatch.
func decodeJSONResponse[T any](t *testing.T, rec *httptest.ResponseRecorder, wantStatus int) T {
	t.Helper()

	// The handler must answer with the expected status.
	if rec.Code != wantStatus {
		t.Fatalf("status = %d, want %d", rec.Code, wantStatus)
	}

	// The response must be declared as JSON.
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", ct)
	}

	// The body must be valid JSON that fits T.
	var body T
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	return body
}
