package billingcalc

import (
	"testing"

	"github.com/luxus-connect/telefonia/api/internal/precision"
)

func TestApportionGlobalDiscount_ExactDivision(t *testing.T) {
	targets := []ApportionmentTarget{
		{ID: "line1", Amount: 100.00},
		{ID: "line2", Amount: 200.00},
		{ID: "line3", Amount: 300.00},
	}
	globalDiscount := 100.00

	results, err := ApportionGlobalDiscount(globalDiscount, targets)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var sumDiscount float64
	for _, r := range results {
		sumDiscount = precision.SumCents(sumDiscount, r.AllocatedDiscount)
	}

	if sumDiscount != globalDiscount {
		t.Errorf("sum of allocated discounts = %v; want %v", sumDiscount, globalDiscount)
	}

	// Validações individuais
	if results[0].AllocatedDiscount != 16.67 {
		t.Errorf("line1 discount = %v; want 16.67", results[0].AllocatedDiscount)
	}
	if results[1].AllocatedDiscount != 33.33 {
		t.Errorf("line2 discount = %v; want 33.33", results[1].AllocatedDiscount)
	}
	if results[2].AllocatedDiscount != 50.00 {
		t.Errorf("line3 discount = %v; want 50.00", results[2].AllocatedDiscount)
	}
}

func TestApportionGlobalDiscount_OddDivisionWithRemainder(t *testing.T) {
	targets := []ApportionmentTarget{
		{ID: "line1", Amount: 50.00},
		{ID: "line2", Amount: 50.00},
		{ID: "line3", Amount: 50.00},
	}
	globalDiscount := 10.00

	results, err := ApportionGlobalDiscount(globalDiscount, targets)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var sumDiscount float64
	for _, r := range results {
		sumDiscount = precision.SumCents(sumDiscount, r.AllocatedDiscount)
	}

	if sumDiscount != globalDiscount {
		t.Errorf("sum of allocated discounts = %v; want %v", sumDiscount, globalDiscount)
	}
}

func TestApportionGlobalDiscount_ZeroDiscount(t *testing.T) {
	targets := []ApportionmentTarget{
		{ID: "line1", Amount: 100.00},
	}
	results, err := ApportionGlobalDiscount(0, targets)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results[0].AllocatedDiscount != 0 || results[0].FinalAmount != 100.00 {
		t.Errorf("expected 0 discount and 100 final amount, got %v", results[0])
	}
}
