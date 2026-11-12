package httpmw

import (
	"context"
	"net/http"
	"time"
)

// Deadline gives every request a context that expires after d, so that a slow
// upstream call is cut instead of holding the connection.
func Deadline(d time.Duration) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), d)
			defer cancel()
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
