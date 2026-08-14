package handlers

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/models"
	"github.com/luxus-connect/telefonia/api/internal/notifications"
	"github.com/luxus-connect/telefonia/api/internal/services"
)

func (h *Handler) listExceedanceTerms(w http.ResponseWriter, r *http.Request) {
	items, err := h.Svc.ListExceedanceTerms(r.Context())
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, items)
}

func (h *Handler) createExceedanceTerm(w http.ResponseWriter, r *http.Request) {
	var input models.CreateExceedanceTermInput
	if err := decodeJSON(r, &input); err != nil {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Invalid request body"))
		return
	}
	item, err := h.Svc.CreateExceedanceTerm(r.Context(), input)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, item)
}

func (h *Handler) updateExceedanceTerm(w http.ResponseWriter, r *http.Request) {
	var input models.UpdateExceedanceTermInput
	if err := decodeJSON(r, &input); err != nil {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Invalid request body"))
		return
	}
	item, err := h.Svc.UpdateExceedanceTerm(r.Context(), chi.URLParam(r, "id"), input)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) updatePhoneLineExceedanceSettings(w http.ResponseWriter, r *http.Request) {
	var input models.UpdatePhoneLineExceedanceInput
	if err := decodeJSON(r, &input); err != nil {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Invalid request body"))
		return
	}
	item, err := h.Svc.UpdatePhoneLineExceedanceSettings(r.Context(), chi.URLParam(r, "id"), input)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) getLineFidelity(w http.ResponseWriter, r *http.Request) {
	item, err := h.Svc.GetLineFidelity(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) upsertLineFidelity(w http.ResponseWriter, r *http.Request) {
	var input models.UpsertLineFidelityInput
	if err := decodeJSON(r, &input); err != nil {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Invalid request body"))
		return
	}
	item, err := h.Svc.UpsertLineFidelity(r.Context(), chi.URLParam(r, "id"), input)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) decideLineFidelityRenewal(w http.ResponseWriter, r *http.Request) {
	var input models.FidelityRenewalDecisionInput
	if err := decodeJSON(r, &input); err != nil {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Invalid request body"))
		return
	}
	item, err := h.Svc.DecideFidelityRenewal(r.Context(), chi.URLParam(r, "id"), input)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) listFidelityRenewalTriggers(w http.ResponseWriter, r *http.Request) {
	items, err := h.Svc.ListFidelityRenewalTriggers(r.Context())
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, items)
}

func (h *Handler) updateFidelityRenewalTrigger(w http.ResponseWriter, r *http.Request) {
	var input models.UpdateFidelityRenewalTriggerInput
	if err := decodeJSON(r, &input); err != nil {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Invalid request body"))
		return
	}
	item, err := h.Svc.UpdateFidelityRenewalTrigger(r.Context(), chi.URLParam(r, "id"), input)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) listPhoneLineGeneratedContracts(w http.ResponseWriter, r *http.Request) {
	items, err := h.Svc.ListGeneratedContractsForPhoneLine(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, items)
}

func (h *Handler) listCustomerGeneratedContracts(w http.ResponseWriter, r *http.Request) {
	items, err := h.Svc.ListGeneratedContractsForCustomer(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, items)
}

func (h *Handler) getFinancialExport(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		id = r.URL.Query().Get("processing_month_id")
	}
	data, err := h.Svc.ExportFinancialBilling(r.Context(), id)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	if strings.EqualFold(r.URL.Query().Get("format"), "csv") {
		csv := services.FinancialExportCSV(data)
		w.Header().Set("Content-Type", "text/csv; charset=utf-8")
		w.Header().Set("Content-Disposition", `attachment; filename="export-financeiro.csv"`)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(csv)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, data)
}
