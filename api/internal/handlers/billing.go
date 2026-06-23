package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/models"
	"github.com/luxus-connect/telefonia/api/internal/notifications"
)

func (h *Handler) listInvoiceEmailTemplates(w http.ResponseWriter, r *http.Request) {
	page := httputil.ParsePagination(r)
	items, total, err := h.Svc.ListInvoiceEmailTemplates(r.Context(), queryParam(r, "kind"), page)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WritePaged(w, items, total)
}

func (h *Handler) getInvoiceEmailTemplate(w http.ResponseWriter, r *http.Request) {
	item, err := h.Svc.GetInvoiceEmailTemplate(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) createInvoiceEmailTemplate(w http.ResponseWriter, r *http.Request) {
	var input models.CreateInvoiceEmailTemplateInput
	if err := decodeJSON(r, &input); err != nil {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Invalid request body"))
		return
	}
	item, err := h.Svc.CreateInvoiceEmailTemplate(r.Context(), input)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, item)
}

func (h *Handler) updateInvoiceEmailTemplate(w http.ResponseWriter, r *http.Request) {
	var input models.UpdateInvoiceEmailTemplateInput
	if err := decodeJSON(r, &input); err != nil {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Invalid request body"))
		return
	}
	item, err := h.Svc.UpdateInvoiceEmailTemplate(r.Context(), chi.URLParam(r, "id"), input)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) listCustomerBillingDocuments(w http.ResponseWriter, r *http.Request) {
	page := httputil.ParsePagination(r)
	items, total, err := h.Svc.ListCustomerBillingDocuments(r.Context(), queryParam(r, "status"), queryParam(r, "customer_id"), page)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WritePaged(w, items, total)
}

func (h *Handler) getCustomerBillingDocument(w http.ResponseWriter, r *http.Request) {
	item, err := h.Svc.GetCustomerBillingDocument(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) createCustomerBillingDocumentFromReceivable(w http.ResponseWriter, r *http.Request) {
	var input struct {
		TemplateCode string `json:"template_code"`
	}
	if err := decodeJSON(r, &input); err != nil {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Invalid request body"))
		return
	}
	item, err := h.Svc.CreateCustomerBillingDocumentFromReceivable(r.Context(), chi.URLParam(r, "receivableId"), input.TemplateCode)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, item)
}

func (h *Handler) updateCustomerBillingDocument(w http.ResponseWriter, r *http.Request) {
	var input models.UpdateCustomerBillingDocumentInput
	if err := decodeJSON(r, &input); err != nil {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Invalid request body"))
		return
	}
	item, err := h.Svc.UpdateCustomerBillingDocument(r.Context(), chi.URLParam(r, "id"), input)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) sendCustomerBillingDocument(w http.ResponseWriter, r *http.Request) {
	result, err := h.Svc.SendCustomerBillingDocument(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) listCustomerBillingSendLog(w http.ResponseWriter, r *http.Request) {
	items, err := h.Svc.ListCustomerBillingSendLog(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, items)
}

func (h *Handler) listOverdueReceivables(w http.ResponseWriter, r *http.Request) {
	page := httputil.ParsePagination(r)
	items, total, err := h.Svc.ListOverdueReceivables(r.Context(), page)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WritePaged(w, items, total)
}

func (h *Handler) sendCollectionReminder(w http.ResponseWriter, r *http.Request) {
	var input models.SendCollectionReminderInput
	if err := decodeJSON(r, &input); err != nil {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Invalid request body"))
		return
	}
	result, err := h.Svc.SendCollectionReminder(r.Context(), input)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, result)
}
