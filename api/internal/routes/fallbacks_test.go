package routes

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/gabrielroriz/payout-orchestrator/api/internal/httpjson"
	"github.com/gabrielroriz/payout-orchestrator/api/internal/middleware"
)

// itemsRouter returns the shared router with two test routes on /items, GET and POST, each answering 200.
func itemsRouter() *chi.Mux {
	r := newRouter(slog.New(slog.NewTextHandler(io.Discard, nil)))
	ok := func(w http.ResponseWriter, r *http.Request) {}
	r.Get("/items", ok)
	r.Post("/items", ok)
	return r
}

// decodeError checks that rec holds a JSON error with wantStatus and returns its detail.
func decodeError(t *testing.T, rec *httptest.ResponseRecorder, wantStatus int) httpjson.ErrorDetail {
	t.Helper()

	// The response must carry the expected status.
	if rec.Code != wantStatus {
		t.Fatalf("status = %d, want %d", rec.Code, wantStatus)
	}

	// The response must be declared as JSON.
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", ct)
	}

	// The body must decode into an ErrorResponse.
	var body httpjson.ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	return body.Error
}

func TestNotFound(t *testing.T) {
	// --- Arrange ---
	// Use a recorder to capture the response in memory.
	rec := httptest.NewRecorder()

	// --- Act ---
	// Ask for a path no route matches.
	itemsRouter().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/nope", nil))

	// --- Assert ---
	// The client must get a JSON 404 with the not_found code.
	got := decodeError(t, rec, http.StatusNotFound)
	want := httpjson.ErrorDetail{Code: httpjson.CodeNotFound, Message: "not found"}
	if got != want {
		t.Errorf("error = %+v, want %+v", got, want)
	}

	// The 404 must still pass through the middleware, so it carries a request ID.
	if rec.Header().Get(middleware.RequestIDHeader) == "" {
		t.Error("404 has no request ID")
	}
}

func TestMethodNotAllowed(t *testing.T) {
	// --- Arrange ---
	// Use a recorder to capture the response in memory.
	rec := httptest.NewRecorder()

	// --- Act ---
	// Use a method /items does not accept.
	itemsRouter().ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/items", nil))

	// --- Assert ---
	// The client must get a JSON 405 with the method_not_allowed code.
	got := decodeError(t, rec, http.StatusMethodNotAllowed)
	want := httpjson.ErrorDetail{Code: httpjson.CodeMethodNotAllowed, Message: "method not allowed"}
	if got != want {
		t.Errorf("error = %+v, want %+v", got, want)
	}

	// The Allow header must list exactly the methods /items accepts.
	if allow := rec.Header().Get("Allow"); allow != "GET, POST" {
		t.Errorf("Allow = %q, want %q", allow, "GET, POST")
	}
}

func TestMatchedRouteSkipsFallbacks(t *testing.T) {
	// --- Arrange ---
	// Use a recorder to capture the response in memory.
	rec := httptest.NewRecorder()

	// --- Act ---
	// Use a method /items accepts.
	itemsRouter().ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/items", nil))

	// --- Assert ---
	// The route must answer, not a fallback.
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}
