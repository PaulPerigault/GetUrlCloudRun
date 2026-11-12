package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// maxArchive bounds the download (the real archives are about 30 MB) and
// maxBinary bounds the extracted file (the real binary is about 150 MB).
var (
	maxArchive = 100 << 20
	maxBinary  = 512 << 20
)

type installSpec struct {
	Base     string // release download base URL
	Version  string
	OS, Arch string
	SumsFile string
	BinDir   string
	Client   *http.Client // nil means http.DefaultClient
}

func readVersion(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read pinned version: %w", err)
	}
	v := strings.TrimSpace(string(data))
	if v == "" || strings.ContainsAny(v, " /\\\n") {
		return "", fmt.Errorf("%s: %q is not a version", path, v)
	}
	return v, nil
}

func binPath(dir, goos string) string {
	name := "tofu"
	if goos == "windows" {
		name += ".exe"
	}
	return filepath.Join(dir, name)
}

// parseSums reads "<sha256>  <file>" lines.
func parseSums(r io.Reader) map[string]string {
	sums := map[string]string{}
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		if f := strings.Fields(sc.Text()); len(f) == 2 {
			sums[f[1]] = strings.ToLower(f[0])
		}
	}
	return sums
}

func install(ctx context.Context, s installSpec) error {
	archive := fmt.Sprintf("tofu_%s_%s_%s.zip", s.Version, s.OS, s.Arch)
	f, err := os.Open(s.SumsFile)
	if err != nil {
		return fmt.Errorf("open checksums: %w", err)
	}
	defer func() { _ = f.Close() }()
	want, ok := parseSums(f)[archive]
	if !ok {
		return fmt.Errorf("%s has no checksum for %s: refusing to download without a fingerprint", s.SumsFile, archive)
	}
	data, err := download(ctx, s.Client, fmt.Sprintf("%s/v%s/%s", s.Base, s.Version, archive))
	if err != nil {
		return err
	}
	if sum := sha256.Sum256(data); hex.EncodeToString(sum[:]) != want {
		return fmt.Errorf("checksum mismatch for %s: refusing to install", archive)
	}
	return extract(data, binPath(s.BinDir, s.OS))
}

func download(ctx context.Context, c *http.Client, url string) ([]byte, error) {
	if c == nil {
		c = http.DefaultClient
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, err
	}
	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("download %s: %w", url, err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download %s: status %d", url, resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, int64(maxArchive)+1))
	if err != nil {
		return nil, fmt.Errorf("download %s: %w", url, err)
	}
	if len(data) > maxArchive {
		return nil, fmt.Errorf("download %s: larger than %d bytes", url, maxArchive)
	}
	return data, nil
}

// extract writes the single binary of the archive to dest, whatever its
// directory in the archive, and ignores every other entry.
func extract(data []byte, dest string) error {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return fmt.Errorf("open archive: %w", err)
	}
	base := filepath.Base(dest)
	for _, zf := range zr.File {
		if zf.Name != base {
			continue
		}
		rc, err := zf.Open()
		if err != nil {
			return err
		}
		defer func() { _ = rc.Close() }()
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return err
		}
		out, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
		if err != nil {
			return err
		}
		n, cpErr := io.Copy(out, io.LimitReader(rc, int64(maxBinary)+1))
		if cpErr == nil && n > int64(maxBinary) {
			cpErr = fmt.Errorf("%s is larger than %d bytes", base, maxBinary)
		}
		return errors.Join(cpErr, out.Close())
	}
	return fmt.Errorf("archive has no %s", base)
}

// installedVersion asks the binary itself (`tofu version -json`).
func installedVersion(bin string) (string, error) {
	out, err := exec.Command(bin, "version", "-json").Output()
	if err != nil {
		return "", fmt.Errorf("run %s: %w", bin, err)
	}
	var v struct {
		Version string `json:"terraform_version"`
	}
	if err := json.Unmarshal(out, &v); err != nil || v.Version == "" {
		return "", fmt.Errorf("%s: unreadable version output", bin)
	}
	return v.Version, nil
}
