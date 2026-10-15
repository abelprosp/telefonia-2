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
	bulk := false
	unassignedID, foreignID, foreignOrg := uuid.NewString(), uuid.NewString(), uuid.NewString()
	for _, testID := range []string{unassignedID, foreignID} {
		defer st.Pool().Exec(ctx, `DELETE FROM auth_users WHERE id=$1`, testID)
	}
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
			records := []keycloak.UserRecord{{ID: id, Username: name, Email: "legacy@example.com", Enabled: true, Attributes: attrs}}
			if bulk {
				records = append(records, keycloak.UserRecord{ID: unassignedID, Username: "unassigned-" + unassignedID, Email: "unassigned@example.com", Enabled: true}, keycloak.UserRecord{ID: foreignID, Username: "foreign-" + foreignID, Email: "foreign@example.com", Enabled: true, Attributes: map[string][]string{"organization_id": {foreignOrg}}})
			}
			json.NewEncoder(w).Encode(records)
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
	if e := importLegacyAccounts(ctx, st, cfg, name+"="+org); e != nil {
		t.Fatal(e)
	}
	missingOrg = false
	if e := importLegacyAccounts(ctx, st, cfg, name+"="+uuid.NewString()); e == nil {
		t.Fatal("explicit assignment overwrote an existing tenant")
	}
	if e := importLegacyAccounts(ctx, st, cfg, "misspelled-user="+org); e == nil {
		t.Fatal("unknown explicit assignment ignored")
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
	bulk = true
	fallback := uuid.NewString()
	if e = importLegacyAccountsWithDefault(ctx, st, cfg, fallback); e != nil {
		t.Fatal(e)
	}
	unassigned, _ := st.LocalUserByID(ctx, unassignedID)
	foreign, _ := st.LocalUserByID(ctx, foreignID)
	if unassigned == nil || unassigned.OrganizationID != fallback {
		t.Fatal("explicit batch destination not applied")
	}
	if foreign == nil || foreign.OrganizationID != foreignOrg {
		t.Fatal("batch migration reassigned existing tenant")
	}
	u, _ = st.LocalUserByID(ctx, id)
	if u.OrganizationID != org || u.PasswordHash != hash {
		t.Fatal("batch migration overwrote existing local identity")
	}
}

func TestParseOrganizationAssignmentsRejectsAmbiguity(t *testing.T) {
	org := uuid.NewString()
	for _, values := range [][]string{{"arturabel"}, {"= " + org}, {"arturabel=invalid"}, {"arturabel=" + uuid.Nil.String()}, {"arturabel=" + org, "arturabel=" + org}} {
		if _, e := parseOrganizationAssignments(values); e == nil {
			t.Fatalf("invalid assignment accepted: %v", values)
		}
	}
	mappings, e := parseOrganizationAssignments([]string{"arturabel=" + org})
	if e != nil || mappings["arturabel"] != org {
		t.Fatal("valid explicit assignment rejected")
	}
}

func TestBatchOrganizationRejectsInvalidDestination(t *testing.T) {
	for _, id := range []string{"invalid", uuid.Nil.String()} {
		if e := importLegacyAccountsWithDefault(context.Background(), nil, config.Config{}, id); e == nil || !strings.Contains(e.Error(), "UUID") {
			t.Fatal("invalid destination was not rejected before importing")
		}
	}
}
