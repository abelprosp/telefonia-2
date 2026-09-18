package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/luxus-connect/telefonia/api/internal/auth"
	"github.com/luxus-connect/telefonia/api/internal/config"
	"github.com/luxus-connect/telefonia/api/internal/observability"
	"github.com/luxus-connect/telefonia/api/internal/services"
	"github.com/luxus-connect/telefonia/api/internal/statemachine"
	"github.com/luxus-connect/telefonia/api/internal/store"
)

const testOrgID = "00000000-0000-0000-0000-000000000001"

func testAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u := &auth.User{
			ID:       "11111111-1111-1111-1111-111111111111",
			Email:    "dev@luxus.local",
			Name:     "Dev Luxus",
			Username: "dev",
			Roles:    []string{"master", "admin", "user", "financial", "partner", "operator", "employee"},
		}
		org := &auth.Organization{ID: testOrgID, Name: "Luxus Connect", Alias: "luxus"}
		ctx := auth.WithUser(r.Context(), u)
		ctx = auth.WithOrganization(ctx, org)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", ".."))
}

func openTestStore(t *testing.T) *store.Store {
	t.Helper()
	raw := os.Getenv("DATABASE_URL")
	if raw == "" {
		raw = "postgres://postgres:postgres@127.0.0.1:5433/luxus_connect_dev?sslmode=disable"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()
	st, err := store.New(ctx, config.NormalizeDatabaseURL(raw))
	if err != nil {
		t.Skipf("postgres indisponível: %v", err)
	}
	t.Cleanup(st.Close)
	ensureSchema(t, st)
	return st
}

func ensureSchema(t *testing.T, st *store.Store) {
	t.Helper()
	var exists bool
	err := st.Pool().QueryRow(context.Background(), `SELECT EXISTS (
		SELECT 1 FROM information_schema.tables WHERE table_schema='public' AND table_name='Providers')`).Scan(&exists)
	if err != nil {
		t.Fatalf("schema probe: %v", err)
	}
	if exists {
		return
	}
	root := repoRoot(t)
	files, err := filepath.Glob(filepath.Join(root, "db", "migrations", "[0-9][0-9][0-9]_*.sql"))
	if err != nil || len(files) == 0 {
		t.Skip("schema ausente e migrações não encontradas")
	}
	for _, f := range files {
		cmd := exec.Command("docker", "exec", "-i", "postgres.connect.luxus", "psql", "-U", "postgres", "-d", "luxus_connect_dev", "-v", "ON_ERROR_STOP=1")
		in, err := os.Open(f)
		if err != nil {
			t.Fatalf("open %s: %v", f, err)
		}
		cmd.Stdin = in
		out, err := cmd.CombinedOutput()
		_ = in.Close()
		if err != nil {
			t.Fatalf("migração %s falhou: %v\n%s", filepath.Base(f), err, out)
		}
	}
}

func newTestServer(t *testing.T, st *store.Store) http.Handler {
	t.Helper()
	svc := &services.Service{
		Store:        st,
		StateMachine: statemachine.NewEngine(st),
	}
	h := &Handler{Svc: svc, Presigned: &services.PresignedService{}}
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Get("/health/live", observability.LivenessHandler())
	h.RegisterRoutes(r, testAuth, testAuth, testAuth, testAuth, testAuth, testAuth)
	return r
}

func doJSON(t *testing.T, srv http.Handler, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var rdr io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		rdr = bytes.NewReader(b)
	}
	req := httptest.NewRequest(method, path, rdr)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	return rec
}

func mustStatus(t *testing.T, rec *httptest.ResponseRecorder, method, path string, allowed ...int) {
	t.Helper()
	for _, code := range allowed {
		if rec.Code == code {
			return
		}
	}
	t.Fatalf("%s %s: expected %v, got %d body=%s", method, path, allowed, rec.Code, truncate(rec.Body.String(), 500))
}

func truncate(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func TestHealthLive(t *testing.T) {
	st := openTestStore(t)
	srv := newTestServer(t, st)
	rec := doJSON(t, srv, http.MethodGet, "/health/live", nil)
	mustStatus(t, rec, http.MethodGet, "/health/live", http.StatusOK)
}

func TestPublicOrganizationSettings(t *testing.T) {
	st := openTestStore(t)
	srv := newTestServer(t, st)
	for _, path := range []string{
		"/v1/organization-settings",
		"/v1/company-settings",
		"/v1/whitelabel-settings",
		"/v1/system-settings",
	} {
		rec := doJSON(t, srv, http.MethodGet, path, nil)
		mustStatus(t, rec, http.MethodGet, path, http.StatusOK)
	}
}

func TestListEndpointsOK(t *testing.T) {
	st := openTestStore(t)
	srv := newTestServer(t, st)
	gets := []string{
		"/v1/me",
		"/v1/stats/dashboard",
		"/v1/stats/operational-dashboard",
		"/v1/inapp-notifications",
		"/v1/providers",
		"/v1/provider-invoices",
		"/v1/customers",
		"/v1/phone-lines",
		"/v1/device-stock",
		"/v1/billing-cycles",
		"/v1/processing-months",
		"/v1/contracts/expiring",
		"/v1/state-transitions",
		"/v1/cost-centers",
		"/v1/reports/line-movements",
		"/v1/reports/financial-summary",
		"/v1/reports/customer-profitability",
		"/v1/exceedance-terms",
		"/v1/fidelity-renewal-triggers",
		"/v1/divergences",
		"/v1/tickets",
		"/v1/approvals",
		"/v1/webhooks",
		"/v1/inventory/devices",
		"/v1/phone-line-operation-requests",
		"/v1/contract-templates",
		"/v1/sales",
		"/v1/financial/summary",
		"/v1/accounts-payable",
		"/v1/accounts-receivable",
		"/v1/partner-sales",
		"/v1/partner-commission-settings",
		"/v1/invoice-email-templates",
		"/v1/invoice-layout-templates",
		"/v1/customer-billing-documents",
		"/v1/collections/overdue",
		"/v1/collections/sicredi/status",
		"/v1/users",
		"/v1/partner/stats/dashboard",
		"/v1/partner/providers",
		"/v1/partner/customers",
		"/v1/partner/phone-lines",
		"/v1/partner/phone-line-operation-requests",
		"/v1/partner/financial/summary",
		"/v1/partner/sales",
		"/v1/partner/contract-templates",
		"/v1/partner/commercial-sales",
		"/v1/portal/me",
		"/v1/portal/lines",
		"/v1/portal/invoices",
		"/v1/portal/contracts",
		"/v1/portal/tickets",
	}
	for _, path := range gets {
		rec := doJSON(t, srv, http.MethodGet, path, nil)
		if rec.Code == http.StatusInternalServerError {
			t.Errorf("GET %s returned %d body=%s", path, rec.Code, truncate(rec.Body.String(), 400))
		}
	}
}

func TestMissingResourceIsNot500(t *testing.T) {
	st := openTestStore(t)
	srv := newTestServer(t, st)
	id := uuid.NewString()
	gets := []string{
		"/v1/providers/" + id,
		"/v1/customers/" + id,
		"/v1/phone-lines/" + id,
		"/v1/device-stock/" + id,
		"/v1/billing-cycles/" + id,
		"/v1/processing-months/" + id,
		"/v1/provider-invoices/" + id,
		"/v1/tickets/" + id,
		"/v1/divergences/" + id,
		"/v1/sales/" + id,
		"/v1/contract-templates/" + id,
		"/v1/invoice-email-templates/" + id,
		"/v1/invoice-layout-templates/" + id,
		"/v1/customer-billing-documents/" + id,
		"/v1/customers/" + id + "/full-360",
		"/v1/phone-lines/" + id + "/full-360",
		"/v1/phone-lines/" + id + "/fidelity",
		"/v1/phone-lines/" + id + "/timeline",
		"/v1/customers/" + id + "/phone-lines",
		"/v1/customers/" + id + "/devices",
		"/v1/customers/" + id + "/attachments",
		"/v1/customers/" + id + "/provider-links",
		"/v1/customers/" + id + "/personal-data",
	}
	for _, path := range gets {
		rec := doJSON(t, srv, http.MethodGet, path, nil)
		if rec.Code == http.StatusInternalServerError {
			t.Errorf("GET %s returned %d body=%s", path, rec.Code, truncate(rec.Body.String(), 400))
		}
	}
}

func TestInvalidCreateBodiesAreClientErrors(t *testing.T) {
	st := openTestStore(t)
	srv := newTestServer(t, st)
	posts := []string{
		"/v1/providers",
		"/v1/customers",
		"/v1/phone-lines/stock",
		"/v1/billing-cycles",
		"/v1/processing-months",
		"/v1/device-stock",
		"/v1/tickets",
		"/v1/sales",
		"/v1/contract-templates",
		"/v1/exceedance-terms",
		"/v1/accounts-payable",
		"/v1/accounts-receivable",
		"/v1/webhooks",
	}
	for _, path := range posts {
		rec := doJSON(t, srv, http.MethodPost, path, map[string]any{})
		if rec.Code == http.StatusInternalServerError {
			t.Errorf("POST %s {} returned %d body=%s", path, rec.Code, truncate(rec.Body.String(), 400))
		}
		if rec.Code < 400 {
			t.Errorf("POST %s {} should be 4xx, got %d body=%s", path, rec.Code, truncate(rec.Body.String(), 400))
		}
	}
}

func TestCRUDProviderCustomerHappyPath(t *testing.T) {
	st := openTestStore(t)
	srv := newTestServer(t, st)
	suffix := strings.ReplaceAll(uuid.NewString(), "-", "")[:12]

	rec := doJSON(t, srv, http.MethodPost, "/v1/providers", map[string]any{
		"name": "Operadora QA " + suffix,
		"slug": "qa-" + suffix,
	})
	mustStatus(t, rec, http.MethodPost, "/v1/providers", http.StatusCreated)
	var provider map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &provider); err != nil {
		t.Fatalf("provider decode: %v body=%s", err, rec.Body.String())
	}
	providerID, _ := provider["id"].(string)
	if providerID == "" {
		t.Fatalf("provider id missing: %s", rec.Body.String())
	}

	rec = doJSON(t, srv, http.MethodGet, "/v1/providers/"+providerID, nil)
	mustStatus(t, rec, http.MethodGet, "/v1/providers/{id}", http.StatusOK)

	rec = doJSON(t, srv, http.MethodPatch, "/v1/providers/"+providerID, map[string]any{
		"name": "Operadora QA " + suffix + " Atualizada",
		"slug": "qa-" + suffix,
	})
	mustStatus(t, rec, http.MethodPatch, "/v1/providers/{id}", http.StatusAccepted, http.StatusOK)

	rec = doJSON(t, srv, http.MethodPost, "/v1/providers/"+providerID+"/plans", map[string]any{
		"code": "QA-PLAN-" + suffix,
		"name": "Plano QA",
	})
	mustStatus(t, rec, http.MethodPost, "/v1/providers/{id}/plans", http.StatusCreated, http.StatusOK)
	var plan map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &plan)
	planID, _ := plan["id"].(string)

	if planID != "" {
		rec = doJSON(t, srv, http.MethodPost, "/v1/providers/"+providerID+"/plans/"+planID+"/services", map[string]any{
			"name":         "Dados QA",
			"service_type": "data",
			"recurring":    true,
			"price":        10.5,
		})
		mustStatus(t, rec, http.MethodPost, "/v1/providers/{id}/plans/{planId}/services", http.StatusCreated, http.StatusOK, http.StatusAccepted)
	}

	legal := "Empresa QA LTDA"
	rec = doJSON(t, srv, http.MethodPost, "/v1/customers", map[string]any{
		"provider_id": providerID,
		"type":        "PJ",
		"name":        "Cliente QA " + suffix,
		"document":    fmt.Sprintf("11%012d", time.Now().UnixNano()%1e12),
		"legal_name":  legal,
	})
	mustStatus(t, rec, http.MethodPost, "/v1/customers", http.StatusCreated)
	var customer map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &customer); err != nil {
		t.Fatalf("customer decode: %v body=%s", err, rec.Body.String())
	}
	customerID, _ := customer["id"].(string)
	if customerID == "" {
		t.Fatalf("customer id missing: %s", rec.Body.String())
	}

	rec = doJSON(t, srv, http.MethodGet, "/v1/customers/"+customerID, nil)
	mustStatus(t, rec, http.MethodGet, "/v1/customers/{id}", http.StatusOK)

	rec = doJSON(t, srv, http.MethodGet, "/v1/customers/"+customerID+"/full-360", nil)
	mustStatus(t, rec, http.MethodGet, "/v1/customers/{id}/full-360", http.StatusOK)

	rec = doJSON(t, srv, http.MethodPatch, "/v1/customers/"+customerID, map[string]any{
		"name": "Cliente QA " + suffix + " Editado",
	})
	mustStatus(t, rec, http.MethodPatch, "/v1/customers/{id}", http.StatusAccepted, http.StatusOK)

	rec = doJSON(t, srv, http.MethodGet, "/v1/customers/"+customerID+"/phone-lines", nil)
	mustStatus(t, rec, http.MethodGet, "/v1/customers/{id}/phone-lines", http.StatusOK)

	rec = doJSON(t, srv, http.MethodGet, "/v1/customers/"+customerID+"/personal-data", nil)
	mustStatus(t, rec, http.MethodGet, "/v1/customers/{id}/personal-data", http.StatusOK)

	rec = doJSON(t, srv, http.MethodPost, "/v1/processing-months", map[string]any{
		"provider_id":  providerID,
		"year":         2026,
		"month":        8,
		"display_name": "Ago/2026 QA " + suffix,
	})
	mustStatus(t, rec, http.MethodPost, "/v1/processing-months", http.StatusCreated, http.StatusOK, http.StatusConflict)

	rec = doJSON(t, srv, http.MethodPost, "/v1/billing-cycles", map[string]any{
		"provider_id": providerID,
		"code":        "QA-" + suffix[:8],
		"name":        "Ciclo QA " + suffix,
		"start_date":  "2026-08-01",
		"end_date":    "2026-08-31",
	})
	mustStatus(t, rec, http.MethodPost, "/v1/billing-cycles", http.StatusCreated, http.StatusOK, http.StatusConflict, http.StatusBadRequest)

	rec = doJSON(t, srv, http.MethodDelete, "/v1/customers/"+customerID, nil)
	mustStatus(t, rec, http.MethodDelete, "/v1/customers/{id}", http.StatusAccepted, http.StatusOK)
}

func TestAllRegisteredRoutesDoNotPanic(t *testing.T) {
	st := openTestStore(t)
	srv := newTestServer(t, st)
	r, ok := srv.(*chi.Mux)
	if !ok {
		t.Fatal("expected chi mux")
	}
	missing := uuid.NewString()
	skipPrefixes := []string{
		"/setup/keycloak",
	}
	err := chi.Walk(r, func(method, route string, _ http.Handler, _ ...func(http.Handler) http.Handler) error {
		for _, p := range skipPrefixes {
			if strings.HasPrefix(route, p) {
				return nil
			}
		}
		path := route
		path = strings.ReplaceAll(path, "{id}", missing)
		path = strings.ReplaceAll(path, "{planId}", missing)
		path = strings.ReplaceAll(path, "{serviceId}", missing)
		path = strings.ReplaceAll(path, "{deviceLinkId}", missing)
		path = strings.ReplaceAll(path, "{attachmentId}", missing)
		path = strings.ReplaceAll(path, "{processingMonthId}", missing)
		path = strings.ReplaceAll(path, "{contractId}", missing)
		path = strings.ReplaceAll(path, "{processingId}", missing)
		path = strings.ReplaceAll(path, "{itemId}", missing)
		path = strings.ReplaceAll(path, "{invoiceId}", missing)
		path = strings.ReplaceAll(path, "{receivableId}", missing)
		path = strings.ReplaceAll(path, "{messageId}", missing)
		if strings.Contains(path, "{") {
			return nil
		}
		var body any
		switch method {
		case http.MethodGet, http.MethodDelete, http.MethodHead:
			body = nil
		default:
			body = map[string]any{}
		}
		rec := doJSON(t, srv, method, path, body)
		if rec.Code == http.StatusInternalServerError {
			t.Errorf("%s %s returned %d body=%s", method, path, rec.Code, truncate(rec.Body.String(), 300))
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk: %v", err)
	}
}
