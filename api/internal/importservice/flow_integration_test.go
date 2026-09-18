package importservice

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/luxus-connect/telefonia/api/internal/dbmigrate"
	"github.com/luxus-connect/telefonia/api/internal/models"
	"github.com/luxus-connect/telefonia/api/internal/store"
)

type fixtureStorage struct{ raw []byte }

func (s fixtureStorage) GetObject(context.Context, string, string) ([]byte, error) { return s.raw, nil }

func fixtureRecord(fields map[int]string) string {
	b := []byte(strings.Repeat(" ", 1900))
	for offset, value := range fields {
		copy(b[offset:], value)
	}
	return string(b)
}

func fixtureInvoice(account, total, plan string) []byte {
	return []byte(strings.Join([]string{
		fixtureRecord(map[int]string{0: account, 110: "010D", 205: "202608", 229: "20260801", 247: "20260825", 353: "20260801", 361: "20260831", 540: total}),
		fixtureRecord(map[int]string{110: "011D", 199: "Empresa Teste", 1846: "12345678000190"}),
		fixtureRecord(map[int]string{110: "110D", 178: account, 248: plan, 318: total}),
	}, "\n"))
}

// Run only against a disposable database with migrations 001–020 applied.
func TestImportFlowPostgres(t *testing.T) {
	url := os.Getenv("TEST_IMPORT_DATABASE_URL")
	if url == "" {
		t.Skip("set TEST_IMPORT_DATABASE_URL to a disposable database with migrations 001–020")
	}
	ctx := context.Background()
	st, err := store.New(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	// The startup migrations must also be safe on the next restart.
	if err := dbmigrate.Apply(ctx, st.Pool()); err != nil {
		t.Fatal(err)
	}
	org, provider, month := uuid.NewString(), uuid.NewString(), uuid.NewString()
	if err := st.CreateProvider(ctx, org, provider, "VIVO Test", provider); err != nil {
		t.Fatal(err)
	}
	if err := st.CreateProcessingMonth(ctx, org, month, provider, "Agosto 2026", 2026, 8); err != nil {
		t.Fatal(err)
	}
	account := fmt.Sprintf("%010d", time.Now().UnixNano()%10000000000)
	newRequest := func(substitute bool) string {
		t.Helper()
		id := uuid.NewString()
		if err := st.CreateImportRequest(ctx, store.ImportRequestRow{ID: id, OrganizationID: org, ProviderID: provider, ProcessingMonthID: month, StorageBucket: "test", StorageObjectKey: org + "/invoice.txt", CreatedBy: "test", AllowSubstitute: substitute}); err != nil {
			t.Fatal(err)
		}
		return id
	}
	checkStatus := func(id string, status int) {
		t.Helper()
		row, err := st.GetImportRequest(ctx, id)
		if err != nil || row == nil {
			t.Fatalf("request: %v / %v", row, err)
		}
		if row.Status != status || row.CompletedAt == nil {
			t.Fatalf("status: %+v", row)
		}
	}
	p := &Processor{Store: st, Storage: fixtureStorage{fixtureInvoice(account, "100.00", "Plano Teste")}}
	id := newRequest(false)
	// Duplicate deliveries of the same request must both finish without a second invoice.
	var wg sync.WaitGroup
	results := make(chan error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); results <- p.ProcessImport(ctx, id) }()
	}
	wg.Wait()
	close(results)
	for err := range results {
		if err != nil {
			t.Fatal(err)
		}
	}
	checkStatus(id, 2)
	input := models.ProviderInvoiceImportRequestInput{ProviderID: provider, ProcessingMonthID: month, StorageBucket: "test", StorageObjectKey: org + "/invoice.txt", AllowSubstitute: true}
	preview, err := p.PreviewImport(ctx, org, input)
	if err != nil || preview.IsValid {
		t.Fatalf("duplicate preview: %+v / %v", preview, err)
	}
	dup := newRequest(true)
	if err := p.ProcessImport(ctx, dup); err == nil {
		t.Fatal("identical hash accepted as replacement")
	}
	checkStatus(dup, 3)
	// A failure after marking the old invoice substituted must restore the old invoice.
	p.Storage = fixtureStorage{fixtureInvoice(account, "110.00", "")}
	bad := newRequest(true)
	if err := p.ProcessImport(ctx, bad); err == nil {
		t.Fatal("missing plan accepted")
	}
	checkStatus(bad, 3)
	var active, all int
	counts := func() {
		t.Helper()
		if err := st.Pool().QueryRow(ctx, `SELECT COUNT(*) FILTER (WHERE "Status" <> 'substituted'), COUNT(*) FROM "ProviderInvoices" WHERE "ProcessingMonthId"=$1`, month).Scan(&active, &all); err != nil {
			t.Fatal(err)
		}
	}
	counts()
	if active != 1 || all != 1 {
		t.Fatalf("failed replacement was not rolled back: %d/%d", active, all)
	}
	p.Storage = fixtureStorage{fixtureInvoice(account, "120.00", "Plano Teste")}
	replacement := newRequest(true)
	if err := p.ProcessImport(ctx, replacement); err != nil {
		t.Fatal(err)
	}
	checkStatus(replacement, 2)
	counts()
	if active != 1 || all != 2 {
		t.Fatalf("replacement counts: %d/%d", active, all)
	}
	var impact float64
	if err := st.Pool().QueryRow(ctx, `SELECT "SubstitutionImpact" FROM "ProviderInvoices" WHERE "ProcessingMonthId"=$1 AND "ParentInvoiceId" IS NOT NULL`, month).Scan(&impact); err != nil || impact != 20 {
		t.Fatalf("impact %v / %v", impact, err)
	}
	p.Storage = fixtureStorage{[]byte("%PDF-1.7")}
	pdf := newRequest(false)
	if err := p.ProcessImport(ctx, pdf); err == nil {
		t.Fatal("PDF accepted")
	}
	checkStatus(pdf, 4)
}

func TestImportErrorDoesNotExposeSQL(t *testing.T) {
	got := importErrorMessage(errors.New(`ERROR: column "ContentSHA256" does not exist (SQLSTATE 42703)`))
	if strings.Contains(got, "SQLSTATE") || strings.Contains(got, "ContentSHA256") {
		t.Fatal(got)
	}
}
