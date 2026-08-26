package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/luxus-connect/telefonia/api/internal/auth"
	"github.com/luxus-connect/telefonia/api/internal/services"
)

func TestFinancialAgentAuth(t *testing.T) {
	h := &Handler{
		Svc:      &services.Service{},
		AgentKey: "secret-agent-key",
		AgentOrg: testOrgID,
	}
	r := chi.NewRouter()
	r.Route("/v1", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(auth.AgentAuthenticate(h.AgentKey, h.AgentOrg))
			r.Get("/agent/financial/health", h.agentFinancialHealth)
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/agent/financial/health", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without key, got %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/v1/agent/financial/health", nil)
	req.Header.Set("X-Agent-Key", "wrong")
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 with wrong key, got %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/v1/agent/financial/health", nil)
	req.Header.Set("X-Agent-Key", "secret-agent-key")
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 with valid key, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestFinancialAgentNotConfigured(t *testing.T) {
	r := chi.NewRouter()
	r.Use(auth.AgentAuthenticate("", ""))
	r.Get("/x", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("X-Agent-Key", "any")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}
}
