package auth

import (
	"net/http"

	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/notifications"
)

// These guards run after Authenticate, which enforces session and MFA checks.
func ApprovalReadAccess(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !CanAccessOperational(r.Context()) && !CanApproveTwoLevel(r.Context()) {
			httputil.WriteFail(w, http.StatusForbidden, notifications.N("FORBIDDEN", "Seu perfil não permite consultar aprovações."))
			return
		}
		next.ServeHTTP(w, r)
	})
}

func ApprovalDecisionAccess(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !CanApproveTwoLevel(r.Context()) {
			httputil.WriteFail(w, http.StatusForbidden, notifications.N("APPROVAL_ROLE_REQUIRED", "Apenas master/admin/financeiro podem aprovar ou recusar."))
			return
		}
		next.ServeHTTP(w, r)
	})
}
