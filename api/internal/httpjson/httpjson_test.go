package httpjson

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// checkJSONHeader stops the test unless rec carries wantStatus and a JSON Content-Type.
func checkJSONHeader(t *testing.T, rec *httptest.ResponseRecorder, wantStatus int) {
	t.Helper()

	// The response must carry the expected status.
	if rec.Code != wantStatus {
		t.Fatalf("status = %d, want %d", rec.Code, wantStatus)
	}

	// The response must be declared as JSON.
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", ct)
	}
}

func TestWrite(t *testing.T) {
	// --- Arrange ---
	// Use a recorder to capture the response in memory.
	rec := httptest.NewRecorder()

	// --- Act ---
	// Write a small payload with a non-default status.
	err := Write(rec, http.StatusCreated, map[string]string{"id": "42"})

	// --- Assert ---
	// Encoding a plain map must not fail.
	if err != nil {
		t.Fatalf("Write: %v", err)
	}

	// The response must carry the given status and the JSON header.
	checkJSONHeader(t, rec, http.StatusCreated)

	// The body must hold the payload.
	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body["id"] != "42" {
		t.Errorf(`body["id"] = %q, want "42"`, body["id"])
	}
}

func TestWriteUnencodable(t *testing.T) {
	// --- Arrange ---
	// Use a recorder to capture the response in memory.
	rec := httptest.NewRecorder()

	// --- Act ---
	// Pass a channel, which encoding/json cannot encode.
	err := Write(rec, http.StatusOK, make(chan int))

	// --- Assert ---
	// The encoding error must reach the caller so it can log it.
	if err == nil {
		t.Fatal("Write returned nil, want an encoding error")
	}
}

func TestRead(t *testing.T) {
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
	err := Read(req, &dst)

	// --- Assert ---
	// A well-formed body must decode without error.
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	// Each field must hold the value sent in the body.
	if dst.Amount != 100 || dst.Note != "hi" {
		t.Errorf("dst = %+v, want {Amount:100 Note:hi}", dst)
	}
}

func TestReadRejectsBadBodies(t *testing.T) {
	// --- Arrange ---
	// List bodies that Read must refuse, each named for the reason it is bad.
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
			err := Read(req, &dst)

			// --- Assert ---
			// The bad body must be refused.
			if err == nil {
				t.Fatal("Read returned nil, want an error")
			}
		})
	}
}
