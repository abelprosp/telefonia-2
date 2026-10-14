package auth

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestIntakeKeyPinsTenantWithoutGrantingMaster(t *testing.T) {
	key := strings.Repeat("a", 64)
	for _, test := range []struct {
		key, org, provided string
		status             int
	}{{key, "tenant", "bad", 401}, {"", "tenant", key, 503}, {key, "", key, 503}, {key, "tenant", key, 204}} {
		req := httptest.NewRequest("POST", "/v1/intake/customers", nil)
		req.Header.Set("X-API-Key", test.provided)
		rec := httptest.NewRecorder()
		IntakeAuthenticate(test.key, test.org)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if OrganizationFromContext(r.Context()).ID != "tenant" || IsMaster(r.Context()) || CanAccessFinancial(r.Context()) {
				t.Fatal("wrong tenant or excessive permission")
			}
			w.WriteHeader(204)
		})).ServeHTTP(rec, req)
		if rec.Code != test.status {
			t.Fatalf("got %d want %d", rec.Code, test.status)
		}
	}
}
