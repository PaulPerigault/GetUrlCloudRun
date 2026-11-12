package httpmw

import (
	"log/slog"
	"net/http"
	"runtime/debug"
)

// Recover turns a panic into a logged error and a generic 500.
func Recover(log *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					log.ErrorContext(r.Context(), "panic in handler",
						"panic", rec, "path", r.URL.Path, "stack", string(debug.Stack()))
					Error(w, http.StatusInternalServerError, "internal error")
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
