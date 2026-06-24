package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/models"
	"github.com/luxus-connect/telefonia/api/internal/notifications"
	"github.com/luxus-connect/telefonia/api/internal/services"
)

func (h *Handler) sicrediWebhook(w http.ResponseWriter, r *http.Request) {
	if err := h.Svc.HandleSicrediWebhook(r.Context(), r); err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]bool{"received": true})
}

func (h *Handler) getSicrediStatus(w http.ResponseWriter, r *http.Request) {
	item, err := h.Svc.GetSicrediStatus(r.Context())
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) registerSicrediWebhook(w http.ResponseWriter, r *http.Request) {
	var input models.RegisterSicrediWebhookInput
	_ = decodeJSON(r, &input)
	item, err := h.Svc.RegisterSicrediWebhook(r.Context(), &input)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) setupSicrediProduction(w http.ResponseWriter, r *http.Request) {
	var input models.RegisterSicrediWebhookInput
	_ = decodeJSON(r, &input)
	item, err := h.Svc.SetupSicrediProduction(r.Context(), &input)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	status := http.StatusOK
	if !item.Success {
		status = http.StatusConflict
	}
	httputil.WriteJSON(w, status, item)
}

func (h *Handler) testSicrediConnection(w http.ResponseWriter, r *http.Request) {
	item, err := h.Svc.TestSicrediConnection(r.Context())
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	status := http.StatusOK
	if !item.Success {
		status = http.StatusConflict
	}
	httputil.WriteJSON(w, status, item)
}

func (h *Handler) getSicrediBoletoPDF(w http.ResponseWriter, r *http.Request) {
	pdf, filename, err := h.Svc.GetSicrediBoletoPDF(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "inline; filename=\""+filename+"\"")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(pdf)
}

func (h *Handler) cancelSicrediBoleto(w http.ResponseWriter, r *http.Request) {
	item, err := h.Svc.CancelSicrediBoleto(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) alterSicrediBoletoDueDate(w http.ResponseWriter, r *http.Request) {
	var input services.AlterSicrediBoletoDueDateInput
	if err := decodeJSON(r, &input); err != nil {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Invalid request body"))
		return
	}
	item, err := h.Svc.AlterSicrediBoletoDueDate(r.Context(), chi.URLParam(r, "id"), input)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}
