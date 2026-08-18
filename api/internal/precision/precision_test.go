package precision

import (
	"testing"
)

func TestRoundHalfUp(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		decimals int
		expected float64
	}{
		{"0.005 to 2 decimals", 0.005, 2, 0.01},
		{"0.004 to 2 decimals", 0.004, 2, 0.00},
		{"33.3333 to 2 decimals", 33.3333, 2, 33.33},
		{"33.3366 to 2 decimals", 33.3366, 2, 33.34},
		{"-0.005 to 2 decimals", -0.005, 2, -0.01},
		{"-0.004 to 2 decimals", -0.004, 2, 0.00},
		{"rate 0.01255 to 4 decimals", 0.01255, 4, 0.0126},
		{"rate 0.01254 to 4 decimals", 0.01254, 4, 0.0125},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RoundHalfUp(tt.input, tt.decimals)
			if got != tt.expected {
				t.Errorf("RoundHalfUp(%v, %d) = %v; want %v", tt.input, tt.decimals, got, tt.expected)
			}
		})
	}
}

func TestSplitInstallments(t *testing.T) {
	tests := []struct {
		name         string
		total        float64
		count        int
		expectedList []float64
	}{
		{
			name:         "100 divided by 3",
			total:        100.00,
			count:        3,
			expectedList: []float64{33.33, 33.33, 33.34},
		},
		{
			name:         "1000 divided by 6",
			total:        1000.00,
			count:        6,
			expectedList: []float64{166.66, 166.66, 166.66, 166.66, 166.66, 166.70},
		},
		{
			name:         "50 divided by 2",
			total:        50.00,
			count:        2,
			expectedList: []float64{25.00, 25.00},
		},
		{
			name:         "10 divided by 1",
			total:        10.00,
			count:        1,
			expectedList: []float64{10.00},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			installments, err := SplitInstallments(tt.total, tt.count)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(installments) != len(tt.expectedList) {
				t.Fatalf("expected %d installments, got %d", len(tt.expectedList), len(installments))
			}

			var sum float64
			for i, v := range installments {
				if v != tt.expectedList[i] {
					t.Errorf("installment[%d] = %v; want %v", i, v, tt.expectedList[i])
				}
				sum = SumCents(sum, v)
			}

			if sum != Round2(tt.total) {
				t.Errorf("sum of installments = %v; want %v", sum, tt.total)
			}
		})
	}
}

func TestCalculateInstallmentAmount(t *testing.T) {
	// 100.00 em 3 parcelas
	if got := CalculateInstallmentAmount(100.00, 3, 1); got != 33.33 {
		t.Errorf("parcela 1 de 3 = %v; want 33.33", got)
	}
	if got := CalculateInstallmentAmount(100.00, 3, 2); got != 33.33 {
		t.Errorf("parcela 2 de 3 = %v; want 33.33", got)
	}
	if got := CalculateInstallmentAmount(100.00, 3, 3); got != 33.34 {
		t.Errorf("parcela 3 de 3 (última) = %v; want 33.34", got)
	}
}

func TestCalculateProRata(t *testing.T) {
	tests := []struct {
		name       string
		amount     float64
		cycleDays  int
		activeDays int
		expected   float64
	}{
		{"full cycle 30 days", 30.00, 30, 30, 30.00},
		{"half cycle 15 days of 30.00", 30.00, 30, 15, 15.00},
		{"10 days of 30.00", 30.00, 30, 10, 10.00},
		{"20 days of 60.00", 60.00, 30, 20, 40.00},
		{"zero active days", 60.00, 30, 0, 0.00},
		{"negative active days", 60.00, 30, -5, 0.00},
		{"active days exceed cycle", 60.00, 30, 35, 60.00},
		{"odd division 49.90 for 13 days of 30", 49.90, 30, 13, 21.62}, // 49.90 * 13 / 30 = 21.62333 -> 21.62
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateProRata(tt.amount, tt.cycleDays, tt.activeDays)
			if got != tt.expected {
				t.Errorf("CalculateProRata(%v, %d, %d) = %v; want %v", tt.amount, tt.cycleDays, tt.activeDays, got, tt.expected)
			}
		})
	}
}

func TestSumCents(t *testing.T) {
	// Floating point drift test: 0.1 + 0.2 in standard IEEE-754 is 0.30000000000000004
	values := []float64{0.1, 0.2, 33.33, 33.33, 33.34}
	got := SumCents(values...)
	want := 100.30

	if got != want {
		t.Errorf("SumCents(%v) = %v; want %v", values, got, want)
	}
}
