package services

import (
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/models"
	"github.com/luxus-connect/telefonia/api/internal/notifications"
)

func registrationError(message string) error {
	return httputil.ValidationError(notifications.N("CUSTOMER_REGISTRATION_INVALID", message))
}

func normalizeCustomerRegistration(typ string, p *models.CustomerRegistrationProfile, addresses []models.CreateCustomerAddressInput, billingEmail *string) error {
	if billingEmail != nil {
		*billingEmail = strings.TrimSpace(*billingEmail)
		if *billingEmail != "" && !validRegistrationEmail(*billingEmail) {
			return registrationError("Informe um e-mail de cobrança válido.")
		}
	}
	if len(addresses) > 10 {
		return registrationError("Informe no máximo 10 endereços.")
	}
	for i := range addresses {
		a := &addresses[i]
		for value, max := range map[*string]int{&a.Street: 256, &a.Number: 16, &a.Neighborhood: 128, &a.City: 128, &a.State: 64, &a.ZipCode: 10, &a.Country: 32} {
			*value = strings.TrimSpace(*value)
			if utf8.RuneCountInString(*value) > max {
				return registrationError("Um campo do endereço excede o tamanho permitido.")
			}
		}
		a.ZipCode = httputil.NormalizeDigits(a.ZipCode)
		a.State = strings.ToUpper(a.State)
		if a.ZipCode != "" && len(a.ZipCode) != 8 {
			return registrationError("O CEP deve ter 8 dígitos.")
		}
		if a.Complement != nil && utf8.RuneCountInString(*a.Complement) > 128 {
			return registrationError("O complemento deve ter no máximo 128 caracteres.")
		}
		if a.State != "" && !validRegistrationState(a.State) {
			return registrationError("Informe uma UF válida no endereço.")
		}
	}
	if p == nil {
		return nil
	}
	for _, value := range []*string{&p.ServiceUnit, &p.UA, &p.Account, &p.Birthplace, &p.IdentityType, &p.IdentityIssuer, &p.IdentityState, &p.RG, &p.FatherName, &p.MotherName, &p.ContactName, &p.Profession, &p.Sex, &p.MaritalStatus, &p.IdentityIssuedOn} {
		*value = strings.TrimSpace(*value)
		if utf8.RuneCountInString(*value) > 256 {
			return registrationError("Os campos cadastrais devem ter no máximo 256 caracteres.")
		}
	}
	p.Notes = strings.TrimSpace(p.Notes)
	if utf8.RuneCountInString(p.Notes) > 4000 {
		return registrationError("A observação deve ter no máximo 4.000 caracteres.")
	}
	if strings.EqualFold(typ, "pj") {
		if p.Birthplace != "" || p.Sex != "" || p.IdentityType != "" || p.IdentityIssuedOn != "" || p.IdentityIssuer != "" || p.IdentityState != "" || p.RG != "" || p.FatherName != "" || p.MotherName != "" || p.MaritalStatus != "" || p.SpouseCPF != "" || p.Profession != "" {
			return registrationError("Dados pessoais, identidade e família são exclusivos de pessoa física.")
		}
	}
	for _, phone := range []*string{&p.ContactPhone, &p.Phone, &p.Mobile} {
		*phone = httputil.NormalizeDigits(*phone)
		if *phone != "" && (len(*phone) < 10 || len(*phone) > 15) {
			return registrationError("Informe telefones com DDD, entre 10 e 15 dígitos.")
		}
	}
	if len(p.Emails) > 10 {
		return registrationError("Informe no máximo 10 e-mails de contato.")
	}
	emails := []string{}
	seen := map[string]bool{}
	for _, value := range p.Emails {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if !validRegistrationEmail(value) {
			return registrationError("Informe e-mails válidos, separados por vírgula.")
		}
		if !seen[strings.ToLower(value)] {
			emails = append(emails, value)
			seen[strings.ToLower(value)] = true
		}
	}
	p.Emails = emails
	if !oneOf(p.Sex, "", "female", "male", "other", "not_informed") {
		return registrationError("Selecione uma opção válida para sexo.")
	}
	if !oneOf(p.MaritalStatus, "", "single", "married", "divorced", "widowed", "separated", "civil_union") {
		return registrationError("Selecione um estado civil válido.")
	}
	if !oneOf(p.IdentityType, "", "rg", "cnh", "passport", "rne", "other") {
		return registrationError("Selecione um tipo de documento válido.")
	}
	p.IdentityState = strings.ToUpper(p.IdentityState)
	if p.IdentityState != "" && !validRegistrationState(p.IdentityState) {
		return registrationError("Informe uma UF válida para o documento.")
	}
	if p.IdentityIssuedOn != "" {
		issued, err := time.Parse("2006-01-02", p.IdentityIssuedOn)
		if err != nil || issued.After(time.Now()) {
			return registrationError("Informe uma data de emissão válida, que não seja futura.")
		}
	}
	p.SpouseCPF = httputil.NormalizeDigits(p.SpouseCPF)
	if p.SpouseCPF != "" && !validSpouseCPF(p.SpouseCPF) {
		return registrationError("Informe um CPF válido para o cônjuge ou companheiro.")
	}
	return nil
}

func validRegistrationEmail(value string) bool {
	a, err := mail.ParseAddress(value)
	return err == nil && a.Address == value && len(value) <= 256
}
func oneOf(value string, options ...string) bool {
	for _, option := range options {
		if value == option {
			return true
		}
	}
	return false
}
func validRegistrationState(value string) bool {
	return oneOf(value, "AC", "AL", "AP", "AM", "BA", "CE", "DF", "ES", "GO", "MA", "MT", "MS", "MG", "PA", "PB", "PR", "PE", "PI", "RJ", "RN", "RS", "RO", "RR", "SC", "SP", "SE", "TO")
}
func validSpouseCPF(value string) bool {
	if len(value) != 11 || value == strings.Repeat(value[:1], 11) {
		return false
	}
	for size := 9; size <= 10; size++ {
		sum := 0
		for i := 0; i < size; i++ {
			sum += int(value[i]-'0') * (size + 1 - i)
		}
		digit := (sum * 10) % 11
		if digit == 10 {
			digit = 0
		}
		if digit != int(value[size]-'0') {
			return false
		}
	}
	return true
}
