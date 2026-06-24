package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/models"
	"github.com/luxus-connect/telefonia/api/internal/notifications"
)

func (h *Handler) listCustomers(w http.ResponseWriter, r *http.Request) {
	page := httputil.ParsePagination(r)
	items, total, err := h.Svc.ListCustomers(r.Context(), queryParam(r, "provider_id"), page)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WritePaged(w, items, total)
}

func (h *Handler) getCustomer(w http.ResponseWriter, r *http.Request) {
	item, err := h.Svc.GetCustomer(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) createCustomer(w http.ResponseWriter, r *http.Request) {
	var input models.CreateCustomerInput
	if err := decodeJSON(r, &input); err != nil {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Invalid request body"))
		return
	}
	item, err := h.Svc.CreateCustomer(r.Context(), input)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, item)
}

func (h *Handler) updateCustomer(w http.ResponseWriter, r *http.Request) {
	var input models.UpdateCustomerInput
	if err := decodeJSON(r, &input); err != nil {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Invalid request body"))
		return
	}
	if err := h.Svc.UpdateCustomer(r.Context(), chi.URLParam(r, "id"), input); err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (h *Handler) inactivateCustomer(w http.ResponseWriter, r *http.Request) {
	if err := h.Svc.InactivateCustomer(r.Context(), chi.URLParam(r, "id")); err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (h *Handler) listCustomerProviderLinks(w http.ResponseWriter, r *http.Request) {
	items, err := h.Svc.ListCustomerProviderLinks(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, items)
}

func (h *Handler) listCustomerPhoneLines(w http.ResponseWriter, r *http.Request) {
	page := httputil.ParsePagination(r)
	items, total, err := h.Svc.ListCustomerPhoneLines(r.Context(), chi.URLParam(r, "id"), page)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WritePaged(w, items, total)
}

func (h *Handler) listCustomerDevices(w http.ResponseWriter, r *http.Request) {
	page := httputil.ParsePagination(r)
	items, total, err := h.Svc.ListCustomerDevices(r.Context(), chi.URLParam(r, "id"), page)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WritePaged(w, items, total)
}

func (h *Handler) assignCustomerDevice(w http.ResponseWriter, r *http.Request) {
	var input models.AssignCustomerDeviceInput
	if err := decodeJSON(r, &input); err != nil {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Invalid request body"))
		return
	}
	item, err := h.Svc.AssignCustomerDevice(r.Context(), chi.URLParam(r, "id"), input)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, item)
}

func (h *Handler) updateCustomerDevice(w http.ResponseWriter, r *http.Request) {
	var input models.UpdateCustomerDeviceInput
	if err := decodeJSON(r, &input); err != nil {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Invalid request body"))
		return
	}
	item, err := h.Svc.UpdateCustomerDevice(r.Context(), chi.URLParam(r, "id"), chi.URLParam(r, "deviceLinkId"), input)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) unassignCustomerDevice(w http.ResponseWriter, r *http.Request) {
	var input models.UnassignCustomerDeviceInput
	if err := decodeJSON(r, &input); err != nil {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Invalid request body"))
		return
	}
	if err := h.Svc.UnassignCustomerDevice(r.Context(), chi.URLParam(r, "id"), chi.URLParam(r, "deviceLinkId"), input); err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (h *Handler) listCustomerAttachments(w http.ResponseWriter, r *http.Request) {
	items, err := h.Svc.ListCustomerAttachments(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, items)
}

func (h *Handler) createCustomerAttachment(w http.ResponseWriter, r *http.Request) {
	var input models.RegisterCustomerAttachmentInput
	if err := decodeJSON(r, &input); err != nil {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Invalid request body"))
		return
	}
	item, err := h.Svc.CreateCustomerAttachment(r.Context(), chi.URLParam(r, "id"), input)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, item)
}

func (h *Handler) deleteCustomerAttachment(w http.ResponseWriter, r *http.Request) {
	if err := h.Svc.DeleteCustomerAttachment(r.Context(), chi.URLParam(r, "id"), chi.URLParam(r, "attachmentId")); err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (h *Handler) getBillingReadiness(w http.ResponseWriter, r *http.Request) {
	item, err := h.Svc.GetBillingReadiness(r.Context(), chi.URLParam(r, "id"), chi.URLParam(r, "processingMonthId"))
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) manualReleaseCustomer(w http.ResponseWriter, r *http.Request) {
	var input models.ManuallyReleaseCustomerInput
	if err := decodeJSON(r, &input); err != nil {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Invalid request body"))
		return
	}
	item, err := h.Svc.ManualReleaseCustomer(r.Context(), chi.URLParam(r, "id"), chi.URLParam(r, "processingMonthId"), input)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}
