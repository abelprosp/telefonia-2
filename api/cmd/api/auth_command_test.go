package main

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/luxus-connect/telefonia/api/internal/config"
	"github.com/luxus-connect/telefonia/api/internal/keycloak"
	"github.com/luxus-connect/telefonia/api/internal/localauth"
	"github.com/luxus-connect/telefonia/api/internal/store"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestLegacyImportPreservesIdentityTenantAndLocalCredentials(t *testing.T) {
	database := os.Getenv("AUTH_TEST_DATABASE_URL")
	if database == "" {
		t.Skip("set AUTH_TEST_DATABASE_URL to a disposable database")
	}
	ctx := context.Background()
	st, e := store.New(ctx, database)
	if e != nil {
		t.Fatal(e)
	}
	defer st.Close()
	id := uuid.NewString()
	org := uuid.NewString()
	name := "legacy-" + id
	defer st.Pool().Exec(ctx, `DELETE FROM auth_users WHERE id=$1`, id)
	missingOrg := true
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/realms/master/protocol/openid-connect/token":
			json.NewEncoder(w).Encode(map[string]any{"access_token": "test-admin-token", "expires_in": 300})
		case r.URL.Path == "/admin/realms/luxus/users":
			if r.URL.Query().Get("first") != "0" || r.URL.Query().Get("max") != "100" {
				t.Error("migration pagination missing")
			}
			attrs := map[string][]string{"document": {"52998224725"}}
			if !missingOrg {
				attrs["organization_id"] = []string{org}
			}
			json.NewEncoder(w).Encode([]keycloak.UserRecord{{ID: id, Username: name, Email: "legacy@example.com", Enabled: true, Attributes: attrs}})
		case r.URL.Path == "/admin/realms/luxus/clients":
			if r.URL.Query().Get("clientId") != "connect-cli" {
				t.Error("foreign application roles requested")
			}
			json.NewEncoder(w).Encode([]map[string]string{{"id": "connect-client", "clientId": "connect-cli"}})
		case strings.HasSuffix(r.URL.Path, "/role-mappings/realm"), strings.HasSuffix(r.URL.Path, "/role-mappings/realm/composite"):
			json.NewEncoder(w).Encode([]map[string]string{{"name": "user"}})
		case strings.HasSuffix(r.URL.Path, "/role-mappings/clients/connect-client/composite"):
			json.NewEncoder(w).Encode([]map[string]string{{"name": "sales"}})
		default:
			t.Errorf("unexpected migration request %s", r.URL)
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	cfg := config.Config{KeycloakAuthServerURL: server.URL, KeycloakRealm: "luxus", KeycloakResource: "connect-cli", KeycloakAdminUsername: "admin", KeycloakAdminPassword: "test-only"}
	if e := importLegacyAccounts(ctx, st, cfg); e == nil {
		t.Fatal("missing tenant silently assigned")
	}
	u, _ := st.LocalUserByID(ctx, id)
	if u != nil {
		t.Fatal("partial migration after missing tenant")
	}
	missingOrg = false
	if e := importLegacyAccounts(ctx, st, cfg); e != nil {
		t.Fatal(e)
	}
	u, e = st.LocalUserByID(ctx, id)
	if e != nil || u == nil || u.OrganizationID != org || u.Username != name || u.Document != "52998224725" || u.PasswordHash != "" {
		t.Fatal("identity changed or default password created", e)
	}
	if strings.Join(u.Roles, ",") != "user,sales" {
		t.Fatal("application roles lost", u.Roles)
	}
	hash, _ := localauth.HashPassword("new secure local password")
	u.PasswordHash = hash
	if e = st.UpdateLocalUser(ctx, u); e != nil {
		t.Fatal(e)
	}
	if e = importLegacyAccounts(ctx, st, cfg); e != nil {
		t.Fatal(e)
	}
	u, _ = st.LocalUserByID(ctx, id)
	if u.PasswordHash != hash {
		t.Fatal("repeat import overwrote local credentials")
	}
}
