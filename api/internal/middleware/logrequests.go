package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

// LogRequests logs one line per request once it has been served, with its
// method, path, status, response size and duration. Server errors (5xx) are
// logged at error level, everything else at info. The query string is left out
// because it may carry sensitive values.
func LogRequests(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := WrapWriter(w)

			next.ServeHTTP(ww, r)

			level := slog.LevelInfo
			if ww.Status() >= http.StatusInternalServerError {
				level = slog.LevelError
			}

			logger.LogAttrs(r.Context(), level, "request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", ww.Status()),
				slog.Int("bytes", ww.BytesWritten()),
				slog.Duration("duration", time.Since(start)),
				slog.String("remote_addr", r.RemoteAddr),
			)
		})
	}
}
