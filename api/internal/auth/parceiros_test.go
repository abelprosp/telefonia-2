package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParceirosAuthenticate(t *testing.T) {
	ok := ParceirosAuthenticate("", "org-1")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("não deveria chamar o handler")
	}))
	rec := httptest.NewRecorder()
	ok.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/v1/integrations/luxus-parceiros/sales", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("sem configuração = %d", rec.Code)
	}

	called := false
	h := ParceirosAuthenticate("secret", "org-1")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if OrganizationFromContext(r.Context()).ID != "org-1" {
			t.Fatal("organização")
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/v1/integrations/luxus-parceiros/sales", nil))
	if rec.Code != http.StatusUnauthorized || called {
		t.Fatalf("sem chave = %d called %v", rec.Code, called)
	}

	req := httptest.NewRequest(http.MethodPost, "/v1/integrations/luxus-parceiros/sales", nil)
	req.Header.Set(ParceirosKeyHeader, "secret")
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent || !called {
		t.Fatalf("chave certa = %d called %v", rec.Code, called)
	}
}
