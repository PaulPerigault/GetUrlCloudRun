// Package gcp reads Cloud Run services through the Cloud Run Admin API v2.
// Authentication is Application Default Credentials: on Cloud Run it is the
// service account of the service, locally it is `gcloud auth application-default
// login`. No credential file is read by this code.
package gcp

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	run "cloud.google.com/go/run/apiv2"
	"cloud.google.com/go/run/apiv2/runpb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/PaulPerigault/GetUrlCloudRun/internal/provider"
)

// Provider implements provider.Provider on top of the Cloud Run API.
type Provider struct {
	client *run.ServicesClient
	parent string
}

// New creates a provider for one project and one region. Extra client options
// are for tests (local endpoint, no authentication).
func New(ctx context.Context, projectID, region string, opts ...option.ClientOption) (*Provider, error) {
	client, err := run.NewServicesClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("gcp: create services client: %w", err)
	}
	return &Provider{
		client: client,
		parent: fmt.Sprintf("projects/%s/locations/%s", projectID, region),
	}, nil
}

// Close releases the underlying connection.
func (p *Provider) Close() error {
	if err := p.client.Close(); err != nil {
		return fmt.Errorf("gcp: close client: %w", err)
	}
	return nil
}

// Name implements provider.Provider.
func (p *Provider) Name() string { return "gcp" }

// ListServices implements provider.Provider.
func (p *Provider) ListServices(ctx context.Context) ([]provider.Service, error) {
	it := p.client.ListServices(ctx, &runpb.ListServicesRequest{Parent: p.parent})
	var out []provider.Service
	for {
		s, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("gcp: list services: %w", err)
		}
		out = append(out, toService(s))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

// GetService implements provider.Provider. One API call, no listing.
func (p *Provider) GetService(ctx context.Context, name string) (provider.Service, error) {
	if !provider.ValidName(name) {
		return provider.Service{}, fmt.Errorf("%w: %q is not a valid service name", provider.ErrNotFound, name)
	}
	s, err := p.client.GetService(ctx, &runpb.GetServiceRequest{Name: p.parent + "/services/" + name})
	if status.Code(err) == codes.NotFound {
		return provider.Service{}, fmt.Errorf("%w: %s", provider.ErrNotFound, name)
	}
	if err != nil {
		return provider.Service{}, fmt.Errorf("gcp: get service: %w", err)
	}
	return toService(s), nil
}

// toService keeps the short name: the last segment of the resource name.
func toService(s *runpb.Service) provider.Service {
	name := s.GetName()
	if i := strings.LastIndex(name, "/"); i >= 0 {
		name = name[i+1:]
	}
	return provider.Service{Name: name, URL: s.GetUri()}
}
