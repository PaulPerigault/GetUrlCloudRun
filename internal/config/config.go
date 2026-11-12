// Package config reads the service configuration from environment variables.
// It is a leaf package: it knows nothing about HTTP or GCP.
package config

import (
	"errors"
	"fmt"
	"regexp"
	"time"
)

// Provider names accepted by PROVIDER.
const (
	ProviderFake = "fake"
	ProviderGCP  = "gcp"
)

var (
	projectIDRe = regexp.MustCompile(`^[a-z][-a-z0-9]{4,28}[a-z0-9]$`)
	regionRe    = regexp.MustCompile(`^[a-z]+-[a-z]+[0-9]+$`)
)

// Config is the validated configuration.
type Config struct {
	Provider        string
	ProjectID       string
	Region          string
	Port            int
	AllowedOrigins  []string
	RequestTimeout  time.Duration
	ShutdownTimeout time.Duration
	RateLimit       float64 // requests per second, global
	RateBurst       int
	LogLevel        string
	LogFormat       string
}

// Load reads and validates the configuration. getenv is os.Getenv in
// production and a map lookup in tests. Every problem is reported, not only
// the first one.
func Load(getenv func(string) string) (Config, error) {
	p := &parser{getenv: getenv}
	c := Config{
		Provider:        p.str("PROVIDER", ProviderGCP),
		ProjectID:       p.str("GCP_PROJECT_ID", ""),
		Region:          p.str("GCP_REGION", "us-central1"),
		Port:            p.intIn("PORT", 8080, 1, 65535),
		AllowedOrigins:  p.origins("ALLOWED_ORIGINS"),
		RequestTimeout:  p.duration("REQUEST_TIMEOUT", 5*time.Second),
		ShutdownTimeout: p.duration("SHUTDOWN_TIMEOUT", 10*time.Second),
		RateLimit:       p.float("RATE_LIMIT_RPS", 20),
		RateBurst:       p.intIn("RATE_LIMIT_BURST", 40, 1, 100000),
		LogLevel:        p.oneOf("LOG_LEVEL", "info", "debug", "info", "warn", "error"),
		LogFormat:       p.oneOf("LOG_FORMAT", "json", "json", "text"),
	}
	c.validateProvider(p)
	if err := errors.Join(p.errs...); err != nil {
		return Config{}, fmt.Errorf("config: %w", err)
	}
	return c, nil
}

func (c Config) validateProvider(p *parser) {
	switch c.Provider {
	case ProviderFake:
	case ProviderGCP:
		if !projectIDRe.MatchString(c.ProjectID) {
			p.errs = append(p.errs, errors.New("GCP_PROJECT_ID is required for PROVIDER=gcp and must be a valid project ID"))
		}
		if !regionRe.MatchString(c.Region) {
			p.errs = append(p.errs, fmt.Errorf("GCP_REGION %q is not a valid region", c.Region))
		}
	default:
		p.errs = append(p.errs, fmt.Errorf("PROVIDER %q must be %q or %q", c.Provider, ProviderFake, ProviderGCP))
	}
}
