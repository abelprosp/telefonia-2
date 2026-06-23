package httputil

import (
	"encoding/json"
	"net/http"

	"github.com/luxus-connect/telefonia/api/internal/notifications"
)

type MessageResponse struct {
	Messages []notifications.Notification `json:"messages"`
}

type PagedResponse[T any] struct {
	Items      []T   `json:"items"`
	TotalCount int64 `json:"total_count"`
}

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func WriteError(w http.ResponseWriter, err *AppError) {
	WriteJSON(w, err.StatusCode(), MessageResponse{Messages: err.Notifications})
}

func WriteFail(w http.ResponseWriter, status int, n ...notifications.Notification) {
	WriteJSON(w, status, MessageResponse{Messages: n})
}

func WritePaged[T any](w http.ResponseWriter, items []T, total int64) {
	if items == nil {
		items = []T{}
	}
	WriteJSON(w, http.StatusOK, PagedResponse[T]{Items: items, TotalCount: total})
}

func HandleServiceError(w http.ResponseWriter, err error) {
	if ae, ok := err.(*AppError); ok {
		WriteError(w, ae)
		return
	}
	WriteFail(w, http.StatusInternalServerError, notifications.SharedUnexpectedError(err.Error()))
}
