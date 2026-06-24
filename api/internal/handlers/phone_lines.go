package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/models"
	"github.com/luxus-connect/telefonia/api/internal/notifications"
)

func (h *Handler) listPhoneLines(w http.ResponseWriter, r *http.Request) {
	page := httputil.ParsePagination(r)
	items, total, err := h.Svc.ListPhoneLines(r.Context(), queryParam(r, "status"), page)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WritePaged(w, items, total)
}

func (h *Handler) createStockPhoneLine(w http.ResponseWriter, r *http.Request) {
	var input models.CreateStockPhoneLineInput
	if err := decodeJSON(r, &input); err != nil {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Invalid request body"))
		return
	}
	item, err := h.Svc.CreateStockPhoneLine(r.Context(), input)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, item)
}

func (h *Handler) getPhoneLine(w http.ResponseWriter, r *http.Request) {
	item, err := h.Svc.GetPhoneLine(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) listPhoneLineCustomerLinks(w http.ResponseWriter, r *http.Request) {
	items, err := h.Svc.ListPhoneLineCustomerLinks(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, items)
}

func (h *Handler) assignPhoneLineCustomer(w http.ResponseWriter, r *http.Request) {
	var input models.AssignPhoneLineCustomerInput
	if err := decodeJSON(r, &input); err != nil {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Invalid request body"))
		return
	}
	item, err := h.Svc.AssignPhoneLineCustomer(r.Context(), chi.URLParam(r, "id"), input)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) transferPhoneLineCustomer(w http.ResponseWriter, r *http.Request) {
	var input models.TransferPhoneLineCustomerInput
	if err := decodeJSON(r, &input); err != nil {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Invalid request body"))
		return
	}
	item, err := h.Svc.TransferPhoneLineCustomer(r.Context(), chi.URLParam(r, "id"), input)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) unassignPhoneLineCustomer(w http.ResponseWriter, r *http.Request) {
	var input models.UnassignPhoneLineCustomerInput
	if err := decodeJSON(r, &input); err != nil {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Invalid request body"))
		return
	}
	if err := h.Svc.UnassignPhoneLineCustomer(r.Context(), chi.URLParam(r, "id"), input); err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (h *Handler) updateActivePhoneLineCustomerLink(w http.ResponseWriter, r *http.Request) {
	var input models.UpdateActivePhoneLineCustomerLinkInput
	if err := decodeJSON(r, &input); err != nil {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Invalid request body"))
		return
	}
	item, err := h.Svc.UpdateActivePhoneLineCustomerLink(r.Context(), chi.URLParam(r, "id"), input)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) listBillingCycles(w http.ResponseWriter, r *http.Request) {
	page := httputil.ParsePagination(r)
	items, total, err := h.Svc.ListBillingCycles(r.Context(), page)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WritePaged(w, items, total)
}

func (h *Handler) getBillingCycle(w http.ResponseWriter, r *http.Request) {
	item, err := h.Svc.GetBillingCycle(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) createBillingCycle(w http.ResponseWriter, r *http.Request) {
	var input models.CreateBillingCycleInput
	if err := decodeJSON(r, &input); err != nil {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Invalid request body"))
		return
	}
	item, err := h.Svc.CreateBillingCycle(r.Context(), input)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, item)
}

func (h *Handler) updateBillingCycle(w http.ResponseWriter, r *http.Request) {
	var input models.UpdateBillingCycleInput
	if err := decodeJSON(r, &input); err != nil {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Invalid request body"))
		return
	}
	if err := h.Svc.UpdateBillingCycle(r.Context(), chi.URLParam(r, "id"), input); err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (h *Handler) listProcessingMonths(w http.ResponseWriter, r *http.Request) {
	page := httputil.ParsePagination(r)
	items, total, err := h.Svc.ListProcessingMonths(r.Context(), page)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WritePaged(w, items, total)
}

func (h *Handler) getProcessingMonth(w http.ResponseWriter, r *http.Request) {
	item, err := h.Svc.GetProcessingMonth(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) createProcessingMonth(w http.ResponseWriter, r *http.Request) {
	var input models.CreateProcessingMonthInput
	if err := decodeJSON(r, &input); err != nil {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Invalid request body"))
		return
	}
	item, err := h.Svc.CreateProcessingMonth(r.Context(), input)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, item)
}

func (h *Handler) closeProcessingMonth(w http.ResponseWriter, r *http.Request) {
	item, err := h.Svc.CloseProcessingMonth(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusAccepted, item)
}

func (h *Handler) closeProcessingMonthContingency(w http.ResponseWriter, r *http.Request) {
	var input models.CloseProcessingMonthContingencyInput
	if err := decodeJSON(r, &input); err != nil {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Invalid request body"))
		return
	}
	item, err := h.Svc.CloseProcessingMonthContingency(r.Context(), chi.URLParam(r, "id"), input)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusAccepted, item)
}

func (h *Handler) listCostCenters(w http.ResponseWriter, r *http.Request) {
	page := httputil.ParsePagination(r)
	items, total, err := h.Svc.ListCostCenters(r.Context(), page)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WritePaged(w, items, total)
}
