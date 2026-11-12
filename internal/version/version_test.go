package version_test

import (
	"strings"
	"testing"

	"github.com/PaulPerigault/GetUrlCloudRun/internal/version"
)

func TestString(t *testing.T) {
	got := version.String()
	for _, want := range []string{"geturlcloudrun", version.Version, version.Commit, version.Date} {
		if !strings.Contains(got, want) {
			t.Errorf("String() = %q, missing %q", got, want)
		}
	}
}
