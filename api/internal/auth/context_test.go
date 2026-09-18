package auth

import (
	"testing"
)

func TestParseOrganizationFromClaims_string(t *testing.T) {
	raw := `{"luxus":{"id":"00000000-0000-0000-0000-000000000001","name":["Luxus Connect"]}}`
	org, err := ParseOrganizationFromClaims(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if org == nil || org.ID != "00000000-0000-0000-0000-000000000001" || org.Alias != "luxus" {
		t.Fatalf("unexpected org: %+v", org)
	}
}

func TestParseOrganizationFromClaims_map(t *testing.T) {
	claim := map[string]interface{}{
		"luxus": map[string]interface{}{
			"id":   "00000000-0000-0000-0000-000000000001",
			"name": []interface{}{"Luxus Connect"},
		},
	}
	org, err := ParseOrganizationFromClaims(claim)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if org == nil || org.Name != "Luxus Connect" {
		t.Fatalf("unexpected org: %+v", org)
	}
}

func TestParseOrganizationFromClaims_mapWithoutID(t *testing.T) {
	claim := map[string]interface{}{
		"luxus": map[string]interface{}{
			"name": "Luxus Connect",
		},
	}
	org, err := ParseOrganizationFromClaims(claim)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	org = normalizeOrganization(org)
	if org == nil || org.ID != DefaultLuxusOrganizationID {
		t.Fatalf("expected luxus default id, got %+v", org)
	}
}

func TestNormalizeOrganization_uuidAlias(t *testing.T) {
	org := normalizeOrganization(&Organization{
		Alias: "00000000-0000-0000-0000-000000000001",
		Name:  "Acme",
	})
	if org == nil || org.ID != "00000000-0000-0000-0000-000000000001" {
		t.Fatalf("expected uuid alias as id, got %+v", org)
	}
}

func TestNormalizeOrganization_luxusConnectAlias(t *testing.T) {
	org := normalizeOrganization(&Organization{
		Alias: "luxus-connect",
		Name:  "Telefonia",
	})
	if org == nil || org.ID != DefaultLuxusOrganizationID {
		t.Fatalf("expected luxus default id, got %+v", org)
	}
}
