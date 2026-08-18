package billingcalc

import (
	"math"
	"time"

	"github.com/luxus-connect/telefonia/api/internal/precision"
)

// FidelityPenaltyResult encapsula o cálculo detalhado de multa por rescisão antecipada dentro da fidelidade.
type FidelityPenaltyResult struct {
	FidelityStartDate time.Time `json:"fidelity_start_date"`
	PredictedEndDate  time.Time `json:"predicted_end_date"`
	CancelDate        time.Time `json:"cancel_date"`
	TotalMonths       int       `json:"total_months"`
	MonthsServed      int       `json:"months_served"`
	MonthsRemaining   int       `json:"months_remaining"`
	MonthlyAmount     float64   `json:"monthly_amount"`
	PenaltyPercentage float64   `json:"penalty_percentage"`
	PenaltyAmount     float64   `json:"penalty_amount"`
	IsExempt          bool      `json:"is_exempt"`
}

// CalculateFidelityPenalty calcula a multa proporcional conforme normas de telecomunicações.
// Se cancelDate >= predictedEndDate, a fidelidade já foi cumprida e a multa é R$ 0,00.
func CalculateFidelityPenalty(monthlyAmount float64, startDate, predictedEndDate, cancelDate time.Time, penaltyPercent float64) FidelityPenaltyResult {
	start := startDate.UTC()
	end := predictedEndDate.UTC()
	cancel := cancelDate.UTC()

	if penaltyPercent <= 0 {
		penaltyPercent = 30.0 // 30% padrão sobre as parcelas restantes
	}

	result := FidelityPenaltyResult{
		FidelityStartDate: start,
		PredictedEndDate:  end,
		CancelDate:        cancel,
		MonthlyAmount:     monthlyAmount,
		PenaltyPercentage: penaltyPercent,
	}

	// Se a data de cancelamento for após ou no término da fidelidade, isento
	if !cancel.Before(end) {
		result.IsExempt = true
		return result
	}

	// Calcula os meses totais e restantes
	daysTotal := end.Sub(start).Hours() / 24.0
	daysServed := cancel.Sub(start).Hours() / 24.0
	if daysServed < 0 {
		daysServed = 0
	}

	totalMonths := int(math.Round(daysTotal / 30.0))
	if totalMonths <= 0 {
		totalMonths = 12
	}
	monthsServed := int(daysServed / 30.0)
	monthsRemaining := totalMonths - monthsServed
	if monthsRemaining <= 0 {
		result.IsExempt = true
		return result
	}

	result.TotalMonths = totalMonths
	result.MonthsServed = monthsServed
	result.MonthsRemaining = monthsRemaining

	// Multa = ValorMensalidade * MesesRestantes * (Percentual / 100)
	rawPenalty := monthlyAmount * float64(monthsRemaining) * (penaltyPercent / 100.0)
	result.PenaltyAmount = precision.Round2(rawPenalty)

	return result
}
