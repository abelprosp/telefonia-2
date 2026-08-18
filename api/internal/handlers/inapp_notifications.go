package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/luxus-connect/telefonia/api/internal/auth"
	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/notifications"
)

func (h *Handler) ListInAppNotifications(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	userID := "anonymous"
	if user != nil {
		userID = user.ID
	}

	res, err := h.Svc.ListInAppNotifications(r.Context(), userID)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, res)
}

func (h *Handler) MarkNotificationAsRead(w http.ResponseWriter, r *http.Request) {
	notifID := chi.URLParam(r, "id")
	if notifID == "" {
		httputil.WriteFail(w, http.StatusBadRequest, notifications.N("VALIDATION_ERROR", "Notification ID is required"))
		return
	}

	user := auth.UserFromContext(r.Context())
	userID := "anonymous"
	if user != nil {
		userID = user.ID
	}

	if err := h.Svc.MarkNotificationAsRead(r.Context(), userID, notifID); err != nil {
		httputil.HandleServiceError(w, err)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]bool{"success": true})
}

func (h *Handler) MarkAllNotificationsAsRead(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	userID := "anonymous"
	if user != nil {
		userID = user.ID
	}

	if err := h.Svc.MarkAllNotificationsAsRead(r.Context(), userID); err != nil {
		httputil.HandleServiceError(w, err)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]bool{"success": true})
}
