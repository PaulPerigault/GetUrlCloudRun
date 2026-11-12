package httpmw_test

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/PaulPerigault/GetUrlCloudRun/internal/httpmw"
)

var ok = http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })

func do(h http.Handler, method, path string, headers map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestChainOrder(t *testing.T) {
	var order []string
	mk := func(name string) httpmw.Middleware {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, name)
				next.ServeHTTP(w, r)
			})
		}
	}
	do(httpmw.Chain(ok, mk("a"), mk("b")), http.MethodGet, "/", nil)
	if strings.Join(order, ",") != "a,b" {
		t.Errorf("order = %v, want a then b", order)
	}
}

func TestCORS(t *testing.T) {
	const listed = "https://app.example.com"
	tests := []struct {
		name       string
		allowed    []string
		method     string
		headers    map[string]string
		wantOrigin string
		wantStatus int
	}{
		{"no origin header", []string{listed}, http.MethodGet, nil, "", 200},
		{"listed origin", []string{listed}, http.MethodGet, map[string]string{"Origin": listed}, listed, 200},
		{"unlisted origin gets nothing", []string{listed}, http.MethodGet, map[string]string{"Origin": "https://evil.example"}, "", 200},
		{"empty list denies everything", nil, http.MethodGet, map[string]string{"Origin": listed}, "", 200},
		{"preflight listed", []string{listed}, http.MethodOptions,
			map[string]string{"Origin": listed, "Access-Control-Request-Method": "GET"}, listed, 204},
		{"preflight unlisted", []string{listed}, http.MethodOptions,
			map[string]string{"Origin": "https://evil.example", "Access-Control-Request-Method": "GET"}, "", 200},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rec := do(httpmw.CORS(tc.allowed)(ok), tc.method, "/", tc.headers)
			if got := rec.Header().Get("Access-Control-Allow-Origin"); got != tc.wantOrigin {
				t.Errorf("Allow-Origin = %q, want %q", got, tc.wantOrigin)
			}
			if rec.Code != tc.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tc.wantStatus)
			}
			if tc.wantStatus == 204 && rec.Header().Get("Access-Control-Allow-Methods") != "GET" {
				t.Errorf("preflight must offer GET only, got %q", rec.Header().Get("Access-Control-Allow-Methods"))
			}
		})
	}
}

type clock struct{ t time.Time }

func (c *clock) now() time.Time { return c.t }

func TestLimiterRefills(t *testing.T) {
	c := &clock{t: time.Unix(1000, 0)}
	l := httpmw.NewLimiter(2, 2, c.now)
	first, second := l.Allow(), l.Allow()
	if !first || !second {
		t.Fatal("the burst of 2 must pass")
	}
	if l.Allow() {
		t.Fatal("third call must be refused")
	}
	c.t = c.t.Add(500 * time.Millisecond) // one token at 2/s
	if !l.Allow() {
		t.Error("a token must be back after 500ms")
	}
	if l.Allow() {
		t.Error("only one token was refilled")
	}
	c.t = c.t.Add(time.Hour) // refill is capped by the burst
	a, b, c3 := l.Allow(), l.Allow(), l.Allow()
	if !a || !b || c3 {
		t.Error("refill must be capped at the burst")
	}
}

func TestRateLimitMiddleware(t *testing.T) {
	c := &clock{t: time.Unix(1000, 0)}
	h := httpmw.RateLimit(httpmw.NewLimiter(1, 1, c.now), "/health")(ok)
	if got := do(h, http.MethodGet, "/api", nil).Code; got != 200 {
		t.Fatalf("first request = %d", got)
	}
	rec := do(h, http.MethodGet, "/api", nil)
	if rec.Code != http.StatusTooManyRequests || rec.Header().Get("Retry-After") != "1" {
		t.Errorf("second request = %d, Retry-After %q", rec.Code, rec.Header().Get("Retry-After"))
	}
	if !strings.Contains(rec.Body.String(), "rate limit exceeded") {
		t.Errorf("body = %q", rec.Body.String())
	}
	if got := do(h, http.MethodGet, "/health", nil).Code; got != 200 {
		t.Errorf("exempt path = %d, want 200", got)
	}
}

func TestRecover(t *testing.T) {
	var logs bytes.Buffer
	boom := http.HandlerFunc(func(http.ResponseWriter, *http.Request) { panic("secret detail") })
	rec := do(httpmw.Recover(slog.New(slog.NewJSONHandler(&logs, nil)))(boom), http.MethodGet, "/x", nil)
	if rec.Code != 500 || strings.Contains(rec.Body.String(), "secret detail") {
		t.Errorf("status %d body %q: the panic value must not reach the client", rec.Code, rec.Body.String())
	}
	if !strings.Contains(logs.String(), "secret detail") {
		t.Errorf("the panic must be logged server side: %q", logs.String())
	}
}

func TestAccessLog(t *testing.T) {
	var logs bytes.Buffer
	c := &clock{t: time.Unix(1000, 0)}
	slow := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		c.t = c.t.Add(30 * time.Millisecond)
		w.WriteHeader(http.StatusTeapot)
	})
	h := httpmw.AccessLog(slog.New(slog.NewJSONHandler(&logs, nil)), c.now)(slow)
	do(h, http.MethodGet, "/api/v1/services?token=abc", map[string]string{"Authorization": "Bearer abc"})
	line := logs.String()
	for _, want := range []string{`"status":418`, `"duration_ms":30`, `"path":"/api/v1/services"`, `"method":"GET"`} {
		if !strings.Contains(line, want) {
			t.Errorf("log line %q lacks %s", line, want)
		}
	}
	if strings.Contains(line, "abc") {
		t.Errorf("query string or header leaked into the log: %q", line)
	}
}

func TestDeadline(t *testing.T) {
	var got time.Time
	h := httpmw.Deadline(time.Minute)(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		got, _ = r.Context().Deadline()
	}))
	do(h, http.MethodGet, "/", nil)
	if time.Until(got) <= 0 || time.Until(got) > time.Minute {
		t.Errorf("deadline = %v, want within one minute", got)
	}
}

func TestErrorBody(t *testing.T) {
	rec := httptest.NewRecorder()
	httpmw.Error(rec, http.StatusBadGateway, "upstream error")
	if rec.Code != 502 || rec.Header().Get("Content-Type") != "application/json; charset=utf-8" ||
		strings.TrimSpace(rec.Body.String()) != `{"error":"upstream error"}` {
		t.Errorf("unexpected response: %d %q %q", rec.Code, rec.Header().Get("Content-Type"), rec.Body.String())
	}
}
