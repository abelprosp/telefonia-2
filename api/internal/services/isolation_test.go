package services

import (
	"context"
	"testing"

	"github.com/luxus-connect/telefonia/api/internal/auth"
	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/keycloak"
	"github.com/luxus-connect/telefonia/api/internal/notifications"
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

func TestToListUser_readsOrganizationAttributes(t *testing.T) {
	orgID := "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	u := keycloak.UserRecord{
		ID:       "user-1",
		Username: "acme",
		Enabled:  true,
		Roles:    []string{auth.RoleMaster, auth.RoleUser},
		Attributes: map[string][]string{
			"organization": {
				`{"acme":{"id":"` + orgID + `","name":["Acme Telecom"]}}`,
			},
			"organization_id": {orgID},
		},
	}
	item := toListUser(u)
	if item.OrganizationID != orgID {
		t.Fatalf("expected org id %s, got %s", orgID, item.OrganizationID)
	}
	if item.OrganizationName != "Acme Telecom" {
		t.Fatalf("expected name Acme Telecom, got %q", item.OrganizationName)
	}
}

func TestToListUser_fallsBackToOrganizationIDAttribute(t *testing.T) {
	orgID := "bbbbbbbb-bbbb-cccc-dddd-eeeeeeeeeeee"
	u := keycloak.UserRecord{
		ID:       "user-2",
		Username: "solo",
		Enabled:  true,
		Attributes: map[string][]string{
			"organization_id": {orgID},
		},
	}
	item := toListUser(u)
	if item.OrganizationID != orgID {
		t.Fatalf("expected org id fallback %s, got %s", orgID, item.OrganizationID)
	}
}
