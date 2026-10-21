package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/luxus-connect/telefonia/api/internal/auth"
)

func approvalOK(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) }

func TestApprovalGuards(t *testing.T) {
	cases := []struct {
		role      string
		canRead   bool
		canDecide bool
	}{
		{auth.RoleMaster, true, true},
		{auth.RoleAdmin, true, true},
		{auth.RoleFinancial, true, true},
		{auth.RoleEmployee, true, false},
		{auth.RoleOperator, true, false},
		{auth.RoleSales, true, false},
		{auth.RoleViewer, false, false},
		{auth.RolePartner, false, false},
		{auth.RoleUser, false, false},
	}
	for _, tc := range cases {
		t.Run(tc.role, func(t *testing.T) {
			ctx := auth.WithUser(context.Background(), &auth.User{ID: "reviewer", Roles: []string{tc.role}})
			check := func(name string, guard func(http.Handler) http.Handler, allowed bool) {
				rec := httptest.NewRecorder()
				guard(http.HandlerFunc(approvalOK)).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/v1/approvals", nil).WithContext(ctx))
				want := http.StatusForbidden
				if allowed {
					want = http.StatusNoContent
				}
				if rec.Code != want {
					t.Errorf("%s: got %d, want %d", name, rec.Code, want)
				}
			}
			check("read", auth.ApprovalReadAccess, tc.canRead)
			check("decision", auth.ApprovalDecisionAccess, tc.canDecide)
		})
	}
}

// TestApprovalRoutesRegistered guards against a later r.Route("/approvals")
// mounting a subrouter and silently dropping earlier GET routes.
func TestApprovalRoutesRegistered(t *testing.T) {
	identity := func(next http.Handler) http.Handler { return next }
	r := chi.NewRouter()
	(&Handler{}).RegisterRoutes(r, identity, identity, identity, identity, identity, identity)

	found := map[string]bool{}
	if err := chi.Walk(r, func(method, route string, _ http.Handler, _ ...func(http.Handler) http.Handler) error {
		found[method+" "+route] = true
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	for _, route := range []string{
		"GET /v1/approvals",
		"GET /v1/approvals/",
		"POST /v1/approvals/{id}/approve",
		"POST /v1/approvals/{id}/reject",
		"POST /v1/approvals",
		"POST /v1/approvals/",
	} {
		if !found[route] {
			t.Errorf("route %s not registered", route)
		}
	}
}
