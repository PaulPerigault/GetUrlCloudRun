// Package app wires the bricks together: configuration, provider, HTTP
// handler, server lifecycle. It is the only package that knows every other
// one; main only calls Run.
package app

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/PaulPerigault/GetUrlCloudRun/internal/version"
)

// Commands lists the sub-commands Run accepts, for the documentation check.
var Commands = []string{"serve", "healthcheck", "version"}

// Run executes the command line: `serve` (default), `healthcheck` or
// `version`. ctx cancellation (SIGINT, SIGTERM) triggers a graceful shutdown.
// Log lines go to stdout, the version and health results go to out.
func Run(ctx context.Context, args []string, getenv func(string) string, out io.Writer) error {
	cmd := "serve"
	if len(args) > 0 {
		cmd = args[0]
	}
	switch cmd {
	case "serve":
		return serve(ctx, getenv, out)
	case "healthcheck":
		return healthcheck(ctx, getenv, out)
	case "version", "--version", "-v":
		_, err := fmt.Fprintln(out, version.String())
		return err
	default:
		return errors.New("unknown command " + cmd + ` (use "serve", "healthcheck" or "version")`)
	}
}
