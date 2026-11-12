// Package api is the HTTP surface of the service: routes and handlers that
// translate provider results into JSON. Errors are logged in full on the
// server and reduced to a fixed message for the client.
package api

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/PaulPerigault/GetUrlCloudRun/internal/httpmw"
	"github.com/PaulPerigault/GetUrlCloudRun/internal/provider"
)

// Server answers the HTTP API on top of a provider.
type Server struct {
	provider provider.Provider
	log      *slog.Logger
}

// New returns the router. It serves:
//
//	GET /health
//	GET /api/v1/services
//	GET /api/v1/services/{name}
func New(p provider.Provider, log *slog.Logger) http.Handler {
	s := &Server{provider: p, log: log}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.health)
	mux.HandleFunc("GET /api/v1/services", s.list)
	mux.HandleFunc("GET /api/v1/services/{name}", s.get)
	// Same paths without a method: any other method gets 405 (the method
	// patterns above are more specific, so GET never reaches these).
	for _, path := range []string{"/health", "/api/v1/services", "/api/v1/services/{name}"} {
		mux.HandleFunc(path, methodNotAllowed)
	}
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		httpmw.Error(w, http.StatusNotFound, "not found")
	})
	return mux
}

func methodNotAllowed(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Allow", http.MethodGet)
	httpmw.Error(w, http.StatusMethodNotAllowed, "method not allowed")
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "provider": s.provider.Name()})
}

func (s *Server) list(w http.ResponseWriter, r *http.Request) {
	services, err := s.provider.ListServices(r.Context())
	if err != nil {
		s.fail(w, r, err)
		return
	}
	if services == nil {
		services = []provider.Service{} // encode [] rather than null
	}
	writeJSON(w, http.StatusOK, services)
}

func (s *Server) get(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if !provider.ValidName(name) {
		httpmw.Error(w, http.StatusBadRequest, "invalid service name")
		return
	}
	svc, err := s.provider.GetService(r.Context(), name)
	if err != nil {
		s.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"url": svc.URL})
}

// fail logs err and answers with a generic message chosen by its kind.
func (s *Server) fail(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, provider.ErrNotFound):
		httpmw.Error(w, http.StatusNotFound, "service not found")
	case errors.Is(err, context.DeadlineExceeded):
		s.log.WarnContext(r.Context(), "upstream timeout", "path", r.URL.Path, "error", err)
		httpmw.Error(w, http.StatusGatewayTimeout, "upstream timeout")
	case errors.Is(err, context.Canceled):
		httpmw.Error(w, http.StatusServiceUnavailable, "request cancelled")
	default:
		s.log.ErrorContext(r.Context(), "provider error", "path", r.URL.Path, "error", err)
		httpmw.Error(w, http.StatusBadGateway, "upstream error")
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
