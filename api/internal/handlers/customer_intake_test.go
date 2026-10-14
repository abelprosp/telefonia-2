package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/luxus-connect/telefonia/api/internal/auth"
	"github.com/luxus-connect/telefonia/api/internal/models"
	"github.com/luxus-connect/telefonia/api/internal/services"
)

func TestIntakeKeyCannotAuthorizeOperationalReads(t *testing.T) {
	key := strings.Repeat("a", 64)
	h := &Handler{Svc: &services.Service{}, IntakeKey: key, IntakeOrg: uuid.NewString()}
	deny := func(_ http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(401) })
	}
	r := chi.NewRouter()
	h.RegisterRoutes(r, deny, deny, deny, deny, deny, deny)
	for _, test := range []struct {
		method, path string
		want         int
	}{{"POST", "/v1/intake/customers", 400}, {"GET", "/v1/customers", 401}, {"POST", "/v1/customers", 401}, {"GET", "/v1/sales", 401}} {
		req := httptest.NewRequest(test.method, test.path, strings.NewReader(`{"client":{},"sale":{}}`))
		req.Header.Set("X-API-Key", key)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != test.want {
			t.Fatalf("key escaped intake route %s %s: %d", test.method, test.path, w.Code)
		}
	}
}

func TestCreateCustomerAcceptsLegacyAndEnvelopeValidation(t *testing.T) {
	h := &Handler{Svc: &services.Service{}}
	ctx := auth.WithOrganization(context.Background(), &auth.Organization{ID: uuid.NewString()})
	for _, test := range []struct{ body, code string }{
		{`{"type":"PF","name":"","document":"02113419084","addresses":[]}`, "CUSTOMER_NAME"},
		{`{"client":{},"sale":{}}`, "CUSTOMER"},
	} {
		rec := httptest.NewRecorder()
		h.createCustomer(rec, httptest.NewRequest("POST", "/v1/customers", strings.NewReader(test.body)).WithContext(ctx))
		if rec.Code != 400 || !strings.Contains(rec.Body.String(), test.code) {
			t.Fatalf("legacy/envelope validation: %d %s", rec.Code, rec.Body.String())
		}
	}
	for _, body := range []string{`{"client":{},"sale":{},"organizationId":"other"}`, `{"name":"a"} {}`, `{"sale":{"value":79.999},"client":{}}`} {
		rec := httptest.NewRecorder()
		h.createCustomer(rec, httptest.NewRequest("POST", "/v1/customers", strings.NewReader(body)))
		if rec.Code != 400 {
			t.Fatalf("accepted invalid/unknown fields: %s", body)
		}
	}
}

type intakeTestStorage struct {
	uploads int
	fail    bool
}

func (s *intakeTestStorage) PutIntakeObject(_ context.Context, _, _, _ string, _ []byte) error {
	s.uploads++
	if s.fail {
		return fmt.Errorf("storage unavailable")
	}
	return nil
}

// Uses only the repository's explicitly configured local test database. Skips
// with the existing suite when PostgreSQL is unavailable.
func TestIntakeTransactionReplayAndTenantIsolation(t *testing.T) {
	st := openTestStore(t)
	ctx := auth.WithOrganization(context.Background(), &auth.Organization{ID: uuid.NewString()})
	ctx = auth.WithUser(ctx, &auth.User{ID: "intake-test", Roles: []string{auth.RoleEmployee}})
	storage := &intakeTestStorage{}
	svc := &services.Service{Store: st, IntakeStorage: storage, IntakeBucket: "files"}
	var input models.CustomerIntakeInput
	body := []byte(`{"sourceSystem":"intake-test","sale":{"id":"` + uuid.NewString() + `","protocol":"VND-TEST","soldAt":"2026-10-14T14:30:00Z","value":79.9,"commissionValue":10,"planName":"Controle"},"client":{"name":"Cliente importado","document":"02113419084","documentType":"CPF","phone":"51999998888"},"attachments":[{"id":"pdf","type":"CONTRACT","name":"contract.pdf","mimeType":"application/pdf","contentBase64":"JVBERi0xLjQKJSVFT0Y="}]}`)
	if err := json.Unmarshal(body, &input); err != nil {
		t.Fatal(err)
	}
	first, err := svc.ImportCustomerSale(ctx, input)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		c, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		org := auth.OrganizationFromContext(ctx).ID
		st.Pool().Exec(c, `DELETE FROM "AuditLogs" WHERE "EntityName"='CustomerIntakes' AND "KeyValues" IN (SELECT "Id" FROM "CustomerIntakes" WHERE "OrganizationId"=$1)`, org)
		st.Pool().Exec(c, `DELETE FROM "CustomerIntakes" WHERE "OrganizationId"=$1`, org)
		st.Pool().Exec(c, `DELETE FROM "Sales" WHERE "OrganizationId"=$1`, org)
		st.Pool().Exec(c, `DELETE FROM "Customers" WHERE "OrganizationId"=$1`, org)
	})
	second, err := svc.ImportCustomerSale(ctx, input)
	if err != nil || !second.Replayed || second.CustomerID != first.CustomerID || second.SaleID != first.SaleID || storage.uploads != 1 {
		t.Fatalf("replay duplicated data or upload: %+v %v %d", second, err, storage.uploads)
	}
	sale, err := st.GetSaleInOrg(ctx, auth.OrganizationFromContext(ctx).ID, first.SaleID, nil)
	if err != nil || sale.Intake == nil || sale.TotalAmount != 79.9 || *sale.Intake.Sale.CommissionValue != 1000 || sale.Contract != nil || sale.PaidAt != nil {
		t.Fatalf("money, metadata or side effects: %+v %v", sale, err)
	}
	if len(sale.Intake.Attachments) != 1 {
		t.Fatal("attachment metadata lost")
	}
	value := models.IntakeMoney(8000)
	input.Sale.Value = &value
	if _, err := svc.ImportCustomerSale(ctx, input); err == nil {
		t.Fatal("changed replay accepted")
	}
	input.Sale.ID = uuid.NewString()
	storage.fail = true
	if _, err := svc.ImportCustomerSale(ctx, input); err == nil {
		t.Fatal("storage failure accepted")
	}
	storage.fail = false
	third, err := svc.ImportCustomerSale(ctx, input)
	if err != nil || third.CustomerID != first.CustomerID || third.SaleID == first.SaleID {
		t.Fatalf("new sale didn't reuse customer after rollback: %+v %v", third, err)
	}
	wrongCtx := auth.WithOrganization(ctx, &auth.Organization{ID: uuid.NewString()})
	if _, err := svc.ListCustomerIntakes(wrongCtx, first.CustomerID); err == nil {
		t.Fatal("cross-tenant read allowed")
	}
	items, err := svc.ListCustomerIntakes(ctx, first.CustomerID)
	if err != nil || len(items) != 2 {
		t.Fatalf("failed upload left database records: %d %v", len(items), err)
	}
	// The authenticated legacy customer endpoint also consumes the envelope.
	h := &Handler{Svc: svc}
	rec := httptest.NewRecorder()
	h.createCustomer(rec, httptest.NewRequest(http.MethodPost, "/v1/customers", bytes.NewReader(body)).WithContext(ctx))
	if rec.Code != 200 {
		t.Fatalf("envelope replay status: %d %s", rec.Code, rec.Body.String())
	}
}
