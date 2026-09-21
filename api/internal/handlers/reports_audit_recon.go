package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/models"
	"github.com/luxus-connect/telefonia/api/internal/notifications"
	"github.com/luxus-connect/telefonia/api/internal/services"
	"github.com/luxus-connect/telefonia/api/internal/store"
)

func (h *Handler) listDomainAuditEvents(w http.ResponseWriter, r *http.Request) {
	page := httputil.ParsePagination(r)
	var from, to *time.Time
	if v := r.URL.Query().Get("from"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			from = &t
		}
	}
	if v := r.URL.Query().Get("to"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			to = &t
		}
	}
	items, total, err := h.Svc.ListDomainAuditEvents(
		r.Context(),
		r.URL.Query().Get("entity_type"),
		r.URL.Query().Get("entity_id"),
		r.URL.Query().Get("action"),
		r.URL.Query().Get("actor_user_id"),
		r.URL.Query().Get("phone_line_id"),
		r.URL.Query().Get("customer_id"),
		from, to, page,
	)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WritePaged(w, items, total)
}

func (h *Handler) listLineConsumptionReport(w http.ResponseWriter, r *http.Request) {
	page := httputil.ParsePagination(r)
	f := store.ConsumptionReportFilter{
		ProcessingMonthID: r.URL.Query().Get("processing_month_id"),
		CustomerID:        r.URL.Query().Get("customer_id"),
		PhoneLineID:       r.URL.Query().Get("phone_line_id"),
		ProviderID:        r.URL.Query().Get("provider_id"),
		Status:            r.URL.Query().Get("status"),
	}
	resp, err := h.Svc.ListLineConsumptionReport(r.Context(), f, page)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) exportLineConsumptionReport(w http.ResponseWriter, r *http.Request) {
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "txt"
	}
	f := store.ConsumptionReportFilter{
		ProcessingMonthID: r.URL.Query().Get("processing_month_id"),
		CustomerID:        r.URL.Query().Get("customer_id"),
		PhoneLineID:       r.URL.Query().Get("phone_line_id"),
		ProviderID:        r.URL.Query().Get("provider_id"),
		Status:            r.URL.Query().Get("status"),
	}
	data, contentType, err := h.Svc.ExportLineConsumptionReport(r.Context(), f, format)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	filename := "consumo-linhas." + format
	if format == "xlsx" {
		filename = "consumo-linhas.xlsx"
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", "attachment; filename="+strconv.Quote(filename))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func (h *Handler) listReconciliationMissing(w http.ResponseWriter, r *http.Request) {
	partial, _ := strconv.ParseBool(r.URL.Query().Get("is_partial_source"))
	items, err := h.Svc.ListMissingInOperator(
		r.Context(),
		r.URL.Query().Get("source"),
		r.URL.Query().Get("provider_id"),
		r.URL.Query().Get("import_job_id"),
		partial,
	)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, items)
}

func (h *Handler) listReconciliationCancelledExternal(w http.ResponseWriter, r *http.Request) {
	items, err := h.Svc.ListCancelledExternallyActiveInternally(
		r.Context(),
		r.URL.Query().Get("source"),
		r.URL.Query().Get("import_job_id"),
	)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, items)
}

func (h *Handler) createExternalLineImportJob(w http.ResponseWriter, r *http.Request) {
	var input models.ExternalLineImportJobInput
	if err := decodeJSON(r, &input); err != nil {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Invalid request body"))
		return
	}
	item, err := h.Svc.RegisterExternalLineImportJob(r.Context(), input)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusAccepted, item)
}

func (h *Handler) applyCancelledExternallyActive(w http.ResponseWriter, r *http.Request) {
	var input services.ApplyCancelledExternallyInput
	if err := decodeJSON(r, &input); err != nil {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Invalid request body"))
		return
	}
	item, err := h.Svc.ApplyCancelledExternallyActive(r.Context(), input)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}
