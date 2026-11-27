package httpjson

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&body), "decode body")
	assert.Equal(t, ErrorDetail{Code: "invalid_amount", Message: "amount is required"}, body.Error)
}
