package handlers

import (
	"net/http"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/luxus-connect/telefonia/api/internal/services"
)

// TestReq08RoutesRegistered verifies REQ08 surface routes exist without needing Postgres.
func TestReq08RoutesRegistered(t *testing.T) {
	h := &Handler{Svc: &services.Service{}, Presigned: &services.PresignedService{}}
	r := chi.NewRouter()
	h.RegisterRoutes(r, testAuth, testAuth, testAuth, testAuth, testAuth, testAuth)

	want := []string{
		"/v1/reports/line-consumption",
		"/v1/reports/line-consumption/export",
		"/v1/audit/events",
		"/v1/reconciliation/missing-in-operator",
		"/v1/reconciliation/cancelled-externally-active",
		"/v1/reconciliation/cancelled-externally-active/apply",
	}
	found := map[string]bool{}
	_ = chi.Walk(r, func(method, route string, _ http.Handler, _ ...func(http.Handler) http.Handler) error {
		path := strings.TrimSuffix(route, "/")
		for _, w := range want {
			if path == w || route == w {
				found[w] = true
			}
		}
		return nil
	})
	for _, w := range want {
		if !found[w] {
			t.Errorf("missing route %s", w)
		}
	}
}
