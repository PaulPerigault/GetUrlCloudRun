package app_test

import (
	"bytes"
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/PaulPerigault/GetUrlCloudRun/internal/app"
)

func env(m map[string]string) func(string) string {
	return func(k string) string { return m[k] }
}

func freePort(t *testing.T) string {
	t.Helper()
	ln, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = ln.Close() }()
	return strconv.Itoa(ln.Addr().(*net.TCPAddr).Port)
}

func get(t *testing.T, rawURL string) (int, string) {
	t.Helper()
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, rawURL, http.NoBody)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err.Error()
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, strings.TrimSpace(string(body))
}

func waitFor(t *testing.T, rawURL string) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		if code, _ := get(t, rawURL); code == http.StatusOK {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("%s did not answer within 10s", rawURL)
}

func TestVersionAndUnknownCommand(t *testing.T) {
	var out bytes.Buffer
	for _, arg := range []string{"version", "--version", "-v"} {
		out.Reset()
		if err := app.Run(context.Background(), []string{arg}, env(nil), &out); err != nil {
			t.Fatal(err)
		}
		if !strings.HasPrefix(out.String(), "geturlcloudrun ") {
			t.Errorf("%s printed %q", arg, out.String())
		}
	}
	err := app.Run(context.Background(), []string{"frobnicate"}, env(nil), &out)
	if err == nil || !strings.Contains(err.Error(), "unknown command") {
		t.Errorf("err = %v", err)
	}
}

func TestServeRejectsBadConfigBeforeListening(t *testing.T) {
	err := app.Run(context.Background(), nil, env(map[string]string{}), io.Discard)
	if err == nil || !strings.Contains(err.Error(), "GCP_PROJECT_ID") {
		t.Errorf("err = %v, want a configuration error", err)
	}
}

func TestServeGCPWithoutCredentialsFailsClosed(t *testing.T) {
	t.Setenv("GOOGLE_APPLICATION_CREDENTIALS", t.TempDir()+"/absent.json")
	err := app.Run(context.Background(), []string{"serve"},
		env(map[string]string{"PROVIDER": "gcp", "GCP_PROJECT_ID": "demo-project-1", "PORT": freePort(t)}), io.Discard)
	if err == nil || !strings.Contains(err.Error(), "provider") {
		t.Errorf("err = %v, want a provider error", err)
	}
}

func TestServeListenFailure(t *testing.T) {
	ln, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = ln.Close() }()
	port := strconv.Itoa(ln.Addr().(*net.TCPAddr).Port)
	err = app.Run(context.Background(), nil, env(map[string]string{"PROVIDER": "fake", "PORT": port}), io.Discard)
	if err == nil || !strings.Contains(err.Error(), "listen") {
		t.Errorf("err = %v, want a listen error", err)
	}
}

func TestServeEndToEndWithFakeProvider(t *testing.T) {
	port := freePort(t)
	base := "http://127.0.0.1:" + port
	ctx, cancel := context.WithCancel(context.Background())
	var logs bytes.Buffer
	done := make(chan error, 1)
	go func() {
		done <- app.Run(ctx, []string{"serve"}, env(map[string]string{
			"PROVIDER": "fake", "PORT": port, "ALLOWED_ORIGINS": "https://app.example.com",
			"RATE_LIMIT_RPS": "1000", "RATE_LIMIT_BURST": "1000", "LOG_FORMAT": "text",
		}), &logs)
	}()
	waitFor(t, base+"/health")

	if code, body := get(t, base+"/api/v1/services/hello"); code != 200 || body != `{"url":"https://hello-demo-ew.a.run.app"}` {
		t.Errorf("GET hello = %d %q", code, body)
	}
	if code, body := get(t, base+"/api/v1/services/absent"); code != 404 || !strings.Contains(body, "service not found") {
		t.Errorf("GET absent = %d %q", code, body)
	}
	var healthOut bytes.Buffer
	if err := app.Run(ctx, []string{"healthcheck"}, env(map[string]string{"PORT": port}), &healthOut); err != nil || healthOut.String() != "ok\n" {
		t.Errorf("healthcheck = %v %q", err, healthOut.String())
	}

	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("graceful shutdown returned %v", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("server did not stop within 10s of cancellation")
	}
	for _, want := range []string{"listening", "provider=fake", "shutting down"} {
		if !strings.Contains(logs.String(), want) {
			t.Errorf("logs lack %q: %s", want, logs.String())
		}
	}
}

func TestServeRateLimits(t *testing.T) {
	port := freePort(t)
	base := "http://127.0.0.1:" + port
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- app.Run(ctx, nil, env(map[string]string{
			"PROVIDER": "fake", "PORT": port, "RATE_LIMIT_RPS": "0.001", "RATE_LIMIT_BURST": "1",
		}), io.Discard)
	}()
	waitFor(t, base+"/health")
	if code, _ := get(t, base+"/api/v1/services"); code != 200 {
		t.Errorf("first request = %d", code)
	}
	if code, _ := get(t, base+"/api/v1/services"); code != http.StatusTooManyRequests {
		t.Errorf("second request = %d, want 429", code)
	}
	if code, _ := get(t, base+"/health"); code != 200 {
		t.Errorf("health must stay reachable, got %d", code)
	}
	cancel()
	if err := <-done; err != nil {
		t.Errorf("shutdown returned %v", err)
	}
}

func TestHealthcheckFailures(t *testing.T) {
	bad := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer bad.Close()
	u, _ := url.Parse(bad.URL)

	tests := []struct {
		name string
		port string
		want string
	}{
		{"unhealthy server", u.Port(), "status 500"},
		{"nothing listening", freePort(t), "healthcheck"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := app.Run(context.Background(), []string{"healthcheck"}, env(map[string]string{"PORT": tc.port}), io.Discard)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Errorf("err = %v, want %q", err, tc.want)
			}
		})
	}
}

func TestEveryListedCommandIsAccepted(t *testing.T) {
	for _, c := range app.Commands {
		err := app.Run(context.Background(), []string{c}, env(map[string]string{"PROVIDER": "aws"}), io.Discard)
		if err != nil && strings.Contains(err.Error(), "unknown command") {
			t.Errorf("app.Commands lists %q but Run rejects it", c)
		}
	}
}
