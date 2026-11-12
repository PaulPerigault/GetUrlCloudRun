package httpmw

import (
	"net/http"
	"slices"
)

// CORS allows cross-origin reads from an explicit list of origins and from
// nobody else: an empty list means no CORS header is ever sent. Only GET is
// offered. A request from an origin that is not listed is still served (CORS
// is enforced by the browser), but it gets no Access-Control-* header.
func CORS(allowed []string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" && slices.Contains(allowed, origin) {
				w.Header().Add("Vary", "Origin")
				w.Header().Set("Access-Control-Allow-Origin", origin)
				if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
					w.Header().Set("Access-Control-Allow-Methods", http.MethodGet)
					w.Header().Set("Access-Control-Max-Age", "600")
					w.WriteHeader(http.StatusNoContent)
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
