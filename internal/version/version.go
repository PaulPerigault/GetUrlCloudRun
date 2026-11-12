// Package version holds the build metadata injected with -ldflags.
package version

// Set at link time: -X <module>/internal/version.Version=...
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

// String returns a one-line description of the build.
func String() string {
	return "geturlcloudrun " + Version + " (commit " + Commit + ", built " + Date + ")"
}
