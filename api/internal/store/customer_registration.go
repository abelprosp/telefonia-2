package store

import (
	"context"
	"encoding/json"

	"github.com/luxus-connect/telefonia/api/internal/models"
)

func (s *Store) SetCustomerRegistration(ctx context.Context, id string, profile *models.CustomerRegistrationProfile, addresses *[]models.CreateCustomerAddressInput) error {
	if profile != nil {
		body, err := json.Marshal(profile)
		if err != nil {
			return err
		}
		if _, err = s.q(ctx).Exec(ctx, `UPDATE "Customers" SET "RegistrationProfile"=$2::jsonb WHERE "Id"=$1`, id, body); err != nil {
			return err
		}
	}
	if addresses != nil {
		if _, err := s.q(ctx).Exec(ctx, `DELETE FROM "CustomerAddresses" WHERE "CustomerId"=$1`, id); err != nil {
			return err
		}
		for _, addr := range *addresses {
			if err := s.CreateCustomerAddress(ctx, id, addr); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Store) loadCustomerRegistration(ctx context.Context, customer *models.ListCustomerResponse) error {
	var profile models.CustomerRegistrationProfile
	if err := s.q(ctx).QueryRow(ctx, `SELECT "RegistrationProfile" FROM "Customers" WHERE "Id"=$1`, customer.ID).Scan(&profile); err != nil {
		return err
	}
	customer.Profile = &profile
	rows, err := s.q(ctx).Query(ctx, `SELECT "Street","Number","Neighborhood","City","State","ZipCode","Complement","Country" FROM "CustomerAddresses" WHERE "CustomerId"=$1 ORDER BY "Id"`, customer.ID)
	if err != nil {
		return err
	}
	defer rows.Close()
	customer.Addresses = []models.CreateCustomerAddressInput{}
	for rows.Next() {
		var a models.CreateCustomerAddressInput
		if err := rows.Scan(&a.Street, &a.Number, &a.Neighborhood, &a.City, &a.State, &a.ZipCode, &a.Complement, &a.Country); err != nil {
			return err
		}
		customer.Addresses = append(customer.Addresses, a)
	}
	return rows.Err()
}
