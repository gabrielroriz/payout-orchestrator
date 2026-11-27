package httpjson

import "net/http"

// Code is a machine-readable error identifier that clients branch on.
// A code must never change once published; only the message may be reworded.
type Code string

// Codes shared by every route; domain codes, such as insufficient_funds, belong next to their handlers.
const (
	CodeNotFound         Code = "not_found"
	CodeMethodNotAllowed Code = "method_not_allowed"
	CodeBodyTooLarge     Code = "body_too_large"
	CodeInternal         Code = "internal_error"
)

// ErrorResponse is the body of every error response, such as {"error": {"code": "not_found", "message": "not found"}}.
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail describes one error: a Code for clients to handle and a Message for people to read.
type ErrorDetail struct {
	Code    Code   `json:"code"`
	Message string `json:"message"`
}

// Error sends an ErrorResponse with the given status, code and message.
func Error(w http.ResponseWriter, status int, code Code, message string) {
	// Ignore write failures: they mean the client has gone, and there is no one left to tell.
	_ = Write(w, status, ErrorResponse{Error: ErrorDetail{Code: code, Message: message}})
}
