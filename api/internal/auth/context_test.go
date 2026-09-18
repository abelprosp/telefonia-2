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

func TestParseOrganizationFromClaims_directObject(t *testing.T) {
	claim := map[string]interface{}{"id": "11111111-1111-1111-1111-111111111111", "name": "Empresa"}
	org, err := ParseOrganizationFromClaims(claim)
	if err != nil || org == nil || org.ID != claim["id"] || org.Name != "Empresa" {
		t.Fatalf("unexpected org: %+v, %v", org, err)
	}
}

func TestParseOrganizationFromClaims_list(t *testing.T) {
	claim := []interface{}{map[string]interface{}{"id": "22222222-2222-2222-2222-222222222222", "name": []interface{}{"Empresa"}}}
	org, err := ParseOrganizationFromClaims(claim)
	if err != nil || org == nil || org.ID != "22222222-2222-2222-2222-222222222222" {
		t.Fatalf("unexpected org: %+v, %v", org, err)
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
	if org == nil || org.ID != "" {
		t.Fatalf("expected no luxus remap for luxus-connect alias, got %+v", org)
	}
}

func TestNormalizeOrganization_legacyLuxusID(t *testing.T) {
	org := normalizeOrganization(&Organization{
		ID:    "luxus",
		Alias: "luxus",
		Name:  "Luxus Telefonia",
	})
	if org == nil || org.ID != DefaultLuxusOrganizationID {
		t.Fatalf("expected luxus id remap, got %+v", org)
	}
}
