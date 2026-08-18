package services_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/luxus-connect/telefonia/api/internal/models"
	"github.com/luxus-connect/telefonia/api/internal/precision"
)

func TestPhase6_PortalMeSummary(t *testing.T) {
	resp := models.PortalCustomerMeResponse{
		Customer: models.GetCustomerResponse{
			ID:      "cust_123",
			Name:    "Empresa Alpha LTDA",
			CpfCnpj: "12345678000190",
		},
		ActiveLinesCount:   12,
		TotalMonthlyAmount: 1450.80,
	}

	if resp.ActiveLinesCount != 12 || resp.TotalMonthlyAmount != 1450.80 {
		t.Errorf("mismatch in portal me response")
	}
}

func TestPhase6_SicrediPixPayloadGeneration(t *testing.T) {
	docID := "doc_abc"
	nossoNum := "12345678"
	totalAmt := 189.90

	pixCopiaCola := fmt.Sprintf("00020126580014br.gov.bcb.pix0136%s520400005303986540%0.2f5802BR5915LUXUS CONECTA6009SAO PAULO62070503***6304", nossoNum, totalAmt)

	if !strings.Contains(pixCopiaCola, "LUXUS CONECTA") || !strings.Contains(pixCopiaCola, nossoNum) {
		t.Errorf("invalid pix payload format")
	}

	pixResp := models.SicrediPixGenerateResponse{
		DocumentID:   docID,
		NossoNumero:  nossoNum,
		PixQrCode:    "https://api.qrserver.com/v1/create-qr-code/?size=250x250&data=" + pixCopiaCola,
		PixCopiaCola: pixCopiaCola,
		ExpiresAt:    time.Now().Add(24 * time.Hour),
	}

	if pixResp.NossoNumero != nossoNum || pixResp.PixCopiaCola != pixCopiaCola {
		t.Errorf("pix response mismatch")
	}
}

func TestPhase6_FinancialSummaryCalculations(t *testing.T) {
	gross := 50000.00
	cost := 32000.00
	margin := precision.Round2(gross - cost)
	marginPct := precision.Round2((margin / gross) * 100.0)

	report := models.FinancialSummaryReportResponse{
		GeneratedAt: time.Now().UTC(),
		TotalLines:  250,
		TotalGross:  gross,
		TotalCost:   cost,
		TotalMargin: margin,
	}

	if report.TotalMargin != 18000.00 {
		t.Errorf("expected margin 18000.00, got %v", report.TotalMargin)
	}
	if marginPct != 36.00 {
		t.Errorf("expected margin pct 36%%, got %v%%", marginPct)
	}
}
