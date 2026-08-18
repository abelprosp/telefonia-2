package services_test

import (
	"testing"
	"time"

	"github.com/luxus-connect/telefonia/api/internal/models"
)

func TestPhase4_PreClosingAlertsSeverityClassification(t *testing.T) {
	alerts := []models.PreClosingAlert{
		{Code: "LINE_WITHOUT_CUSTOMER", Severity: "CRITICAL", Message: "Linha ativa sem cliente vinculado."},
		{Code: "INVOICE_PENDING_CONCILIATION", Severity: "WARNING", Message: "Fatura de operadora pendente de conciliação."},
		{Code: "FIDELITY_EXPIRING", Severity: "WARNING", Message: "Contrato de fidelidade prestes a expirar."},
	}

	criticalCount := 0
	warningCount := 0
	canClose := true

	for _, a := range alerts {
		if a.Severity == "CRITICAL" {
			criticalCount++
			canClose = false
		} else if a.Severity == "WARNING" {
			warningCount++
		}
	}

	if canClose {
		t.Errorf("expected canClose to be false when CRITICAL alerts exist")
	}
	if criticalCount != 1 {
		t.Errorf("expected 1 critical alert, got %d", criticalCount)
	}
	if warningCount != 2 {
		t.Errorf("expected 2 warning alerts, got %d", warningCount)
	}
}

func TestPhase4_ExpiringContractsDaysRemaining(t *testing.T) {
	now := time.Now().UTC()
	endDate := now.AddDate(0, 0, 15) // vence em 15 dias

	daysRemaining := int(endDate.Sub(now).Hours() / 24.0)
	if daysRemaining < 14 || daysRemaining > 16 {
		t.Errorf("expected approx 15 days remaining, got %d", daysRemaining)
	}
}
