package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

// LogRequests logs one line per request once it has been served.
func LogRequests(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Start the clock and wrap the writer, so the status and size are known afterwards.
			start := time.Now()
			ww := wrap(w, r)

			next.ServeHTTP(ww, r)

			// Log server errors (5xx) at error level and everything else at info.
			level := slog.LevelInfo
			if statusOf(ww) >= http.StatusInternalServerError {
				level = slog.LevelError
			}

			// Log the path without its query string, which may carry sensitive values.
			logger.LogAttrs(r.Context(), level, "request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", statusOf(ww)),
				slog.Int("bytes", ww.BytesWritten()),
				slog.Duration("duration", time.Since(start)),
				slog.String("remote_addr", r.RemoteAddr),
				slog.String("request_id", RequestIDFrom(r.Context())),
			)
		})
	}
}
