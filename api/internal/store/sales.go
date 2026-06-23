package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/models"
)

type CustomerContractData struct {
	Name      string
	LegalName *string
	Document  string
	Type      string
	Street    string
	Number    string
	Neighborhood string
	Complement   *string
	City      string
	State     string
	ZipCode   string
	Country   string
}

func (s *Store) ListContractTemplates(ctx context.Context, orgID string, activeOnly bool, page httputil.PageSearch) ([]models.ListContractTemplateResponse, int64, error) {
	base := ` FROM "ContractTemplates" WHERE "OrganizationId" = $1`
	args := []any{orgID}
	if activeOnly {
		base += ` AND "Active" = true`
	}

	var total int64
	if err := s.q(ctx).QueryRow(ctx, `SELECT COUNT(*) `+base, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	selectQ := `
		SELECT "Id", "Name", "Code", "Active", "CreatedAt", "UpdatedAt"
		` + base + `
		ORDER BY "Name" ASC
		OFFSET $` + itoa(len(args)+1) + ` LIMIT $` + itoa(len(args)+2)
	args = append(args, page.Offset(), page.Limit())

	rows, err := s.q(ctx).Query(ctx, selectQ, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []models.ListContractTemplateResponse
	for rows.Next() {
		var item models.ListContractTemplateResponse
		if err := rows.Scan(&item.ID, &item.Name, &item.Code, &item.Active, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (s *Store) GetContractTemplate(ctx context.Context, orgID, id string) (*models.GetContractTemplateResponse, error) {
	var item models.GetContractTemplateResponse
	err := s.q(ctx).QueryRow(ctx, `
		SELECT "Id", "Name", "Code", "Active", "CreatedAt", "UpdatedAt", "BodyTemplate"
		FROM "ContractTemplates"
		WHERE "OrganizationId" = $1 AND "Id" = $2`, orgID, id).Scan(
		&item.ID, &item.Name, &item.Code, &item.Active, &item.CreatedAt, &item.UpdatedAt, &item.BodyTemplate,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *Store) CreateContractTemplate(ctx context.Context, id, orgID, name, code, body string, active bool, now time.Time) error {
	_, err := s.q(ctx).Exec(ctx, `
		INSERT INTO "ContractTemplates" ("Id", "OrganizationId", "Name", "Code", "BodyTemplate", "Active", "CreatedAt", "UpdatedAt")
		VALUES ($1, $2, $3, $4, $5, $6, $7, $7)`,
		id, orgID, name, code, body, active, now)
	return err
}

func (s *Store) UpdateContractTemplate(ctx context.Context, orgID, id string, name, code, body *string, active *bool, now time.Time) error {
	q := s.q(ctx)
	if name != nil {
		if _, err := q.Exec(ctx, `UPDATE "ContractTemplates" SET "Name" = $1, "UpdatedAt" = $2 WHERE "OrganizationId" = $3 AND "Id" = $4`,
			*name, now, orgID, id); err != nil {
			return err
		}
	}
	if code != nil {
		if _, err := q.Exec(ctx, `UPDATE "ContractTemplates" SET "Code" = $1, "UpdatedAt" = $2 WHERE "OrganizationId" = $3 AND "Id" = $4`,
			*code, now, orgID, id); err != nil {
			return err
		}
	}
	if body != nil {
		if _, err := q.Exec(ctx, `UPDATE "ContractTemplates" SET "BodyTemplate" = $1, "UpdatedAt" = $2 WHERE "OrganizationId" = $3 AND "Id" = $4`,
			*body, now, orgID, id); err != nil {
			return err
		}
	}
	if active != nil {
		if _, err := q.Exec(ctx, `UPDATE "ContractTemplates" SET "Active" = $1, "UpdatedAt" = $2 WHERE "OrganizationId" = $3 AND "Id" = $4`,
			*active, now, orgID, id); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) ContractTemplateCodeExists(ctx context.Context, orgID, code string, excludeID *string) (bool, error) {
	q := `SELECT EXISTS(SELECT 1 FROM "ContractTemplates" WHERE "OrganizationId" = $1 AND "Code" = $2`
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

func (s *Store) NextSaleNumber(ctx context.Context, orgID string) (string, error) {
	var count int64
	err := s.q(ctx).QueryRow(ctx, `
		SELECT COUNT(*) FROM "Sales" WHERE "OrganizationId" = $1 AND "CreatedAt" >= date_trunc('day', NOW() AT TIME ZONE 'UTC')`,
		orgID).Scan(&count)
	if err != nil {
		return "", err
	}
	now := time.Now().UTC()
	return fmt.Sprintf("VND-%s-%04d", now.Format("20060102"), count+1), nil
}

func (s *Store) ListSales(ctx context.Context, orgID string, status *string, salespersonUserID *string, page httputil.PageSearch) ([]models.ListSaleResponse, int64, error) {
	base := `
		FROM "Sales" s
		JOIN "Customers" c ON c."Id" = s."CustomerId"
		LEFT JOIN "ContractTemplates" ct ON ct."Id" = s."ContractTemplateId"
		WHERE s."OrganizationId" = $1`
	args := []any{orgID}
	if status != nil && *status != "" {
		base += ` AND s."Status" = $` + itoa(len(args)+1) + `::sale_status`
		args = append(args, *status)
	}
	if salespersonUserID != nil && *salespersonUserID != "" {
		base += ` AND s."SalespersonUserId" = $` + itoa(len(args)+1)
		args = append(args, *salespersonUserID)
	}

	var total int64
	if err := s.q(ctx).QueryRow(ctx, `SELECT COUNT(*) `+base, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	selectQ := `
		SELECT s."Id", s."SaleNumber", s."CustomerId", c."Name", s."SalespersonUserId",
			s."ContractTemplateId", ct."Name", s."Status"::text, s."SoldAt", s."TotalAmount",
			s."Notes", s."CreatedAt", s."UpdatedAt"
		` + base + `
		ORDER BY s."CreatedAt" DESC
		OFFSET $` + itoa(len(args)+1) + ` LIMIT $` + itoa(len(args)+2)
	args = append(args, page.Offset(), page.Limit())

	rows, err := s.q(ctx).Query(ctx, selectQ, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []models.ListSaleResponse
	for rows.Next() {
		var item models.ListSaleResponse
		if err := rows.Scan(
			&item.ID, &item.SaleNumber, &item.CustomerID, &item.CustomerName, &item.SalespersonUserID,
			&item.ContractTemplateID, &item.ContractTemplateName, &item.Status, &item.SoldAt, &item.TotalAmount,
			&item.Notes, &item.CreatedAt, &item.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (s *Store) GetSaleInOrg(ctx context.Context, orgID, id string, salespersonUserID *string) (*models.GetSaleResponse, error) {
	base := `
		FROM "Sales" s
		JOIN "Customers" c ON c."Id" = s."CustomerId"
		LEFT JOIN "ContractTemplates" ct ON ct."Id" = s."ContractTemplateId"
		WHERE s."OrganizationId" = $1 AND s."Id" = $2`
	args := []any{orgID, id}
	if salespersonUserID != nil && *salespersonUserID != "" {
		base += ` AND s."SalespersonUserId" = $3`
		args = append(args, *salespersonUserID)
	}

	var sale models.GetSaleResponse
	err := s.q(ctx).QueryRow(ctx, `
		SELECT s."Id", s."SaleNumber", s."CustomerId", c."Name", s."SalespersonUserId",
			s."ContractTemplateId", ct."Name", s."Status"::text, s."SoldAt", s."TotalAmount",
			s."Notes", s."CreatedAt", s."UpdatedAt"
		`+base, args...).Scan(
		&sale.ID, &sale.SaleNumber, &sale.CustomerID, &sale.CustomerName, &sale.SalespersonUserID,
		&sale.ContractTemplateID, &sale.ContractTemplateName, &sale.Status, &sale.SoldAt, &sale.TotalAmount,
		&sale.Notes, &sale.CreatedAt, &sale.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	items, err := s.listSaleLineItems(ctx, id)
	if err != nil {
		return nil, err
	}
	sale.Items = items

	contract, err := s.getGeneratedContractBySale(ctx, orgID, id)
	if err != nil {
		return nil, err
	}
	sale.Contract = contract
	return &sale, nil
}

func (s *Store) listSaleLineItems(ctx context.Context, saleID string) ([]models.SaleLineItemResponse, error) {
	rows, err := s.q(ctx).Query(ctx, `
		SELECT "Id", "LineItemType"::text, "Description", "Quantity", "UnitPrice", "TotalPrice",
			"PhoneLineId", "DeviceSku", "SortOrder"
		FROM "SaleLineItems"
		WHERE "SaleId" = $1
		ORDER BY "SortOrder" ASC, "Description" ASC`, saleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.SaleLineItemResponse
	for rows.Next() {
		var item models.SaleLineItemResponse
		if err := rows.Scan(
			&item.ID, &item.LineItemType, &item.Description, &item.Quantity, &item.UnitPrice, &item.TotalPrice,
			&item.PhoneLineID, &item.DeviceSku, &item.SortOrder,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) getGeneratedContractBySale(ctx context.Context, orgID, saleID string) (*models.GeneratedContractResponse, error) {
	var c models.GeneratedContractResponse
	err := s.q(ctx).QueryRow(ctx, `
		SELECT "Id", "ContractTemplateId", "Status"::text, "RenderedHtml", "GeneratedAt"
		FROM "GeneratedContracts"
		WHERE "OrganizationId" = $1 AND "SaleId" = $2
		ORDER BY "CreatedAt" DESC
		LIMIT 1`, orgID, saleID).Scan(&c.ID, &c.ContractTemplateID, &c.Status, &c.RenderedHTML, &c.GeneratedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (s *Store) CreateSale(ctx context.Context, id, orgID, customerID, salespersonUserID, saleNumber string,
	contractTemplateID *string, notes *string, now time.Time) error {
	_, err := s.q(ctx).Exec(ctx, `
		INSERT INTO "Sales" ("Id", "OrganizationId", "CustomerId", "SalespersonUserId", "ContractTemplateId",
			"Status", "SaleNumber", "Notes", "TotalAmount", "CreatedAt", "UpdatedAt")
		VALUES ($1, $2, $3, $4, $5, 'draft'::sale_status, $6, $7, 0, $8, $8)`,
		id, orgID, customerID, salespersonUserID, contractTemplateID, saleNumber, notes, now)
	return err
}

func (s *Store) AddSaleLineItem(ctx context.Context, id, saleID, lineItemType, description string,
	quantity, unitPrice, totalPrice float64, phoneLineID, deviceSku *string, sortOrder int32) error {
	_, err := s.q(ctx).Exec(ctx, `
		INSERT INTO "SaleLineItems" ("Id", "SaleId", "LineItemType", "Description", "Quantity",
			"UnitPrice", "TotalPrice", "PhoneLineId", "DeviceSku", "SortOrder")
		VALUES ($1, $2, $3::sale_line_item_type, $4, $5, $6, $7, $8, $9, $10)`,
		id, saleID, lineItemType, description, quantity, unitPrice, totalPrice, phoneLineID, deviceSku, sortOrder)
	return err
}

func (s *Store) RecalculateSaleTotal(ctx context.Context, saleID string, now time.Time) error {
	_, err := s.q(ctx).Exec(ctx, `
		UPDATE "Sales" SET "TotalAmount" = COALESCE((
			SELECT SUM("TotalPrice") FROM "SaleLineItems" WHERE "SaleId" = $1
		), 0), "UpdatedAt" = $2
		WHERE "Id" = $1`, saleID, now)
	return err
}

func (s *Store) DeleteSaleLineItem(ctx context.Context, saleID, itemID string) error {
	tag, err := s.q(ctx).Exec(ctx, `
		DELETE FROM "SaleLineItems" WHERE "SaleId" = $1 AND "Id" = $2`, saleID, itemID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (s *Store) UpdateSaleDraft(ctx context.Context, orgID, saleID string, customerID, contractTemplateID, notes *string, now time.Time) error {
	q := s.q(ctx)
	if customerID != nil {
		if _, err := q.Exec(ctx, `
			UPDATE "Sales" SET "CustomerId" = $1, "UpdatedAt" = $2
			WHERE "OrganizationId" = $3 AND "Id" = $4 AND "Status" = 'draft'::sale_status`,
			*customerID, now, orgID, saleID); err != nil {
			return err
		}
	}
	if contractTemplateID != nil {
		var templateID *string
		if *contractTemplateID != "" {
			templateID = contractTemplateID
		}
		if _, err := q.Exec(ctx, `
			UPDATE "Sales" SET "ContractTemplateId" = $1, "UpdatedAt" = $2
			WHERE "OrganizationId" = $3 AND "Id" = $4 AND "Status" = 'draft'::sale_status`,
			templateID, now, orgID, saleID); err != nil {
			return err
		}
	}
	if notes != nil {
		if _, err := q.Exec(ctx, `
			UPDATE "Sales" SET "Notes" = $1, "UpdatedAt" = $2
			WHERE "OrganizationId" = $3 AND "Id" = $4 AND "Status" = 'draft'::sale_status`,
			*notes, now, orgID, saleID); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) ConfirmSale(ctx context.Context, orgID, saleID string, soldAt time.Time, now time.Time) error {
	tag, err := s.q(ctx).Exec(ctx, `
		UPDATE "Sales" SET "Status" = 'confirmed'::sale_status, "SoldAt" = $1, "UpdatedAt" = $2
		WHERE "OrganizationId" = $3 AND "Id" = $4 AND "Status" = 'draft'::sale_status`,
		soldAt, now, orgID, saleID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (s *Store) CancelSale(ctx context.Context, orgID, saleID string, now time.Time) error {
	tag, err := s.q(ctx).Exec(ctx, `
		UPDATE "Sales" SET "Status" = 'cancelled'::sale_status, "UpdatedAt" = $1
		WHERE "OrganizationId" = $2 AND "Id" = $3 AND "Status" <> 'cancelled'::sale_status`,
		now, orgID, saleID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (s *Store) SaleItemCount(ctx context.Context, saleID string) (int64, error) {
	var count int64
	err := s.q(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM "SaleLineItems" WHERE "SaleId" = $1`, saleID).Scan(&count)
	return count, err
}

func (s *Store) GetCustomerContractData(ctx context.Context, orgID, customerID string) (*CustomerContractData, error) {
	var data CustomerContractData
	err := s.q(ctx).QueryRow(ctx, `
		SELECT c."Name", c."LegalName", c."Type"::text,
			COALESCE((SELECT cd."Number" FROM "CustomerDocuments" cd
				WHERE cd."CustomerId" = c."Id" AND cd."DocumentType" IN ('cpf','cnpj') LIMIT 1), '')
		FROM "Customers" c
		WHERE c."OrganizationId" = $1 AND c."Id" = $2`, orgID, customerID).Scan(
		&data.Name, &data.LegalName, &data.Type, &data.Document,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	addrErr := s.q(ctx).QueryRow(ctx, `
		SELECT "Street", "Number", "Neighborhood", "Complement", "City", "State", "ZipCode", "Country"
		FROM "CustomerAddresses"
		WHERE "CustomerId" = $1
		ORDER BY "Id" ASC
		LIMIT 1`, customerID).Scan(
		&data.Street, &data.Number, &data.Neighborhood, &data.Complement,
		&data.City, &data.State, &data.ZipCode, &data.Country,
	)
	if addrErr != nil && !errors.Is(addrErr, pgx.ErrNoRows) {
		return nil, addrErr
	}
	return &data, nil
}

func (s *Store) SaveGeneratedContract(ctx context.Context, id, orgID, saleID, templateID, renderedHTML string, now time.Time) error {
	_, err := s.q(ctx).Exec(ctx, `
		INSERT INTO "GeneratedContracts" ("Id", "OrganizationId", "SaleId", "ContractTemplateId",
			"Status", "RenderedHtml", "GeneratedAt", "CreatedAt")
		VALUES ($1, $2, $3, $4, 'generated'::generated_contract_status, $5, $6, $6)`,
		id, orgID, saleID, templateID, renderedHTML, now)
	return err
}

func (s *Store) SaveFailedGeneratedContract(ctx context.Context, id, orgID, saleID, templateID string, now time.Time) error {
	_, err := s.q(ctx).Exec(ctx, `
		INSERT INTO "GeneratedContracts" ("Id", "OrganizationId", "SaleId", "ContractTemplateId",
			"Status", "CreatedAt")
		VALUES ($1, $2, $3, $4, 'failed'::generated_contract_status, $5)`,
		id, orgID, saleID, templateID, now)
	return err
}

func (s *Store) GetSaleStatus(ctx context.Context, orgID, saleID string) (string, error) {
	var status string
	err := s.q(ctx).QueryRow(ctx, `
		SELECT "Status"::text FROM "Sales" WHERE "OrganizationId" = $1 AND "Id" = $2`,
		orgID, saleID).Scan(&status)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", pgx.ErrNoRows
	}
	return status, err
}
