// Command toolfetch installs and checks the pinned OpenTofu binary.
//
//	go run ./dev/toolfetch ensure   install .tools/bin/tofu if absent or at another version, then check it
//	go run ./dev/toolfetch check    fail unless .tools/bin/tofu is exactly the pinned version
//
// The version is read from infra/.opentofu-version and the SHA-256 of every
// release archive from infra/opentofu.sha256 (copied from the official
// SHA256SUMS of the release). A download whose digest differs is refused:
// nothing is installed without a fingerprint, and the check fails closed.
package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"runtime"
	"time"
)

const (
	versionFile = "infra/.opentofu-version"
	sumsFile    = "infra/opentofu.sha256"
	binDir      = ".tools/bin"
	releaseBase = "https://github.com/opentofu/opentofu/releases/download"
)

func main() {
	if len(os.Args) != 2 || (os.Args[1] != "ensure" && os.Args[1] != "check") {
		fmt.Fprintln(os.Stderr, "usage: toolfetch ensure|check")
		os.Exit(2)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	err := run(ctx, os.Args[1], options{
		VersionFile: versionFile,
		Spec:        installSpec{Base: releaseBase, OS: runtime.GOOS, Arch: runtime.GOARCH, SumsFile: sumsFile, BinDir: binDir},
		Out:         os.Stdout,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "toolfetch:", err)
		cancel()
		os.Exit(1)
	}
}

type options struct {
	VersionFile string
	Spec        installSpec // Version is filled from VersionFile
	Out         io.Writer
}

func run(ctx context.Context, mode string, o options) error {
	want, err := readVersion(o.VersionFile)
	if err != nil {
		return err
	}
	o.Spec.Version = want
	bin := binPath(o.Spec.BinDir, o.Spec.OS)
	if mode == "ensure" {
		if got, err := installedVersion(bin); err != nil || got != want {
			if err := install(ctx, o.Spec); err != nil {
				return err
			}
		}
	}
	got, err := installedVersion(bin)
	if err != nil {
		return err
	}
	if got != want {
		return fmt.Errorf("%s is OpenTofu %s, the repository pins %s (run `task infra:tool`)", bin, got, want)
	}
	_, err = fmt.Fprintf(o.Out, "OpenTofu %s (%s)\n", got, bin)
	return err
}
