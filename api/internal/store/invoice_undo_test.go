package store_test

import (
	"context"
	"fmt"
	"hash/fnv"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/luxus-connect/telefonia/api/internal/store"
)

// uniqueTestPhone returns a digits-only MSISDN derived from id so NormalizedNumber
// stays unique across parallel fixtures (UUID hex is not digit-safe).
func uniqueTestPhone(id string) string {
	h := fnv.New64a()
	_, _ = h.Write([]byte(id))
	return fmt.Sprintf("55%011d", h.Sum64()%100_000_000_000)
}

// Set INVOICE_UNDO_TEST_DATABASE_URL to a disposable, migrated PostgreSQL database.
// Each fixture uses its own organization; never point this at production.
type undoFixture struct {
	st                                                                                                                               *store.Store
	org, provider, company, account, cycle, month, invoice, customer, plan, line, existing, payable, receivable, document, unrelated string
}

func newUndoFixture(t *testing.T) *undoFixture {
	t.Helper()
	url := os.Getenv("INVOICE_UNDO_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("set INVOICE_UNDO_TEST_DATABASE_URL for PostgreSQL integration tests")
	}
	ctx := context.Background()
	st, err := store.New(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(st.Close)
	f := &undoFixture{st: st}
	for _, p := range []*string{&f.org, &f.provider, &f.company, &f.account, &f.cycle, &f.month, &f.invoice, &f.customer, &f.plan, &f.line, &f.existing, &f.payable, &f.receivable, &f.document, &f.unrelated} {
		*p = uuid.NewString()
	}
	f.exec(t, `INSERT INTO "Providers" ("Id","OrganizationId","Name","Slug") VALUES ($1,$2,'Undo test',$1)`, f.provider, f.org)
	f.exec(t, `INSERT INTO "ContractingCompanies" ("Id","ProviderId","LegalName","TaxId") VALUES ($1,$2,'Undo test','12345678000195')`, f.company, f.provider)
	f.exec(t, `INSERT INTO "ProviderAccounts" ("Id","ContractingCompanyId","AccountNumber") VALUES ($1,$2,$1)`, f.account, f.company)
	f.exec(t, `INSERT INTO "ProviderPlans" ("Id","ProviderId","Name","Code") VALUES ($1,$2,'Test','TEST')`, f.plan, f.provider)
	f.exec(t, `INSERT INTO "BillingCycles" ("Id","OrganizationId","ProviderId","Code","Name","StartDate","EndDate","Status") VALUES ($1,$2,$3,'TEST','Test','2026-08-01','2026-08-31','open')`, f.cycle, f.org, f.provider)
	f.exec(t, `INSERT INTO "ProcessingMonths" ("Id","OrganizationId","ProviderId","Year","Month","DisplayName","Status","ClosedInContingency") VALUES ($1,$2,$3,2026,8,'Test','open',false)`, f.month, f.org, f.provider)
	f.exec(t, `INSERT INTO "Customers" ("Id","OrganizationId","Type","Name","Active") VALUES ($1,$2,'pf','Undo test',false)`, f.customer, f.org)
	err = st.CreateProviderInvoice(ctx, store.ProviderInvoiceInsert{ID: f.invoice, Number: "202608", ProviderAccountID: f.account, ContractingCompanyID: f.company, BillingCycleID: f.cycle, ProcessingMonthID: f.month, IssueDate: time.Date(2026, 8, 18, 0, 0, 0, 0, time.UTC), DueDate: time.Date(2026, 9, 8, 0, 0, 0, 0, time.UTC), TotalAmount: 33.99})
	if err != nil {
		t.Fatal(err)
	}
	if err = st.CreatePhoneLine(ctx, f.existing, f.plan, f.account, uniqueTestPhone(f.existing)); err != nil {
		t.Fatal(err)
	}
	f.exec(t, `UPDATE "PhoneLines" SET "BaseCost"=9, "CostWithConsumption"=9 WHERE "Id"=$1`, f.existing)
	err = st.WithTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		ctx = store.CtxWithTx(ctx, tx)
		if err := st.BeginInvoiceUndoTracking(ctx, f.invoice); err != nil {
			return err
		}
		if err := st.CreatePhoneLine(ctx, f.line, f.plan, f.account, uniqueTestPhone(f.line)); err != nil {
			return err
		}
		if err := st.UpdatePhoneLineCosts(ctx, f.line, 13.99, 13.99, f.invoice); err != nil {
			return err
		}
		if err := st.UpdatePhoneLineCosts(ctx, f.existing, 20, 20, f.invoice); err != nil {
			return err
		}
		if err := st.AssignPhoneLineCustomer(ctx, f.line, f.customer, time.Now(), nil); err != nil {
			return err
		}
		if err := st.ReactivateCustomer(ctx, f.customer); err != nil {
			return err
		}
		if err := st.LinkInvoicePhoneLine(ctx, f.invoice, f.line); err != nil {
			return err
		}
		if err := st.LinkInvoicePhoneLine(ctx, f.invoice, f.existing); err != nil {
			return err
		}
		return st.EndInvoiceUndoTracking(ctx)
	})
	if err != nil {
		t.Fatal(err)
	}
	f.exec(t, `INSERT INTO "AccountsPayable" ("Id","OrganizationId","Description","VendorName","ProviderInvoiceId","IssueDate","DueDate","Amount","CreatedAt","UpdatedAt") VALUES ($1,$2,'Test','Test',$3,current_date,current_date,33.99,now(),now())`, f.payable, f.org, f.invoice)
	for _, id := range []string{f.receivable, f.unrelated} {
		f.exec(t, `INSERT INTO "AccountsReceivable" ("Id","OrganizationId","CustomerId","ProcessingMonthId","Description","IssueDate","DueDate","Amount","CreatedAt","UpdatedAt") VALUES ($1,$2,$3,$4,'Test',current_date,current_date,40,now(),now())`, id, f.org, f.customer, f.month)
	}
	f.exec(t, `INSERT INTO "CustomerBillingDocuments" ("Id","OrganizationId","CustomerId","AccountsReceivableId","ProcessingMonthId","PhoneLineId","InvoiceNumber","IssueDate","DueDate","Amount","RecipientEmail","EmailSubject","EmailBodyHtml","CreatedAt","UpdatedAt") VALUES ($1,$2,$3,$4,$5,$6,'TEST',current_date,current_date,40,'test@example.invalid','Test','Test',now(),now())`, f.document, f.org, f.customer, f.receivable, f.month, f.line)
	return f
}

func (f *undoFixture) exec(t *testing.T, sql string, args ...any) {
	t.Helper()
	if _, err := f.st.Pool().Exec(context.Background(), sql, args...); err != nil {
		t.Fatal(err)
	}
}
func (f *undoFixture) check(t *testing.T, sql string, args ...any) {
	t.Helper()
	var ok bool
	if err := f.st.Pool().QueryRow(context.Background(), sql, args...).Scan(&ok); err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatalf("failed assertion: %s", sql)
	}
}

func TestInvoiceUndo_RemovesOwnedLinesRestoresExistingAndCancelsFinance(t *testing.T) {
	f := newUndoFixture(t)
	ctx := context.Background()
	if err := f.st.CancelProviderInvoice(ctx, f.org, f.invoice, "test"); err != nil {
		t.Fatal(err)
	}
	f.check(t, `SELECT NOT EXISTS(SELECT 1 FROM "PhoneLines" WHERE "Id"=$1)`, f.line)
	f.check(t, `SELECT "BaseCost"=9 AND "CostWithConsumption"=9 AND "LastInvoiceId" IS NULL FROM "PhoneLines" WHERE "Id"=$1`, f.existing)
	f.check(t, `SELECT NOT "Active" FROM "Customers" WHERE "Id"=$1`, f.customer)
	f.check(t, `SELECT "Status"='cancelled' AND "UndoneAt" IS NOT NULL FROM "ProviderInvoices" WHERE "Id"=$1`, f.invoice)
	f.check(t, `SELECT "Status"='cancelled' FROM "AccountsPayable" WHERE "Id"=$1`, f.payable)
	f.check(t, `SELECT "Status"='cancelled' FROM "AccountsReceivable" WHERE "Id"=$1`, f.receivable)
	f.check(t, `SELECT "Status"='cancelled' FROM "CustomerBillingDocuments" WHERE "Id"=$1`, f.document)
	f.check(t, `SELECT "Status"='open' FROM "AccountsReceivable" WHERE "Id"=$1`, f.unrelated)
	if err := f.st.CancelProviderInvoice(ctx, f.org, f.invoice, "test"); err != nil {
		t.Fatalf("idempotent retry: %v", err)
	}
	exists, err := f.st.InvoiceDuplicateExists(ctx, f.account, f.company, f.month, time.Date(2026, 9, 8, 0, 0, 0, 0, time.UTC))
	if err != nil || exists {
		t.Fatalf("reimport blocked: %v %v", exists, err)
	}
}

func TestInvoiceUndo_RollsBackEverythingOnLaterEdit(t *testing.T) {
	f := newUndoFixture(t)
	f.exec(t, `UPDATE "PhoneLines" SET "BaseCost"=77 WHERE "Id"=$1`, f.line)
	if err := f.st.CancelProviderInvoice(context.Background(), f.org, f.invoice, "test"); err == nil {
		t.Fatal("expected conflict")
	}
	f.check(t, `SELECT "Status"='open' FROM "AccountsPayable" WHERE "Id"=$1`, f.payable)
	f.check(t, `SELECT "Status"='open' FROM "AccountsReceivable" WHERE "Id"=$1`, f.receivable)
	f.check(t, `SELECT "BaseCost"=20 FROM "PhoneLines" WHERE "Id"=$1`, f.existing)
	f.check(t, `SELECT "UndoneAt" IS NULL FROM "ProviderInvoices" WHERE "Id"=$1`, f.invoice)
}

func TestInvoiceUndo_BlocksPaidOrBankIssued(t *testing.T) {
	for _, kind := range []string{"payable", "receivable", "boleto"} {
		t.Run(kind, func(t *testing.T) {
			f := newUndoFixture(t)
			switch kind {
			case "payable":
				f.exec(t, `UPDATE "AccountsPayable" SET "PaidAmount"=1 WHERE "Id"=$1`, f.payable)
			case "receivable":
				f.exec(t, `UPDATE "AccountsReceivable" SET "ReceivedAmount"=1 WHERE "Id"=$1`, f.receivable)
			case "boleto":
				f.exec(t, `UPDATE "CustomerBillingDocuments" SET "SicrediNossoNumero"='test',"SicrediBoletoStatus"='issued' WHERE "Id"=$1`, f.document)
			}
			if err := f.st.CancelProviderInvoice(context.Background(), f.org, f.invoice, "test"); err == nil {
				t.Fatal("expected block")
			}
			f.check(t, `SELECT "UndoneAt" IS NULL FROM "ProviderInvoices" WHERE "Id"=$1`, f.invoice)
			f.check(t, `SELECT EXISTS(SELECT 1 FROM "PhoneLines" WHERE "Id"=$1)`, f.line)
		})
	}
}

func TestInvoiceUndo_RejectsOtherOrganization(t *testing.T) {
	f := newUndoFixture(t)
	if err := f.st.CancelProviderInvoice(context.Background(), uuid.NewString(), f.invoice, "test"); err == nil {
		t.Fatal("expected not found")
	}
	f.check(t, `SELECT "Status"='open' FROM "AccountsPayable" WHERE "Id"=$1`, f.payable)
}

func TestInvoiceUndo_LegacyCreationAudit(t *testing.T) {
	f := newUndoFixture(t)
	f.exec(t, `DELETE FROM "PhoneLines" WHERE "Id"=$1`, f.existing)
	f.exec(t, `UPDATE "ProviderInvoices" SET "UndoTracked"=false,"Status"='cancelled' WHERE "Id"=$1`, f.invoice)
	f.exec(t, `DELETE FROM "InvoiceImportChanges" WHERE "InvoiceId"=$1`, f.invoice)
	f.exec(t, `INSERT INTO "AuditLogs" ("Id","ChangeType","EntityName","KeyValues","ChangedBy","NewValues","Timestamp")
 VALUES ($1,'Create','PhoneLine',$2,'import','{"message":"Linha criada em estoque automaticamente a partir da fatura 202608 / 08/01/2026."}',now())`, uuid.NewString(), f.line)
	// Incorrect provenance must not delete a line, even if LastInvoiceId matches.
	if err := f.st.CancelProviderInvoice(context.Background(), f.org, f.invoice, "test"); err == nil {
		t.Fatal("expected provenance block")
	}
	f.exec(t, `UPDATE "AuditLogs" SET "NewValues"='{"message":"Linha criada em estoque automaticamente a partir da fatura 202608 / 08/09/2026."}' WHERE "KeyValues"=$1`, f.line)
	if err := f.st.CancelProviderInvoice(context.Background(), f.org, f.invoice, "test"); err != nil {
		t.Fatal(err)
	}
	f.check(t, `SELECT NOT EXISTS(SELECT 1 FROM "PhoneLines" WHERE "Id"=$1)`, f.line)
	f.check(t, `SELECT "Status"='cancelled' FROM "AccountsReceivable" WHERE "Id"=$1`, f.receivable)
}

func TestInvoiceUndo_CancelsTrackedReceivableWithoutDocument(t *testing.T) {
	f := newUndoFixture(t)
	ctx := context.Background()
	orphan := uuid.NewString()
	if err := f.st.CreateBillingReceivable(ctx, orphan, f.org, f.customer, "Test", &f.month, &f.line, time.Now(), time.Now(), 15, time.Now()); err != nil {
		t.Fatal(err)
	}
	if err := f.st.CancelProviderInvoice(ctx, f.org, f.invoice, "test"); err != nil {
		t.Fatal(err)
	}
	f.check(t, `SELECT "Status"='cancelled' FROM "AccountsReceivable" WHERE "Id"=$1`, orphan)
	// Cancelled outputs must never be reopened by a delayed payment request.
	if err := f.st.RegisterReceivablePayment(ctx, uuid.NewString(), f.org, orphan, "test", 1, time.Now(), nil, nil, time.Now()); err == nil {
		t.Fatal("payment accepted for cancelled receivable")
	}
}

func TestInvoiceUndo_RestoresAbsentLine(t *testing.T) {
	f := newUndoFixture(t)
	ctx := context.Background()
	absent := uuid.NewString()
	if err := f.st.CreatePhoneLine(ctx, absent, f.plan, f.account, uniqueTestPhone(absent)); err != nil {
		t.Fatal(err)
	}
	if err := f.st.WithTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		ctx = store.CtxWithTx(ctx, tx)
		if err := f.st.BeginInvoiceUndoTracking(ctx, f.invoice); err != nil {
			return err
		}
		if err := f.st.UpdatePhoneLineStatus(ctx, absent, "inactive"); err != nil {
			return err
		}
		return f.st.EndInvoiceUndoTracking(ctx)
	}); err != nil {
		t.Fatal(err)
	}
	if err := f.st.CancelProviderInvoice(ctx, f.org, f.invoice, "test"); err != nil {
		t.Fatal(err)
	}
	f.check(t, `SELECT "Status"='in_stock' FROM "PhoneLines" WHERE "Id"=$1`, absent)
}
