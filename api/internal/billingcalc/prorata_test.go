package billingcalc

import (
	"testing"
	"time"
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
