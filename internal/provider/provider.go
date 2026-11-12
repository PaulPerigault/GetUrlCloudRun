// Package provider defines the contract between the HTTP layer and the source
// of Cloud Run service data. Implementations live in provider/gcp (real API)
// and provider/fake (deterministic fixtures). This package is a leaf.
package provider

import (
	"context"
	"errors"
	"regexp"
)

// ErrNotFound is returned by GetService when no service has the given name.
var ErrNotFound = errors.New("service not found")

// serviceName is the Cloud Run service name grammar (1 to 49 characters).
var serviceName = regexp.MustCompile(`^[a-z]([-a-z0-9]{0,47}[a-z0-9])?$`)

// ValidName reports whether name is a well-formed Cloud Run service name.
func ValidName(name string) bool { return serviceName.MatchString(name) }

// Service is a Cloud Run service: its short name and its public URL.
type Service struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// Provider reads Cloud Run services. Implementations must honour ctx
// cancellation and deadlines.
type Provider interface {
	// ListServices returns every service, sorted by name.
	ListServices(ctx context.Context) ([]Service, error)
	// GetService returns one service by its short name, or ErrNotFound.
	GetService(ctx context.Context, name string) (Service, error)
	// Name identifies the implementation ("fake" or "gcp") for logs and /health.
	Name() string
}
