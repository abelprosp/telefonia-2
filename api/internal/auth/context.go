package auth

import (
	"context"
	"encoding/json"
	"fmt"
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
	Name []string `json:"name"`
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
		return &Organization{ID: entry.ID, Name: name, Alias: alias}, nil
	}
	return nil, nil
}
