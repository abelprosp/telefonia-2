package services

import (
	"testing"

	"github.com/luxus-connect/telefonia/api/internal/models"
)

func TestPickDashboardMonth(t *testing.T) {
	months := []models.ListProcessingMonthResponse{
		{ID: "aug", Year: 2026, Month: 8, DisplayName: "08/2026"},
		{ID: "jul", Year: 2026, Month: 7, DisplayName: "07/2026"},
	}

	if got := pickDashboardMonth(months, ""); got == nil || got.ID != "aug" {
		t.Fatalf("expected latest month when id is empty, got %#v", got)
	}
	if got := pickDashboardMonth(months, "jul"); got == nil || got.ID != "jul" {
		t.Fatalf("expected selected month, got %#v", got)
	}
	if got := pickDashboardMonth(months, "missing"); got != nil {
		t.Fatalf("expected nil for unknown month, got %#v", got)
	}
	if got := pickDashboardMonth(nil, ""); got != nil {
		t.Fatalf("expected nil for empty list, got %#v", got)
	}
}
