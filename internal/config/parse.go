package config

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// parser accumulates errors so that Load can report all of them at once.
type parser struct {
	getenv func(string) string
	errs   []error
}

func (p *parser) str(key, def string) string {
	if v := strings.TrimSpace(p.getenv(key)); v != "" {
		return v
	}
	return def
}

func (p *parser) oneOf(key, def string, allowed ...string) string {
	v := p.str(key, def)
	for _, a := range allowed {
		if v == a {
			return v
		}
	}
	p.errs = append(p.errs, fmt.Errorf("%s %q must be one of %v", key, v, allowed))
	return def
}

func (p *parser) intIn(key string, def, lo, hi int) int {
	raw := p.str(key, "")
	if raw == "" {
		return def
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < lo || n > hi {
		p.errs = append(p.errs, fmt.Errorf("%s %q must be an integer between %d and %d", key, raw, lo, hi))
		return def
	}
	return n
}

func (p *parser) float(key string, def float64) float64 {
	raw := p.str(key, "")
	if raw == "" {
		return def
	}
	f, err := strconv.ParseFloat(raw, 64)
	if err != nil || f <= 0 || f > 1e6 {
		p.errs = append(p.errs, fmt.Errorf("%s %q must be a number above 0 and at most 1000000", key, raw))
		return def
	}
	return f
}

func (p *parser) duration(key string, def time.Duration) time.Duration {
	raw := p.str(key, "")
	if raw == "" {
		return def
	}
	d, err := time.ParseDuration(raw)
	if err != nil || d <= 0 || d > time.Minute {
		p.errs = append(p.errs, fmt.Errorf("%s %q must be a duration above 0 and at most 1m", key, raw))
		return def
	}
	return d
}

// origins parses a comma-separated list of exact origins (scheme://host[:port]).
// The wildcard is refused on purpose: CORS is deny-by-default.
func (p *parser) origins(key string) []string {
	var out []string
	for _, raw := range strings.Split(p.getenv(key), ",") {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		if !isOrigin(raw) {
			p.errs = append(p.errs, fmt.Errorf("%s entry %q must be an origin like https://app.example.com", key, raw))
			continue
		}
		out = append(out, strings.ToLower(raw))
	}
	return out
}

func isOrigin(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" || u.User != nil {
		return false
	}
	schemeOK := u.Scheme == "http" || u.Scheme == "https"
	return schemeOK && u.Path == "" && u.RawQuery == "" && u.Fragment == ""
}
