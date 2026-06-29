package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/models"
)

func (s *Store) ListCustomers(ctx context.Context, orgID string, providerID *string, salespersonUserID *string, page httputil.PageSearch) ([]models.ListCustomerResponse, int64, error) {
	q := s.q(ctx)
	base := `
		FROM "Customers" c
		LEFT JOIN "CustomerDocuments" cd ON cd."CustomerId" = c."Id" AND cd."DocumentType" IN ('cpf', 'cnpj')
		LEFT JOIN "CustomerDocuments" sr ON sr."CustomerId" = c."Id" AND sr."DocumentType" = 'state_registration'
		WHERE c."OrganizationId" = $1`
	args := []any{orgID}
	if providerID != nil && *providerID != "" {
		base += ` AND EXISTS (
			SELECT 1 FROM "CustomerProviderLinks" cpl
			WHERE cpl."CustomerId" = c."Id" AND cpl."ProviderId" = $` + itoa(len(args)+1) + ` AND cpl."EndDate" IS NULL)`
		args = append(args, *providerID)
	}
	if salespersonUserID != nil && *salespersonUserID != "" {
		base += ` AND c."ResponsibleSalespersonUserId" = $` + itoa(len(args)+1)
		args = append(args, *salespersonUserID)
	}

	var total int64
	countQuery := `SELECT COUNT(DISTINCT c."Id") ` + base
	if err := q.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	selectQuery := `
		SELECT DISTINCT c."Id", c."Active", c."Type"::text, c."Name",
			COALESCE(cd."Number", ''), sr."Number", c."LegalName",
			c."BirthOrOpeningDate", c."ResponsibleSalespersonUserId", c."BillingEmail",
			COALESCE(c."IsReseller", false)
		` + base + `
		ORDER BY c."Name"
		OFFSET $` + itoa(len(args)+1) + ` LIMIT $` + itoa(len(args)+2)
	args = append(args, page.Offset(), page.Limit())

	rows, err := q.Query(ctx, selectQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []models.ListCustomerResponse
	for rows.Next() {
		var item models.ListCustomerResponse
		if err := rows.Scan(&item.ID, &item.Active, &item.Type, &item.Name, &item.CpfCnpj,
			&item.StateRegistration, &item.LegalName, &item.BirthOrOpeningDate, &item.ResponsibleSalespersonUserID, &item.BillingEmail, &item.IsReseller); err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (s *Store) GetCustomer(ctx context.Context, id string) (*models.ListCustomerResponse, error) {
	return s.GetCustomerInOrg(ctx, "", id, nil)
}

func (s *Store) GetCustomerInOrg(ctx context.Context, orgID, id string, salespersonUserID *string) (*models.ListCustomerResponse, error) {
	q := s.q(ctx)
	query := `
		SELECT c."Id", c."Active", c."Type"::text, c."Name",
			COALESCE((SELECT cd."Number" FROM "CustomerDocuments" cd
				WHERE cd."CustomerId" = c."Id" AND cd."DocumentType" IN ('cpf','cnpj') LIMIT 1), ''),
			(SELECT sr."Number" FROM "CustomerDocuments" sr
				WHERE sr."CustomerId" = c."Id" AND sr."DocumentType" = 'state_registration' LIMIT 1),
			c."LegalName", c."BirthOrOpeningDate", c."ResponsibleSalespersonUserId", c."BillingEmail",
			COALESCE(c."IsReseller", false)
		FROM "Customers" c
		WHERE c."Id" = $1`
	args := []any{id}
	if orgID != "" {
		query += ` AND c."OrganizationId" = $2`
		args = append(args, orgID)
	}
	if salespersonUserID != nil && *salespersonUserID != "" {
		query += ` AND c."ResponsibleSalespersonUserId" = $` + itoa(len(args)+1)
		args = append(args, *salespersonUserID)
	}
	var item models.ListCustomerResponse
	err := q.QueryRow(ctx, query, args...).
		Scan(&item.ID, &item.Active, &item.Type, &item.Name, &item.CpfCnpj,
			&item.StateRegistration, &item.LegalName, &item.BirthOrOpeningDate, &item.ResponsibleSalespersonUserID, &item.BillingEmail, &item.IsReseller)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *Store) CustomerOwnedBySalesperson(ctx context.Context, orgID, customerID, salespersonUserID string) (bool, error) {
	var exists bool
	err := s.q(ctx).QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM "Customers"
			WHERE "OrganizationId" = $1 AND "Id" = $2 AND "ResponsibleSalespersonUserId" = $3)`,
		orgID, customerID, salespersonUserID).Scan(&exists)
	return exists, err
}

func (s *Store) CustomerDocumentExists(ctx context.Context, orgID, document string) (bool, error) {
	var exists bool
	err := s.q(ctx).QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM "CustomerDocuments" cd
			JOIN "Customers" c ON c."Id" = cd."CustomerId"
			WHERE c."OrganizationId" = $1 AND cd."Number" = $2)`, orgID, document).Scan(&exists)
	return exists, err
}

func (s *Store) CreateCustomer(ctx context.Context, orgID, id, providerID, customerType, name, document, docType string,
	legalName, stateReg, salesperson *string, birth *time.Time) error {
	q := s.q(ctx)
	if _, err := q.Exec(ctx, `
		INSERT INTO "Customers" ("Id", "OrganizationId", "Name", "LegalName", "Type", "Active",
			"BirthOrOpeningDate", "ResponsibleSalespersonUserId")
		VALUES ($1, $2, $3, $4, $5::customer_type, true, $6, $7)`,
		id, orgID, name, legalName, customerType, birth, salesperson); err != nil {
		return err
	}
	docID := newUUID()
	if _, err := q.Exec(ctx, `
		INSERT INTO "CustomerDocuments" ("Id", "CustomerId", "DocumentType", "Number")
		VALUES ($1, $2, $3::customer_document_type, $4)`, docID, id, docType, document); err != nil {
		return err
	}
	if stateReg != nil && *stateReg != "" {
		if _, err := q.Exec(ctx, `
			INSERT INTO "CustomerDocuments" ("Id", "CustomerId", "DocumentType", "Number")
			VALUES ($1, $2, 'state_registration'::customer_document_type, $3)`, newUUID(), id, *stateReg); err != nil {
			return err
		}
	}
	linkID := newUUID()
	now := time.Now().UTC()
	if _, err := q.Exec(ctx, `
		INSERT INTO "CustomerProviderLinks" ("Id", "CustomerId", "ProviderId", "StartDate")
		VALUES ($1, $2, $3, $4)`, linkID, id, providerID, now); err != nil {
		return err
	}
	return nil
}

func (s *Store) CreateCustomerAddress(ctx context.Context, customerID string, addr models.CreateCustomerAddressInput) error {
	country := addr.Country
	if country == "" {
		country = "Brasil"
	}
	_, err := s.q(ctx).Exec(ctx, `
		INSERT INTO "CustomerAddresses" ("Id", "CustomerId", "Street", "Number", "Neighborhood",
			"City", "State", "ZipCode", "Complement", "Country")
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		newUUID(), customerID, addr.Street, addr.Number, addr.Neighborhood,
		addr.City, addr.State, addr.ZipCode, addr.Complement, country)
	return err
}

func (s *Store) UpdateCustomer(ctx context.Context, id, name string, legalName, stateReg, salesperson *string, birth *time.Time) error {
	tag, err := s.q(ctx).Exec(ctx, `
		UPDATE "Customers" SET "Name" = $2, "LegalName" = $3,
			"BirthOrOpeningDate" = $4, "ResponsibleSalespersonUserId" = $5
		WHERE "Id" = $1`, id, name, legalName, birth, salesperson)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	if stateReg != nil {
		_, _ = s.q(ctx).Exec(ctx, `
			DELETE FROM "CustomerDocuments"
			WHERE "CustomerId" = $1 AND "DocumentType" = 'state_registration'::customer_document_type`, id)
		if *stateReg != "" {
			_, err = s.q(ctx).Exec(ctx, `
				INSERT INTO "CustomerDocuments" ("Id", "CustomerId", "DocumentType", "Number")
				VALUES ($1, $2, 'state_registration'::customer_document_type, $3)`, newUUID(), id, *stateReg)
		}
	}
	return err
}

func (s *Store) InactivateCustomer(ctx context.Context, id string) error {
	tag, err := s.q(ctx).Exec(ctx, `UPDATE "Customers" SET "Active" = false WHERE "Id" = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (s *Store) ListCustomerProviderLinks(ctx context.Context, orgID, customerID string) ([]models.CustomerProviderLinkResponse, error) {
	rows, err := s.q(ctx).Query(ctx, `
		SELECT cpl."CustomerId", cpl."ProviderId", p."Name", cpl."StartDate", cpl."EndDate"
		FROM "CustomerProviderLinks" cpl
		JOIN "Customers" c ON c."Id" = cpl."CustomerId"
		JOIN "Providers" p ON p."Id" = cpl."ProviderId"
		WHERE c."OrganizationId" = $1 AND cpl."CustomerId" = $2
		ORDER BY cpl."StartDate" DESC`, orgID, customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []models.CustomerProviderLinkResponse
	for rows.Next() {
		var item models.CustomerProviderLinkResponse
		var endDate *time.Time
		if err := rows.Scan(&item.CustomerID, &item.ProviderID, &item.ProviderName, &item.StartDate, &endDate); err != nil {
			return nil, err
		}
		item.EndDate = endDate
		item.IsActive = endDate == nil
		items = append(items, item)
	}
	if items == nil {
		items = []models.CustomerProviderLinkResponse{}
	}
	return items, rows.Err()
}

func (s *Store) ListCustomerPhoneLines(ctx context.Context, orgID, customerID string, page httputil.PageSearch) ([]models.CustomerPhoneLineLinkResponse, int64, error) {
	q := s.q(ctx)
	var total int64
	if err := q.QueryRow(ctx, `
		SELECT COUNT(*) FROM "PhoneLineCustomerLinks" l
		JOIN "Customers" c ON c."Id" = l."CustomerId"
		WHERE c."OrganizationId" = $1 AND l."CustomerId" = $2`, orgID, customerID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := q.Query(ctx, `
		SELECT l."CustomerId", l."PhoneLineId", pl."Number", pl."Status"::text,
			pl."LineClassification"::text, l."StartDate", l."EndDate", l."MonthlyAmount",
			pl."BaseCost", pl."CostWithConsumption"
		FROM "PhoneLineCustomerLinks" l
		JOIN "Customers" c ON c."Id" = l."CustomerId"
		JOIN "PhoneLines" pl ON pl."Id" = l."PhoneLineId"
		WHERE c."OrganizationId" = $1 AND l."CustomerId" = $2
		ORDER BY l."StartDate" DESC
		OFFSET $3 LIMIT $4`, orgID, customerID, page.Offset(), page.Limit())
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []models.CustomerPhoneLineLinkResponse
	for rows.Next() {
		var item models.CustomerPhoneLineLinkResponse
		var endDate *time.Time
		if err := rows.Scan(&item.CustomerID, &item.PhoneLineID, &item.PhoneLineNumber,
			&item.PhoneLineStatus, &item.LineClassification, &item.StartDate, &endDate,
			&item.MonthlyAmount, &item.BaseCost, &item.CostWithConsumption); err != nil {
			return nil, 0, err
		}
		item.EndDate = endDate
		item.IsActive = endDate == nil
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (s *Store) CustomerExistsInOrg(ctx context.Context, orgID, customerID string) (bool, error) {
	var exists bool
	err := s.q(ctx).QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM "Customers" WHERE "OrganizationId" = $1 AND "Id" = $2)`,
		orgID, customerID).Scan(&exists)
	return exists, err
}

func (s *Store) CustomerHasActiveProvider(ctx context.Context, customerID, providerID string) (bool, error) {
	var exists bool
	err := s.q(ctx).QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM "CustomerProviderLinks"
			WHERE "CustomerId" = $1 AND "ProviderId" = $2 AND "EndDate" IS NULL)`,
		customerID, providerID).Scan(&exists)
	return exists, err
}

func (s *Store) ListCustomersByDocument(ctx context.Context, orgID, document string) ([]string, error) {
	rows, err := s.q(ctx).Query(ctx, `
		SELECT c."Id" FROM "Customers" c
		JOIN "CustomerDocuments" cd ON cd."CustomerId" = c."Id"
		WHERE c."OrganizationId" = $1 AND cd."Number" = $2 AND c."Active" = true`, orgID, document)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (s *Store) GetCustomerCNPJ(ctx context.Context, customerID string) (string, error) {
	var doc string
	err := s.q(ctx).QueryRow(ctx, `
		SELECT "Number" FROM "CustomerDocuments"
		WHERE "CustomerId" = $1 AND "DocumentType" = 'cnpj'::customer_document_type LIMIT 1`, customerID).Scan(&doc)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}
	return doc, err
}

func (s *Store) GetCustomerPrimaryDocument(ctx context.Context, customerID string) (string, error) {
	var doc string
	err := s.q(ctx).QueryRow(ctx, `
		SELECT "Number" FROM "CustomerDocuments"
		WHERE "CustomerId" = $1 AND "DocumentType" IN ('cpf','cnpj') LIMIT 1`, customerID).Scan(&doc)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}
	return doc, err
}

func (s *Store) ReactivateCustomer(ctx context.Context, customerID string) error {
	_, err := s.q(ctx).Exec(ctx, `UPDATE "Customers" SET "Active" = true WHERE "Id" = $1`, customerID)
	return err
}

func (s *Store) AddCustomerProviderLink(ctx context.Context, customerID, providerID string, start time.Time) error {
	_, err := s.q(ctx).Exec(ctx, `
		INSERT INTO "CustomerProviderLinks" ("Id", "CustomerId", "ProviderId", "StartDate")
		VALUES ($1, $2, $3, $4)
		ON CONFLICT DO NOTHING`, newUUID(), customerID, providerID, start)
	return err
}

func (s *Store) CountDashboardCustomers(ctx context.Context, orgID string) (int32, error) {
	var n int32
	err := s.q(ctx).QueryRow(ctx, `SELECT COUNT(*)::int FROM "Customers" WHERE "OrganizationId" = $1`, orgID).Scan(&n)
	return n, err
}
