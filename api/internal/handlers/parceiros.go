package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	connectauth "github.com/luxus-connect/telefonia/api/internal/auth"
	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/notifications"
	"github.com/luxus-connect/telefonia/api/internal/parceiros"
)

func (h *Handler) importParceirosSale(w http.ResponseWriter, r *http.Request) {
	org := connectauth.OrganizationFromContext(r.Context())
	if org == nil || org.ID == "" {
		httputil.WriteFail(w, http.StatusServiceUnavailable, notifications.N("PARCEIROS_INTAKE_NOT_CONFIGURED", "A entrada do Luxus Parceiros não está configurada."))
		return
	}
	var input parceiros.SaleIntake
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("INVALID_BODY", "JSON inválido."))
		return
	}
	result, err := h.Svc.ImportParceirosSale(r.Context(), org.ID, input)
	if err != nil {
		var validation *parceiros.ValidationError
		if errors.As(err, &validation) {
			httputil.WriteFail(w, http.StatusBadRequest, notifications.N("VALIDATION", validation.Error()))
			return
		}
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, result)
}
