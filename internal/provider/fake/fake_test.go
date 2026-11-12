package fake_test

import (
	"context"
	"errors"
	"testing"

	"github.com/PaulPerigault/GetUrlCloudRun/internal/provider"
	"github.com/PaulPerigault/GetUrlCloudRun/internal/provider/fake"
)

func TestNewLoadsSortedFixtures(t *testing.T) {
	p, err := fake.New()
	if err != nil {
		t.Fatal(err)
	}
	got, err := p.ListServices(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 5 {
		t.Fatalf("got %d services, want 5", len(got))
	}
	for i := 1; i < len(got); i++ {
		if got[i-1].Name >= got[i].Name {
			t.Errorf("not sorted: %q before %q", got[i-1].Name, got[i].Name)
		}
	}
	if p.Name() != "fake" {
		t.Errorf("Name() = %q", p.Name())
	}
}

func TestGetService(t *testing.T) {
	p := fake.NewFrom([]provider.Service{{Name: "b", URL: "https://b"}, {Name: "a", URL: "https://a"}})
	tests := []struct {
		name    string
		query   string
		want    string
		wantErr error
	}{
		{"found", "a", "https://a", nil},
		{"missing", "zzz", "", provider.ErrNotFound},
		{"empty", "", "", provider.ErrNotFound},
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

func TestListReturnsCopy(t *testing.T) {
	p := fake.NewFrom([]provider.Service{{Name: "a", URL: "u"}})
	first, _ := p.ListServices(context.Background())
	first[0].URL = "mutated"
	second, _ := p.ListServices(context.Background())
	if second[0].URL != "u" {
		t.Error("caller mutation leaked into the provider")
	}
}

func TestCancelledContext(t *testing.T) {
	p := fake.NewFrom(nil)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := p.ListServices(ctx); !errors.Is(err, context.Canceled) {
		t.Errorf("ListServices err = %v", err)
	}
	if _, err := p.GetService(ctx, "a"); !errors.Is(err, context.Canceled) {
		t.Errorf("GetService err = %v", err)
	}
}
