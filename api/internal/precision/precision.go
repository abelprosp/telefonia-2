package precision

import (
	"fmt"
	"math"
)

// RoundHalfUp arredonda o valor financeiro usando a regra Half-Up (>= 0.5 arredonda para cima).
// Utiliza conversão segura para evitar imprecisões de representação IEEE-754.
func RoundHalfUp(val float64, decimals int) float64 {
	if math.IsNaN(val) || math.IsInf(val, 0) {
		return 0
	}
	pow := math.Pow10(decimals)
	if val < 0 {
		return math.Ceil(val*pow-0.5) / pow
	}
	return math.Floor(val*pow+0.5) / pow
}

// Round2 é o atalho para arredondamento padrão em 2 casas decimais (centavos de Real).
func Round2(val float64) float64 {
	return RoundHalfUp(val, 2)
}

// Round4 é o atalho para precisão intermediária (taxas de juros, multas e coeficientes).
func Round4(val float64) float64 {
	return RoundHalfUp(val, 4)
}

// ToCents converte um valor em reais para um inteiro de centavos (evitando floats acumulativos).
func ToCents(val float64) int64 {
	return int64(math.Round(val * 100))
}

// FromCents converte um valor de centavos de volta para float64 com 2 casas.
func FromCents(cents int64) float64 {
	return float64(cents) / 100.0
}

// SumCents efetua o somatório de múltiplos valores em centavos inteiros, eliminando drift de ponto flutuante.
func SumCents(values ...float64) float64 {
	var totalCents int64
	for _, v := range values {
		totalCents += ToCents(v)
	}
	return FromCents(totalCents)
}

// SplitInstallments divide um valor total em N parcelas, distribuindo o resíduo de centavos na última parcela.
// Garante que a soma de todas as parcelas seja estritamente igual a Round2(totalAmount).
func SplitInstallments(totalAmount float64, count int) ([]float64, error) {
	if count <= 0 {
		return nil, fmt.Errorf("número de parcelas deve ser maior que zero (recebido: %d)", count)
	}
	totalCents := ToCents(totalAmount)
	baseInstallmentCents := totalCents / int64(count)
	remainderCents := totalCents % int64(count)

	installments := make([]float64, count)
	for i := 0; i < count; i++ {
		if i == count-1 {
			// A última parcela absorve o resíduo
			installments[i] = FromCents(baseInstallmentCents + remainderCents)
		} else {
			installments[i] = FromCents(baseInstallmentCents)
		}
	}
	return installments, nil
}

// CalculateInstallmentAmount calcula o valor de uma parcela específica (1 a count).
// A última parcela absorve o resíduo de arredondamento.
func CalculateInstallmentAmount(totalAmount float64, count, current int) float64 {
	if count <= 0 || current <= 0 || current > count {
		if count > 0 {
			return Round2(totalAmount / float64(count))
		}
		return Round2(totalAmount)
	}
	installments, err := SplitInstallments(totalAmount, count)
	if err != nil {
		return Round2(totalAmount)
	}
	return installments[current-1]
}

// CalculateProRata calcula o valor proporcional estrito de um serviço ou composição com base no ciclo.
func CalculateProRata(amount float64, cycleDays, activeDays int) float64 {
	if amount == 0 || activeDays <= 0 || cycleDays <= 0 {
		return 0
	}
	if activeDays >= cycleDays {
		return Round2(amount)
	}
	// Multiplica primeiro e divide em centavos para máxima acurácia
	amountCents := ToCents(amount)
	proRataCents := RoundHalfUp(float64(amountCents*int64(activeDays))/float64(cycleDays), 0)
	return FromCents(int64(proRataCents))
}
