package services

import (
	"strings"
	"testing"
	"time"

	"github.com/luxus-connect/telefonia/api/internal/models"
)

func TestCalculateLateCharges_InDueDate(t *testing.T) {
	due := time.Date(2026, 8, 26, 0, 0, 0, 0, time.UTC)
	got := CalculateLateCharges(100, due, due, 2, 1)
	if got.Overdue || got.DaysOverdue != 0 || got.TotalAmount != 100 {
		t.Fatalf("expected in-due invoice, got %+v", got)
	}
}

func TestCalculateLateCharges_Overdue(t *testing.T) {
	due := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	asOf := time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)
	got := CalculateLateCharges(100, due, asOf, 2, 1)
	if !got.Overdue || got.DaysOverdue != 30 {
		t.Fatalf("expected 30 days overdue, got %+v", got)
	}
	if got.LateFeeAmount != 2 {
		t.Fatalf("expected multa 2.00, got %.2f", got.LateFeeAmount)
	}
	if got.InterestAmount != 1 {
		t.Fatalf("expected juros 1.00 (1%% a.m. em 30 dias), got %.2f", got.InterestAmount)
	}
	if got.TotalAmount != 103 {
		t.Fatalf("expected total 103.00, got %.2f", got.TotalAmount)
	}
}

func TestScoreReceiptMatch_PixAndAmount(t *testing.T) {
	tx := "E123PIX"
	inv := models.FinancialAgentInvoice{
		InvoiceNumber:      "FAT-001",
		Amount:             150.40,
		CustomerName:       "Maria Silva",
		SicrediPixTxID:     &tx,
		SicrediNossoNumero: agentStrPtr("000123"),
	}
	score, reasons := scoreReceiptMatch(models.FinancialAgentVerifyReceiptInput{
		Amount:    150.40,
		PixTxID:   "E123PIX",
		PayerName: "Maria Silva",
	}, inv, nil)
	if score < 70 {
		t.Fatalf("expected high score, got %d (%v)", score, reasons)
	}
}

func TestParseAgentDate(t *testing.T) {
	d, err := parseAgentDate("26/08/2026")
	if err != nil {
		t.Fatal(err)
	}
	if d.Day() != 26 || d.Month() != 8 || d.Year() != 2026 {
		t.Fatalf("unexpected %s", d)
	}
}

func TestAmountsClose(t *testing.T) {
	if !amountsClose(10.00, 10.04) {
		t.Fatal("expected close amounts")
	}
	if amountsClose(10.00, 10.20) {
		t.Fatal("did not expect close amounts")
	}
}

func TestPixIDsMatch(t *testing.T) {
	tx := "abc123xyz"
	inv := models.FinancialAgentInvoice{SicrediPixTxID: &tx}
	if !pixIDsMatch(models.FinancialAgentVerifyReceiptInput{PixTxID: "ABC123XYZ"}, inv) {
		t.Fatal("expected pix match")
	}
	if pixIDsMatch(models.FinancialAgentVerifyReceiptInput{PixTxID: "other"}, inv) {
		t.Fatal("did not expect pix match")
	}
}

func TestToAgentInvoiceOverdue(t *testing.T) {
	doc := models.ListCustomerBillingDocumentResponse{
		ID:            "1",
		CustomerName:  "Cliente",
		InvoiceNumber: "FAT-9",
		DueDate:       time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
		Amount:        50,
		Status:        "sent",
	}
	inv := toAgentInvoice(doc, time.Date(2026, 8, 10, 0, 0, 0, 0, time.UTC))
	if !inv.Overdue || inv.DaysOverdue != 9 || inv.Paid {
		t.Fatalf("unexpected invoice %+v", inv)
	}
	if !strings.Contains(CalculateLateCharges(50, doc.DueDate, time.Date(2026, 8, 10, 0, 0, 0, 0, time.UTC), 2, 1).Explanation, "Atraso") {
		t.Fatal("expected overdue explanation")
	}
}

func agentStrPtr(v string) *string { return &v }
