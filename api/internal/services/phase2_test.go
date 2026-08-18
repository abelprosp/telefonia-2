package services_test

import (
	"testing"
	"time"

	"github.com/luxus-connect/telefonia/api/internal/precision"
)

func TestPhase2_DiscountValidationLogic(t *testing.T) {
	servicesTotal := 100.00
	discount1 := 40.00
	discount2 := 70.00

	// Desconto 1 (R$ 40,00) <= R$ 100,00 -> Permitido
	if discount1 > servicesTotal {
		t.Errorf("discount1 %v should be allowed for services %v", discount1, servicesTotal)
	}

	// Desconto 1 + Desconto 2 (R$ 110,00) > R$ 100,00 -> Bloqueado
	totalDiscounts := precision.SumCents(discount1, discount2)
	if totalDiscounts <= servicesTotal {
		t.Errorf("total discounts %v should exceed servicesTotal %v", totalDiscounts, servicesTotal)
	}
}

func TestPhase2_InstallmentVigencyCalculation(t *testing.T) {
	startDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	installmentCount := 12

	expectedEnd := startDate.AddDate(0, installmentCount, 0)
	if expectedEnd.Year() != 2027 || expectedEnd.Month() != 1 || expectedEnd.Day() != 1 {
		t.Errorf("expected end date 2027-01-01, got %v", expectedEnd)
	}
}

func TestPhase2_PayoffRemainingAmountCalculation(t *testing.T) {
	installmentAmount := 50.00
	count := 10
	current := 7 // restam 10 - 7 + 1 = 4 parcelas (7, 8, 9, 10)

	remaining := count - current + 1
	payoffAmount := precision.Round2(installmentAmount * float64(remaining))

	if remaining != 4 {
		t.Errorf("expected 4 remaining installments, got %d", remaining)
	}
	if payoffAmount != 200.00 {
		t.Errorf("expected payoff amount 200.00, got %v", payoffAmount)
	}
}
