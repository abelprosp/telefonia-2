package auth

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/notifications"
)

const ParceirosKeyHeader = "X-Parceiros-Key"

// ParceirosAuthenticate libera só a porta de entrada das vendas concluídas do Luxus Parceiros.
func ParceirosAuthenticate(apiKey, orgID string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.TrimSpace(apiKey) == "" || strings.TrimSpace(orgID) == "" {
				httputil.WriteFail(w, http.StatusServiceUnavailable, notifications.N("PARCEIROS_INTAKE_NOT_CONFIGURED", "A entrada do Luxus Parceiros não está configurada."))
				return
			}
			provided := strings.TrimSpace(r.Header.Get(ParceirosKeyHeader))
			if subtle.ConstantTimeCompare([]byte(provided), []byte(apiKey)) != 1 {
				httputil.WriteFail(w, http.StatusUnauthorized, notifications.N("UNAUTHORIZED", "Chave do Luxus Parceiros inválida."))
				return
			}
			ctx := WithUser(r.Context(), &User{
				ID:       "luxus-parceiros",
				Email:    "parceiros@luxus.local",
				Name:     "Luxus Parceiros",
				Username: "luxus-parceiros",
				Roles:    []string{RoleOperator},
			})
			ctx = WithOrganization(ctx, &Organization{ID: orgID, Name: "Luxus Connect", Alias: "luxus"})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
