package auth

import (
	"context"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

func (m *Middleware) enrichUserIdentity(ctx context.Context, tokenStr string, claims jwt.MapClaims, user *User) *User {
	if user == nil {
		return user
	}

	if user.ID == "" {
		if sid, ok := claims["sid"].(string); ok && sid != "" && m.keycloak != nil && m.keycloak.Enabled() {
			if userID, err := m.keycloak.ResolveUserIDBySessionSID(ctx, sid, m.cfg.KeycloakResource); err == nil {
				user.ID = userID
			} else {
				m.logger.Warn("failed to resolve user id from keycloak session", "error", err)
			}
		}
	}

	return m.enrichUserFromUserInfo(ctx, tokenStr, user)
}

func (m *Middleware) resolveOrganizationFromKeycloak(ctx context.Context, userID string) *Organization {
	if m.keycloak == nil || !m.keycloak.Enabled() || strings.TrimSpace(userID) == "" {
		return nil
	}

	raw, err := m.keycloak.GetUserOrganizationAttribute(ctx, userID)
	if err != nil {
		m.logger.Warn("failed to resolve organization from keycloak", "error", err)
		return nil
	}

	org, err := ParseOrganizationClaim(raw)
	if err != nil {
		m.logger.Warn("failed to parse organization attribute", "error", err)
		return nil
	}
	return org
}
