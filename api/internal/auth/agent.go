package auth

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/notifications"
)

const AgentKeyHeader = "X-Agent-Key"

// AgentAuthenticate valida a chave estática do agente n8n e injeta org + usuário de serviço.
func AgentAuthenticate(apiKey, orgID string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.TrimSpace(apiKey) == "" {
				httputil.WriteFail(w, http.StatusServiceUnavailable, notifications.FinancialAgentNotConfigured)
				return
			}
			provided := strings.TrimSpace(r.Header.Get(AgentKeyHeader))
			if provided == "" {
				if auth := r.Header.Get("Authorization"); strings.HasPrefix(auth, "Bearer ") {
					provided = strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))
				}
			}
			if subtle.ConstantTimeCompare([]byte(provided), []byte(apiKey)) != 1 {
				httputil.WriteFail(w, http.StatusUnauthorized, notifications.N("UNAUTHORIZED", "Invalid financial agent key"))
				return
			}
			if strings.TrimSpace(orgID) == "" {
				httputil.WriteFail(w, http.StatusServiceUnavailable, notifications.FinancialAgentOrgRequired)
				return
			}
			ctx := WithUser(r.Context(), &User{
				ID:       "financial-agent",
				Email:    "agente-financeiro@luxus.local",
				Name:     "Agente Financeiro",
				Username: "financial-agent",
				Roles:    []string{RoleFinancial, RoleMaster},
			})
			ctx = WithOrganization(ctx, &Organization{ID: orgID, Name: "Luxus Connect", Alias: "luxus"})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
