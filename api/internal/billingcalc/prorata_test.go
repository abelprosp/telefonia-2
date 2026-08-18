package billingcalc

import (
	"testing"
	"time"

	"github.com/luxus-connect/telefonia/api/internal/precision"
)

func TestProRataMidCycleStart(t *testing.T) {
	cycle := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	start := time.Date(2026, 7, 16, 0, 0, 0, 0, time.UTC)
	got := ProRataAmount(30, &cycle, &start, nil)
	if got != 15 {
		t.Fatalf("expected 15, got %v", got)
	}
}

func TestProRataMidCycleEnd(t *testing.T) {
	cycle := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 7, 10, 0, 0, 0, 0, time.UTC)
	got := ProRataAmount(30, &cycle, nil, &end)
	if got != 10 {
		t.Fatalf("expected 10, got %v", got)
	}
}

func TestProRataFullCycle(t *testing.T) {
	cycle := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	got := ProRataAmount(60, &cycle, nil, nil)
	if got != 60 {
		t.Fatalf("expected 60, got %v", got)
	}
}

func TestProRataConfigurableDivisor(t *testing.T) {
	cycle := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	start := time.Date(2026, 2, 15, 0, 0, 0, 0, time.UTC)
	amount := 280.0

	got28 := ProRataAmountWithDivisor(amount, 28, &cycle, &start, nil)
	// Feb 15–28 inclusive = 14 days of 28 → 140
	if got28 != 140 {
		t.Fatalf("divisor 28: expected 140, got %v", got28)
	}

	got30 := ProRataAmountWithDivisor(amount, 30, &cycle, &start, nil)
	// Feb 15–Mar 2 (30-day cycle from Feb 1) = 16 days of 30
	want30 := precision.CalculateProRata(amount, 30, ActiveDaysWithDivisor(30, &cycle, &start, nil))
	if got30 != want30 {
		t.Fatalf("divisor 30: expected %v, got %v", want30, got30)
	}

	got31 := ProRataAmountWithDivisor(310, 31, &cycle, nil, nil)
	if got31 != 310 {
		t.Fatalf("divisor 31 full cycle: expected 310, got %v", got31)
	}
}
