package httputil

import "strings"

func ImportRequestStatusString(status int) string {
	switch status {
	case 0:
		return "pending"
	case 1:
		return "processing"
	case 2:
		return "completed"
	case 3:
		return "failed"
	default:
		return "pending"
	}
}

func NormalizeDigits(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func CustomerTypeFromInput(t string) string {
	switch strings.ToUpper(strings.TrimSpace(t)) {
	case "PF":
		return "pf"
	case "PJ":
		return "pj"
	default:
		return strings.ToLower(t)
	}
}

func DocumentTypeForCustomer(customerType string) string {
	if customerType == "pj" {
		return "cnpj"
	}
	return "cpf"
}

// NormalizePhoneLineStatus maps API/query values (e.g. IN_STOCK) to PostgreSQL enum wire values.
func NormalizePhoneLineStatus(status string) string {
	return strings.ToLower(strings.TrimSpace(status))
}
