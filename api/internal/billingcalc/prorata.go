package billingcalc

import (
	"time"

	"github.com/luxus-connect/telefonia/api/internal/precision"
)

const CycleDays = 30

func NormalizeDivisor(divisor int) int {
	if divisor < 1 {
		return CycleDays
	}
	if divisor > 31 {
		return 31
	}
	return divisor
}

// ProRataAmount applies daily pro-rata on a 30-day cycle using deterministic financial precision.
func ProRataAmount(amount float64, cycleStart, itemStart, itemEnd *time.Time) float64 {
	return ProRataAmountWithDivisor(amount, CycleDays, cycleStart, itemStart, itemEnd)
}

func ProRataAmountWithDivisor(amount float64, divisor int, cycleStart, itemStart, itemEnd *time.Time) float64 {
	if amount == 0 {
		return 0
	}
	divisor = NormalizeDivisor(divisor)
	days := ActiveDaysWithDivisor(divisor, cycleStart, itemStart, itemEnd)
	return precision.CalculateProRata(amount, divisor, days)
}

func ActiveDays(cycleStart, itemStart, itemEnd *time.Time) int {
	return ActiveDaysWithDivisor(CycleDays, cycleStart, itemStart, itemEnd)
}

func ActiveDaysWithDivisor(divisor int, cycleStart, itemStart, itemEnd *time.Time) int {
	divisor = NormalizeDivisor(divisor)
	start := dateOnly(cycleStart)
	if start.IsZero() {
		start = time.Date(time.Now().UTC().Year(), time.Now().UTC().Month(), 1, 0, 0, 0, 0, time.UTC)
	}
	cycleEnd := start.AddDate(0, 0, divisor-1)

	from := start
	if itemStart != nil {
		is := dateOnly(itemStart)
		if is.After(cycleEnd) {
			return 0
		}
		if is.After(from) {
			from = is
		}
	}
	to := cycleEnd
	if itemEnd != nil {
		ie := dateOnly(itemEnd)
		if ie.Before(start) {
			return 0
		}
		if ie.Before(to) {
			to = ie
		}
	}
	if to.Before(from) {
		return 0
	}
	days := int(to.Sub(from).Hours()/24) + 1
	if days > divisor {
		return divisor
	}
	return days
}

func dateOnly(t *time.Time) time.Time {
	if t == nil || t.IsZero() {
		return time.Time{}
	}
	y, m, d := t.UTC().Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func CycleStartFromYearMonth(year, month int) time.Time {
	return time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
}

func SignedItemAmount(itemType string, amount, quantity float64) float64 {
	if quantity <= 0 {
		quantity = 1
	}
	total := precision.Round2(amount * quantity)
	if itemType == "discount" {
		return -total
	}
	return total
}

func round2(v float64) float64 {
	return precision.Round2(v)
}
