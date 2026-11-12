package main

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// buildServer compiles the real binary into a temporary directory.
func buildServer(t *testing.T) string {
	t.Helper()
	exe := filepath.Join(t.TempDir(), "geturlcloudrun")
	if runtime.GOOS == "windows" {
		exe += ".exe"
	}
	cmd := exec.Command("go", "build", "-o", exe, "../../cmd/geturlcloudrun")
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, out)
	}
	return exe
}

func TestDemoAgainstTheRealBinary(t *testing.T) {
	exe := buildServer(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	var out bytes.Buffer
	if err := run(ctx, exe, &out); err != nil {
		t.Fatalf("run: %v\n%s", err, out.String())
	}
	for _, want := range []string{"/api/v1/services/hello", "demo: ok", "-> 404"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("output lacks %q:\n%s", want, out.String())
		}
	}
}

func TestDemoFailsWhenTheBinaryIsMissing(t *testing.T) {
	err := run(context.Background(), filepath.Join(t.TempDir(), "absent"), &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "start") {
		t.Errorf("err = %v", err)
	}
}

func TestWaitHealthyGivesUp(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	if err := waitHealthy(ctx, "http://127.0.0.1:1/health"); err == nil {
		t.Error("an unreachable server must end in an error")
	}
}
