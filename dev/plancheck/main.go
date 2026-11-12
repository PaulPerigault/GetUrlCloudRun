// Command plancheck refuses an OpenTofu plan that destroys or replaces a
// resource which is not listed in infra/allowed-destroys.txt.
//
// The owner produces the plan JSON by hand, outside CI (this repository never
// plans or applies from a pipeline):
//
//	tofu -chdir=infra plan -out=plan.bin
//	tofu -chdir=infra show -json plan.bin > plan.json
//	task infra:plan:check PLAN=infra/plan.json
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"slices"
	"sort"
	"strings"
)

const allowedFile = "infra/allowed-destroys.txt"

type plan struct {
	ResourceChanges []struct {
		Address string `json:"address"`
		Change  struct {
			Actions []string `json:"actions"`
		} `json:"change"`
	} `json:"resource_changes"`
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: plancheck <plan.json>")
		os.Exit(2)
	}
	if err := run(os.Args[1], allowedFile, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "plancheck:", err)
		os.Exit(1)
	}
}

func run(planPath, allowedPath string, out io.Writer) error {
	pf, err := os.Open(planPath)
	if err != nil {
		return fmt.Errorf("open plan: %w", err)
	}
	defer func() { _ = pf.Close() }()
	af, err := os.Open(allowedPath)
	if err != nil {
		return fmt.Errorf("open allowed destroys: %w", err)
	}
	defer func() { _ = af.Close() }()
	destroyed, err := destroys(pf)
	if err != nil {
		return err
	}
	allowed := parseAllowed(af)
	var refused []string
	for _, addr := range destroyed {
		if !slices.Contains(allowed, addr) {
			refused = append(refused, addr)
		}
	}
	if len(refused) > 0 {
		return fmt.Errorf("the plan destroys or replaces %s, which %s does not allow", strings.Join(refused, ", "), allowedPath)
	}
	_, err = fmt.Fprintf(out, "plancheck: %d destroy action(s), all allowed\n", len(destroyed))
	return err
}

// destroys returns the addresses whose actions include delete (a replacement
// is delete + create), sorted.
func destroys(r io.Reader) ([]string, error) {
	var p plan
	if err := json.NewDecoder(r).Decode(&p); err != nil {
		return nil, fmt.Errorf("decode plan: %w", err)
	}
	var out []string
	for _, rc := range p.ResourceChanges {
		if slices.Contains(rc.Change.Actions, "delete") {
			out = append(out, rc.Address)
		}
	}
	sort.Strings(out)
	return out, nil
}

func parseAllowed(r io.Reader) []string {
	var out []string
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			out = append(out, line)
		}
	}
	return out
}
