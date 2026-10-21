package main

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/luxus-connect/telefonia/api/internal/localauth"
	"github.com/luxus-connect/telefonia/api/internal/store"
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
	file := t.TempDir() + "/users.json"
	writeUsers := func() {
		attrs := map[string][]string{"document": {"52998224725"}}
		if !missingOrg {
			attrs["organization_id"] = []string{org}
		}
		records := []importedAccount{{ID: id, Username: name, Email: "legacy@example.com", Enabled: true, Attributes: attrs, ClientRoles: map[string][]string{"connect-cli": {"sales"}, "foreign-client": {"master"}}}}
		if bulk {
			records = append(records, importedAccount{ID: unassignedID, Username: "unassigned-" + unassignedID, Enabled: true}, importedAccount{ID: foreignID, Username: "foreign-" + foreignID, Enabled: true, OrganizationID: foreignOrg})
		}
		raw, err := json.Marshal(map[string]any{"users": records})
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(file, raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
	writeUsers()
	if e := importAccountsFile(ctx, st, file, ""); e == nil {
		t.Fatal("missing tenant silently assigned")
	}
	u, _ := st.LocalUserByID(ctx, id)
	if u != nil {
		t.Fatal("partial migration after missing tenant")
	}
	if e := importAccountsFile(ctx, st, file, "", name+"="+org); e != nil {
		t.Fatal(e)
	}
	missingOrg = false
	writeUsers()
	if e := importAccountsFile(ctx, st, file, "", name+"="+uuid.NewString()); e == nil {
		t.Fatal("explicit assignment overwrote an existing tenant")
	}
	if e := importAccountsFile(ctx, st, file, "", "misspelled-user="+org); e == nil {
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
	if e = importAccountsFile(ctx, st, file, ""); e != nil {
		t.Fatal(e)
	}
	u, _ = st.LocalUserByID(ctx, id)
	if u.PasswordHash != hash {
		t.Fatal("repeat import overwrote local credentials")
	}
	bulk = true
	writeUsers()
	fallback := uuid.NewString()
	if e = importAccountsFile(ctx, st, file, fallback); e != nil {
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
		if _, e := readImportedAccounts(strings.NewReader("[]"), id); e == nil || !strings.Contains(e.Error(), "UUID") {
			t.Fatal("invalid destination was not rejected before importing")
		}
	}
}

func TestOfflineImportAcceptsExportAndKeepsApplicationRoles(t *testing.T) {
	id, org := uuid.NewString(), uuid.NewString()
	record := importedAccount{ID: id, Username: "existing", Enabled: true, Attributes: map[string][]string{
		"organization": {`{"company":{"id":"` + org + `","name":["Company"]}}`},
		"document":     {"52998224725"},
	}, RealmRoles: []string{"user"}, ClientRoles: map[string][]string{"connect-cli": {"sales", "sales"}, "unrelated": {"master"}}}
	for _, value := range []any{[]importedAccount{record}, map[string]any{"users": []importedAccount{record}}} {
		data, _ := json.Marshal(value)
		users, err := readImportedAccounts(strings.NewReader(string(data)), "")
		if err != nil {
			t.Fatal(err)
		}
		u := users[0]
		if u.ID != id || u.OrganizationID != org || u.OrganizationName != "Company" || u.Document != "52998224725" || strings.Join(u.Roles, ",") != "user,sales" {
			t.Fatalf("identity, organization or roles changed: %+v", u)
		}
		if u.PasswordHash != "" || u.PasskeyCount != 0 {
			t.Fatal("import assigned credentials")
		}
	}
}

func TestOfflineImportValidatesEntireFileBeforeWriting(t *testing.T) {
	id, org := uuid.NewString(), uuid.NewString()
	valid := `{"id":"` + id + `","username":"one","organizationId":"` + org + `","enabled":true}`
	for _, raw := range []string{
		``, `[]`, `{}`, `null`, `[` + valid + `] {}`,
		`[` + valid + `,` + valid + `]`,
		`[{"id":"invalid","username":"one","organizationId":"` + org + `"}]`,
		`[{"id":"` + id + `","username":"with space","organizationId":"` + org + `"}]`,
		`[{"id":"` + id + `","username":"one"}]`,
		`[{"id":"` + id + `","username":"one","attributes":{"organization":["{\"a\":{\"id\":\"` + org + `\"},\"b\":{\"id\":\"` + uuid.NewString() + `\"}}"]}}]`,
	} {
		if _, err := readImportedAccounts(strings.NewReader(raw), ""); err == nil {
			t.Fatalf("invalid import accepted: %s", raw)
		}
	}
	if _, err := readImportedAccounts(strings.NewReader(`[`+valid+`]`), "", "one="+uuid.NewString()); err == nil {
		t.Fatal("explicit assignment replaced an existing tenant")
	}
	if _, err := readImportedAccounts(strings.NewReader(`[`+valid+`]`), "", "missing="+org); err == nil {
		t.Fatal("unknown explicit assignment accepted")
	}
	// A shell-selected default applies only to accounts with no tenant.
	other := uuid.NewString()
	users, err := readImportedAccounts(strings.NewReader(`[`+valid+`,{"id":"`+uuid.NewString()+`","username":"two"}]`), other)
	if err != nil || users[0].OrganizationID != org || users[1].OrganizationID != other {
		t.Fatal("batch assignment changed a known tenant", err)
	}
}
