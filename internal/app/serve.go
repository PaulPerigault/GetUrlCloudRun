package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/PaulPerigault/GetUrlCloudRun/internal/api"
	"github.com/PaulPerigault/GetUrlCloudRun/internal/config"
	"github.com/PaulPerigault/GetUrlCloudRun/internal/httpmw"
)

func serve(ctx context.Context, getenv func(string) string, out io.Writer) error {
	cfg, err := config.Load(getenv)
	if err != nil {
		return err
	}
	log := newLogger(cfg, out)
	prov, closeProv, err := newProvider(ctx, cfg)
	if err != nil {
		return err
	}
	defer closeProv()

	ln, err := (&net.ListenConfig{}).Listen(ctx, "tcp", ":"+strconv.Itoa(cfg.Port))
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	log.Info("listening", "addr", ln.Addr().String(), "provider", prov.Name(),
		"cors_origins", len(cfg.AllowedOrigins), "request_timeout", cfg.RequestTimeout.String())
	return (&runner{cfg: cfg, log: log}).run(ctx, newHandler(cfg, api.New(prov, log), log), ln)
}

// newHandler wraps the API with the middlewares, outermost first.
func newHandler(cfg config.Config, h http.Handler, log *slog.Logger) http.Handler {
	limiter := httpmw.NewLimiter(cfg.RateLimit, cfg.RateBurst, time.Now)
	return httpmw.Chain(h,
		httpmw.Recover(log),
		httpmw.AccessLog(log, time.Now),
		httpmw.CORS(cfg.AllowedOrigins),
		httpmw.RateLimit(limiter, "/health"),
		httpmw.Deadline(cfg.RequestTimeout),
	)
}

// runner owns the HTTP server lifecycle.
type runner struct {
	cfg config.Config
	log *slog.Logger
}

// run serves on ln until ctx is cancelled, then drains in-flight requests for
// at most cfg.ShutdownTimeout.
func (r *runner) run(ctx context.Context, h http.Handler, ln net.Listener) error {
	srv := &http.Server{
		Handler:           h,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      r.cfg.RequestTimeout + 5*time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    8 << 10,
	}
	errc := make(chan error, 1)
	go func() { errc <- srv.Serve(ln) }()

	select {
	case err := <-errc:
		return fmt.Errorf("serve: %w", err)
	case <-ctx.Done():
	}
	r.log.Info("shutting down", "timeout", r.cfg.ShutdownTimeout.String())
	shutCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), r.cfg.ShutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutCtx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}
	if err := <-errc; err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("serve: %w", err)
	}
	return nil
}
