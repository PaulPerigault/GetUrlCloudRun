package httpmw

import (
	"log/slog"
	"net/http"
	"time"
)

// AccessLog writes one structured line per request: method, path, status and
// duration. It records no client address, header or query string.
func AccessLog(log *slog.Logger, now func() time.Time) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := now()
			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rec, r)
			log.InfoContext(r.Context(), "request",
				"method", r.Method, "path", r.URL.Path, "status", rec.status,
				"duration_ms", now().Sub(start).Milliseconds())
		})
	}
}
