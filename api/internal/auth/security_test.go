package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/luxus-connect/telefonia/api/internal/config"
)

func TestClientRolesCannotGrantPrivilegesAcrossApplications(t *testing.T) {
	claims := jwt.MapClaims{"resource_access": map[string]interface{}{
		"other-app":   map[string]interface{}{"roles": []interface{}{"master", "admin"}},
		"connect-cli": map[string]interface{}{"roles": []interface{}{"sales"}},
	}}
	roles := extractRoles(claims, "connect-cli")
	if len(roles) != 1 || roles[0] != "sales" {
		t.Fatalf("unexpected roles: %v", roles)
	}
}

func TestCommercialWritesCannotAdministerOperations(t *testing.T) {
	for _, test := range []struct {
		method, path string
		allowed      bool
	}{
		{"POST", "/v1/customers", true}, {"PATCH", "/v1/customers/id", true},
		{"POST", "/v1/sales/id/confirm", true}, {"POST", "/v1/sales/id/mark-paid", false},
		{"DELETE", "/v1/customers/id", false}, {"POST", "/v1/providers", false},
		{"POST", "/v1/phone-lines/stock", false}, {"POST", "/v1/processing-months/id/close", false},
		{"POST", "/v1/contract-templates", false}, {"GET", "/v1/providers", true},
	} {
		req := httptest.NewRequest(test.method, test.path, nil)
		ctx := WithUser(context.Background(), &User{Roles: []string{RoleSales}})
		rec := httptest.NewRecorder()
		OperationalMutationGuard(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(204) })).ServeHTTP(rec, req.WithContext(ctx))
		if (rec.Code == 204) != test.allowed {
			t.Errorf("%s %s returned %d", test.method, test.path, rec.Code)
		}
	}
}

func TestProductionPrivilegesRequireVerifiedMFA(t *testing.T) {
	mw := &Middleware{cfg: config.Config{Environment: "Production"}}
	for _, amr := range [][]string{nil, {"pwd"}, {"pwd", "otp"}} {
		ctx := WithUser(context.Background(), &User{Roles: []string{RoleMaster}, Amr: amr})
		rec := httptest.NewRecorder()
		mw.RequireMaster(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(204) })).ServeHTTP(rec, httptest.NewRequest("GET", "/", nil).WithContext(ctx))
		if (rec.Code == 204) != (len(amr) == 2) {
			t.Errorf("amr=%v status=%d", amr, rec.Code)
		}
	}
}
