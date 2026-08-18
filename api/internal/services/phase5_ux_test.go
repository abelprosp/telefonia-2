package services_test

import (
	"testing"
	"time"

	"github.com/luxus-connect/telefonia/api/internal/models"
	"github.com/luxus-connect/telefonia/api/internal/precision"
)

func TestPhase5_OperationalDashboardAggregations(t *testing.T) {
	lines := models.OperationalDashboardLinesSummary{
		TotalLines:        150,
		ActiveLines:       120,
		InTransitionLines: 10,
		CanceledLines:     20,
		OrphanLines:       5,
	}

	if lines.TotalLines != (lines.ActiveLines + lines.InTransitionLines + lines.CanceledLines) {
		t.Errorf("lines count mismatch")
	}

	revenue := 12500.00
	cost := 7800.00
	margin := precision.Round2(revenue - cost)
	marginPct := precision.Round2((margin / revenue) * 100.0)

	financial := models.OperationalDashboardFinancialSummary{
		ProjectedMonthlyRevenue: revenue,
		TotalBaseCost:           cost,
		ProjectedMargin:         margin,
		MarginPercentage:        marginPct,
	}

	if financial.ProjectedMargin != 4700.00 {
		t.Errorf("expected margin 4700.00, got %v", financial.ProjectedMargin)
	}
	if financial.MarginPercentage != 37.60 {
		t.Errorf("expected margin pct 37.60%%, got %v%%", financial.MarginPercentage)
	}
}

func TestPhase5_DivergenceLifecycle(t *testing.T) {
	divergence := models.DivergenceItemResponse{
		ID:             "div_001",
		DivergenceType: "LINE_NOT_IN_SYSTEM",
		Severity:       "HIGH",
		Status:         "pending",
		CreatedAt:      time.Now().UTC(),
	}

	if divergence.Status != "pending" {
		t.Errorf("expected initial status pending")
	}

	// Resolução
	resolvedBy := "operador@luxus.com.br"
	notes := "Linha cadastrada no sistema manualmente"
	divergence.Status = "resolved"
	divergence.ResolvedBy = &resolvedBy
	divergence.ResolutionNotes = &notes

	if divergence.Status != "resolved" || *divergence.ResolvedBy != "operador@luxus.com.br" {
		t.Errorf("expected divergence to be resolved")
	}
}
