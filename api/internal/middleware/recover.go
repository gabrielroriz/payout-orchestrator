package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/gabrielroriz/payout-orchestrator/api/internal/httpjson"
)

// Recover turns a panic in a handler into a logged error and, when it still can, a 500 JSON response.
func Recover(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Wrap the writer to learn whether the response has started; this reuses LogRequests' wrapper.
			ww := wrap(w, r)

			// Catch a panic on the way out; without this, net/http drops the connection and the client gets no answer.
			defer func() {
				v := recover()
				if v == nil {
					return
				}

				// http.ErrAbortHandler is a deliberate abort, not a bug: let net/http handle it quietly.
				if v == http.ErrAbortHandler {
					panic(v)
				}

				// Log the panic with its stack trace and request ID, to find the bug.
				logger.LogAttrs(r.Context(), slog.LevelError, "panic",
					slog.Any("panic", v),
					slog.String("stack", string(debug.Stack())),
					slog.String("method", r.Method),
					slog.String("path", r.URL.Path),
					slog.String("request_id", RequestIDFrom(r.Context())),
				)

				// If the response has started, the status can no longer change.
				// Abort it, so the client sees a broken response instead of a cut-off one that looks complete.
				if headerSent(ww) {
					panic(http.ErrAbortHandler)
				}

				// Drop the headers the handler set for the response it never finished, such as Content-Length, keeping only the request ID.
				h := ww.Header()
				id := h.Get(RequestIDHeader)
				clear(h)
				if id != "" {
					h.Set(RequestIDHeader, id)
				}

				// Nothing was sent yet, so answer with a clean 500.
				httpjson.Error(ww, http.StatusInternalServerError, httpjson.CodeInternal, "internal server error")
			}()

			next.ServeHTTP(ww, r)
		})
	}
}
