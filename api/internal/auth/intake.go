package auth

import (
	"crypto/sha256"
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/notifications"
)

// IntakeAuthenticate is only mounted on the intake write route, never on staff APIs.
func IntakeAuthenticate(key, org string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.TrimSpace(key) == "" || strings.TrimSpace(org) == "" {
				httputil.WriteFail(w, 503, notifications.N("INTAKE_DISABLED", "Integração de clientes não configurada."))
				return
			}
			provided := r.Header.Get("X-API-Key")
			if provided == "" && strings.HasPrefix(r.Header.Get("Authorization"), "Bearer ") {
				provided = strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
			}
			expectedHash, providedHash := sha256.Sum256([]byte(key)), sha256.Sum256([]byte(provided))
			if subtle.ConstantTimeCompare(expectedHash[:], providedHash[:]) != 1 {
				httputil.WriteFail(w, 401, notifications.N("UNAUTHORIZED", "Chave de integração inválida."))
				return
			}
			ctx := WithOrganization(r.Context(), &Organization{ID: org})
			ctx = WithUser(ctx, &User{ID: "customer-intake", Username: "customer-intake", Name: "Integração de clientes", Roles: []string{RoleEmployee}})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
