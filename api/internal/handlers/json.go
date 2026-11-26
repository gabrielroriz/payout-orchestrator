package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

// writeJSON sends data as a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}

// readJSON decodes a JSON request body into dst. It rejects fields that dst
// does not declare and trailing data after the first JSON value. Body size is
// capped by middleware.LimitBody, which wraps every route.
func readJSON(r *http.Request, dst any) error {
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
