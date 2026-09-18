package store_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/luxus-connect/telefonia/api/internal/models"
	"github.com/luxus-connect/telefonia/api/internal/store"
)

func TestCustomerRegistrationRoundTrip(t *testing.T) {
	f := newUndoFixture(t)
	ctx := context.Background()
	profile := &models.CustomerRegistrationProfile{ServiceUnit: "Unidade Centro", UA: "UA-1", ShowUA: true, Account: "123", ShowAccount: true,
		Birthplace: "Porto Alegre / RS", Sex: "female", IdentityType: "rg", IdentityIssuedOn: "2020-01-02", IdentityIssuer: "SSP", IdentityState: "RS", RG: "1234567",
		FatherName: "José", MotherName: "Maria", ContactPhone: "5133334444", Phone: "5133335555", Mobile: "51999998888", Emails: []string{"contato@example.com", "outro@example.com"},
		ContactName: "Ana", MaritalStatus: "married", SpouseCPF: "52998224725", Profession: "Engenheira", Notes: "Atender pela manhã", ShowNotes: true}
	address := []models.CreateCustomerAddressInput{{Street: "Rua Teste", Number: "10", Neighborhood: "Centro", City: "Porto Alegre", State: "RS", ZipCode: "90000000", Country: "Brasil"}}
	err := f.st.WithTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		return f.st.SetCustomerRegistration(store.CtxWithTx(ctx, tx), f.customer, profile, &address)
	})
	if err != nil {
		t.Fatal(err)
	}
	customer, err := f.st.GetCustomerInOrg(ctx, f.org, f.customer, nil)
	if err != nil {
		t.Fatal(err)
	}
	if customer.Profile == nil || customer.Profile.MotherName != "Maria" || len(customer.Profile.Emails) != 2 || !customer.Profile.ShowUA || len(customer.Addresses) != 1 || customer.Addresses[0].City != "Porto Alegre" {
		t.Fatalf("profile/address lost: %+v", customer)
	}
	// A PATCH with no profile/address must retain the details already saved.
	if err := f.st.SetCustomerRegistration(ctx, f.customer, nil, nil); err != nil {
		t.Fatal(err)
	}
	customer, err = f.st.GetCustomerInOrg(ctx, f.org, f.customer, nil)
	if err != nil {
		t.Fatal(err)
	}
	if customer.Profile.RG != "1234567" {
		t.Fatal("omitted profile was erased")
	}
	// Empty values explicitly clear optional details.
	empty := []models.CreateCustomerAddressInput{}
	err = f.st.WithTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		return f.st.SetCustomerRegistration(store.CtxWithTx(ctx, tx), f.customer, &models.CustomerRegistrationProfile{}, &empty)
	})
	if err != nil {
		t.Fatal(err)
	}
	customer, err = f.st.GetCustomerInOrg(ctx, f.org, f.customer, nil)
	if err != nil {
		t.Fatal(err)
	}
	if customer.Profile.RG != "" || len(customer.Addresses) != 0 {
		t.Fatal("explicit clear not persisted")
	}
}
