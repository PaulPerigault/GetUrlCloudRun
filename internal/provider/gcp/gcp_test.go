package gcp_test

import (
	"context"
	"errors"
	"net"
	"testing"

	"cloud.google.com/go/run/apiv2/runpb"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	"github.com/PaulPerigault/GetUrlCloudRun/internal/provider"
	"github.com/PaulPerigault/GetUrlCloudRun/internal/provider/gcp"
)

const parent = "projects/demo/locations/europe-west1"

// stubServer is an in-process Cloud Run API served on a loopback port.
type stubServer struct {
	runpb.UnimplementedServicesServer
	services map[string]string
	failList bool
	failGet  bool
}

func (s *stubServer) ListServices(_ context.Context, req *runpb.ListServicesRequest) (*runpb.ListServicesResponse, error) {
	if s.failList {
		return nil, status.Error(codes.PermissionDenied, "denied")
	}
	if req.GetParent() != parent {
		return nil, status.Error(codes.InvalidArgument, "bad parent "+req.GetParent())
	}
	var out []*runpb.Service
	for name, url := range s.services {
		out = append(out, &runpb.Service{Name: parent + "/services/" + name, Uri: url})
	}
	return &runpb.ListServicesResponse{Services: out}, nil
}

func (s *stubServer) GetService(_ context.Context, req *runpb.GetServiceRequest) (*runpb.Service, error) {
	if s.failGet {
		return nil, status.Error(codes.Unavailable, "down")
	}
	for name, url := range s.services {
		if req.GetName() == parent+"/services/"+name {
			return &runpb.Service{Name: req.GetName(), Uri: url}, nil
		}
	}
	return nil, status.Error(codes.NotFound, "no such service")
}

func newProvider(t *testing.T, stub *stubServer) *gcp.Provider {
	t.Helper()
	lis, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	srv := grpc.NewServer()
	runpb.RegisterServicesServer(srv, stub)
	go func() { _ = srv.Serve(lis) }()
	t.Cleanup(srv.Stop)

	p, err := gcp.New(context.Background(), "demo", "europe-west1",
		option.WithEndpoint(lis.Addr().String()),
		option.WithoutAuthentication(),
		option.WithGRPCDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = p.Close() })
	return p
}

func TestListServicesSortedShortNames(t *testing.T) {
	p := newProvider(t, &stubServer{services: map[string]string{"zeta": "https://z", "alpha": "https://a"}})
	got, err := p.ListServices(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	want := []provider.Service{{Name: "alpha", URL: "https://a"}, {Name: "zeta", URL: "https://z"}}
	if len(got) != 2 || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("got %v, want %v", got, want)
	}
	if p.Name() != "gcp" {
		t.Errorf("Name() = %q", p.Name())
	}
}

func TestGetService(t *testing.T) {
	p := newProvider(t, &stubServer{services: map[string]string{"hello": "https://hello"}})
	tests := []struct {
		name    string
		query   string
		want    string
		wantErr error
	}{
		{"found", "hello", "https://hello", nil},
		{"absent", "missing", "", provider.ErrNotFound},
		{"invalid grammar", "Bad_Name", "", provider.ErrNotFound},
		{"too long", "a123456789012345678901234567890123456789012345678", "", provider.ErrNotFound},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s, err := p.GetService(context.Background(), tc.query)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("err = %v, want %v", err, tc.wantErr)
			}
			if s.URL != tc.want {
				t.Errorf("URL = %q, want %q", s.URL, tc.want)
			}
		})
	}
}

func TestAPIErrorsAreWrapped(t *testing.T) {
	p := newProvider(t, &stubServer{failList: true, failGet: true})
	if _, err := p.ListServices(context.Background()); err == nil || errors.Is(err, provider.ErrNotFound) {
		t.Errorf("ListServices err = %v, want a non-NotFound error", err)
	}
	if _, err := p.GetService(context.Background(), "hello"); err == nil || errors.Is(err, provider.ErrNotFound) {
		t.Errorf("GetService err = %v, want a non-NotFound error", err)
	}
}
