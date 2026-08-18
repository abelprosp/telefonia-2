package services

import (
	"testing"

	"github.com/luxus-connect/telefonia/api/internal/models"
)

func TestValidateCreateCustomerAllowsMissingProvider(t *testing.T) {
	legal := "Empresa LTDA"
	err := validateCreateCustomer(models.CreateCustomerInput{
		Type:      "PJ",
		Name:      "Empresa",
		Document:  "12345678000199",
		LegalName: &legal,
	})
	if err != nil {
		t.Fatalf("expected create without provider_id to be valid, got %v", err)
	}
}
