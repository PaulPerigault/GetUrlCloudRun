package main

import (
	"bytes"
	"strings"
	"testing"
)

const profile = `mode: atomic
m/internal/a/a.go:1.1,2.2 4 1
m/internal/a/a.go:3.1,4.2 1 0
m/internal/b/b.go:1.1,2.2 10 3
m/dev/tool/main.go:1.1,2.2 10 0
`

func TestCheck(t *testing.T) {
	tests := []struct {
		name    string
		profile string
		min     float64
		want    string
	}{
		{"a is 80 percent, b is 100", profile, 75, ""},
		{"a fails at 85", profile, 85, "m/internal/a (80.0%)"},
		{"dev packages are ignored", profile, 100, "m/internal/a"},
		{"empty profile fails closed", "mode: atomic\n", 85, "no package"},
		{"garbage line fails", "mode: atomic\nm/internal/a/a.go:1.1,2.2 x y\n", 85, "unreadable"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var out bytes.Buffer
			err := check(strings.NewReader(tc.profile), "internal/", tc.min, &out)
			if tc.want == "" {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("err = %v, want %q", err, tc.want)
			}
		})
	}
}

func TestCheckPrintsEveryPackage(t *testing.T) {
	var out bytes.Buffer
	_ = check(strings.NewReader(profile), "internal/", 0, &out)
	if !strings.Contains(out.String(), "m/internal/a") || !strings.Contains(out.String(), "m/internal/b") || strings.Contains(out.String(), "dev") {
		t.Errorf("output = %q", out.String())
	}
}
