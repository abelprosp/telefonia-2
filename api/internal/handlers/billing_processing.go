package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/models"
	"github.com/luxus-connect/telefonia/api/internal/notifications"
)

func (h *Handler) listLineBillingProcessings(w http.ResponseWriter, r *http.Request) {
	item, err := h.Svc.ListLineBillingProcessings(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) enableEndUserBillingProcessing(w http.ResponseWriter, r *http.Request) {
	item, err := h.Svc.EnableEndUserProcessing(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) updateLineBillingProcessing(w http.ResponseWriter, r *http.Request) {
	var input models.UpdateLineBillingProcessingInput
	if err := decodeJSON(r, &input); err != nil {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Invalid request body"))
		return
	}
	item, err := h.Svc.UpdateLineBillingProcessing(r.Context(), chi.URLParam(r, "id"), chi.URLParam(r, "processingId"), input)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) mirrorLineBillingProcessing(w http.ResponseWriter, r *http.Request) {
	item, err := h.Svc.MirrorProcessingFromPrimary(r.Context(), chi.URLParam(r, "id"), chi.URLParam(r, "processingId"))
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) createLineBillingCompositionItem(w http.ResponseWriter, r *http.Request) {
	var input models.CreateLineBillingCompositionItemInput
	if err := decodeJSON(r, &input); err != nil {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Invalid request body"))
		return
	}
	item, err := h.Svc.CreateLineBillingCompositionItem(r.Context(), chi.URLParam(r, "id"), chi.URLParam(r, "processingId"), input)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, item)
}

func (h *Handler) updateLineBillingCompositionItem(w http.ResponseWriter, r *http.Request) {
	var input models.UpdateLineBillingCompositionItemInput
	if err := decodeJSON(r, &input); err != nil {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Invalid request body"))
		return
	}
	item, err := h.Svc.UpdateLineBillingCompositionItem(r.Context(), chi.URLParam(r, "id"), chi.URLParam(r, "processingId"), chi.URLParam(r, "itemId"), input)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) deleteLineBillingCompositionItem(w http.ResponseWriter, r *http.Request) {
	if err := h.Svc.DeleteLineBillingCompositionItem(r.Context(), chi.URLParam(r, "id"), chi.URLParam(r, "processingId"), chi.URLParam(r, "itemId")); err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) listLineBillingProcessingAudit(w http.ResponseWriter, r *http.Request) {
	items, err := h.Svc.ListLineBillingProcessingAudit(r.Context(), chi.URLParam(r, "id"), chi.URLParam(r, "processingId"))
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, items)
}
