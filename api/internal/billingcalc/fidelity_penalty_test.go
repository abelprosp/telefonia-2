package billingcalc

import (
	"testing"
	"time"
)

func TestCalculateFidelityPenalty_ActiveBreak(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(2, 0, 0) // 24 meses (2028-01-01)
	cancel := start.AddDate(0, 6, 0) // 6 meses depois (restam 18 meses)
	monthlyAmount := 100.00
	penaltyPercent := 30.0 // 30%

	result := CalculateFidelityPenalty(monthlyAmount, start, end, cancel, penaltyPercent)

	if result.IsExempt {
		t.Errorf("expected penalty not to be exempt")
	}
	if result.MonthsRemaining != 18 {
		t.Errorf("expected 18 months remaining, got %d", result.MonthsRemaining)
	}
	// 18 * 100 * 0.30 = 540.00
	if result.PenaltyAmount != 540.00 {
		t.Errorf("expected penalty amount 540.00, got %v", result.PenaltyAmount)
	}
}

func TestCalculateFidelityPenalty_ExpiredExempt(t *testing.T) {
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(1, 0, 0) // 12 meses (2025-01-01)
	cancel := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)

	result := CalculateFidelityPenalty(80.00, start, end, cancel, 30.0)

	if !result.IsExempt {
		t.Errorf("expected penalty to be exempt after fidelity expiration")
	}
	if result.PenaltyAmount != 0 {
		t.Errorf("expected 0 penalty amount, got %v", result.PenaltyAmount)
	}
}
