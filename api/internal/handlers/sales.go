package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/models"
	"github.com/luxus-connect/telefonia/api/internal/notifications"
)

func (h *Handler) listContractTemplates(w http.ResponseWriter, r *http.Request) {
	page := httputil.ParsePagination(r)
	activeOnly := r.URL.Query().Get("active_only") == "true"
	items, total, err := h.Svc.ListContractTemplates(r.Context(), activeOnly, page)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WritePaged(w, items, total)
}

func (h *Handler) getContractTemplate(w http.ResponseWriter, r *http.Request) {
	item, err := h.Svc.GetContractTemplate(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) createContractTemplate(w http.ResponseWriter, r *http.Request) {
	var input models.CreateContractTemplateInput
	if err := decodeJSON(r, &input); err != nil {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Invalid request body"))
		return
	}
	item, err := h.Svc.CreateContractTemplate(r.Context(), input)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, item)
}

func (h *Handler) updateContractTemplate(w http.ResponseWriter, r *http.Request) {
	var input models.UpdateContractTemplateInput
	if err := decodeJSON(r, &input); err != nil {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Invalid request body"))
		return
	}
	item, err := h.Svc.UpdateContractTemplate(r.Context(), chi.URLParam(r, "id"), input)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) listSales(w http.ResponseWriter, r *http.Request) {
	page := httputil.ParsePagination(r)
	items, total, err := h.Svc.ListSales(r.Context(), queryParam(r, "status"), page)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WritePaged(w, items, total)
}

func (h *Handler) getSale(w http.ResponseWriter, r *http.Request) {
	item, err := h.Svc.GetSale(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) createSale(w http.ResponseWriter, r *http.Request) {
	var input models.CreateSaleInput
	if err := decodeJSON(r, &input); err != nil {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Invalid request body"))
		return
	}
	item, err := h.Svc.CreateSale(r.Context(), input, false)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, item)
}

func (h *Handler) updateSale(w http.ResponseWriter, r *http.Request) {
	var input models.UpdateSaleInput
	if err := decodeJSON(r, &input); err != nil {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Invalid request body"))
		return
	}
	item, err := h.Svc.UpdateSale(r.Context(), chi.URLParam(r, "id"), input, false)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) addSaleLineItem(w http.ResponseWriter, r *http.Request) {
	var input models.AddSaleLineItemInput
	if err := decodeJSON(r, &input); err != nil {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Invalid request body"))
		return
	}
	item, err := h.Svc.AddSaleLineItem(r.Context(), chi.URLParam(r, "id"), input, false)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) deleteSaleLineItem(w http.ResponseWriter, r *http.Request) {
	item, err := h.Svc.DeleteSaleLineItem(r.Context(), chi.URLParam(r, "id"), chi.URLParam(r, "itemId"), false)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) confirmSale(w http.ResponseWriter, r *http.Request) {
	item, err := h.Svc.ConfirmSale(r.Context(), chi.URLParam(r, "id"), false)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) cancelSale(w http.ResponseWriter, r *http.Request) {
	item, err := h.Svc.CancelSale(r.Context(), chi.URLParam(r, "id"), false)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) partnerListCommercialSales(w http.ResponseWriter, r *http.Request) {
	page := httputil.ParsePagination(r)
	items, total, err := h.Svc.PartnerListCommercialSales(r.Context(), queryParam(r, "status"), page)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WritePaged(w, items, total)
}

func (h *Handler) partnerGetCommercialSale(w http.ResponseWriter, r *http.Request) {
	item, err := h.Svc.PartnerGetCommercialSale(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) partnerCreateCommercialSale(w http.ResponseWriter, r *http.Request) {
	var input models.CreateSaleInput
	if err := decodeJSON(r, &input); err != nil {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Invalid request body"))
		return
	}
	item, err := h.Svc.CreateSale(r.Context(), input, true)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, item)
}

func (h *Handler) partnerUpdateCommercialSale(w http.ResponseWriter, r *http.Request) {
	var input models.UpdateSaleInput
	if err := decodeJSON(r, &input); err != nil {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Invalid request body"))
		return
	}
	item, err := h.Svc.UpdateSale(r.Context(), chi.URLParam(r, "id"), input, true)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) partnerAddSaleLineItem(w http.ResponseWriter, r *http.Request) {
	var input models.AddSaleLineItemInput
	if err := decodeJSON(r, &input); err != nil {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Invalid request body"))
		return
	}
	item, err := h.Svc.AddSaleLineItem(r.Context(), chi.URLParam(r, "id"), input, true)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) partnerDeleteSaleLineItem(w http.ResponseWriter, r *http.Request) {
	item, err := h.Svc.DeleteSaleLineItem(r.Context(), chi.URLParam(r, "id"), chi.URLParam(r, "itemId"), true)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) partnerConfirmCommercialSale(w http.ResponseWriter, r *http.Request) {
	item, err := h.Svc.ConfirmSale(r.Context(), chi.URLParam(r, "id"), true)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) partnerCancelCommercialSale(w http.ResponseWriter, r *http.Request) {
	item, err := h.Svc.CancelSale(r.Context(), chi.URLParam(r, "id"), true)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) partnerListContractTemplates(w http.ResponseWriter, r *http.Request) {
	page := httputil.ParsePagination(r)
	items, total, err := h.Svc.PartnerListContractTemplates(r.Context(), page)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WritePaged(w, items, total)
}
