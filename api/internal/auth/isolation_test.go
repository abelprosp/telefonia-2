package auth

import (
	"context"
	"testing"
)

func TestNormalizeOrganization_rejectsUnknownAliasWithoutID(t *testing.T) {
	org := NormalizeOrganization(&Organization{Alias: "acme-telecom", Name: "Acme"})
	if org == nil || org.ID != "" {
		t.Fatalf("expected empty id for unknown alias, got %+v", org)
	}
}

func TestOrganizationFromContext_emptyWithoutSet(t *testing.T) {
	if OrganizationFromContext(context.Background()) != nil {
		t.Fatal("expected nil org")
	}
}
