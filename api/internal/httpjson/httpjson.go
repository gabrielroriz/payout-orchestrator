// Package httpjson reads and writes JSON bodies, so every handler and middleware answers in the same shape.
package httpjson

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

// Write sends data as a JSON response with the given status code.
func Write(w http.ResponseWriter, status int, data any) error {
	// Set the header before the status, since headers are frozen once the status is sent.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}

// Read decodes a JSON request body into dst; middleware.LimitBody caps its size.
func Read(r *http.Request, dst any) error {
	// Reject fields that dst does not declare, so typos in a request fail loudly.
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return err
	}

	// A second Decode must hit EOF; anything else means the body held more than one value.
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return errors.New("body must contain a single JSON value")
	}
	return nil
}
