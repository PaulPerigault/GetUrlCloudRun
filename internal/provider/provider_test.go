package provider_test

import (
	"strings"
	"testing"

	"github.com/PaulPerigault/GetUrlCloudRun/internal/provider"
)

func TestValidName(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want bool
	}{
		{"simple", "hello", true},
		{"with dash and digits", "billing-api-2", true},
		{"single letter", "a", true},
		{"max length 49", "a" + strings.Repeat("b", 48), true},
		{"too long 50", "a" + strings.Repeat("b", 49), false},
		{"empty", "", false},
		{"upper case", "Hello", false},
		{"starts with digit", "1abc", false},
		{"ends with dash", "abc-", false},
		{"underscore", "a_b", false},
		{"path traversal", "../x", false},
		{"slash", "a/b", false},
		{"newline", "abc\n", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := provider.ValidName(tc.in); got != tc.want {
				t.Errorf("ValidName(%q) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}
