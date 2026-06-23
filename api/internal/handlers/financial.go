package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/models"
	"github.com/luxus-connect/telefonia/api/internal/notifications"
)

func (h *Handler) getFinancialSummary(w http.ResponseWriter, r *http.Request) {
	summary, err := h.Svc.GetFinancialSummary(r.Context())
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, summary)
}

func (h *Handler) listAccountsPayable(w http.ResponseWriter, r *http.Request) {
	page := httputil.ParsePagination(r)
	items, total, err := h.Svc.ListAccountsPayable(r.Context(), queryParam(r, "status"), page)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WritePaged(w, items, total)
}

func (h *Handler) createAccountPayable(w http.ResponseWriter, r *http.Request) {
	var input models.CreateAccountPayableInput
	if err := decodeJSON(r, &input); err != nil {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Invalid request body"))
		return
	}
	item, err := h.Svc.CreateAccountPayable(r.Context(), input)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, item)
}

func (h *Handler) createAccountPayableFromInvoice(w http.ResponseWriter, r *http.Request) {
	item, err := h.Svc.CreateAccountPayableFromInvoice(r.Context(), chi.URLParam(r, "invoiceId"))
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, item)
}

func (h *Handler) updateAccountPayable(w http.ResponseWriter, r *http.Request) {
	var input models.UpdateAccountPayableInput
	if err := decodeJSON(r, &input); err != nil {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Invalid request body"))
		return
	}
	if err := h.Svc.UpdateAccountPayable(r.Context(), chi.URLParam(r, "id"), input); err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (h *Handler) registerPayablePayment(w http.ResponseWriter, r *http.Request) {
	var input models.RegisterFinancialPaymentInput
	if err := decodeJSON(r, &input); err != nil {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Invalid request body"))
		return
	}
	if err := h.Svc.RegisterPayablePayment(r.Context(), chi.URLParam(r, "id"), input); err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (h *Handler) listAccountsReceivable(w http.ResponseWriter, r *http.Request) {
	page := httputil.ParsePagination(r)
	items, total, err := h.Svc.ListAccountsReceivable(r.Context(), queryParam(r, "customer_id"), queryParam(r, "status"), page)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WritePaged(w, items, total)
}

func (h *Handler) createAccountReceivable(w http.ResponseWriter, r *http.Request) {
	var input models.CreateAccountReceivableInput
	if err := decodeJSON(r, &input); err != nil {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Invalid request body"))
		return
	}
	item, err := h.Svc.CreateAccountReceivable(r.Context(), input)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, item)
}

func (h *Handler) updateAccountReceivable(w http.ResponseWriter, r *http.Request) {
	var input models.UpdateAccountReceivableInput
	if err := decodeJSON(r, &input); err != nil {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Invalid request body"))
		return
	}
	if err := h.Svc.UpdateAccountReceivable(r.Context(), chi.URLParam(r, "id"), input); err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (h *Handler) registerReceivablePayment(w http.ResponseWriter, r *http.Request) {
	var input models.RegisterFinancialPaymentInput
	if err := decodeJSON(r, &input); err != nil {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Invalid request body"))
		return
	}
	if err := h.Svc.RegisterReceivablePayment(r.Context(), chi.URLParam(r, "id"), input); err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (h *Handler) listPartnerSales(w http.ResponseWriter, r *http.Request) {
	page := httputil.ParsePagination(r)
	items, total, err := h.Svc.ListPartnerSales(r.Context(), queryParam(r, "salesperson_user_id"), queryParam(r, "status"), page)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WritePaged(w, items, total)
}

func (h *Handler) syncPartnerSales(w http.ResponseWriter, r *http.Request) {
	var input models.SyncPartnerSalesInput
	if err := decodeJSON(r, &input); err != nil {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Invalid request body"))
		return
	}
	result, err := h.Svc.SyncPartnerSales(r.Context(), input)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) updatePartnerSaleStatus(w http.ResponseWriter, r *http.Request) {
	var input models.UpdatePartnerSaleStatusInput
	if err := decodeJSON(r, &input); err != nil {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Invalid request body"))
		return
	}
	item, err := h.Svc.UpdatePartnerSaleStatus(r.Context(), chi.URLParam(r, "id"), input)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) getPartnerCommissionSettings(w http.ResponseWriter, r *http.Request) {
	item, err := h.Svc.GetPartnerCommissionSettings(r.Context())
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) updatePartnerCommissionSettings(w http.ResponseWriter, r *http.Request) {
	var input models.UpdatePartnerCommissionSettingsInput
	if err := decodeJSON(r, &input); err != nil {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Invalid request body"))
		return
	}
	item, err := h.Svc.UpdatePartnerCommissionSettings(r.Context(), input)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) partnerGetFinancialSummary(w http.ResponseWriter, r *http.Request) {
	summary, err := h.Svc.PartnerGetFinancialSummary(r.Context())
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, summary)
}

func (h *Handler) partnerListSales(w http.ResponseWriter, r *http.Request) {
	page := httputil.ParsePagination(r)
	items, total, err := h.Svc.PartnerListSales(r.Context(), page)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WritePaged(w, items, total)
}
