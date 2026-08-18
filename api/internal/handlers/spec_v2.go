package handlers

import (
	"net/http"
	"strconv"
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

func (h *Handler) estimateLineFidelityPenalty(w http.ResponseWriter, r *http.Request) {
	cancelDate := r.URL.Query().Get("cancel_date")
	var cancelPtr *string
	if cancelDate != "" {
		cancelPtr = &cancelDate
	}
	item, err := h.Svc.EstimateLineFidelityPenalty(r.Context(), chi.URLParam(r, "id"), cancelPtr, nil)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) applyLineFidelityPenalty(w http.ResponseWriter, r *http.Request) {
	var input models.ApplyFidelityPenaltyInput
	if err := decodeJSON(r, &input); err != nil {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Invalid request body"))
		return
	}
	item, err := h.Svc.ApplyLineFidelityPenalty(r.Context(), chi.URLParam(r, "id"), input)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, item)
}

func (h *Handler) apportionProviderInvoiceDiscount(w http.ResponseWriter, r *http.Request) {
	var input models.ApportionGlobalDiscountInput
	if err := decodeJSON(r, &input); err != nil {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Invalid request body"))
		return
	}
	item, err := h.Svc.ApportionProviderInvoiceDiscount(r.Context(), chi.URLParam(r, "id"), input)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) listExpiringContracts(w http.ResponseWriter, r *http.Request) {
	daysStr := r.URL.Query().Get("days")
	days := 30
	if daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 {
			days = d
		}
	}
	items, err := h.Svc.ListExpiringContracts(r.Context(), days)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, items)
}

func (h *Handler) getPreClosingAlerts(w http.ResponseWriter, r *http.Request) {
	items, err := h.Svc.GetPreClosingAlerts(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, items)
}

func (h *Handler) getPhoneLineTimeline(w http.ResponseWriter, r *http.Request) {
	item, err := h.Svc.GetPhoneLineTimeline(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) anonymizeCustomer(w http.ResponseWriter, r *http.Request) {
	item, err := h.Svc.AnonymizeCustomer(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) exportCustomerPersonalData(w http.ResponseWriter, r *http.Request) {
	item, err := h.Svc.ExportCustomerPersonalData(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) simulateBillingImpact(w http.ResponseWriter, r *http.Request) {
	item, err := h.Svc.SimulateBillingImpact(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) getProcessingMonthLineReadiness(w http.ResponseWriter, r *http.Request) {
	item, err := h.Svc.GetProcessingMonthLineReadiness(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) getLineBillingExplanation(w http.ResponseWriter, r *http.Request) {
	monthID := r.URL.Query().Get("processing_month_id")
	item, err := h.Svc.GetLineBillingExplanation(r.Context(), chi.URLParam(r, "id"), monthID)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) closeProcessingMonthWithHash(w http.ResponseWriter, r *http.Request) {
	item, err := h.Svc.CloseProcessingMonthWithHash(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) getOperationalDashboard(w http.ResponseWriter, r *http.Request) {
	item, err := h.Svc.GetOperationalDashboard(r.Context())
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) getPhoneLine360(w http.ResponseWriter, r *http.Request) {
	item, err := h.Svc.GetPhoneLine360(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) getCustomer360(w http.ResponseWriter, r *http.Request) {
	item, err := h.Svc.GetCustomer360(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) listDivergences(w http.ResponseWriter, r *http.Request) {
	monthID := r.URL.Query().Get("processing_month_id")
	items, err := h.Svc.ListDivergences(r.Context(), monthID)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, items)
}

func (h *Handler) resolveDivergence(w http.ResponseWriter, r *http.Request) {
	var input models.ResolveDivergenceInput
	if err := decodeJSON(r, &input); err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	item, err := h.Svc.ResolveDivergence(r.Context(), chi.URLParam(r, "id"), input)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) portalGetMe(w http.ResponseWriter, r *http.Request) {
	item, err := h.Svc.PortalGetMe(r.Context())
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) portalListLines(w http.ResponseWriter, r *http.Request) {
	items, err := h.Svc.PortalListLines(r.Context())
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, items)
}

func (h *Handler) portalListInvoices(w http.ResponseWriter, r *http.Request) {
	items, err := h.Svc.PortalListInvoices(r.Context())
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, items)
}

func (h *Handler) generateSicrediPix(w http.ResponseWriter, r *http.Request) {
	item, err := h.Svc.GenerateSicrediPix(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) getFinancialSummaryReport(w http.ResponseWriter, r *http.Request) {
	item, err := h.Svc.GetFinancialSummaryReport(r.Context())
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) getCustomerProfitabilityReport(w http.ResponseWriter, r *http.Request) {
	item, err := h.Svc.GetCustomerProfitabilityReport(r.Context())
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) listWebhooks(w http.ResponseWriter, r *http.Request) {
	items, err := h.Svc.ListWebhooks(r.Context())
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, items)
}

func (h *Handler) createWebhook(w http.ResponseWriter, r *http.Request) {
	var input models.CreateWebhookSubscriptionInput
	if err := decodeJSON(r, &input); err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	item, err := h.Svc.CreateWebhook(r.Context(), input)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, item)
}

func (h *Handler) deleteWebhook(w http.ResponseWriter, r *http.Request) {
	if err := h.Svc.DeleteWebhook(r.Context(), chi.URLParam(r, "id")); err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) testWebhook(w http.ResponseWriter, r *http.Request) {
	item, err := h.Svc.TestWebhook(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) listInventoryDevices(w http.ResponseWriter, r *http.Request) {
	items, err := h.Svc.ListInventoryDevices(r.Context())
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, items)
}

func (h *Handler) createInventoryDevice(w http.ResponseWriter, r *http.Request) {
	var input models.CreateInventoryDeviceInput
	if err := decodeJSON(r, &input); err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	item, err := h.Svc.CreateInventoryDevice(r.Context(), input)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, item)
}

func (h *Handler) updateInventoryDevice(w http.ResponseWriter, r *http.Request) {
	var input models.UpdateInventoryDeviceInput
	if err := decodeJSON(r, &input); err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	item, err := h.Svc.UpdateInventoryDevice(r.Context(), chi.URLParam(r, "id"), input)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) exportOrganizationData(w http.ResponseWriter, r *http.Request) {
	item, err := h.Svc.ExportOrganizationData(r.Context())
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}
