// Package phone provides central telephone number normalization for comparisons,
// imports, search and reconciliations. It strips presentation characters only;
// it does not strip DDI 55 or other significant digits automatically.
package phone

import (
	"strings"
	"unicode"
)

// Result holds the original input (trimmed) and the digit-only normalized form.
type Result struct {
	Original   string
	Normalized string
}

// Normalize removes masks, spaces, parentheses, hyphens and other non-digit
// presentation characters. Empty input yields empty Normalized.
//
// Example: "(51) 99999-8888" → "51999998888"
func Normalize(raw string) Result {
	original := strings.TrimSpace(raw)
	if original == "" {
		return Result{}
	}
	var b strings.Builder
	b.Grow(len(original))
	for _, r := range original {
		if unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return Result{Original: original, Normalized: b.String()}
}

// Digits is a convenience wrapper returning only the normalized digit string.
func Digits(raw string) string {
	return Normalize(raw).Normalized
}

// Equal compares two numbers after normalization.
func Equal(a, b string) bool {
	na, nb := Digits(a), Digits(b)
	if na == "" || nb == "" {
		return false
	}
	return na == nb
}

// IsValidBasic rejects empty and non-digit-only results shorter than minDigits
// or longer than maxDigits. Brazilian mobile/landline without DDI is typically
// 10–11 digits; with DDI 55 it may be 12–13. Callers choose bounds.
func IsValidBasic(raw string, minDigits, maxDigits int) bool {
	n := Digits(raw)
	if n == "" {
		return false
	}
	if minDigits > 0 && len(n) < minDigits {
		return false
	}
	if maxDigits > 0 && len(n) > maxDigits {
		return false
	}
	return true
}

// Canonical SIM form factors used across cadastro, import and reconciliation.
const (
	SimPhysical = "PHYSICAL"
	SimEsim     = "ESIM"
	SimUnknown  = "UNKNOWN"
)

// NormalizeSimType maps free-text labels to PHYSICAL | ESIM | UNKNOWN.
func NormalizeSimType(raw string) string {
	v := strings.ToUpper(strings.TrimSpace(raw))
	v = strings.ReplaceAll(v, "-", "")
	v = strings.ReplaceAll(v, "_", "")
	v = strings.ReplaceAll(v, " ", "")
	switch v {
	case "PHYSICAL", "SIM", "SIMPHYSICAL", "FISICO", "FÍSICO", "CHIP":
		return SimPhysical
	case "ESIM", "ESIMDIGITAL", "DIGITAL", "EMBEDDED":
		return SimEsim
	case "", "UNKNOWN", "DESCONHECIDO", "N/A", "NA":
		return SimUnknown
	default:
		return SimUnknown
	}
}
