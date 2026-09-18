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
		return ParseOrganizationClaim(v)
	case map[string]interface{}:
		raw, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		return ParseOrganizationClaim(string(raw))
	default:
		return nil, fmt.Errorf("unsupported organization claim type %T", claim)
	}
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

// normalizeOrganization fills a missing org ID from alias/name heuristics.
// Keycloak Organizations claims usually look like {"alias":{"name":["..."]}}
// without a nested "id"; newer user attributes use the UUID as the map key.
func normalizeOrganization(org *Organization) *Organization {
	if org == nil {
		return nil
	}
	if strings.TrimSpace(org.ID) != "" {
		return org
	}
	aliasRaw := strings.TrimSpace(org.Alias)
	alias := strings.ToLower(aliasRaw)
	name := strings.ToLower(strings.TrimSpace(org.Name))
	if looksLikeUUID(aliasRaw) {
		org.ID = aliasRaw
		return org
	}
	if alias == "luxus" || strings.Contains(alias, "luxus") || strings.Contains(name, "luxus") {
		org.ID = DefaultLuxusOrganizationID
	}
	return org
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
