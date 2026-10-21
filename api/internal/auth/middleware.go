package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

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
var errMFA = errors.New("confirme sua identidade com uma passkey para continuar")

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
				httputil.WriteFail(w, http.StatusForbidden, notifications.N("MFA_REQUIRED", "Confirme sua identidade com uma passkey na página de segurança da conta e entre novamente."))
				return
			}
			httputil.WriteFail(w, http.StatusUnauthorized, notifications.N("UNAUTHORIZED", err.Error()))
			return
		}
		u := UserFromContext(req.Context())
		// Defense in depth: production MFA applies to every account, not only privileged roles.
		if m.cfg.IsProduction() && (u == nil || !MFAVerifiedFromClaims(u.Acr, u.Amr)) && strings.TrimSuffix(r.URL.Path, "/") != "/v1/me" {
			httputil.WriteFail(w, http.StatusForbidden, notifications.N("MFA_REQUIRED", "Confirme sua identidade com uma passkey na página de segurança da conta e entre novamente."))
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
	if session.Verified {
		amr = append(amr, "passkey")
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
		// Defense in depth: production MFA applies to every account, not only privileged roles.
		if m.cfg.IsProduction() && (u == nil || !MFAVerifiedFromClaims(u.Acr, u.Amr)) {
			httputil.WriteFail(w, http.StatusForbidden, notifications.N("MFA_REQUIRED", "Confirme sua identidade com uma passkey na página de segurança da conta e entre novamente."))
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

func MFAVerifiedFromClaims(acr string, amr []string) bool {
	for _, item := range amr {
		switch strings.ToLower(strings.TrimSpace(item)) {
		case "otp", "mfa", "totp", "hwk", "passkey", "webauthn":
			return true
		}
	}
	switch strings.TrimSpace(acr) {
	case "2", "loa2", "http://schemas.openid.net/pape/policies/2007/06/multi-factor":
		return true
	}
	return false
}

// BackgroundContext returns a context with org for background jobs.
func BackgroundContext(orgID string) context.Context {
	ctx := context.Background()
	return WithOrganization(ctx, &Organization{ID: orgID})
}
