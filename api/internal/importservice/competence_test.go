package importservice

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/store"
	"github.com/luxus-connect/telefonia/api/internal/vivo"
)

func TestParseReferenceMonthYM(t *testing.T) {
	y, m, ok := parseReferenceMonthYM("202406")
	if !ok || y != 2024 || m != 6 {
		t.Fatalf("got %d/%d ok=%v", m, y, ok)
	}
	y, m, ok = parseReferenceMonthYM("07/2026")
	if !ok || y != 2026 || m != 7 {
		t.Fatalf("got %d/%d ok=%v", m, y, ok)
	}
	if _, _, ok := parseReferenceMonthYM("bad"); ok {
		t.Fatal("expected fail")
	}
}

func TestValidateInvoiceCompetence_match(t *testing.T) {
	header := &vivo.Line010DHeader{ReferenceMonth: "202607"}
	month := &store.ProcessingMonthRow{Year: 2026, Month: 7}
	if err := validateInvoiceCompetence(header, month); err != nil {
		t.Fatal(err)
	}
}

func TestValidateInvoiceCompetence_mismatch(t *testing.T) {
	header := &vivo.Line010DHeader{ReferenceMonth: "202609"}
	month := &store.ProcessingMonthRow{Year: 2026, Month: 7}
	err := validateInvoiceCompetence(header, month)
	if err == nil {
		t.Fatal("expected mismatch")
	}
	var app *httputil.AppError
	if !errors.As(err, &app) {
		t.Fatalf("expected AppError, got %T: %v", err, err)
	}
	if !strings.Contains(app.Error(), "diverge") {
		t.Fatalf("unexpected message: %s", app.Error())
	}
}

func TestValidateInvoiceCompetence_undetermined(t *testing.T) {
	header := &vivo.Line010DHeader{ReferenceMonth: "xx"}
	month := &store.ProcessingMonthRow{Year: 2026, Month: 7}
	if err := validateInvoiceCompetence(header, month); err == nil {
		t.Fatal("expected undetermined")
	}
}

func TestValidateInvoiceCompetence_billingEndFallback(t *testing.T) {
	header := &vivo.Line010DHeader{
		ReferenceMonth: "",
		BillingEndDate: time.Date(2026, 9, 15, 0, 0, 0, 0, time.UTC),
	}
	month := &store.ProcessingMonthRow{Year: 2026, Month: 9}
	if err := validateInvoiceCompetence(header, month); err != nil {
		t.Fatal(err)
	}
	month.Month = 7
	if err := validateInvoiceCompetence(header, month); err == nil {
		t.Fatal("expected mismatch on billing end")
	}
}
