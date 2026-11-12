package httpmw

import (
	"net/http"
	"slices"
	"sync"
	"time"
)

// Limiter is a token bucket shared by all clients. Cloud Run terminates TLS
// in front of the container, so a per-client key would be spoofable; a global
// bucket bounds the load on the Cloud Run Admin API, which is what matters.
type Limiter struct {
	mu     sync.Mutex
	rate   float64 // tokens per second
	burst  float64
	tokens float64
	last   time.Time
	now    func() time.Time
}

// NewLimiter returns a full bucket. now is injected for tests.
func NewLimiter(rps float64, burst int, now func() time.Time) *Limiter {
	return &Limiter{rate: rps, burst: float64(burst), tokens: float64(burst), last: now(), now: now}
}

// Allow takes one token if there is one.
func (l *Limiter) Allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	t := l.now()
	l.tokens = min(l.burst, l.tokens+t.Sub(l.last).Seconds()*l.rate)
	l.last = t
	if l.tokens < 1 {
		return false
	}
	l.tokens--
	return true
}

// RateLimit answers 429 when the bucket is empty. Paths in exempt (health
// checks) never consume a token.
func RateLimit(l *Limiter, exempt ...string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !slices.Contains(exempt, r.URL.Path) && !l.Allow() {
				w.Header().Set("Retry-After", "1")
				Error(w, http.StatusTooManyRequests, "rate limit exceeded")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
