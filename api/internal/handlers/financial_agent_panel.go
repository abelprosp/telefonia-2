package handlers

import (
	"net/http"

	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/models"
	"github.com/luxus-connect/telefonia/api/internal/notifications"
)

func (h *Handler) getFinancialAgentPanel(w http.ResponseWriter, r *http.Request) {
	item, err := h.Svc.GetFinancialAgentPanel(r.Context())
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) updateFinancialAgentSettings(w http.ResponseWriter, r *http.Request) {
	var input models.FinancialAgentSettingsInput
	if err := decodeJSON(r, &input); err != nil {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Corpo da requisição inválido."))
		return
	}
	item, err := h.Svc.UpdateFinancialAgentSettings(r.Context(), input)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) getFinancialAgentWhatsApp(w http.ResponseWriter, r *http.Request) {
	item, err := h.Svc.GetFinancialAgentWhatsApp(r.Context())
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) connectFinancialAgentWhatsApp(w http.ResponseWriter, r *http.Request) {
	item, err := h.Svc.ConnectFinancialAgentWhatsApp(r.Context())
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) disconnectFinancialAgentWhatsApp(w http.ResponseWriter, r *http.Request) {
	item, err := h.Svc.DisconnectFinancialAgentWhatsApp(r.Context())
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) listFinancialAgentEvents(w http.ResponseWriter, r *http.Request) {
	items, err := h.Svc.ListFinancialAgentEvents(r.Context())
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WritePaged(w, items, int64(len(items)))
}
