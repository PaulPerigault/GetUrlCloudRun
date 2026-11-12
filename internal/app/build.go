package app

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/PaulPerigault/GetUrlCloudRun/internal/config"
	"github.com/PaulPerigault/GetUrlCloudRun/internal/provider"
	"github.com/PaulPerigault/GetUrlCloudRun/internal/provider/fake"
	"github.com/PaulPerigault/GetUrlCloudRun/internal/provider/gcp"
)

func newLogger(cfg config.Config, w io.Writer) *slog.Logger {
	var level slog.Level
	_ = level.UnmarshalText([]byte(cfg.LogLevel)) // validated by config.Load
	opts := &slog.HandlerOptions{Level: level}
	if cfg.LogFormat == "text" {
		return slog.New(slog.NewTextHandler(w, opts))
	}
	return slog.New(slog.NewJSONHandler(w, opts))
}

// newProvider builds the provider chosen by PROVIDER and the function that
// releases it.
func newProvider(ctx context.Context, cfg config.Config) (provider.Provider, func(), error) {
	if cfg.Provider == config.ProviderFake {
		p, err := fake.New()
		if err != nil {
			return nil, nil, fmt.Errorf("provider: %w", err)
		}
		return p, func() {}, nil
	}
	p, err := gcp.New(ctx, cfg.ProjectID, cfg.Region)
	if err != nil {
		return nil, nil, fmt.Errorf("provider: %w", err)
	}
	return p, func() { _ = p.Close() }, nil
}
