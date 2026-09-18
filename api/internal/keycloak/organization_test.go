package keycloak

import (
	"strings"
	"testing"
)

func TestDefaultOrganizationAttribute_usesSlugAndID(t *testing.T) {
	orgID := "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	attrs := DefaultOrganizationAttribute(orgID, `Acme Telecom "SP"`)
	raw := attrs["organization"][0]
	if !strings.Contains(raw, `"id":"`+orgID+`"`) {
		t.Fatalf("expected nested id, got %s", raw)
	}
	if !strings.Contains(raw, `"acme-telecom-sp"`) {
		t.Fatalf("expected slug alias, got %s", raw)
	}
	if attrs["organization_id"][0] != orgID {
		t.Fatalf("expected organization_id attribute, got %v", attrs["organization_id"])
	}
	if strings.Contains(raw, `""`) {
		t.Fatalf("unexpected empty quotes in %s", raw)
	}
}

func TestDefaultOrganizationAttribute_luxusAlias(t *testing.T) {
	attrs := DefaultOrganizationAttribute("00000000-0000-0000-0000-000000000001", "Luxus Connect")
	raw := attrs["organization"][0]
	if !strings.HasPrefix(raw, `{"luxus":`) {
		t.Fatalf("expected luxus alias, got %s", raw)
	}
}

func TestDefaultOrganizationAttribute_avoidsLuxusAliasForOtherTenants(t *testing.T) {
	orgID := "11111111-2222-3333-4444-555555555555"
	attrs := DefaultOrganizationAttribute(orgID, "Luxus Partner Co")
	raw := attrs["organization"][0]
	if strings.HasPrefix(raw, `{"luxus":`) || strings.HasPrefix(raw, `{"luxus-`) {
		t.Fatalf("must not use luxus* alias for other tenants: %s", raw)
	}
	if !strings.Contains(raw, `"id":"`+orgID+`"`) {
		t.Fatalf("expected nested id, got %s", raw)
	}
}
