package handlers

import (
	"net/http"

	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/models"
	"github.com/luxus-connect/telefonia/api/internal/notifications"
)

func (h *Handler) getCurrentUserProfile(w http.ResponseWriter, r *http.Request) {
	profile, err := h.Svc.GetCurrentUserProfile(r.Context())
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, profile)
}

func (h *Handler) updateCurrentUserProfile(w http.ResponseWriter, r *http.Request) {
	var input models.UpdateUserProfileInput
	if err := decodeJSON(r, &input); err != nil {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Corpo da requisição inválido."))
		return
	}
	profile, err := h.Svc.UpdateCurrentUserProfile(r.Context(), input)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, profile)
}

func (h *Handler) getOrganizationSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := h.Svc.GetOrganizationSettings(r.Context())
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, settings)
}

func (h *Handler) updateCompanySettings(w http.ResponseWriter, r *http.Request) {
	var input models.UpdateCompanySettingsInput
	if err := decodeJSON(r, &input); err != nil {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Corpo da requisição inválido."))
		return
	}
	settings, err := h.Svc.UpdateCompanySettings(r.Context(), input)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, settings)
}

func (h *Handler) updateWhitelabelSettings(w http.ResponseWriter, r *http.Request) {
	var input models.UpdateWhitelabelSettingsInput
	if err := decodeJSON(r, &input); err != nil {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Corpo da requisição inválido."))
		return
	}
	settings, err := h.Svc.UpdateWhitelabelSettings(r.Context(), input)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, settings)
}

func (h *Handler) updateSystemSettings(w http.ResponseWriter, r *http.Request) {
	var input models.UpdateSystemSettingsInput
	if err := decodeJSON(r, &input); err != nil {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Corpo da requisição inválido."))
		return
	}
	settings, err := h.Svc.UpdateSystemSettings(r.Context(), input)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, settings)
}
