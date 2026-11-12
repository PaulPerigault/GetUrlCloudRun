package api_test

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/PaulPerigault/GetUrlCloudRun/internal/api"
	"github.com/PaulPerigault/GetUrlCloudRun/internal/provider"
	"github.com/PaulPerigault/GetUrlCloudRun/internal/provider/fake"
)

// failing is a provider whose every call returns err.
type failing struct{ err error }

func (f failing) Name() string { return "failing" }
func (f failing) ListServices(context.Context) ([]provider.Service, error) {
	return nil, f.err
}
func (f failing) GetService(context.Context, string) (provider.Service, error) {
	return provider.Service{}, f.err
}

func serve(t *testing.T, p provider.Provider, method, path string) (*httptest.ResponseRecorder, string) {
	t.Helper()
	var logs bytes.Buffer
	h := api.New(p, slog.New(slog.NewJSONHandler(&logs, nil)))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(method, path, nil))
	return rec, logs.String()
}

func fixtures(t *testing.T) provider.Provider {
	t.Helper()
	return fake.NewFrom([]provider.Service{
		{Name: "hello", URL: "https://hello.example"},
		{Name: "catalog", URL: "https://catalog.example"},
	})
}

func TestRoutes(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
		wantBody   string
	}{
		{"health", "GET", "/health", 200, `{"provider":"fake","status":"ok"}`},
		{"list", "GET", "/api/v1/services", 200,
			`[{"name":"catalog","url":"https://catalog.example"},{"name":"hello","url":"https://hello.example"}]`},
		{"get by name", "GET", "/api/v1/services/hello", 200, `{"url":"https://hello.example"}`},
		{"unknown name", "GET", "/api/v1/services/absent", 404, `{"error":"service not found"}`},
		{"invalid name", "GET", "/api/v1/services/Bad_Name", 400, `{"error":"invalid service name"}`},
		{"unknown route", "GET", "/nope", 404, `{"error":"not found"}`},
		{"post refused", "POST", "/api/v1/services", 405, `{"error":"method not allowed"}`},
		{"delete refused on item", "DELETE", "/api/v1/services/hello", 405, `{"error":"method not allowed"}`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rec, _ := serve(t, fixtures(t), tc.method, tc.path)
			if rec.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tc.wantStatus)
			}
			if tc.wantBody != "" && strings.TrimSpace(rec.Body.String()) != tc.wantBody {
				t.Errorf("body = %q, want %q", strings.TrimSpace(rec.Body.String()), tc.wantBody)
			}
		})
	}
}

func TestJSONHeaders(t *testing.T) {
	rec, _ := serve(t, fixtures(t), "GET", "/api/v1/services")
	for k, want := range map[string]string{
		"Content-Type":           "application/json; charset=utf-8",
		"Cache-Control":          "no-store",
		"X-Content-Type-Options": "nosniff",
	} {
		if got := rec.Header().Get(k); got != want {
			t.Errorf("%s = %q, want %q", k, got, want)
		}
	}
}

func TestEmptyListIsAnArrayNotNull(t *testing.T) {
	rec, _ := serve(t, fake.NewFrom(nil), "GET", "/api/v1/services")
	if strings.TrimSpace(rec.Body.String()) != "[]" {
		t.Errorf("body = %q, want []", rec.Body.String())
	}
}

func TestErrorsAreGenericForClientsAndDetailedInLogs(t *testing.T) {
	secret := "rpc error: code = Internal desc = projects/my-secret-project exploded"
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantBody   string
		wantLogged bool
	}{
		{"generic failure", errors.New(secret), 502, `{"error":"upstream error"}`, true},
		{"deadline", context.DeadlineExceeded, 504, `{"error":"upstream timeout"}`, true},
		{"cancelled", context.Canceled, 503, `{"error":"request cancelled"}`, false},
		{"not found from provider", provider.ErrNotFound, 404, `{"error":"service not found"}`, false},
	}
	for _, tc := range tests {
		for _, path := range []string{"/api/v1/services", "/api/v1/services/hello"} {
			t.Run(tc.name+" "+path, func(t *testing.T) {
				rec, logs := serve(t, failing{err: tc.err}, "GET", path)
				if rec.Code != tc.wantStatus || strings.TrimSpace(rec.Body.String()) != tc.wantBody {
					t.Errorf("got %d %q, want %d %q", rec.Code, rec.Body.String(), tc.wantStatus, tc.wantBody)
				}
				if strings.Contains(rec.Body.String(), "my-secret-project") {
					t.Error("provider error text leaked to the client")
				}
				if tc.wantLogged && !strings.Contains(logs, "path") {
					t.Errorf("error not logged server side: %q", logs)
				}
			})
		}
	}
}

func TestProviderDetailIsLoggedOnGenericFailure(t *testing.T) {
	_, logs := serve(t, failing{err: errors.New("boom detail")}, "GET", "/api/v1/services")
	if !strings.Contains(logs, "boom detail") {
		t.Errorf("log = %q, want the provider error", logs)
	}
}
