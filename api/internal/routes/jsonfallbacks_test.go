package routes

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gabrielroriz/payout-orchestrator/api/internal/httpjson"
)

// fallbackMux returns jsonFallbacks around a mux with one route, GET /items, that answers 200 with the plain-text body "items".
func fallbackMux() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /items", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("items"))
	})
	return jsonFallbacks(mux)
}

func TestJSONFallbacksUnmatched(t *testing.T) {
	// --- Arrange ---
	// List requests no route matches, with the status and error they must get.
	tests := []struct {
		name, method, path string
		wantStatus         int
		wantError          httpjson.ErrorDetail
	}{
		{"unknown path", http.MethodGet, "/nope", http.StatusNotFound,
			httpjson.ErrorDetail{Code: httpjson.CodeNotFound, Message: "not found"}},
		{"wrong method", http.MethodPost, "/items", http.StatusMethodNotAllowed,
			httpjson.ErrorDetail{Code: httpjson.CodeMethodNotAllowed, Message: "method not allowed"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// --- Arrange ---
			// Use a recorder to capture the response in memory.
			rec := httptest.NewRecorder()

			// --- Act ---
			// Send the request through the fallbacks.
			fallbackMux().ServeHTTP(rec, httptest.NewRequest(tt.method, tt.path, nil))

			// --- Assert ---
			// The status must be the one net/http chose.
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tt.wantStatus)
			}

			// The response must be declared as JSON.
			if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
				t.Errorf("Content-Type = %q, want application/json", ct)
			}

			// The body must be exactly one JSON error, with no plain text left over.
			var body httpjson.ErrorResponse
			dec := json.NewDecoder(rec.Body)
			if err := dec.Decode(&body); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			if body.Error != tt.wantError {
				t.Errorf("error = %+v, want %+v", body.Error, tt.wantError)
			}
			if dec.More() {
				t.Error("body has data after the JSON error")
			}
		})
	}
}

func TestJSONFallbacksKeepsAllowHeader(t *testing.T) {
	// --- Arrange ---
	// Use a recorder to capture the response in memory.
	rec := httptest.NewRecorder()

	// --- Act ---
	// Send a request with a method the route does not accept.
	fallbackMux().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/items", nil))

	// --- Assert ---
	// The Allow header must still list the accepted methods.
	if got := rec.Header().Get("Allow"); got == "" {
		t.Error("Allow header is missing")
	}
}

func TestJSONFallbacksLeavesMatchedRoutesAlone(t *testing.T) {
	// --- Arrange ---
	// Use a recorder to capture the response in memory.
	rec := httptest.NewRecorder()

	// --- Act ---
	// Send a request the route handles.
	fallbackMux().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/items", nil))

	// --- Assert ---
	// The route's own response must pass through unchanged.
	if rec.Code != http.StatusOK || rec.Body.String() != "items" {
		t.Errorf("response = %d %q, want 200 \"items\"", rec.Code, rec.Body.String())
	}
}

func TestJSONErrorWriterWriteWithoutHeader(t *testing.T) {
	// --- Arrange ---
	// Wrap a recorder in the error writer.
	rec := httptest.NewRecorder()
	w := &jsonErrorWriter{ResponseWriter: rec}

	// --- Act ---
	// Write a body without setting a status first.
	n, err := w.Write([]byte("plain text"))

	// --- Assert ---
	// The write must look successful to the caller.
	if err != nil || n != len("plain text") {
		t.Errorf("Write() = %d, %v, want %d, nil", n, err, len("plain text"))
	}

	// The body must be a JSON error with the default status, not the plain text.
	var body httpjson.ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	want := httpjson.ErrorDetail{Code: "ok", Message: "ok"}
	if rec.Code != http.StatusOK || body.Error != want {
		t.Errorf("response = %d %+v, want 200 %+v", rec.Code, body.Error, want)
	}
}

func TestJSONErrorWriterIgnoresSecondWriteHeader(t *testing.T) {
	// --- Arrange ---
	// Wrap a recorder in the error writer.
	rec := httptest.NewRecorder()
	w := &jsonErrorWriter{ResponseWriter: rec}

	// --- Act ---
	// Set the status twice.
	w.WriteHeader(http.StatusNotFound)
	w.WriteHeader(http.StatusTeapot)

	// --- Assert ---
	// Only the first status must be sent.
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}

func TestCodeFor(t *testing.T) {
	// --- Arrange ---
	// List statuses and the code each must map to.
	tests := []struct {
		status int
		want   httpjson.Code
	}{
		{http.StatusNotFound, httpjson.CodeNotFound},
		{http.StatusMethodNotAllowed, httpjson.CodeMethodNotAllowed},
		{http.StatusBadRequest, "bad_request"},
		{http.StatusServiceUnavailable, "service_unavailable"},
	}

	for _, tt := range tests {
		t.Run(http.StatusText(tt.status), func(t *testing.T) {
			// --- Act ---
			// Map the status to a code.
			got := codeFor(tt.status)

			// --- Assert ---
			// The code must match the expected one.
			if got != tt.want {
				t.Errorf("codeFor(%d) = %q, want %q", tt.status, got, tt.want)
			}
		})
	}
}
