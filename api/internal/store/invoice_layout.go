package store

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/models"
)

func (s *Store) ListInvoiceLayoutTemplates(ctx context.Context, orgID string, page httputil.PageSearch) ([]models.ListInvoiceLayoutTemplateResponse, int64, error) {
	base := ` FROM "InvoiceLayoutTemplates" WHERE "OrganizationId" = $1`
	args := []any{orgID}
	var total int64
	if err := s.q(ctx).QueryRow(ctx, `SELECT COUNT(*) `+base, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	selectQ := `
		SELECT "Id", "Name", "Code", "Active", "CreatedAt", "UpdatedAt"
		` + base + `
		ORDER BY "Name"
		OFFSET $` + itoa(len(args)+1) + ` LIMIT $` + itoa(len(args)+2)
	args = append(args, page.Offset(), page.Limit())
	rows, err := s.q(ctx).Query(ctx, selectQ, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []models.ListInvoiceLayoutTemplateResponse
	for rows.Next() {
		var item models.ListInvoiceLayoutTemplateResponse
		if err := rows.Scan(&item.ID, &item.Name, &item.Code, &item.Active, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (s *Store) GetInvoiceLayoutTemplate(ctx context.Context, orgID, id string) (*models.GetInvoiceLayoutTemplateResponse, error) {
	var item models.GetInvoiceLayoutTemplateResponse
	var configRaw []byte
	err := s.q(ctx).QueryRow(ctx, `
		SELECT "Id", "Name", "Code", "ConfigJson", "Active", "CreatedAt", "UpdatedAt"
		FROM "InvoiceLayoutTemplates"
		WHERE "OrganizationId" = $1 AND "Id" = $2`, orgID, id).Scan(
		&item.ID, &item.Name, &item.Code, &configRaw, &item.Active, &item.CreatedAt, &item.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	item.ConfigJson = json.RawMessage(configRaw)
	return &item, nil
}

func (s *Store) GetInvoiceLayoutTemplateByCode(ctx context.Context, orgID, code string) (*models.GetInvoiceLayoutTemplateResponse, error) {
	var item models.GetInvoiceLayoutTemplateResponse
	var configRaw []byte
	err := s.q(ctx).QueryRow(ctx, `
		SELECT "Id", "Name", "Code", "ConfigJson", "Active", "CreatedAt", "UpdatedAt"
		FROM "InvoiceLayoutTemplates"
		WHERE "OrganizationId" = $1 AND "Code" = $2 AND "Active" = true`, orgID, code).Scan(
		&item.ID, &item.Name, &item.Code, &configRaw, &item.Active, &item.CreatedAt, &item.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	item.ConfigJson = json.RawMessage(configRaw)
	return &item, nil
}

func (s *Store) InvoiceLayoutTemplateCodeExists(ctx context.Context, orgID, code string, excludeID *string) (bool, error) {
	q := `SELECT EXISTS(SELECT 1 FROM "InvoiceLayoutTemplates" WHERE "OrganizationId" = $1 AND "Code" = $2`
	args := []any{orgID, code}
	if excludeID != nil && *excludeID != "" {
		q += ` AND "Id" <> $3`
		args = append(args, *excludeID)
	}
	q += `)`
	var exists bool
	err := s.q(ctx).QueryRow(ctx, q, args...).Scan(&exists)
	return exists, err
}

func (s *Store) CreateInvoiceLayoutTemplate(ctx context.Context, id, orgID, name, code string, config json.RawMessage, active bool, now time.Time) error {
	_, err := s.q(ctx).Exec(ctx, `
		INSERT INTO "InvoiceLayoutTemplates" (
			"Id", "OrganizationId", "Name", "Code", "ConfigJson", "Active", "CreatedAt", "UpdatedAt"
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $7)`,
		id, orgID, name, code, config, active, now)
	return err
}

func (s *Store) UpdateInvoiceLayoutTemplate(ctx context.Context, orgID, id, name string, config json.RawMessage, active bool, now time.Time) error {
	tag, err := s.q(ctx).Exec(ctx, `
		UPDATE "InvoiceLayoutTemplates"
		SET "Name" = $3, "ConfigJson" = $4, "Active" = $5, "UpdatedAt" = $6
		WHERE "OrganizationId" = $1 AND "Id" = $2`, orgID, id, name, config, active, now)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

type CustomerAddressParts struct {
	Street       string
	Number       string
	Neighborhood string
	City         string
	State        string
	ZipCode      string
}

func (s *Store) GetCustomerAddressParts(ctx context.Context, customerID string) (*CustomerAddressParts, error) {
	var item CustomerAddressParts
	err := s.q(ctx).QueryRow(ctx, `
		SELECT "Street", "Number", "Neighborhood", "City", "State", "ZipCode"
		FROM "CustomerAddresses"
		WHERE "CustomerId" = $1
		ORDER BY "Id"
		LIMIT 1`, customerID).Scan(&item.Street, &item.Number, &item.Neighborhood, &item.City, &item.State, &item.ZipCode)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *Store) GetCustomerAddressForBilling(ctx context.Context, customerID string) (string, error) {
	var street, number, neighborhood, city, state, zip string
	err := s.q(ctx).QueryRow(ctx, `
		SELECT "Street", "Number", "Neighborhood", "City", "State", "ZipCode"
		FROM "CustomerAddresses"
		WHERE "CustomerId" = $1
		ORDER BY "Id"
		LIMIT 1`, customerID).Scan(&street, &number, &neighborhood, &city, &state, &zip)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return street + ", " + number + ", " + neighborhood + " - " + city + " - " + state + " - CEP: " + zip, nil
}

func (s *Store) GetCustomerPhoneForBilling(ctx context.Context, customerID string) (string, error) {
	var phone string
	err := s.q(ctx).QueryRow(ctx, `
		SELECT COALESCE(pl."Number", '')
		FROM "PhoneLineCustomerLinks" plcl
		JOIN "PhoneLines" pl ON pl."Id" = plcl."PhoneLineId"
		WHERE plcl."CustomerId" = $1 AND plcl."EndDate" IS NULL
		ORDER BY plcl."StartDate" DESC
		LIMIT 1`, customerID).Scan(&phone)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return phone, nil
}
