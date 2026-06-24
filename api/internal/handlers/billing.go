package handlers

import (
	"net/http"
	"strconv"
	"strings"

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

func (h *Handler) downloadCustomerBillingDocument(w http.ResponseWriter, r *http.Request) {
	html, filename, err := h.Svc.GetCustomerBillingDocumentDownload(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(html)
}

func (h *Handler) bulkBillingPreview(w http.ResponseWriter, r *http.Request) {
	monthID := ""
	if v := queryParam(r, "processing_month_id"); v != nil {
		monthID = *v
	}
	var customerIDs []string
	if v := queryParam(r, "customer_ids"); v != nil && *v != "" {
		customerIDs = strings.Split(*v, ",")
	}
	item, err := h.Svc.BulkBillingPreview(r.Context(), monthID, customerIDs)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) manualBillingPreview(w http.ResponseWriter, r *http.Request) {
	var customerIDs []string
	if v := queryParam(r, "customer_ids"); v != nil && *v != "" {
		customerIDs = strings.Split(*v, ",")
	}
	item, err := h.Svc.ManualBillingPreview(r.Context(), customerIDs)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) manualGenerateBillingDocuments(w http.ResponseWriter, r *http.Request) {
	var input models.ManualGenerateBillingDocumentsInput
	if err := decodeJSON(r, &input); err != nil {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Invalid request body"))
		return
	}
	item, err := h.Svc.ManualGenerateBillingDocuments(r.Context(), input)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) generateCustomerBillingDocument(w http.ResponseWriter, r *http.Request) {
	var input models.GenerateCustomerBillingDocumentInput
	if err := decodeJSON(r, &input); err != nil {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Invalid request body"))
		return
	}
	item, err := h.Svc.GenerateCustomerBillingDocument(r.Context(), chi.URLParam(r, "id"), input)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, item)
}

func (h *Handler) bulkGenerateBillingDocuments(w http.ResponseWriter, r *http.Request) {
	var input models.BulkGenerateBillingDocumentsInput
	if err := decodeJSON(r, &input); err != nil {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Invalid request body"))
		return
	}
	item, err := h.Svc.BulkGenerateBillingDocuments(r.Context(), input)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) createCustomerBillingDocumentFromReceivable(w http.ResponseWriter, r *http.Request) {
	var input struct {
		TemplateCode       string `json:"template_code"`
		LayoutTemplateCode string `json:"layout_template_code"`
	}
	if err := decodeJSON(r, &input); err != nil {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Invalid request body"))
		return
	}
	item, err := h.Svc.CreateCustomerBillingDocumentFromReceivable(r.Context(), chi.URLParam(r, "receivableId"), input.TemplateCode, input.LayoutTemplateCode)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, item)
}

func (h *Handler) issueSicrediBoleto(w http.ResponseWriter, r *http.Request) {
	item, err := h.Svc.IssueSicrediBoleto(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
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

func (h *Handler) syncSicrediPayment(w http.ResponseWriter, r *http.Request) {
	item, err := h.Svc.SyncSicrediPaymentForDocument(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) syncSicrediPayments(w http.ResponseWriter, r *http.Request) {
	daysBack := 7
	if v := queryParam(r, "days_back"); v != nil && *v != "" {
		if n, err := strconv.Atoi(*v); err == nil && n > 0 {
			daysBack = n
		}
	}
	item, err := h.Svc.SyncSicrediPayments(r.Context(), daysBack)
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

func (h *Handler) listInvoiceLayoutTemplates(w http.ResponseWriter, r *http.Request) {
	page := httputil.ParsePagination(r)
	items, total, err := h.Svc.ListInvoiceLayoutTemplates(r.Context(), page)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WritePaged(w, items, total)
}

func (h *Handler) getInvoiceLayoutTemplate(w http.ResponseWriter, r *http.Request) {
	item, err := h.Svc.GetInvoiceLayoutTemplate(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) createInvoiceLayoutTemplate(w http.ResponseWriter, r *http.Request) {
	var input models.CreateInvoiceLayoutTemplateInput
	if err := decodeJSON(r, &input); err != nil {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Invalid request body"))
		return
	}
	item, err := h.Svc.CreateInvoiceLayoutTemplate(r.Context(), input)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, item)
}

func (h *Handler) updateInvoiceLayoutTemplate(w http.ResponseWriter, r *http.Request) {
	var input models.UpdateInvoiceLayoutTemplateInput
	if err := decodeJSON(r, &input); err != nil {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Invalid request body"))
		return
	}
	item, err := h.Svc.UpdateInvoiceLayoutTemplate(r.Context(), chi.URLParam(r, "id"), input)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) previewInvoiceLayout(w http.ResponseWriter, r *http.Request) {
	var input models.PreviewInvoiceLayoutInput
	if err := decodeJSON(r, &input); err != nil {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Invalid request body"))
		return
	}
	item, err := h.Svc.PreviewInvoiceLayout(r.Context(), input)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}
