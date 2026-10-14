package billingcalc

import (
	"fmt"
	"sort"

	"github.com/luxus-connect/telefonia/api/internal/precision"
)

type ApportionmentTarget struct {
	ID     string  `json:"id"`
	Amount float64 `json:"amount"`
}

type ApportionedItem struct {
	ID                string  `json:"id"`
	OriginalAmount    float64 `json:"original_amount"`
	AllocatedDiscount float64 `json:"allocated_discount"`
	FinalAmount       float64 `json:"final_amount"`
}

// ApportionGlobalDiscount distribui um desconto global entre múltiplos alvos proporcionalmente aos seus valores.
// O resíduo de centavos é alocado no item de maior valor para garantir precisão exata: Sum(AllocatedDiscount) == globalDiscount.
func ApportionGlobalDiscount(globalDiscount float64, targets []ApportionmentTarget) ([]ApportionedItem, error) {
	if globalDiscount <= 0 {
		result := make([]ApportionedItem, len(targets))
		for i, t := range targets {
			result[i] = ApportionedItem{
				ID:                t.ID,
				OriginalAmount:    t.Amount,
				AllocatedDiscount: 0,
				FinalAmount:       t.Amount,
			}
		}
		return result, nil
	}

	if len(targets) == 0 {
		return nil, fmt.Errorf("nenhum alvo fornecido para rateio de desconto")
	}

	var totalAmount float64

	for _, t := range targets {
		if t.Amount > 0 {
			totalAmount = precision.SumCents(totalAmount, t.Amount)
		}
	}

	if totalAmount <= 0 {
		return nil, fmt.Errorf("o somatório dos valores dos alvos é zero ou negativo")
	}

	results := make([]ApportionedItem, len(targets))
	var totalAllocated float64

	for i, t := range targets {
		if t.Amount <= 0 {
			results[i] = ApportionedItem{
				ID:                t.ID,
				OriginalAmount:    t.Amount,
				AllocatedDiscount: 0,
				FinalAmount:       t.Amount,
			}
			continue
		}

		// Cálculo proporcional: DescontoGlobal * (ValorItem / Total)
		fraction := t.Amount / totalAmount
		discountItem := precision.Round2(globalDiscount * fraction)

		// Desconto não pode exceder o valor do próprio item
		if discountItem > t.Amount {
			discountItem = t.Amount
		}

		results[i] = ApportionedItem{
			ID:                t.ID,
			OriginalAmount:    t.Amount,
			AllocatedDiscount: discountItem,
			FinalAmount:       precision.Round2(t.Amount - discountItem),
		}
		totalAllocated = precision.SumCents(totalAllocated, discountItem)
	}

	// Ajuste do resíduo de centavos para garantir que Sum(AllocatedDiscount) == globalDiscount.
	// Distribui o que sobra (ou retira o excesso) em itens com folga, priorizando os de maior valor.
	order := make([]int, 0, len(results))
	for i := range results {
		if results[i].OriginalAmount > 0 {
			order = append(order, i)
		}
	}
	sort.Slice(order, func(a, b int) bool {
		return results[order[a]].OriginalAmount > results[order[b]].OriginalAmount
	})
	for _, i := range order {
		remainder := precision.Round2(globalDiscount - totalAllocated)
		if remainder == 0 {
			break
		}
		var delta float64
		if remainder > 0 {
			headroom := precision.Round2(results[i].OriginalAmount - results[i].AllocatedDiscount)
			if headroom <= 0 {
				continue
			}
			delta = remainder
			if delta > headroom {
				delta = headroom
			}
		} else {
			capacity := results[i].AllocatedDiscount
			if capacity <= 0 {
				continue
			}
			delta = remainder
			if -delta > capacity {
				delta = -capacity
			}
		}
		results[i].AllocatedDiscount = precision.Round2(results[i].AllocatedDiscount + delta)
		results[i].FinalAmount = precision.Round2(results[i].OriginalAmount - results[i].AllocatedDiscount)
		totalAllocated = precision.SumCents(totalAllocated, delta)
	}

	return results, nil
}
