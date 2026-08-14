package services

import (
	"testing"
	"time"

	"github.com/luxus-connect/telefonia/api/internal/models"
	"github.com/luxus-connect/telefonia/api/internal/store"
)

func TestMatchExceedanceTermPrefersLongest(t *testing.T) {
	terms := []models.ExceedanceTermResponse{
		{Term: "Roaming"},
		{Term: "Roaming Internacional"},
		{Term: "Excedente"},
	}
	got := MatchExceedanceTerm("Roaming Internacional — Europa", terms)
	if got == nil || got.Term != "Roaming Internacional" {
		t.Fatalf("expected Roaming Internacional, got %#v", got)
	}
}

func TestMatchExceedanceTermNone(t *testing.T) {
	terms := []models.ExceedanceTermResponse{{Term: "Roaming"}}
	if MatchExceedanceTerm("Assinatura plano smart", terms) != nil {
		t.Fatal("expected no match")
	}
}

func TestChargedExceedanceAmountTabulated(t *testing.T) {
	tab := 12.5
	term := &models.ExceedanceTermResponse{ChargeType: "tabulated", TabulatedAmount: &tab}
	got, kind := ChargedExceedanceAmount(80, "mirroed", term)
	if got != 12.5 || kind != "tabulated" {
		t.Fatalf("got %v %s", got, kind)
	}
}

func TestChargedExceedanceAmountMirrored(t *testing.T) {
	term := &models.ExceedanceTermResponse{ChargeType: "mirrored"}
	got, kind := ChargedExceedanceAmount(80, "mirroed", term)
	if got != 80 || kind != "mirrored" {
		t.Fatalf("got %v %s", got, kind)
	}
}

func TestPredictedFidelityEnd(t *testing.T) {
	start := mustDate(t, "2026-01-15")
	got := store.PredictedFidelityEnd(start, 12)
	if got.Format("2006-01-02") != "2027-01-15" {
		t.Fatalf("got %s", got.Format("2006-01-02"))
	}
}

func mustDate(t *testing.T, s string) time.Time {
	t.Helper()
	tm, err := time.Parse("2006-01-02", s)
	if err != nil {
		t.Fatal(err)
	}
	return tm
}
