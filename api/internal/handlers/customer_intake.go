package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/models"
	"github.com/luxus-connect/telefonia/api/internal/notifications"
)

func decodeStrictBytes(body []byte, dst any) error {
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return err
	}
	if err := dec.Decode(new(any)); !errors.Is(err, io.EOF) {
		return errors.New("request must contain a single JSON value")
	}
	return nil
}

func intakeDecodeError(w http.ResponseWriter, err error) {
	var tooLarge *http.MaxBytesError
	if errors.As(err, &tooLarge) {
		httputil.WriteFail(w, 413, notifications.N("INTAKE_TOO_LARGE", "Requisição acima do limite de 24 MiB."))
		return
	}
	message := err.Error()
	if len(message) > 300 {
		message = message[:300] + "…"
	}
	httputil.WriteFail(w, 400, notifications.N("REQUEST_VALIDATION", "JSON inválido: "+message))
}

func (h *Handler) customerIntake(w http.ResponseWriter, r *http.Request) {
	var input models.CustomerIntakeInput
	if err := decodeJSON(r, &input); err != nil {
		intakeDecodeError(w, err)
		return
	}
	h.importCustomerSale(w, r, input)
}

func (h *Handler) importCustomerSale(w http.ResponseWriter, r *http.Request, input models.CustomerIntakeInput) {
	result, err := h.Svc.ImportCustomerSale(r.Context(), input)
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	status := http.StatusCreated
	if result.Replayed {
		status = http.StatusOK
	}
	httputil.WriteJSON(w, status, result)
}

func (h *Handler) listCustomerIntakes(w http.ResponseWriter, r *http.Request) {
	items, err := h.Svc.ListCustomerIntakes(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	httputil.WriteJSON(w, 200, items)
}

func (h *Handler) downloadCustomerIntakeAttachment(w http.ResponseWriter, r *http.Request) {
	items, err := h.Svc.ListCustomerAttachments(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		httputil.HandleServiceError(w, err)
		return
	}
	for _, item := range items {
		if item.ID != chi.URLParam(r, "attachmentId") {
			continue
		}
		result, err := h.Presigned.CreateAuthorizedDownloadURL(r.Context(), models.CreatePresignedDownloadURLInput{BucketName: item.StorageBucket, ObjectKey: item.StorageObjectKey})
		if err != nil {
			httputil.HandleServiceError(w, err)
			return
		}
		httputil.WriteJSON(w, 200, result)
		return
	}
	httputil.WriteFail(w, 404, notifications.N("ATTACHMENT_NOT_FOUND", "Anexo não encontrado para este cliente."))
}
