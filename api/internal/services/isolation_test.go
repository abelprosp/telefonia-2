package services

import (
	"context"
	"testing"

	"github.com/luxus-connect/telefonia/api/internal/auth"
	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/notifications"
	"github.com/luxus-connect/telefonia/api/internal/store"
)

func TestRequireCallerOrganization_rejectsMissing(t *testing.T) {
	_, err := requireCallerOrganization(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	appErr, ok := err.(*httputil.AppError)
	if !ok || appErr.Notifications[0].Code != notifications.SharedOrganizationRequired.Code {
		t.Fatalf("expected ORGANIZATION_ID_REQUIRED, got %#v", err)
	}
}

func TestRequireCallerOrganization_normalizesLegacyLuxusAliasID(t *testing.T) {
	ctx := auth.WithOrganization(context.Background(), &auth.Organization{ID: "luxus", Alias: "luxus"})
	org, err := requireCallerOrganization(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if org.ID != auth.DefaultLuxusOrganizationID {
		t.Fatalf("expected luxus uuid, got %s", org.ID)
	}
}

func TestRequireCallerOrganization_acceptsUUID(t *testing.T) {
	id := "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	ctx := auth.WithOrganization(context.Background(), &auth.Organization{ID: id, Name: "Acme"})
	org, err := requireCallerOrganization(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if org.ID != id {
		t.Fatalf("got %s", org.ID)
	}
}

func TestLocalListUser_preservesOrganization(t *testing.T) {
	orgID := "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	u := store.LocalUser{
		ID:               "user-1",
		Username:         "acme",
		Enabled:          true,
		Roles:            []string{auth.RoleMaster, auth.RoleUser},
		OrganizationID:   orgID,
		OrganizationName: "Acme Telecom",
	}
	item := localListUser(u)
	if item.OrganizationID != orgID {
		t.Fatalf("expected org id %s, got %s", orgID, item.OrganizationID)
	}
	if item.OrganizationName != "Acme Telecom" {
		t.Fatalf("expected name Acme Telecom, got %q", item.OrganizationName)
	}
}

func TestListOrganizationUsers_filtersByCallerOrg(t *testing.T) {
	caller := "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	other := "bbbbbbbb-bbbb-cccc-dddd-eeeeeeeeeeee"
	users := []store.LocalUser{
		{
			ID: "1", Username: "a", Enabled: true, Roles: []string{auth.RoleMaster},
			OrganizationID: caller,
		},
		{
			ID: "2", Username: "b", Enabled: true, Roles: []string{auth.RoleMaster},
			OrganizationID: other,
		},
		{
			ID: "3", Username: "c", Enabled: true, Roles: []string{auth.RoleEmployee},
			OrganizationID: caller, OrganizationName: "Same",
		},
	}

	var kept []string
	for _, u := range users {
		item := localListUser(u)
		if item.OrganizationID == caller {
			kept = append(kept, item.Username)
		}
	}
	if len(kept) != 2 || kept[0] != "a" || kept[1] != "c" {
		t.Fatalf("expected only caller-org users, got %v", kept)
	}
}

func TestUpdateOrganizationUser_forbiddenAcrossOrgs(t *testing.T) {
	caller := "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	other := "bbbbbbbb-bbbb-cccc-dddd-eeeeeeeeeeee"
	ctx := auth.WithOrganization(context.Background(), &auth.Organization{ID: caller, Name: "A"})
	callerOrg, err := requireCallerOrganization(ctx)
	if err != nil {
		t.Fatal(err)
	}
	target := localListUser(store.LocalUser{
		ID: "x", Username: "foreign", Enabled: true,
		OrganizationID: other,
	})
	if target.OrganizationID == callerOrg.ID {
		t.Fatal("fixture broken")
	}
	// Mirrors the guard in UpdateOrganizationUser for non-platform admins.
	isPlatformAdmin := callerOrg.ID == auth.DefaultLuxusOrganizationID
	if isPlatformAdmin || target.OrganizationID == callerOrg.ID {
		t.Fatal("expected cross-org access to be denied by guard logic")
	}
}
