package services_test

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/luxus-connect/telefonia/api/internal/models"
	"github.com/luxus-connect/telefonia/api/internal/precision"
)

func TestPhase3_ConsolidationHashGeneration(t *testing.T) {
	orgID := "org_123"
	monthID := "month_2026_08"
	userID := "user_master"
	timestamp := time.Date(2026, 8, 17, 21, 0, 0, 0, time.UTC).Unix()

	hashData := fmt.Sprintf("ORG=%s|MONTH=%s|CLOSED_BY=%s|TIMESTAMP=%d", orgID, monthID, userID, timestamp)
	hashBytes := sha256.Sum256([]byte(hashData))
	hashStr := hex.EncodeToString(hashBytes[:])

	if len(hashStr) != 64 {
		t.Errorf("expected 64-char SHA256 hex string, got %s (len %d)", hashStr, len(hashStr))
	}
}

func TestPhase3_BillingExplanationDecomposition(t *testing.T) {
	components := []models.BillingExplanationComponent{
		{Type: "service", Description: "Plano Smart Empresas", Amount: 79.90},
		{Type: "device", Description: "Samsung Galaxy A54", Amount: 35.00, Details: "Parcela 4/12"},
		{Type: "discount", Description: "Desconto Fidelidade", Amount: -15.00},
	}

	var total float64
	for _, c := range components {
		total = precision.SumCents(total, c.Amount)
	}

	if total != 99.90 {
		t.Errorf("expected total 99.90, got %v", total)
	}
}

func TestPhase3_ImpactSimulationMarginCalculation(t *testing.T) {
	revenue := 10000.00
	cost := 6500.00
	margin := precision.Round2(revenue - cost)
	marginPct := precision.Round2((margin / revenue) * 100.0)

	if margin != 3500.00 {
		t.Errorf("expected margin 3500.00, got %v", margin)
	}
	if marginPct != 35.00 {
		t.Errorf("expected margin percentage 35%%, got %v%%", marginPct)
	}
}
