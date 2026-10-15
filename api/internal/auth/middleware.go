package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"

	"github.com/golang-jwt/jwt/v5"
	"github.com/luxus-connect/telefonia/api/internal/config"
	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/localauth"
	"github.com/luxus-connect/telefonia/api/internal/notifications"
)

type Middleware struct {
	local  *localauth.Manager
	cfg    config.Config
	logger *slog.Logger
}

var errOrigin = errors.New("origem não permitida")
var errMFA = errors.New("confirme o autenticador para continuar")

func NewMiddleware(cfg config.Config, logger *slog.Logger, local *localauth.Manager) (*Middleware, error) {
	return &Middleware{cfg: cfg, logger: logger, local: local}, nil
}

func (m *Middleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		req, err := m.authenticateRequest(r, true)
		if err != nil {
			if errors.Is(err, errOrigin) {
				httputil.WriteFail(w, http.StatusForbidden, notifications.N("CSRF_REJECTED", "Origem da solicitação não permitida."))
				return
			}
			if errors.Is(err, errMFA) {
				httputil.WriteFail(w, http.StatusForbidden, notifications.N("MFA_REQUIRED", "Confirme o autenticador para continuar."))
				return
			}
			httputil.WriteFail(w, http.StatusUnauthorized, notifications.N("UNAUTHORIZED", err.Error()))
			return
		}
		u := UserFromContext(req.Context())
		if m.cfg.IsProduction() && (IsMaster(req.Context()) || IsFinancial(req.Context())) &&
			(u == nil || !MFAVerifiedFromClaims(u.Acr, u.Amr)) && strings.TrimSuffix(r.URL.Path, "/") != "/v1/me" {
			httputil.WriteFail(w, http.StatusForbidden, notifications.N("MFA_REQUIRED", "Confirme sua identidade com autenticação de dois fatores na página de segurança da conta e entre novamente."))
			return
		}
		next.ServeHTTP(w, req)
	})
}

// OptionalAuthenticate resolves a valid session for public branding routes.
// Invalid/missing sessions
// continue without blocking the request.
func (m *Middleware) OptionalAuthenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req, err := m.authenticateRequest(r, false)
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		next.ServeHTTP(w, req)
	})
}

func (m *Middleware) authenticateRequest(r *http.Request, required bool) (*http.Request, error) {
	if m.local == nil {
		return nil, fmt.Errorf("autenticação indisponível")
	}
	session, err := m.local.Resolve(r)
	if err != nil {
		if !required {
			return r, nil
		}
		return nil, err
	}
	if r.Method != "GET" && r.Method != "HEAD" && r.Method != "OPTIONS" && !m.local.AllowedOrigin(r) {
		return nil, errOrigin
	}
	u := session.User
	amr := []string{"pwd"}
	if session.Verified && u.MFASecret != "" {
		amr = append(amr, "otp")
	}
	ctx := WithUser(r.Context(), &User{ID: u.ID, Username: u.Username, Email: u.Email, Name: strings.TrimSpace(u.FirstName + " " + u.LastName), Roles: u.Roles, Document: u.Document, Amr: amr})
	ctx = WithOrganization(ctx, &Organization{ID: u.OrganizationID, Name: u.OrganizationName})
	if m.local.NeedsMFA(session) && strings.TrimSuffix(r.URL.Path, "/") != "/v1/me" {
		return nil, errMFA
	}
	return r.WithContext(ctx), nil
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
		u := UserFromContext(r.Context())
		if m.cfg.IsProduction() && (IsMaster(r.Context()) || IsFinancial(r.Context())) &&
			(u == nil || !MFAVerifiedFromClaims(u.Acr, u.Amr)) {
			httputil.WriteFail(w, http.StatusForbidden, notifications.N("MFA_REQUIRED", "Confirme sua identidade com autenticação de dois fatores na página de segurança da conta e entre novamente."))
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

func extractUser(claims jwt.MapClaims, clientID string) *User {
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
		Roles:    extractRoles(claims, clientID),
		Document: extractDocument(claims),
		Acr:      claimString(claims["acr"]),
		Amr:      parseRoleSlice(claims["amr"]),
	}
}

func claimString(raw any) string {
	if s, ok := raw.(string); ok {
		return s
	}
	if n, ok := raw.(float64); ok {
		return fmt.Sprintf("%.0f", n)
	}
	return ""
}

func (m *Middleware) validateTokenIssuerAudience(claims jwt.MapClaims) error {
	realm := strings.TrimSpace(m.cfg.KeycloakRealm)
	if realm == "" {
		return fmt.Errorf("Keycloak realm is not configured")
	}
	publicURL := strings.TrimRight(strings.TrimSpace(m.cfg.KeycloakPublicAuthServerURL), "/")
	if publicURL == "" {
		publicURL = strings.TrimRight(strings.TrimSpace(m.cfg.KeycloakAuthServerURL), "/")
	}
	expectedIssuer := publicURL + "/realms/" + realm
	// Accept either the public or the internal issuer used by Keycloak behind a reverse proxy.
	altIssuer := strings.TrimRight(strings.TrimSpace(m.cfg.KeycloakAuthServerURL), "/") + "/realms/" + realm
	iss := strings.TrimSpace(claimString(claims["iss"]))
	if iss == "" || (iss != expectedIssuer && iss != altIssuer) {
		return fmt.Errorf("Invalid token issuer")
	}

	clientID := strings.TrimSpace(m.cfg.KeycloakResource)
	if clientID == "" {
		return fmt.Errorf("Keycloak resource is not configured")
	}
	if azp := strings.TrimSpace(claimString(claims["azp"])); azp != "" {
		if azp != clientID {
			return fmt.Errorf("Invalid token audience")
		}
		return nil
	}
	if !audienceContains(claims["aud"], clientID) {
		return fmt.Errorf("Invalid token audience")
	}
	return nil
}

func audienceContains(raw any, clientID string) bool {
	switch v := raw.(type) {
	case string:
		return strings.TrimSpace(v) == clientID
	case []any:
		for _, item := range v {
			if s, ok := item.(string); ok && strings.TrimSpace(s) == clientID {
				return true
			}
		}
	case []string:
		for _, s := range v {
			if strings.TrimSpace(s) == clientID {
				return true
			}
		}
	}
	return false
}

func MFAVerifiedFromClaims(acr string, amr []string) bool {
	for _, item := range amr {
		switch strings.ToLower(strings.TrimSpace(item)) {
		case "otp", "mfa", "totp", "hwk":
			return true
		}
	}
	switch strings.TrimSpace(acr) {
	case "2", "loa2", "http://schemas.openid.net/pape/policies/2007/06/multi-factor":
		return true
	}
	return false
}

func extractDocument(claims jwt.MapClaims) string {
	keys := []string{"cpf_cnpj", "document", "cpf", "cnpj", "tax_id"}
	for _, k := range keys {
		if v, ok := claims[k].(string); ok {
			digits := onlyDigits(v)
			if len(digits) == 11 || len(digits) == 14 {
				return digits
			}
		}
	}
	return ""
}

func onlyDigits(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func extractRoles(claims jwt.MapClaims, clientID string) []string {
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
		if client, ok := resourceAccess[clientID].(map[string]interface{}); ok {
			add(parseRoleSlice(client["roles"]))
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
