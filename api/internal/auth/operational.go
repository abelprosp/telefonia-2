package auth

import (
	"net/http"
	"strings"

	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/notifications"
)

// OperationalMutationGuard separates commercial writes from operational writes.
// Sales staff may maintain customers and sales, but may not administer stock,
// providers, processing months, approvals or contract templates.
func OperationalMutationGuard(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions || CanMutateOperational(r.Context()) {
			next.ServeHTTP(w, r)
			return
		}
		path := strings.TrimSuffix(r.URL.Path, "/")
		// Reading one's notifications is a personal preference, not an operational mutation.
		if r.Method == http.MethodPost && strings.HasPrefix(path, "/v1/inapp-notifications/") {
			next.ServeHTTP(w, r)
			return
		}
		commercial := IsSales(r.Context()) && (path == "/v1/customers" && r.Method == http.MethodPost ||
			strings.HasPrefix(path, "/v1/customers/") && r.Method == http.MethodPatch && strings.Count(path, "/") == 3 ||
			strings.HasPrefix(path, "/v1/sales") && (path == "/v1/sales" || strings.HasPrefix(path, "/v1/sales/")) && !strings.HasSuffix(path, "/mark-paid"))
		if !commercial {
			httputil.WriteFail(w, http.StatusForbidden, notifications.N("FORBIDDEN", "Seu perfil não permite esta alteração operacional."))
			return
		}
		next.ServeHTTP(w, r)
	})
}
