package services_test

import (
	"testing"
	"time"

	"github.com/luxus-connect/telefonia/api/internal/billingcalc"
	"github.com/luxus-connect/telefonia/api/internal/models"
	"github.com/luxus-connect/telefonia/api/internal/precision"
	"github.com/luxus-connect/telefonia/api/internal/services"
)

func TestPhase3_ApportionGlobalDiscountLogic(t *testing.T) {
	targets := []billingcalc.ApportionmentTarget{
		{ID: "line_1", Amount: 200.00},
		{ID: "line_2", Amount: 300.00},
		{ID: "line_3", Amount: 500.00},
	}
	globalDiscount := 100.00 // Total = 1000.00. Pesos: 20%, 30%, 50% -> R$ 20.00, R$ 30.00, R$ 50.00

	items, err := billingcalc.ApportionGlobalDiscount(globalDiscount, targets)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var sumDiscount float64
	for _, it := range items {
		sumDiscount = precision.SumCents(sumDiscount, it.AllocatedDiscount)
	}

	if sumDiscount != globalDiscount {
		t.Errorf("expected sum of discounts %v, got %v", globalDiscount, sumDiscount)
	}
	if items[0].AllocatedDiscount != 20.00 || items[0].FinalAmount != 180.00 {
		t.Errorf("line_1 expected discount 20.00 and final 180.00, got %v", items[0])
	}
	if items[1].AllocatedDiscount != 30.00 || items[1].FinalAmount != 270.00 {
		t.Errorf("line_2 expected discount 30.00 and final 270.00, got %v", items[1])
	}
	if items[2].AllocatedDiscount != 50.00 || items[2].FinalAmount != 450.00 {
		t.Errorf("line_3 expected discount 50.00 and final 450.00, got %v", items[2])
	}
}

func TestPhase3_ExceedanceTermMatching(t *testing.T) {
	tabulatedPrice := 25.00
	terms := []models.ExceedanceTermResponse{
		{ID: "t1", Term: "diaria de dados", ChargeType: "tabulated", TabulatedAmount: &tabulatedPrice, Active: true},
		{ID: "t2", Term: "roaming internacional", ChargeType: "mirrored", Active: true},
	}

	match1 := services.MatchExceedanceTerm("Uso de diaria de dados 500MB no exterior", terms)
	if match1 == nil || match1.ID != "t1" {
		t.Errorf("expected match with t1, got %v", match1)
	}

	amount, chargeType := services.ChargedExceedanceAmount(89.90, "normal", match1)
	if chargeType != "tabulated" || amount != 25.00 {
		t.Errorf("expected tabulated charge of 25.00, got %v / %v", amount, chargeType)
	}
}

func TestPhase3_FidelityPenaltyCalculation(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(1, 0, 0) // 12 meses
	cancel := start.AddDate(0, 4, 0) // 4 meses decorridos (restam 8 meses)
	monthlyAmount := 150.00
	penaltyPercent := 30.0

	res := billingcalc.CalculateFidelityPenalty(monthlyAmount, start, end, cancel, penaltyPercent)

	if res.IsExempt {
		t.Errorf("expected penalty not to be exempt")
	}
	if res.MonthsRemaining != 8 {
		t.Errorf("expected 8 months remaining, got %d", res.MonthsRemaining)
	}
	// 8 * 150 * 0.30 = 360.00
	if res.PenaltyAmount != 360.00 {
		t.Errorf("expected penalty 360.00, got %v", res.PenaltyAmount)
	}
}
