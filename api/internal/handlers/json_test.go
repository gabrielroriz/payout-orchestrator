package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWriteJSON(t *testing.T) {
	// --- Arrange ---
	// Use a recorder to capture the response in memory.
	rec := httptest.NewRecorder()

	// --- Act ---
	// Write a small payload with a non-default status.
	err := writeJSON(rec, http.StatusCreated, map[string]string{"id": "42"})

	// --- Assert ---
	// Encoding a plain map must not fail.
	if err != nil {
		t.Fatalf("writeJSON: %v", err)
	}

	// The response must carry the given status, the JSON header and the payload.
	body := decodeJSONResponse[map[string]string](t, rec, http.StatusCreated)
	if body["id"] != "42" {
		t.Errorf(`body["id"] = %q, want "42"`, body["id"])
	}
}

func TestWriteJSONUnencodable(t *testing.T) {
	// --- Arrange ---
	// Use a recorder to capture the response in memory.
	rec := httptest.NewRecorder()

	// --- Act ---
	// Pass a channel, which encoding/json cannot encode.
	err := writeJSON(rec, http.StatusOK, make(chan int))

	// --- Assert ---
	// The encoding error must reach the caller so it can log it.
	if err == nil {
		t.Fatal("writeJSON returned nil, want an encoding error")
	}
}

func TestReadJSON(t *testing.T) {
	// --- Arrange ---
	// Declare the shape the handler expects.
	var dst struct {
		Amount int    `json:"amount"`
		Note   string `json:"note"`
	}

	// Build a request whose body matches that shape.
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"amount": 100, "note": "hi"}`))

	// --- Act ---
	// Decode the body into dst.
	err := readJSON(req, &dst)

	// --- Assert ---
	// A well-formed body must decode without error.
	if err != nil {
		t.Fatalf("readJSON: %v", err)
	}

	// Each field must hold the value sent in the body.
	if dst.Amount != 100 || dst.Note != "hi" {
		t.Errorf("dst = %+v, want {Amount:100 Note:hi}", dst)
	}
}

func TestReadJSONRejectsBadBodies(t *testing.T) {
	// --- Arrange ---
	// List bodies that readJSON must refuse, each with the reason it is bad.
	tests := []struct {
		name string
		body string
	}{
		{"malformed", `{"amount": `},
		{"empty", ``},
		{"wrong type", `{"amount": "100"}`},
		{"unknown field", `{"amount": 100, "extra": true}`},
		{"two values", `{"amount": 100}{"amount": 200}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// --- Arrange ---
			// Declare the shape the handler expects.
			var dst struct {
				Amount int    `json:"amount"`
				Note   string `json:"note"`
			}

			// Build a request carrying the bad body.
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.body))

			// --- Act ---
			// Try to decode the body.
			err := readJSON(req, &dst)

			// --- Assert ---
			// The bad body must be refused.
			if err == nil {
				t.Fatal("readJSON returned nil, want an error")
			}
		})
	}
}
