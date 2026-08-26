package handlers

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/models"
	"github.com/luxus-connect/telefonia/api/internal/notifications"
)

func decodeAgentJSON(r *http.Request, v any) error {
	defer r.Body.Close()
	if r.Body == nil {
		return nil
	}
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(v); err != nil {
		if err == io.EOF {
			return nil
		}
		return err
	}
	return nil
}

func (h *Handler) agentFinancialHealth(w http.ResponseWriter, r *http.Request) {
	item, err := h.Svc.FinancialAgentHealth(r.Context())
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) agentLookupCustomer(w http.ResponseWriter, r *http.Request) {
	var input models.FinancialAgentLookupInput
	if err := decodeAgentJSON(r, &input); err != nil {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Invalid request body"))
		return
	}
	item, err := h.Svc.FinancialAgentLookupCustomer(r.Context(), input)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) agentListInvoices(w http.ResponseWriter, r *http.Request) {
	var input models.FinancialAgentListInvoicesInput
	if err := decodeAgentJSON(r, &input); err != nil {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Invalid request body"))
		return
	}
	item, err := h.Svc.FinancialAgentListInvoices(r.Context(), input)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) agentInvoicePackage(w http.ResponseWriter, r *http.Request) {
	var input models.FinancialAgentInvoicePackageInput
	if err := decodeAgentJSON(r, &input); err != nil {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Invalid request body"))
		return
	}
	item, err := h.Svc.FinancialAgentInvoicePackage(r.Context(), input)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) agentCheckDelinquency(w http.ResponseWriter, r *http.Request) {
	var input models.FinancialAgentDelinquencyInput
	if err := decodeAgentJSON(r, &input); err != nil {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Invalid request body"))
		return
	}
	item, err := h.Svc.FinancialAgentCheckDelinquency(r.Context(), input)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) agentCalculateInterest(w http.ResponseWriter, r *http.Request) {
	var input models.FinancialAgentInterestInput
	if err := decodeAgentJSON(r, &input); err != nil {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Invalid request body"))
		return
	}
	item, err := h.Svc.FinancialAgentCalculateInterest(r.Context(), input)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) agentSyncSicredi(w http.ResponseWriter, r *http.Request) {
	var input models.FinancialAgentSyncSicrediInput
	_ = decodeAgentJSON(r, &input)
	item, err := h.Svc.FinancialAgentSyncSicredi(r.Context(), input)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) agentVerifyReceipt(w http.ResponseWriter, r *http.Request) {
	var input models.FinancialAgentVerifyReceiptInput
	if err := decodeAgentJSON(r, &input); err != nil {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Invalid request body"))
		return
	}
	item, err := h.Svc.FinancialAgentVerifyReceipt(r.Context(), input)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) agentConfirmPayment(w http.ResponseWriter, r *http.Request) {
	var input models.FinancialAgentConfirmPaymentInput
	if err := decodeAgentJSON(r, &input); err != nil {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Invalid request body"))
		return
	}
	item, err := h.Svc.FinancialAgentConfirmPayment(r.Context(), input)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) agentBoletoPDF(w http.ResponseWriter, r *http.Request) {
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

func (h *Handler) agentInvoiceDownload(w http.ResponseWriter, r *http.Request) {
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
