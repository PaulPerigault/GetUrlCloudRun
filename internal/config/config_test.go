package config_test

import (
	"strings"
	"testing"
	"time"

	"github.com/PaulPerigault/GetUrlCloudRun/internal/config"
)

func env(m map[string]string) func(string) string {
	return func(k string) string { return m[k] }
}

func TestLoadValid(t *testing.T) {
	tests := []struct {
		name  string
		env   map[string]string
		check func(t *testing.T, c config.Config)
	}{
		{
			name: "fake needs nothing else",
			env:  map[string]string{"PROVIDER": "fake"},
			check: func(t *testing.T, c config.Config) {
				if c.Provider != "fake" || c.Port != 8080 || c.RequestTimeout != 5*time.Second || len(c.AllowedOrigins) != 0 {
					t.Errorf("unexpected defaults: %+v", c)
				}
			},
		},
		{
			name: "gcp with overrides",
			env: map[string]string{
				"GCP_PROJECT_ID": "my-project-123", "GCP_REGION": "europe-west1", "PORT": "9090",
				"ALLOWED_ORIGINS": "https://app.example.com, HTTP://localhost:4200,", "REQUEST_TIMEOUT": "2s",
				"RATE_LIMIT_RPS": "2.5", "RATE_LIMIT_BURST": "3", "LOG_LEVEL": "debug", "LOG_FORMAT": "text",
			},
			check: func(t *testing.T, c config.Config) {
				if c.Provider != "gcp" || c.Region != "europe-west1" || c.Port != 9090 || c.RateLimit != 2.5 || c.RateBurst != 3 {
					t.Errorf("unexpected values: %+v", c)
				}
				if len(c.AllowedOrigins) != 2 || c.AllowedOrigins[1] != "http://localhost:4200" {
					t.Errorf("origins = %v", c.AllowedOrigins)
				}
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c, err := config.Load(env(tc.env))
			if err != nil {
				t.Fatal(err)
			}
			tc.check(t, c)
		})
	}
}

func TestLoadRejects(t *testing.T) {
	tests := []struct {
		name string
		env  map[string]string
		want string
	}{
		{"gcp default needs a project", map[string]string{}, "GCP_PROJECT_ID"},
		{"bad project id", map[string]string{"GCP_PROJECT_ID": "X"}, "GCP_PROJECT_ID"},
		{"bad region", map[string]string{"GCP_PROJECT_ID": "my-project-123", "GCP_REGION": "mars"}, "GCP_REGION"},
		{"unknown provider", map[string]string{"PROVIDER": "aws"}, "PROVIDER"},
		{"port out of range", map[string]string{"PROVIDER": "fake", "PORT": "70000"}, "PORT"},
		{"port not a number", map[string]string{"PROVIDER": "fake", "PORT": "abc"}, "PORT"},
		{"wildcard origin", map[string]string{"PROVIDER": "fake", "ALLOWED_ORIGINS": "*"}, "ALLOWED_ORIGINS"},
		{"origin with path", map[string]string{"PROVIDER": "fake", "ALLOWED_ORIGINS": "https://a.example/x"}, "ALLOWED_ORIGINS"},
		{"timeout too long", map[string]string{"PROVIDER": "fake", "REQUEST_TIMEOUT": "1h"}, "REQUEST_TIMEOUT"},
		{"timeout not a duration", map[string]string{"PROVIDER": "fake", "SHUTDOWN_TIMEOUT": "soon"}, "SHUTDOWN_TIMEOUT"},
		{"rate not a number", map[string]string{"PROVIDER": "fake", "RATE_LIMIT_RPS": "fast"}, "RATE_LIMIT_RPS"},
		{"rate zero", map[string]string{"PROVIDER": "fake", "RATE_LIMIT_RPS": "0"}, "RATE_LIMIT_RPS"},
		{"burst zero", map[string]string{"PROVIDER": "fake", "RATE_LIMIT_BURST": "0"}, "RATE_LIMIT_BURST"},
		{"log level", map[string]string{"PROVIDER": "fake", "LOG_LEVEL": "loud"}, "LOG_LEVEL"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := config.Load(env(tc.env))
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("err = %v, want mention of %s", err, tc.want)
			}
		})
	}
}

func TestLoadReportsEveryProblem(t *testing.T) {
	_, err := config.Load(env(map[string]string{"PROVIDER": "fake", "PORT": "0", "LOG_LEVEL": "loud"}))
	if err == nil || !strings.Contains(err.Error(), "PORT") || !strings.Contains(err.Error(), "LOG_LEVEL") {
		t.Fatalf("err = %v, want both PORT and LOG_LEVEL", err)
	}
}
