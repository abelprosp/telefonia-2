package privacy

import (
	"regexp"
	"strings"
)

var (
	cpfDigitsRegex  = regexp.MustCompile(`\D`)
	emailUserRegex  = regexp.MustCompile(`^([^@]+)@(.+)$`)
	phoneDigitsOnly = regexp.MustCompile(`\D`)
)

// MaskDocument mascara CPFs e CNPJs preservando formato e primeiros/últimos dígitos.
func MaskDocument(doc string) string {
	clean := cpfDigitsRegex.ReplaceAllString(doc, "")
	if len(clean) == 11 {
		// CPF: 123.***.***-01
		return clean[:3] + ".***.***-" + clean[9:]
	} else if len(clean) == 14 {
		// CNPJ: 12.***.***/0001-90
		return clean[:2] + ".***.***/" + clean[8:12] + "-" + clean[12:]
	}
	if len(doc) > 4 {
		return doc[:2] + strings.Repeat("*", len(doc)-4) + doc[len(doc)-2:]
	}
	return strings.Repeat("*", len(doc))
}

// MaskEmail mascara o endereço de e-mail preservando os 2 primeiros caracteres e o domínio.
func MaskEmail(email string) string {
	trimmed := strings.TrimSpace(email)
	matches := emailUserRegex.FindStringSubmatch(trimmed)
	if len(matches) < 3 {
		return strings.Repeat("*", len(trimmed))
	}
	user := matches[1]
	domain := matches[2]

	if len(user) <= 2 {
		return user + "***@" + domain
	}
	return user[:2] + strings.Repeat("*", len(user)-2) + "@" + domain
}

// MaskPhone mascara telefones preservando DDD e os últimos 2 dígitos.
func MaskPhone(phone string) string {
	clean := phoneDigitsOnly.ReplaceAllString(phone, "")
	if len(clean) == 11 {
		// (11) 9****-**34
		return "(" + clean[:2] + ") " + clean[2:3] + "****-**" + clean[9:]
	} else if len(clean) == 10 {
		// (11) ****-**34
		return "(" + clean[:2] + ") ****-**" + clean[8:]
	}
	if len(phone) > 4 {
		return phone[:2] + strings.Repeat("*", len(phone)-4) + phone[len(phone)-2:]
	}
	return strings.Repeat("*", len(phone))
}
