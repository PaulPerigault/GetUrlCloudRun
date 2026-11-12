// Package fake is a deterministic in-memory provider backed by recorded
// fixtures. It lets the binary, the tests and the demo run without a GCP
// account or any network access.
package fake

import (
	"context"
	_ "embed" // fixtures are embedded in the binary
	"encoding/json"
	"fmt"
	"sort"

	"github.com/PaulPerigault/GetUrlCloudRun/internal/provider"
)

//go:embed services.json
var fixtures []byte

// Provider serves a fixed list of services.
type Provider struct {
	services []provider.Service
}

// New returns a provider loaded with the embedded fixtures.
func New() (*Provider, error) {
	var services []provider.Service
	if err := json.Unmarshal(fixtures, &services); err != nil {
		return nil, fmt.Errorf("fake: decode fixtures: %w", err)
	}
	return NewFrom(services), nil
}

// NewFrom returns a provider serving the given services, sorted by name.
func NewFrom(services []provider.Service) *Provider {
	sorted := append([]provider.Service(nil), services...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Name < sorted[j].Name })
	return &Provider{services: sorted}
}

// Name implements provider.Provider.
func (p *Provider) Name() string { return "fake" }

// ListServices implements provider.Provider.
func (p *Provider) ListServices(ctx context.Context) ([]provider.Service, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("fake: list services: %w", err)
	}
	return append([]provider.Service(nil), p.services...), nil
}

// GetService implements provider.Provider.
func (p *Provider) GetService(ctx context.Context, name string) (provider.Service, error) {
	if err := ctx.Err(); err != nil {
		return provider.Service{}, fmt.Errorf("fake: get service: %w", err)
	}
	for _, s := range p.services {
		if s.Name == name {
			return s, nil
		}
	}
	return provider.Service{}, fmt.Errorf("%w: %s", provider.ErrNotFound, name)
}
