// Command demo runs the real binary offline with the fake provider, calls
// every route and checks the answers. It needs no GCP account and no network
// beyond the loopback interface.
//
//	go run ./dev/demo -bin bin/geturlcloudrun
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type step struct {
	path     string
	status   int
	contains string
}

var steps = []step{
	{"/health", 200, `"provider":"fake"`},
	{"/api/v1/services", 200, `"name":"billing-api"`},
	{"/api/v1/services/hello", 200, `{"url":"https://hello-demo-ew.a.run.app"}`},
	{"/api/v1/services/absent", 404, `service not found`},
	{"/api/v1/services/Bad_Name", 400, `invalid service name`},
}

func main() {
	bin := flag.String("bin", "bin/geturlcloudrun", "binary to run")
	flag.Parse()
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	err := run(ctx, *bin, os.Stdout)
	cancel()
	if err != nil {
		fmt.Fprintln(os.Stderr, "demo:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, bin string, out io.Writer) error {
	port, err := freePort(ctx)
	if err != nil {
		return err
	}
	cmd := exec.CommandContext(ctx, bin, "serve")
	cmd.Env = append(os.Environ(), "PROVIDER=fake", "PORT="+port, "LOG_LEVEL=warn")
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start %s: %w", bin, err)
	}
	defer func() { _ = cmd.Process.Kill(); _ = cmd.Wait() }()

	base := "http://127.0.0.1:" + port
	if err := waitHealthy(ctx, base+"/health"); err != nil {
		return err
	}
	for _, s := range steps {
		code, body, err := get(ctx, base+s.path)
		if err != nil {
			return err
		}
		fmt.Fprintf(out, "GET %-28s -> %d %s\n", s.path, code, strings.TrimSpace(body))
		if code != s.status || !strings.Contains(body, s.contains) {
			return fmt.Errorf("GET %s: want %d containing %q", s.path, s.status, s.contains)
		}
	}
	fmt.Fprintln(out, "demo: ok")
	return nil
}

func freePort(ctx context.Context) (string, error) {
	ln, err := (&net.ListenConfig{}).Listen(ctx, "tcp", "127.0.0.1:0")
	if err != nil {
		return "", err
	}
	defer func() { _ = ln.Close() }()
	return strconv.Itoa(ln.Addr().(*net.TCPAddr).Port), nil
}

func waitHealthy(ctx context.Context, url string) error {
	for ctx.Err() == nil {
		if code, _, err := get(ctx, url); err == nil && code == http.StatusOK {
			return nil
		}
		time.Sleep(50 * time.Millisecond)
	}
	return fmt.Errorf("server not healthy: %w", ctx.Err())
}

func get(ctx context.Context, url string) (int, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return 0, "", err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer func() { _ = resp.Body.Close() }()
	b, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	return resp.StatusCode, string(b), err
}
