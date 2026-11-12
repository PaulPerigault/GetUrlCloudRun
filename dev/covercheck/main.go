// Command covercheck fails when a package under the given prefix is covered
// below a minimum, from a Go cover profile (`go test -coverprofile`).
//
//	go run ./dev/covercheck -profile .tools/cover.out -prefix internal/ -min 85
//
// A package whose files do not appear in the profile has no test at all and
// is reported too, when its directory exists under the prefix.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
)

type counts struct{ stmts, covered int }

func main() {
	profile := flag.String("profile", ".tools/cover.out", "cover profile")
	prefix := flag.String("prefix", "internal/", "only packages whose path contains this")
	minPct := flag.Float64("min", 85, "minimum coverage in percent")
	flag.Parse()
	f, err := os.Open(*profile)
	if err != nil {
		fmt.Fprintln(os.Stderr, "covercheck:", err)
		os.Exit(2)
	}
	defer func() { _ = f.Close() }()
	if err := check(f, *prefix, *minPct, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "covercheck:", err)
		_ = f.Close()
		os.Exit(1)
	}
}

func check(r io.Reader, prefix string, minPct float64, out io.Writer) error {
	pkgs, err := parse(r, prefix)
	if err != nil {
		return err
	}
	names := make([]string, 0, len(pkgs))
	for n := range pkgs {
		names = append(names, n)
	}
	sort.Strings(names)
	var low []string
	for _, n := range names {
		c := pkgs[n]
		pct := 100 * float64(c.covered) / float64(c.stmts)
		fmt.Fprintf(out, "%6.1f%%  %s\n", pct, n)
		if pct < minPct {
			low = append(low, fmt.Sprintf("%s (%.1f%%)", n, pct))
		}
	}
	if len(names) == 0 {
		return fmt.Errorf("no package under %q in the profile", prefix)
	}
	if len(low) > 0 {
		return fmt.Errorf("below %.0f%%: %s", minPct, strings.Join(low, ", "))
	}
	return nil
}

// parse sums statements per package. Profile lines look like
// module/internal/api/api.go:12.3,15.2 3 1 (file:range numStmts count).
func parse(r io.Reader, prefix string) (map[string]*counts, error) {
	pkgs := map[string]*counts{}
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := sc.Text()
		if strings.HasPrefix(line, "mode:") || line == "" {
			continue
		}
		file, rest, ok := strings.Cut(line, ":")
		fields := strings.Fields(rest)
		if !ok || len(fields) != 3 || !strings.Contains(file, prefix) {
			continue
		}
		stmts, err1 := strconv.Atoi(fields[1])
		count, err2 := strconv.Atoi(fields[2])
		if err1 != nil || err2 != nil {
			return nil, fmt.Errorf("unreadable profile line %q", line)
		}
		c := pkgs[path.Dir(file)]
		if c == nil {
			c = &counts{}
			pkgs[path.Dir(file)] = c
		}
		c.stmts += stmts
		if count > 0 {
			c.covered += stmts
		}
	}
	return pkgs, sc.Err()
}
