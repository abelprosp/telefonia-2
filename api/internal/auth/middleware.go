package auth

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	keyfunc "github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/luxus-connect/telefonia/api/internal/config"
	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/keycloak"
	"github.com/luxus-connect/telefonia/api/internal/notifications"
)

type Middleware struct {
	jwks      keyfunc.Keyfunc
	cfg       config.Config
	logger    *slog.Logger
	keycloak  *keycloak.AdminClient
}

func NewMiddleware(cfg config.Config, logger *slog.Logger, kc *keycloak.AdminClient) (*Middleware, error) {
	jwksURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/certs",
		cfg.KeycloakAuthServerURL, cfg.KeycloakRealm)

	jwks, err := keyfunc.NewDefault([]string{jwksURL})
	if err != nil {
		return nil, fmt.Errorf("jwks: %w", err)
	}

	return &Middleware{jwks: jwks, cfg: cfg, logger: logger, keycloak: kc}, nil
}

func (m *Middleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			httputil.WriteFail(w, http.StatusUnauthorized, notifications.N("UNAUTHORIZED", "Missing or invalid authorization header"))
			return
		}
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		token, err := jwt.Parse(tokenStr, m.jwks.Keyfunc)
		if err != nil || !token.Valid {
			httputil.WriteFail(w, http.StatusUnauthorized, notifications.N("UNAUTHORIZED", "Invalid token"))
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			httputil.WriteFail(w, http.StatusUnauthorized, notifications.N("UNAUTHORIZED", "Invalid token claims"))
			return
		}

		user := extractUser(claims)
		user = m.enrichUserIdentity(r.Context(), tokenStr, claims, user)
		ctx := WithUser(r.Context(), user)

		var org *Organization
		if claim, ok := claims["organization"]; ok && claim != nil {
			parsed, err := ParseOrganizationFromClaims(claim)
			if err != nil {
				m.logger.Warn("failed to parse organization claim", "error", err)
			} else {
				org = parsed
			}
		}
		if org == nil && user != nil {
			org = m.resolveOrganizationFromKeycloak(r.Context(), user.ID)
		}
		if org != nil {
			ctx = WithOrganization(ctx, org)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *Middleware) RequireAdmin(next http.Handler) http.Handler {
	return m.requireRole(next, CanManageUsers, "Master role required")
}

func (m *Middleware) RequireOperational(next http.Handler) http.Handler {
	return m.requireRole(next, CanAccessOperational, "Operational access required")
}

func (m *Middleware) RequireFinancialAccess(next http.Handler) http.Handler {
	return m.requireRole(next, CanAccessFinancial, "Financial access required")
}

func (m *Middleware) RequireMaster(next http.Handler) http.Handler {
	return m.requireRole(next, CanManageUsers, "Master role required")
}

func (m *Middleware) requireRole(next http.Handler, check func(context.Context) bool, message string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !check(r.Context()) {
			httputil.WriteFail(w, http.StatusForbidden, notifications.N("FORBIDDEN", message))
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (m *Middleware) RequirePartner(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !HasRole(r.Context(), RolePartner) {
			httputil.WriteFail(w, http.StatusForbidden, notifications.N("FORBIDDEN", "Partner role required"))
			return
		}
		next.ServeHTTP(w, r)
	})
}

func extractUser(claims jwt.MapClaims) *User {
	sub, _ := claims["sub"].(string)
	if sub == "" {
		if alt, ok := claims["connect_user_id"].(string); ok {
			sub = alt
		}
	}
	email, _ := claims["email"].(string)
	name, _ := claims["name"].(string)
	preferred, _ := claims["preferred_username"].(string)

	return &User{
		ID:       sub,
		Email:    email,
		Name:     name,
		Username: preferred,
		Roles:    extractRoles(claims),
	}
}

func extractRoles(claims jwt.MapClaims) []string {
	seen := make(map[string]struct{})
	var roles []string
	add := func(items []string) {
		for _, r := range items {
			if r == "" {
				continue
			}
			if _, ok := seen[r]; ok {
				continue
			}
			seen[r] = struct{}{}
			roles = append(roles, r)
		}
	}

	add(parseRoleSlice(claims["roles"]))

	if realmAccess, ok := claims["realm_access"].(map[string]interface{}); ok {
		add(parseRoleSlice(realmAccess["roles"]))
	}

	if resourceAccess, ok := claims["resource_access"].(map[string]interface{}); ok {
		for _, raw := range resourceAccess {
			if client, ok := raw.(map[string]interface{}); ok {
				add(parseRoleSlice(client["roles"]))
			}
		}
	}

	return roles
}

func parseRoleSlice(raw any) []string {
	switch v := raw.(type) {
	case []interface{}:
		out := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok {
				out = append(out, s)
			}
		}
		return out
	case []string:
		return v
	default:
		return nil
	}
}

// BackgroundContext returns a context with org for background jobs.
var bgMu sync.Mutex

func BackgroundContext(orgID string) context.Context {
	ctx := context.Background()
	return WithOrganization(ctx, &Organization{ID: orgID})
}

// RefreshJWKS periodically refreshes JWKS in production.
func (m *Middleware) StartJWKSRefresh(ctx context.Context) {
	if !m.cfg.IsProduction() {
		return
	}
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// keyfunc v3 auto-refreshes; noop placeholder for explicit refresh hook
			}
		}
	}()
}
