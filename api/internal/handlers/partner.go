package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/models"
	"github.com/luxus-connect/telefonia/api/internal/notifications"
)

func (h *Handler) partnerGetDashboardStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.Svc.PartnerGetDashboardStats(r.Context())
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, stats)
}

func (h *Handler) partnerListProviders(w http.ResponseWriter, r *http.Request) {
	items, total, err := h.Svc.PartnerListProviders(r.Context())
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WritePaged(w, items, total)
}

func (h *Handler) partnerListCustomers(w http.ResponseWriter, r *http.Request) {
	page := httputil.ParsePagination(r)
	items, total, err := h.Svc.PartnerListCustomers(r.Context(), queryParam(r, "provider_id"), page)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WritePaged(w, items, total)
}

func (h *Handler) partnerGetCustomer(w http.ResponseWriter, r *http.Request) {
	item, err := h.Svc.PartnerGetCustomer(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) partnerCreateCustomer(w http.ResponseWriter, r *http.Request) {
	var input models.CreateCustomerInput
	if err := decodeJSON(r, &input); err != nil {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Invalid request body"))
		return
	}
	item, err := h.Svc.PartnerCreateCustomer(r.Context(), input)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, item)
}

func (h *Handler) partnerUpdateCustomer(w http.ResponseWriter, r *http.Request) {
	var input models.UpdateCustomerInput
	if err := decodeJSON(r, &input); err != nil {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Invalid request body"))
		return
	}
	if err := h.Svc.PartnerUpdateCustomer(r.Context(), chi.URLParam(r, "id"), input); err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (h *Handler) partnerListCustomerPhoneLines(w http.ResponseWriter, r *http.Request) {
	page := httputil.ParsePagination(r)
	items, total, err := h.Svc.PartnerListCustomerPhoneLines(r.Context(), chi.URLParam(r, "id"), page)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WritePaged(w, items, total)
}

func (h *Handler) partnerListPhoneLines(w http.ResponseWriter, r *http.Request) {
	page := httputil.ParsePagination(r)
	items, total, err := h.Svc.PartnerListPhoneLines(r.Context(), page)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WritePaged(w, items, total)
}

func (h *Handler) partnerCreateLineOperationRequest(w http.ResponseWriter, r *http.Request) {
	var input models.CreatePhoneLineOperationRequestInput
	if err := decodeJSON(r, &input); err != nil {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Invalid request body"))
		return
	}
	item, err := h.Svc.PartnerCreateLineOperationRequest(r.Context(), input)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, item)
}

func (h *Handler) partnerListLineOperationRequests(w http.ResponseWriter, r *http.Request) {
	page := httputil.ParsePagination(r)
	items, total, err := h.Svc.PartnerListLineOperationRequests(r.Context(), page)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WritePaged(w, items, total)
}

func (h *Handler) listLineOperationRequests(w http.ResponseWriter, r *http.Request) {
	page := httputil.ParsePagination(r)
	items, total, err := h.Svc.ListLineOperationRequests(r.Context(), page)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WritePaged(w, items, total)
}

func (h *Handler) reviewLineOperationRequest(w http.ResponseWriter, r *http.Request) {
	var input models.ReviewPhoneLineOperationRequestInput
	if err := decodeJSON(r, &input); err != nil {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Invalid request body"))
		return
	}
	item, err := h.Svc.ReviewLineOperationRequest(r.Context(), chi.URLParam(r, "id"), input)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}
