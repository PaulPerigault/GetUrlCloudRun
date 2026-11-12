package main

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func makeZip(t *testing.T, name, content string) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for _, n := range []string{"LICENSE", name} {
		w, err := zw.Create(n)
		if err != nil {
			t.Fatal(err)
		}
		_, _ = w.Write([]byte(content))
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func sumOf(b []byte) string {
	s := sha256.Sum256(b)
	return hex.EncodeToString(s[:])
}

func spec(t *testing.T, base, sums string) installSpec {
	t.Helper()
	dir := t.TempDir()
	sumsPath := filepath.Join(dir, "sums")
	if err := os.WriteFile(sumsPath, []byte(sums), 0o600); err != nil {
		t.Fatal(err)
	}
	return installSpec{Base: base, Version: "9.9.9", OS: "linux", Arch: "amd64", SumsFile: sumsPath, BinDir: filepath.Join(dir, "bin")}
}

func TestInstall(t *testing.T) {
	archive := makeZip(t, "tofu", "binary-bytes")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v9.9.9/tofu_9.9.9_linux_amd64.zip" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write(archive)
	}))
	defer srv.Close()

	tests := []struct {
		name    string
		sums    string
		wantErr string
	}{
		{"matching digest installs", sumOf(archive) + "  tofu_9.9.9_linux_amd64.zip\n", ""},
		{"digest mismatch is refused", strings.Repeat("0", 64) + "  tofu_9.9.9_linux_amd64.zip\n", "checksum mismatch"},
		{"no fingerprint, no download", sumOf(archive) + "  tofu_9.9.9_darwin_arm64.zip\n", "without a fingerprint"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := spec(t, srv.URL, tc.sums)
			err := install(context.Background(), s)
			if tc.wantErr == "" {
				if err != nil {
					t.Fatal(err)
				}
				got, rerr := os.ReadFile(filepath.Join(s.BinDir, "tofu"))
				if rerr != nil || string(got) != "binary-bytes" {
					t.Errorf("installed file = %q, %v", got, rerr)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("err = %v, want %q", err, tc.wantErr)
			}
			if _, statErr := os.Stat(filepath.Join(s.BinDir, "tofu")); statErr == nil {
				t.Error("a binary was written despite the refusal")
			}
		})
	}
}

func TestInstallDownloadFailures(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer srv.Close()
	s := spec(t, srv.URL, strings.Repeat("a", 64)+"  tofu_9.9.9_linux_amd64.zip\n")
	if err := install(context.Background(), s); err == nil || !strings.Contains(err.Error(), "status 502") {
		t.Errorf("err = %v, want status 502", err)
	}
	s.SumsFile = filepath.Join(t.TempDir(), "absent")
	if err := install(context.Background(), s); err == nil {
		t.Error("a missing checksum file must fail")
	}
}

func TestExtractRejectsArchiveWithoutBinary(t *testing.T) {
	data := makeZip(t, "other", "x")
	if err := extract(data, filepath.Join(t.TempDir(), "tofu")); err == nil {
		t.Error("expected an error when the archive has no tofu")
	}
	if err := extract([]byte("not a zip"), filepath.Join(t.TempDir(), "tofu")); err == nil {
		t.Error("expected an error for a corrupt archive")
	}
}

func TestReadVersion(t *testing.T) {
	dir := t.TempDir()
	for content, ok := range map[string]bool{"1.13.9\n": true, "": false, "1.13 9": false, "../x": false} {
		p := filepath.Join(dir, "v")
		_ = os.WriteFile(p, []byte(content), 0o600)
		if _, err := readVersion(p); (err == nil) != ok {
			t.Errorf("readVersion(%q) err = %v, want ok=%v", content, err, ok)
		}
	}
	if _, err := readVersion(filepath.Join(dir, "absent")); err == nil {
		t.Error("a missing file must fail")
	}
}

func TestBinPathAndParseSums(t *testing.T) {
	if got := binPath("d", "windows"); filepath.Base(got) != "tofu.exe" {
		t.Errorf("windows = %q", got)
	}
	if got := binPath("d", "linux"); filepath.Base(got) != "tofu" {
		t.Errorf("linux = %q", got)
	}
	sums := parseSums(strings.NewReader("AB12  a.zip\nbroken line here\n\ncd34  b.zip\n"))
	if sums["a.zip"] != "ab12" || sums["b.zip"] != "cd34" || len(sums) != 2 {
		t.Errorf("sums = %v", sums)
	}
}

func TestInstalledVersionFailsClosed(t *testing.T) {
	if _, err := installedVersion(filepath.Join(t.TempDir(), "absent")); err == nil {
		t.Error("a missing binary must fail")
	}
}

// TestMain turns the test binary into a fake `tofu` when HELPER_TOFU_VERSION
// is set, so that installedVersion and run can be tested without OpenTofu.
func TestMain(m *testing.M) {
	if v := os.Getenv("HELPER_TOFU_VERSION"); v != "" {
		_, _ = os.Stdout.WriteString(`{"terraform_version":"` + v + `"}`)
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func TestInstalledVersionReadsJSON(t *testing.T) {
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("HELPER_TOFU_VERSION", "9.9.9")
	if got, err := installedVersion(exe); err != nil || got != "9.9.9" {
		t.Errorf("got %q, %v", got, err)
	}
}

func TestRunEnsureAndCheck(t *testing.T) {
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	exeBytes, err := os.ReadFile(exe)
	if err != nil {
		t.Fatal(err)
	}
	name := "tofu"
	if runtime.GOOS == "windows" {
		name = "tofu.exe"
	}
	archive := makeZip(t, name, string(exeBytes))
	file := fmt.Sprintf("tofu_9.9.9_%s_%s.zip", runtime.GOOS, runtime.GOARCH)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write(archive) }))
	defer srv.Close()

	dir := t.TempDir()
	vf := filepath.Join(dir, "version")
	_ = os.WriteFile(vf, []byte("9.9.9\n"), 0o600)
	s := spec(t, srv.URL, sumOf(archive)+"  "+file+"\n")
	s.OS, s.Arch = runtime.GOOS, runtime.GOARCH
	o := options{VersionFile: vf, Spec: s, Out: io.Discard}
	t.Setenv("HELPER_TOFU_VERSION", "9.9.9")

	if err := run(context.Background(), "check", o); err == nil {
		t.Error("check must fail while nothing is installed")
	}
	if err := run(context.Background(), "ensure", o); err != nil {
		t.Fatalf("ensure: %v", err)
	}
	if err := run(context.Background(), "check", o); err != nil {
		t.Errorf("check after ensure: %v", err)
	}
	_ = os.WriteFile(vf, []byte("1.0.0\n"), 0o600) // the repository now pins another version
	err = run(context.Background(), "check", o)
	if err == nil || !strings.Contains(err.Error(), "pins 1.0.0") {
		t.Errorf("check with a mismatch: %v", err)
	}
	if err := run(context.Background(), "ensure", o); err == nil {
		t.Error("ensure must fail closed when the checksum file has no entry for the new version")
	}
}

func TestDownloadBounds(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(bytes.Repeat([]byte("x"), 2048))
	}))
	defer srv.Close()
	old := maxArchive
	maxArchive = 1024
	defer func() { maxArchive = old }()
	if _, err := download(context.Background(), nil, srv.URL); err == nil || !strings.Contains(err.Error(), "larger than") {
		t.Errorf("err = %v, want a size refusal", err)
	}
	if _, err := download(context.Background(), nil, "http://127.0.0.1:1/unreachable"); err == nil {
		t.Error("an unreachable server must fail")
	}
	if _, err := download(context.Background(), nil, "http://bad host/"); err == nil {
		t.Error("a malformed URL must fail")
	}
}

func TestExtractToUnwritableDestination(t *testing.T) {
	data := makeZip(t, "tofu", "x")
	blocker := filepath.Join(t.TempDir(), "file")
	_ = os.WriteFile(blocker, []byte("x"), 0o600)
	if err := extract(data, filepath.Join(blocker, "sub", "tofu")); err == nil {
		t.Error("extract under a regular file must fail")
	}
}

func TestExtractRefusesOversizedBinary(t *testing.T) {
	old := maxBinary
	maxBinary = 8
	defer func() { maxBinary = old }()
	if err := extract(makeZip(t, "tofu", "more than eight bytes"), filepath.Join(t.TempDir(), "tofu")); err == nil ||
		!strings.Contains(err.Error(), "larger than") {
		t.Errorf("err = %v, want a size refusal instead of a silent truncation", err)
	}
}
