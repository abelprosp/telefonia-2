package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/models"
)

func (s *Store) ListPhoneLines(ctx context.Context, orgID string, status *string, page httputil.PageSearch) ([]models.ListPhoneLineResponse, int64, error) {
	base := `
		FROM "PhoneLines" pl
		JOIN "ProviderAccounts" pa ON pa."Id" = pl."ProviderAccountId"
		JOIN "ContractingCompanies" cc ON cc."Id" = pa."ContractingCompanyId"
		JOIN "Providers" p ON p."Id" = cc."ProviderId"
		JOIN "ProviderPlans" pp ON pp."Id" = pl."ProviderPlanId"
		LEFT JOIN "CostCenters" cst ON cst."Id" = pl."CostCenterId"
		LEFT JOIN "ProviderInvoices" inv ON inv."Id" = pl."LastInvoiceId"
		LEFT JOIN "PhoneLines" tit ON tit."Id" = pl."TitularLineId"
		WHERE p."OrganizationId" = $1`
	args := []any{orgID}
	if status != nil && *status != "" {
		base += ` AND pl."Status" = $2::phone_line_status`
		args = append(args, *status)
	}

	var total int64
	if err := s.q(ctx).QueryRow(ctx, `SELECT COUNT(*) `+base, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offsetParam := len(args) + 1
	limitParam := len(args) + 2
	selectQ := `
		SELECT pl."Id", pl."ProviderPlanId", pp."Name", pl."ProviderAccountId", pa."AccountNumber",
			pl."CostCenterId", cst."Name", pl."LastInvoiceId", inv."Number",
			pl."TitularLineId", tit."Number", pl."Number", pl."LineClassification"::text,
			pl."Status"::text, pl."TransitionSubStatus"::text, pl."TransitionStartedAt",
			pl."ActivationDate", pl."CancellationDate", pl."BaseCost", pl."CostWithConsumption"
		` + base + `
		ORDER BY pl."Number"
		OFFSET $` + itoa(offsetParam) + ` LIMIT $` + itoa(limitParam)
	args = append(args, page.Offset(), page.Limit())

	rows, err := s.q(ctx).Query(ctx, selectQ, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	return scanPhoneLineList(rows, total)
}

func scanPhoneLineList(rows pgx.Rows, total int64) ([]models.ListPhoneLineResponse, int64, error) {
	var items []models.ListPhoneLineResponse
	for rows.Next() {
		var item models.ListPhoneLineResponse
		if err := rows.Scan(
			&item.ID, &item.ProviderPlanID, &item.ProviderPlanName, &item.ProviderAccountID,
			&item.ProviderAccountNumber, &item.CostCenterID, &item.CostCenterName,
			&item.LastInvoiceID, &item.LastInvoiceNumber, &item.TitularLineID, &item.TitularLineNumber,
			&item.Number, &item.LineClassification, &item.Status, &item.TransitionSubStatus,
			&item.TransitionStartedAt, &item.ActivationDate, &item.CancellationDate,
			&item.BaseCost, &item.CostWithConsumption); err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (s *Store) GetPhoneLine(ctx context.Context, orgID, id string) (*models.GetPhoneLineResponse, error) {
	q := s.q(ctx)
	var item models.GetPhoneLineResponse
	err := q.QueryRow(ctx, `
		SELECT pl."Id", pl."ProviderPlanId", pp."Name", pl."ProviderAccountId", pa."AccountNumber",
			pl."CostCenterId", cst."Name", pl."LastInvoiceId", inv."Number",
			pl."TitularLineId", tit."Number", pl."Number", pl."LineClassification"::text,
			pl."Status"::text, pl."TransitionSubStatus"::text, pl."TransitionStartedAt",
			pl."ActivationDate", pl."CancellationDate", pl."BaseCost", pl."CostWithConsumption"
		FROM "PhoneLines" pl
		JOIN "ProviderAccounts" pa ON pa."Id" = pl."ProviderAccountId"
		JOIN "ContractingCompanies" cc ON cc."Id" = pa."ContractingCompanyId"
		JOIN "Providers" p ON p."Id" = cc."ProviderId"
		JOIN "ProviderPlans" pp ON pp."Id" = pl."ProviderPlanId"
		LEFT JOIN "CostCenters" cst ON cst."Id" = pl."CostCenterId"
		LEFT JOIN "ProviderInvoices" inv ON inv."Id" = pl."LastInvoiceId"
		LEFT JOIN "PhoneLines" tit ON tit."Id" = pl."TitularLineId"
		WHERE p."OrganizationId" = $1 AND pl."Id" = $2`, orgID, id).
		Scan(
			&item.ID, &item.ProviderPlanID, &item.ProviderPlanName, &item.ProviderAccountID,
			&item.ProviderAccountNumber, &item.CostCenterID, &item.CostCenterName,
			&item.LastInvoiceID, &item.LastInvoiceNumber, &item.TitularLineID, &item.TitularLineNumber,
			&item.Number, &item.LineClassification, &item.Status, &item.TransitionSubStatus,
			&item.TransitionStartedAt, &item.ActivationDate, &item.CancellationDate,
			&item.BaseCost, &item.CostWithConsumption)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	svcRows, err := q.Query(ctx, `
		SELECT "Id", "PhoneLineId", "ProviderPlanServiceId", "Name", "Code", "Recurring", "Price", "Active"
		FROM "PhoneLineServices" WHERE "PhoneLineId" = $1`, id)
	if err != nil {
		return nil, err
	}
	defer svcRows.Close()
	for svcRows.Next() {
		var svc models.GetPhoneLineServiceResponse
		if err := svcRows.Scan(&svc.ID, &svc.PhoneLineID, &svc.ProviderPlanServiceID,
			&svc.Name, &svc.Code, &svc.Recurring, &svc.Price, &svc.Active); err != nil {
			return nil, err
		}
		item.Services = append(item.Services, svc)
	}
	if item.Services == nil {
		item.Services = []models.GetPhoneLineServiceResponse{}
	}
	item.Children = []models.GetChildPhoneLineResponse{}
	return &item, nil
}

func (s *Store) ListPhoneLineCustomerLinks(ctx context.Context, orgID, phoneLineID string) ([]models.PhoneLineCustomerLinkResponse, error) {
	rows, err := s.q(ctx).Query(ctx, `
		SELECT l."PhoneLineId", l."CustomerId", c."Name",
			(SELECT cd."Number" FROM "CustomerDocuments" cd
			 WHERE cd."CustomerId" = c."Id" AND cd."DocumentType" IN ('cpf','cnpj') LIMIT 1),
			l."StartDate", l."EndDate"
		FROM "PhoneLineCustomerLinks" l
		JOIN "Customers" c ON c."Id" = l."CustomerId"
		JOIN "PhoneLines" pl ON pl."Id" = l."PhoneLineId"
		JOIN "ProviderAccounts" pa ON pa."Id" = pl."ProviderAccountId"
		JOIN "ContractingCompanies" cc ON cc."Id" = pa."ContractingCompanyId"
		JOIN "Providers" p ON p."Id" = cc."ProviderId"
		WHERE p."OrganizationId" = $1 AND l."PhoneLineId" = $2
		ORDER BY l."StartDate" DESC`, orgID, phoneLineID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []models.PhoneLineCustomerLinkResponse
	for rows.Next() {
		var item models.PhoneLineCustomerLinkResponse
		var endDate *time.Time
		if err := rows.Scan(&item.PhoneLineID, &item.CustomerID, &item.CustomerName,
			&item.CustomerDocument, &item.StartDate, &endDate); err != nil {
			return nil, err
		}
		item.EndDate = endDate
		item.IsActive = endDate == nil
		items = append(items, item)
	}
	if items == nil {
		items = []models.PhoneLineCustomerLinkResponse{}
	}
	return items, rows.Err()
}

func (s *Store) GetActivePhoneLineCustomerLink(ctx context.Context, phoneLineID string) (linkID, customerID string, err error) {
	err = s.q(ctx).QueryRow(ctx, `
		SELECT "Id", "CustomerId" FROM "PhoneLineCustomerLinks"
		WHERE "PhoneLineId" = $1 AND "EndDate" IS NULL`, phoneLineID).Scan(&linkID, &customerID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", "", nil
	}
	return linkID, customerID, err
}

func (s *Store) AssignPhoneLineCustomer(ctx context.Context, phoneLineID, customerID string, start time.Time) error {
	if _, err := s.q(ctx).Exec(ctx, `
		UPDATE "PhoneLineCustomerLinks" SET "EndDate" = $2
		WHERE "PhoneLineId" = $1 AND "EndDate" IS NULL`, phoneLineID, start); err != nil {
		return err
	}
	_, err := s.q(ctx).Exec(ctx, `
		INSERT INTO "PhoneLineCustomerLinks" ("Id", "PhoneLineId", "CustomerId", "StartDate")
		VALUES ($1, $2, $3, $4)`, newUUID(), phoneLineID, customerID, start)
	return err
}

func (s *Store) UnassignPhoneLineCustomer(ctx context.Context, phoneLineID string, end time.Time) error {
	tag, err := s.q(ctx).Exec(ctx, `
		UPDATE "PhoneLineCustomerLinks" SET "EndDate" = $2
		WHERE "PhoneLineId" = $1 AND "EndDate" IS NULL`, phoneLineID, end)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (s *Store) CustomerHasOtherActivePhoneLines(ctx context.Context, orgID, customerID, excludeLineID string) (bool, error) {
	var exists bool
	err := s.q(ctx).QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM "PhoneLineCustomerLinks" l
			JOIN "Customers" c ON c."Id" = l."CustomerId"
			WHERE c."OrganizationId" = $1 AND l."CustomerId" = $2
				AND l."EndDate" IS NULL AND l."PhoneLineId" != $3)`,
		orgID, customerID, excludeLineID).Scan(&exists)
	return exists, err
}

func (s *Store) GetPhoneLineProviderID(ctx context.Context, orgID, phoneLineID string) (string, error) {
	var providerID string
	err := s.q(ctx).QueryRow(ctx, `
		SELECT p."Id" FROM "PhoneLines" pl
		JOIN "ProviderAccounts" pa ON pa."Id" = pl."ProviderAccountId"
		JOIN "ContractingCompanies" cc ON cc."Id" = pa."ContractingCompanyId"
		JOIN "Providers" p ON p."Id" = cc."ProviderId"
		WHERE p."OrganizationId" = $1 AND pl."Id" = $2`, orgID, phoneLineID).Scan(&providerID)
	return providerID, err
}

type PhoneLineRow struct {
	ID                string
	Number            string
	ProviderAccountID string
	ProviderPlanID    string
	Status            string
}

func (s *Store) GetPhoneLineByNumber(ctx context.Context, number string) (*PhoneLineRow, error) {
	var pl PhoneLineRow
	err := s.q(ctx).QueryRow(ctx, `
		SELECT "Id", "Number", "ProviderAccountId", "ProviderPlanId", "Status"::text
		FROM "PhoneLines" WHERE "Number" = $1`, number).
		Scan(&pl.ID, &pl.Number, &pl.ProviderAccountID, &pl.ProviderPlanID, &pl.Status)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &pl, err
}

func (s *Store) CreatePhoneLine(ctx context.Context, id, planID, accountID, number string) error {
	_, err := s.q(ctx).Exec(ctx, `
		INSERT INTO "PhoneLines" ("Id", "Number", "ProviderAccountId", "ProviderPlanId",
			"LineClassification", "Status")
		VALUES ($1, $2, $3, $4, 'normal'::line_classification, 'in_stock'::phone_line_status)`,
		id, number, accountID, planID)
	return err
}

func (s *Store) GetProviderAccountByProviderAndNumber(ctx context.Context, orgID, providerID, accountNumber string) (*ProviderAccountRow, error) {
	var a ProviderAccountRow
	err := s.q(ctx).QueryRow(ctx, `
		SELECT pa."Id", pa."ContractingCompanyId", pa."AccountNumber"
		FROM "ProviderAccounts" pa
		JOIN "ContractingCompanies" cc ON cc."Id" = pa."ContractingCompanyId"
		JOIN "Providers" p ON p."Id" = cc."ProviderId"
		WHERE p."OrganizationId" = $1 AND p."Id" = $2 AND pa."AccountNumber" = $3`,
		orgID, providerID, accountNumber).Scan(&a.ID, &a.ContractingCompanyID, &a.AccountNumber)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (s *Store) ProviderPlanExistsForProvider(ctx context.Context, orgID, providerID, planID string) (bool, error) {
	var exists bool
	err := s.q(ctx).QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM "ProviderPlans" pp
			JOIN "Providers" p ON p."Id" = pp."ProviderId"
			WHERE p."OrganizationId" = $1 AND p."Id" = $2 AND pp."Id" = $3)`,
		orgID, providerID, planID).Scan(&exists)
	return exists, err
}

func (s *Store) UpdatePhoneLineStatus(ctx context.Context, id, status string) error {
	_, err := s.q(ctx).Exec(ctx, `
		UPDATE "PhoneLines" SET "Status" = $2::phone_line_status WHERE "Id" = $1`, id, status)
	return err
}

func (s *Store) UpdatePhoneLineTransition(ctx context.Context, id, status, subStatus string, startedAt time.Time) error {
	_, err := s.q(ctx).Exec(ctx, `
		UPDATE "PhoneLines"
		SET "Status" = $2::phone_line_status,
			"TransitionSubStatus" = $3::transition_sub_status,
			"TransitionStartedAt" = $4
		WHERE "Id" = $1`, id, status, subStatus, startedAt)
	return err
}

func (s *Store) UpdatePhoneLineCosts(ctx context.Context, id string, base, withConsumption float64, lastInvoiceID string) error {
	_, err := s.q(ctx).Exec(ctx, `
		UPDATE "PhoneLines" SET "BaseCost" = $2, "CostWithConsumption" = $3, "LastInvoiceId" = $4
		WHERE "Id" = $1`, id, base, withConsumption, lastInvoiceID)
	return err
}

func (s *Store) ListPhoneLinesByAccount(ctx context.Context, accountID string) ([]PhoneLineRow, error) {
	rows, err := s.q(ctx).Query(ctx, `
		SELECT "Id", "Number", "ProviderAccountId", "ProviderPlanId", "Status"::text
		FROM "PhoneLines" WHERE "ProviderAccountId" = $1`, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []PhoneLineRow
	for rows.Next() {
		var pl PhoneLineRow
		if err := rows.Scan(&pl.ID, &pl.Number, &pl.ProviderAccountID, &pl.ProviderPlanID, &pl.Status); err != nil {
			return nil, err
		}
		items = append(items, pl)
	}
	return items, rows.Err()
}

func (s *Store) CountDashboardPhoneLines(ctx context.Context, orgID string) (int32, error) {
	var n int32
	err := s.q(ctx).QueryRow(ctx, `
		SELECT COUNT(*)::int FROM "PhoneLines" pl
		JOIN "ProviderAccounts" pa ON pa."Id" = pl."ProviderAccountId"
		JOIN "ContractingCompanies" cc ON cc."Id" = pa."ContractingCompanyId"
		JOIN "Providers" p ON p."Id" = cc."ProviderId"
		WHERE p."OrganizationId" = $1`, orgID).Scan(&n)
	return n, err
}
