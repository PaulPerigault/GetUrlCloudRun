package app

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

// healthcheck calls GET /health on the local server. A distroless image has
// no shell, curl or wget, so the binary checks itself (used by Docker Compose).
func healthcheck(ctx context.Context, getenv func(string) string, out io.Writer) error {
	port := getenv("PORT")
	if port == "" {
		port = "8080"
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://127.0.0.1:"+port+"/health", http.NoBody)
	if err != nil {
		return fmt.Errorf("healthcheck: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("healthcheck: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("healthcheck: status %d", resp.StatusCode)
	}
	_, err = fmt.Fprintln(out, "ok")
	return err
}
