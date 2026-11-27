package routes

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
	require.Equal(t, wantStatus, rec.Code, "status")

	// The response must be declared as JSON.
	require.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	// The body must decode into an ErrorResponse.
	var body httpjson.ErrorResponse
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&body), "decode body")
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
	assert.Equal(t, httpjson.ErrorDetail{Code: httpjson.CodeNotFound, Message: "not found"}, got)

	// The 404 must still pass through the middleware, so it carries a request ID.
	assert.NotEmpty(t, rec.Header().Get(middleware.RequestIDHeader), "404 has no request ID")
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
	assert.Equal(t, httpjson.ErrorDetail{Code: httpjson.CodeMethodNotAllowed, Message: "method not allowed"}, got)

	// The Allow header must list exactly the methods /items accepts.
	assert.Equal(t, "GET, POST", rec.Header().Get("Allow"))
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
	assert.Equal(t, http.StatusOK, rec.Code)
}
