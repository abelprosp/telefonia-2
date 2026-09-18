package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

type contextKey int

const (
	userKey contextKey = iota
	orgKey
)

type User struct {
	ID       string
	Email    string
	Name     string
	Username string
	Roles    []string
	Document string
	Acr      string
	Amr      []string
}

type Organization struct {
	ID    string
	Name  string
	Alias string
}

func WithUser(ctx context.Context, u *User) context.Context {
	return context.WithValue(ctx, userKey, u)
}

func WithOrganization(ctx context.Context, o *Organization) context.Context {
	return context.WithValue(ctx, orgKey, o)
}

func UserFromContext(ctx context.Context) *User {
	u, _ := ctx.Value(userKey).(*User)
	return u
}

func OrganizationFromContext(ctx context.Context) *Organization {
	o, _ := ctx.Value(orgKey).(*Organization)
	return o
}

func HasRole(ctx context.Context, role string) bool {
	u := UserFromContext(ctx)
	if u == nil {
		return false
	}
	for _, r := range u.Roles {
		if r == role {
			return true
		}
	}
	return false
}

func IsPartner(ctx context.Context) bool {
	return HasRole(ctx, RolePartner)
}

func IsAdmin(ctx context.Context) bool {
	return IsMaster(ctx)
}

type organizationClaimEntry struct {
	ID   string   `json:"id"`
	Name []string `json:"-"`
}

func (e *organizationClaimEntry) UnmarshalJSON(data []byte) error {
	var raw struct {
		ID   string          `json:"id"`
		Name json.RawMessage `json:"name"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	e.ID = raw.ID
	if len(raw.Name) == 0 {
		return nil
	}
	var names []string
	if err := json.Unmarshal(raw.Name, &names); err == nil {
		e.Name = names
		return nil
	}
	var single string
	if err := json.Unmarshal(raw.Name, &single); err == nil {
		if single != "" {
			e.Name = []string{single}
		}
		return nil
	}
	return nil
}

func ParseOrganizationClaim(raw string) (*Organization, error) {
	var orgs map[string]organizationClaimEntry
	if err := json.Unmarshal([]byte(raw), &orgs); err != nil {
		return nil, err
	}
	return firstOrganization(orgs)
}

func ParseOrganizationFromClaims(claim any) (*Organization, error) {
	switch v := claim.(type) {
	case string:
		if v == "" {
			return nil, nil
		}
		var direct struct {
			ID   string `json:"id"`
			Name any    `json:"name"`
		}
		if err := json.Unmarshal([]byte(v), &direct); err == nil && strings.TrimSpace(direct.ID) != "" {
			return &Organization{ID: strings.TrimSpace(direct.ID), Name: organizationName(direct.Name)}, nil
		}
		return ParseOrganizationClaim(v)
	case map[string]interface{}:
		if id, ok := v["id"].(string); ok && strings.TrimSpace(id) != "" {
			return &Organization{ID: strings.TrimSpace(id), Name: organizationName(v["name"])}, nil
		}
		raw, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		return ParseOrganizationClaim(string(raw))
	case []interface{}:
		for _, entry := range v {
			if org, err := ParseOrganizationFromClaims(entry); err != nil {
				return nil, err
			} else if org != nil {
				return org, nil
			}
		}
		return nil, nil
	default:
		return nil, fmt.Errorf("unsupported organization claim type %T", claim)
	}
}

func organizationName(raw any) string {
	switch value := raw.(type) {
	case string:
		return value
	case []interface{}:
		if len(value) > 0 {
			if name, ok := value[0].(string); ok {
				return name
			}
		}
	case []string:
		if len(value) > 0 {
			return value[0]
		}
	}
	return ""
}

func firstOrganization(orgs map[string]organizationClaimEntry) (*Organization, error) {
	for alias, entry := range orgs {
		name := ""
		if len(entry.Name) > 0 {
			name = entry.Name[0]
		}
		id := strings.TrimSpace(entry.ID)
		return &Organization{ID: id, Name: name, Alias: alias}, nil
	}
	return nil, nil
}

// NormalizeOrganization fills a missing org ID from alias/name heuristics.
// Keycloak Organizations claims usually look like {"alias":{"name":["..."]}}
// without a nested "id"; newer user attributes embed id in the JSON value.
func NormalizeOrganization(org *Organization) *Organization {
	if org == nil {
		return nil
	}
	id := strings.TrimSpace(org.ID)
	// Legacy mistake: attribute used id:"luxus" instead of the real UUID.
	if id == "luxus" || id == "default" {
		org.ID = DefaultLuxusOrganizationID
		if strings.TrimSpace(org.Alias) == "" {
			org.Alias = "luxus"
		}
		return org
	}
	if id != "" {
		return org
	}
	aliasRaw := strings.TrimSpace(org.Alias)
	alias := strings.ToLower(aliasRaw)
	if looksLikeUUID(aliasRaw) {
		org.ID = aliasRaw
		return org
	}
	// Only the exact legacy Luxus alias remaps — never substring matches
	// like "luxus-connect" company names belonging to other tenants.
	if alias == "luxus" {
		org.ID = DefaultLuxusOrganizationID
	}
	return org
}

// normalizeOrganization is kept for internal call sites.
func normalizeOrganization(org *Organization) *Organization {
	return NormalizeOrganization(org)
}

func looksLikeUUID(v string) bool {
	v = strings.TrimSpace(v)
	if len(v) != 36 {
		return false
	}
	for i, r := range v {
		switch i {
		case 8, 13, 18, 23:
			if r != '-' {
				return false
			}
		default:
			if (r < '0' || r > '9') && (r < 'a' || r > 'f') && (r < 'A' || r > 'F') {
				return false
			}
		}
	}
	return true
}

const DefaultLuxusOrganizationID = "00000000-0000-0000-0000-000000000001"
