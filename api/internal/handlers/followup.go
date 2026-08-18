package handlers

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/models"
	"github.com/luxus-connect/telefonia/api/internal/notifications"
)

func (h *Handler) listApprovals(w http.ResponseWriter, r *http.Request) {
	items, err := h.Svc.ListApprovalRequests(r.Context(), r.URL.Query().Get("status"))
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, items)
}

func (h *Handler) createApproval(w http.ResponseWriter, r *http.Request) {
	var input models.CreateApprovalRequestInput
	if err := decodeJSON(r, &input); err != nil {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Invalid request body"))
		return
	}
	item, err := h.Svc.RequestTwoLevelApproval(r.Context(), input)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, item)
}

func (h *Handler) approveApproval(w http.ResponseWriter, r *http.Request) {
	item, err := h.Svc.ApproveRequest(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) rejectApproval(w http.ResponseWriter, r *http.Request) {
	var input models.RejectApprovalInput
	_ = decodeJSON(r, &input)
	item, err := h.Svc.RejectApprovalRequest(r.Context(), chi.URLParam(r, "id"), input.Reason)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) listSupportTickets(w http.ResponseWriter, r *http.Request) {
	page := httputil.ParsePagination(r)
	items, total, err := h.Svc.ListSupportTickets(r.Context(), r.URL.Query().Get("customer_id"), r.URL.Query().Get("status"), page)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WritePaged(w, items, total)
}

func (h *Handler) createSupportTicket(w http.ResponseWriter, r *http.Request) {
	var input models.CreateSupportTicketInput
	if err := decodeJSON(r, &input); err != nil {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Invalid request body"))
		return
	}
	item, err := h.Svc.CreateSupportTicket(r.Context(), input)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, item)
}

func (h *Handler) getSupportTicket(w http.ResponseWriter, r *http.Request) {
	item, err := h.Svc.GetSupportTicket(r.Context(), chi.URLParam(r, "id"), true)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) updateSupportTicket(w http.ResponseWriter, r *http.Request) {
	var input models.UpdateSupportTicketInput
	if err := decodeJSON(r, &input); err != nil {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Invalid request body"))
		return
	}
	item, err := h.Svc.UpdateSupportTicket(r.Context(), chi.URLParam(r, "id"), input)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) addSupportTicketMessage(w http.ResponseWriter, r *http.Request) {
	var input models.AddSupportTicketMessageInput
	if err := decodeJSON(r, &input); err != nil {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Invalid request body"))
		return
	}
	item, err := h.Svc.AddSupportTicketMessage(r.Context(), chi.URLParam(r, "id"), input)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) ticketAttachmentUploadURL(w http.ResponseWriter, r *http.Request) {
	h.writeTicketAttachmentUploadURL(w, r, false)
}

func (h *Handler) portalTicketAttachmentUploadURL(w http.ResponseWriter, r *http.Request) {
	h.writeTicketAttachmentUploadURL(w, r, true)
}

func (h *Handler) writeTicketAttachmentUploadURL(w http.ResponseWriter, r *http.Request, portal bool) {
	if h.Presigned == nil {
		httputil.WriteFail(w, http.StatusServiceUnavailable, notifications.ObjectStorageUnavailable)
		return
	}
	var input models.TicketAttachmentUploadInput
	if err := decodeJSON(r, &input); err != nil {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Invalid request body"))
		return
	}
	ticketID := chi.URLParam(r, "id")
	key, err := h.Svc.PrepareTicketAttachmentKey(r.Context(), ticketID, input, portal)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	result, err := h.Presigned.CreateUploadURL(r.Context(), models.CreatePresignedUploadURLInput{
		BucketName:       input.BucketName,
		ObjectKey:        key,
		ContentType:      input.ContentType,
		ExpiresInSeconds: input.ExpiresInSeconds,
	})
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, models.TicketAttachmentUploadResponse{
		PresignedURLModel: *result,
		ObjectKey:         key,
		BucketName:        strings.TrimSpace(input.BucketName),
	})
}

func (h *Handler) ticketAttachmentDownload(w http.ResponseWriter, r *http.Request) {
	h.writeTicketAttachmentDownload(w, r, false)
}

func (h *Handler) portalTicketAttachmentDownload(w http.ResponseWriter, r *http.Request) {
	h.writeTicketAttachmentDownload(w, r, true)
}

func (h *Handler) writeTicketAttachmentDownload(w http.ResponseWriter, r *http.Request, portal bool) {
	if h.Presigned == nil {
		httputil.WriteFail(w, http.StatusServiceUnavailable, notifications.ObjectStorageUnavailable)
		return
	}
	msg, err := h.Svc.TicketAttachmentLocation(r.Context(), chi.URLParam(r, "id"), chi.URLParam(r, "messageId"), portal)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	bucket := ""
	if msg.AttachmentBucket != nil {
		bucket = strings.TrimSpace(*msg.AttachmentBucket)
	}
	if bucket == "" {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.ImportStorageBucketRequired)
		return
	}
	result, err := h.Presigned.CreateDownloadURL(r.Context(), models.CreatePresignedDownloadURLInput{
		BucketName: bucket,
		ObjectKey:  strings.TrimSpace(*msg.AttachmentKey),
	})
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	name := ""
	if msg.AttachmentName != nil {
		name = *msg.AttachmentName
	}
	ct := ""
	if msg.AttachmentContentType != nil {
		ct = *msg.AttachmentContentType
	}
	httputil.WriteJSON(w, http.StatusOK, models.TicketAttachmentDownloadResponse{
		PresignedURLModel: *result,
		FileName:          name,
		ContentType:       ct,
	})
}

func (h *Handler) getDivergence(w http.ResponseWriter, r *http.Request) {
	item, err := h.Svc.GetDivergence(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) assignDivergence(w http.ResponseWriter, r *http.Request) {
	var input models.AssignDivergenceInput
	if err := decodeJSON(r, &input); err != nil {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Invalid request body"))
		return
	}
	item, err := h.Svc.AssignDivergence(r.Context(), chi.URLParam(r, "id"), input.OwnerUserID)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) commentDivergence(w http.ResponseWriter, r *http.Request) {
	var input models.CommentDivergenceInput
	if err := decodeJSON(r, &input); err != nil {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Invalid request body"))
		return
	}
	item, err := h.Svc.CommentDivergence(r.Context(), chi.URLParam(r, "id"), input.Body)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) runProcessingMonthPipeline(w http.ResponseWriter, r *http.Request) {
	item, err := h.Svc.RunProcessingMonthPipeline(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusAccepted, item)
}

func (h *Handler) listProcessingMonthRuns(w http.ResponseWriter, r *http.Request) {
	items, err := h.Svc.ListProcessingMonthRuns(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, items)
}

func (h *Handler) pushFinancialExportSFTP(w http.ResponseWriter, r *http.Request) {
	item, err := h.Svc.PushFinancialExportSFTP(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) portalListContracts(w http.ResponseWriter, r *http.Request) {
	items, err := h.Svc.PortalListContracts(r.Context())
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, items)
}

func (h *Handler) portalUpdateProfile(w http.ResponseWriter, r *http.Request) {
	var input models.PortalUpdateProfileInput
	if err := decodeJSON(r, &input); err != nil {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Invalid request body"))
		return
	}
	item, err := h.Svc.PortalUpdateProfile(r.Context(), input)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) portalListTickets(w http.ResponseWriter, r *http.Request) {
	page := httputil.ParsePagination(r)
	items, total, err := h.Svc.PortalListTickets(r.Context(), page)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WritePaged(w, items, total)
}

func (h *Handler) portalCreateTicket(w http.ResponseWriter, r *http.Request) {
	var input models.CreateSupportTicketInput
	if err := decodeJSON(r, &input); err != nil {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Invalid request body"))
		return
	}
	item, err := h.Svc.PortalCreateTicket(r.Context(), input)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, item)
}

func (h *Handler) portalGetTicket(w http.ResponseWriter, r *http.Request) {
	item, err := h.Svc.PortalGetSupportTicket(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) portalAddTicketMessage(w http.ResponseWriter, r *http.Request) {
	var input models.AddSupportTicketMessageInput
	if err := decodeJSON(r, &input); err != nil {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("REQUEST_VALIDATION", "Invalid request body"))
		return
	}
	item, err := h.Svc.PortalAddSupportTicketMessage(r.Context(), chi.URLParam(r, "id"), input)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) portalDownloadInvoice(w http.ResponseWriter, r *http.Request) {
	html, filename, err := h.Svc.PortalDownloadInvoice(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(html)
}
