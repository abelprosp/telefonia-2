package services

import (
	"testing"

	"github.com/luxus-connect/telefonia/api/internal/models"
)

func TestCustomerRegistrationSeparatesPFAndPJ(t *testing.T) {
	personal := models.CustomerRegistrationProfile{RG: "1234567", MotherName: "Maria", SpouseCPF: "529.982.247-25", Sex: "female"}
	if err := normalizeCustomerRegistration("pj", &personal, nil, nil); err == nil {
		t.Fatal("PJ accepted personal fields")
	}
	if err := normalizeCustomerRegistration("pf", &personal, nil, nil); err != nil {
		t.Fatal(err)
	}
	if personal.SpouseCPF != "52998224725" {
		t.Fatal("CPF not normalized")
	}
	company := models.CustomerRegistrationProfile{ServiceUnit: " Unidade 1 ", ContactName: "Maria", Phone: "(11) 3333-4444", Emails: []string{" financeiro@example.com ", "financeiro@example.com", "contato@example.com"}}
	if err := normalizeCustomerRegistration("pj", &company, nil, nil); err != nil {
		t.Fatal(err)
	}
	if company.ServiceUnit != "Unidade 1" || company.Phone != "1133334444" || len(company.Emails) != 2 {
		t.Fatalf("unexpected normalization: %+v", company)
	}
}

func TestCustomerRegistrationRejectsInvalidValues(t *testing.T) {
	for name, p := range map[string]models.CustomerRegistrationProfile{
		"invalid email": {Emails: []string{"foo@"}}, "invalid cpf": {SpouseCPF: "11111111111"},
		"invalid date": {IdentityIssuedOn: "2026-02-31"}, "future date": {IdentityIssuedOn: "2099-01-01"},
		"invalid UF": {IdentityState: "XX"}, "invalid phone": {Mobile: "123"},
	} {
		t.Run(name, func(t *testing.T) {
			if err := normalizeCustomerRegistration("pf", &p, nil, nil); err == nil {
				t.Fatal("invalid value accepted")
			}
		})
	}
	addresses := []models.CreateCustomerAddressInput{{ZipCode: "12", State: "SP"}}
	if err := normalizeCustomerRegistration("pf", nil, addresses, nil); err == nil {
		t.Fatal("invalid CEP accepted")
	}
}
