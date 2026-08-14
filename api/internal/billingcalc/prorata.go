package billingcalc

import (
	"math"
	"time"
)

const CycleDays = 30

// ProRataAmount applies daily pro-rata on a 30-day cycle.
// Mid-cycle start uses remaining days (day 16 → 15 days). Mid-cycle end uses days elapsed (day 10 → 10 days).
func ProRataAmount(amount float64, cycleStart, itemStart, itemEnd *time.Time) float64 {
	if amount == 0 {
		return 0
	}
	days := ActiveDays(cycleStart, itemStart, itemEnd)
	if days >= CycleDays {
		return round2(amount)
	}
	if days <= 0 {
		return 0
	}
	return round2(amount / float64(CycleDays) * float64(days))
}

func ActiveDays(cycleStart, itemStart, itemEnd *time.Time) int {
	start := dateOnly(cycleStart)
	if start.IsZero() {
		start = time.Date(time.Now().UTC().Year(), time.Now().UTC().Month(), 1, 0, 0, 0, 0, time.UTC)
	}
	cycleEnd := start.AddDate(0, 0, CycleDays-1)

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
	if days > CycleDays {
		return CycleDays
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
	total := amount * quantity
	if itemType == "discount" {
		return -total
	}
	return total
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}
