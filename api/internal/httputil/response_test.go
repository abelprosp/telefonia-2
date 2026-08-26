package httputil

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/luxus-connect/telefonia/api/internal/notifications"
)

func TestHandleServiceError_WrappedAppError(t *testing.T) {
	rec := httptest.NewRecorder()
	err := fmt.Errorf("contexto: %w", ValidationError(notifications.CustomerNameRequired))
	HandleServiceError(rec, err)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for wrapped validation error, got %d body=%s", rec.Code, rec.Body.String())
	}
	var body MessageResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(body.Messages) == 0 || body.Messages[0].Code != "CUSTOMER_NAME_REQUIRED" {
		t.Fatalf("unexpected messages: %+v", body.Messages)
	}
}

func TestHandleServiceError_Unavailable(t *testing.T) {
	rec := httptest.NewRecorder()
	HandleServiceError(rec, UnavailableError(notifications.N("KEYCLOAK_ADMIN_UNAVAILABLE", "User management is not configured.")))
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}
}

func TestWritePagedNilItems(t *testing.T) {
	rec := httptest.NewRecorder()
	var items []string
	WritePaged(rec, items, 0)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var body PagedResponse[string]
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Items == nil {
		t.Fatal("items should be empty slice, not null")
	}
}
