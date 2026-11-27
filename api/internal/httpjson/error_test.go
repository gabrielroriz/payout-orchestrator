package httpjson

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestError(t *testing.T) {
	// --- Arrange ---
	// Use a recorder to capture the response in memory.
	rec := httptest.NewRecorder()

	// --- Act ---
	// Send an error response.
	Error(rec, http.StatusBadRequest, "invalid_amount", "amount is required")

	// --- Assert ---
	// The response must carry the given status and the JSON header.
	checkJSONHeader(t, rec, http.StatusBadRequest)

	// The body must be an ErrorResponse holding the code and the message.
	var body ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	want := ErrorDetail{Code: "invalid_amount", Message: "amount is required"}
	if body.Error != want {
		t.Errorf("error = %+v, want %+v", body.Error, want)
	}
}
