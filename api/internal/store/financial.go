package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/models"
)

func (s *Store) GetFinancialSummary(ctx context.Context, orgID string) (*models.FinancialSummaryResponse, error) {
	var summary models.FinancialSummaryResponse
	err := s.q(ctx).QueryRow(ctx, `
		SELECT
			COALESCE((SELECT SUM("Amount" - "PaidAmount") FROM "AccountsPayable"
			 WHERE "OrganizationId" = $1 AND "Status" IN ('open', 'partially_settled', 'overdue')), 0),
			COALESCE((SELECT SUM("Amount" - "ReceivedAmount") FROM "AccountsReceivable"
			 WHERE "OrganizationId" = $1 AND "Status" IN ('open', 'partially_settled', 'overdue')), 0),
			COALESCE((SELECT SUM("CommissionAmount") FROM "PartnerSalesRecords"
			 WHERE "OrganizationId" = $1 AND "Status" = 'accrued'), 0),
			COALESCE((SELECT COUNT(*)::int FROM "AccountsPayable"
			 WHERE "OrganizationId" = $1 AND "Status" = 'overdue'), 0),
			COALESCE((SELECT COUNT(*)::int FROM "AccountsReceivable"
			 WHERE "OrganizationId" = $1 AND "Status" = 'overdue'), 0),
			COALESCE((SELECT COUNT(*)::int FROM "ProviderInvoices" i
			 JOIN "ProviderAccounts" pa ON pa."Id" = i."ProviderAccountId"
			 JOIN "ContractingCompanies" cc ON cc."Id" = pa."ContractingCompanyId"
			 JOIN "Providers" p ON p."Id" = cc."ProviderId"
			 WHERE p."OrganizationId" = $1), 0),
			COALESCE((SELECT SUM(i."TotalAmount") FROM "ProviderInvoices" i
			 JOIN "ProviderAccounts" pa ON pa."Id" = i."ProviderAccountId"
			 JOIN "ContractingCompanies" cc ON cc."Id" = pa."ContractingCompanyId"
			 JOIN "Providers" p ON p."Id" = cc."ProviderId"
			 WHERE p."OrganizationId" = $1), 0),
			COALESCE((SELECT COUNT(*)::int FROM "ProviderInvoices" i
			 JOIN "ProviderAccounts" pa ON pa."Id" = i."ProviderAccountId"
			 JOIN "ContractingCompanies" cc ON cc."Id" = pa."ContractingCompanyId"
			 JOIN "Providers" p ON p."Id" = cc."ProviderId"
			 WHERE p."OrganizationId" = $1
			   AND NOT EXISTS (
			     SELECT 1 FROM "AccountsPayable" ap
			     WHERE ap."OrganizationId" = $1 AND ap."ProviderInvoiceId" = i."Id"
			   )), 0),
			COALESCE((SELECT COUNT(*)::int FROM "ProcessingMonths" pm
			 JOIN "Providers" p ON p."Id" = pm."ProviderId"
			 WHERE p."OrganizationId" = $1 AND pm."Status" = 'open'), 0),
			COALESCE((SELECT COUNT(*)::int FROM "CustomerBillingDocuments" WHERE "OrganizationId" = $1 AND "Status" = 'draft'), 0),
			COALESCE((SELECT COUNT(*)::int FROM "CustomerBillingDocuments" WHERE "OrganizationId" = $1 AND "Status" = 'ready'), 0),
			COALESCE((SELECT COUNT(*)::int FROM "CustomerBillingDocuments" WHERE "OrganizationId" = $1 AND "Status" = 'sent'), 0)`,
		orgID).Scan(
		&summary.TotalPayableOpen,
		&summary.TotalReceivableOpen,
		&summary.TotalPartnerCommission,
		&summary.PayableOverdueCount,
		&summary.ReceivableOverdueCount,
		&summary.ProviderInvoicesCount,
		&summary.ProviderInvoicesTotalAmount,
		&summary.ProviderInvoicesWithoutPayableCount,
		&summary.OpenProcessingMonthsCount,
		&summary.BillingDocumentsDraftCount,
		&summary.BillingDocumentsReadyCount,
		&summary.BillingDocumentsSentCount,
	)
	if err != nil {
		return nil, err
	}
	return &summary, nil
}

func (s *Store) ListAccountsPayable(ctx context.Context, orgID string, status *string, page httputil.PageSearch) ([]models.ListAccountPayableResponse, int64, error) {
	base := ` FROM "AccountsPayable" WHERE "OrganizationId" = $1`
	args := []any{orgID}
	if status != nil && *status != "" {
		base += ` AND "Status" = $2::financial_entry_status`
		args = append(args, *status)
	}

	var total int64
	if err := s.q(ctx).QueryRow(ctx, `SELECT COUNT(*) `+base, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	selectQ := `
		SELECT "Id", "Description", "VendorName", "ProviderInvoiceId", "PartnerSalespersonUserId",
			"IssueDate", "DueDate", "Amount", "PaidAmount", "Status"::text, "Notes", "CreatedAt"
		` + base + `
		ORDER BY "DueDate" ASC
		OFFSET $` + itoa(len(args)+1) + ` LIMIT $` + itoa(len(args)+2)
	args = append(args, page.Offset(), page.Limit())

	rows, err := s.q(ctx).Query(ctx, selectQ, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []models.ListAccountPayableResponse
	for rows.Next() {
		var item models.ListAccountPayableResponse
		if err := rows.Scan(
			&item.ID, &item.Description, &item.VendorName, &item.ProviderInvoiceID,
			&item.PartnerSalespersonUserID, &item.IssueDate, &item.DueDate,
			&item.Amount, &item.PaidAmount, &item.Status, &item.Notes, &item.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		item.Balance = item.Amount - item.PaidAmount
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (s *Store) CreateAccountPayable(ctx context.Context, id, orgID, description, vendorName string, providerInvoiceID, partnerUserID *string, issueDate, dueDate time.Time, amount float64, notes *string, now time.Time) error {
	_, err := s.q(ctx).Exec(ctx, `
		INSERT INTO "AccountsPayable" (
			"Id", "OrganizationId", "Description", "VendorName", "ProviderInvoiceId",
			"PartnerSalespersonUserId", "IssueDate", "DueDate", "Amount", "PaidAmount",
			"Status", "Notes", "CreatedAt", "UpdatedAt"
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, 0, 'open'::financial_entry_status, $10, $11, $11)`,
		id, orgID, description, vendorName, providerInvoiceID, partnerUserID,
		issueDate, dueDate, amount, notes, now)
	return err
}

func (s *Store) UpdateAccountPayable(ctx context.Context, orgID, id, description, vendorName string, dueDate time.Time, amount float64, status string, notes *string, now time.Time) error {
	tag, err := s.q(ctx).Exec(ctx, `
		UPDATE "AccountsPayable"
		SET "Description" = $3, "VendorName" = $4, "DueDate" = $5, "Amount" = $6,
			"Status" = $7::financial_entry_status, "Notes" = $8, "UpdatedAt" = $9
		WHERE "OrganizationId" = $1 AND "Id" = $2`,
		orgID, id, description, vendorName, dueDate, amount, status, notes, now)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (s *Store) GetAccountPayableAmounts(ctx context.Context, orgID, id string) (amount, paid float64, err error) {
	err = s.q(ctx).QueryRow(ctx, `
		SELECT "Amount", "PaidAmount" FROM "AccountsPayable"
		WHERE "OrganizationId" = $1 AND "Id" = $2`, orgID, id).Scan(&amount, &paid)
	return amount, paid, err
}

func (s *Store) RegisterPayablePayment(ctx context.Context, paymentID, orgID, accountID, userID string, amount float64, paymentDate time.Time, reference, notes *string, now time.Time) error {
	return s.WithTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		ctx = CtxWithTx(ctx, tx)

		var totalAmount, paidAmount float64
		var dueDate time.Time
		err := tx.QueryRow(ctx, `
			SELECT "Amount", "PaidAmount", "DueDate" FROM "AccountsPayable"
			WHERE "OrganizationId" = $1 AND "Id" = $2 FOR UPDATE`, orgID, accountID).Scan(&totalAmount, &paidAmount, &dueDate)
		if err != nil {
			return err
		}

		newPaid := paidAmount + amount
		status := computeFinancialStatus(totalAmount, newPaid, dueDate)
		_, err = tx.Exec(ctx, `
			UPDATE "AccountsPayable"
			SET "PaidAmount" = $3, "Status" = $4::financial_entry_status, "UpdatedAt" = $5
			WHERE "OrganizationId" = $1 AND "Id" = $2`,
			orgID, accountID, newPaid, status, now)
		if err != nil {
			return err
		}

		_, err = tx.Exec(ctx, `
			INSERT INTO "FinancialPayments" (
				"Id", "OrganizationId", "AccountType", "AccountId", "Amount",
				"PaymentDate", "Reference", "Notes", "CreatedByUserId", "CreatedAt"
			) VALUES ($1, $2, 'payable', $3, $4, $5, $6, $7, $8, $9)`,
			paymentID, orgID, accountID, amount, paymentDate, reference, notes, userID, now)
		return err
	})
}

func (s *Store) ListAccountsReceivable(ctx context.Context, orgID string, customerID, status *string, page httputil.PageSearch) ([]models.ListAccountReceivableResponse, int64, error) {
	base := `
		FROM "AccountsReceivable" ar
		JOIN "Customers" c ON c."Id" = ar."CustomerId"
		WHERE ar."OrganizationId" = $1`
	args := []any{orgID}
	if customerID != nil && *customerID != "" {
		base += ` AND ar."CustomerId" = $` + itoa(len(args)+1)
		args = append(args, *customerID)
	}
	if status != nil && *status != "" {
		base += ` AND ar."Status" = $` + itoa(len(args)+1) + `::financial_entry_status`
		args = append(args, *status)
	}

	var total int64
	if err := s.q(ctx).QueryRow(ctx, `SELECT COUNT(*) `+base, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	selectQ := `
		SELECT ar."Id", ar."CustomerId", c."Name", ar."Description", ar."ProcessingMonthId",
			ar."IssueDate", ar."DueDate", ar."Amount", ar."ReceivedAmount", ar."Status"::text,
			ar."Notes", ar."CreatedAt"
		` + base + `
		ORDER BY ar."DueDate" ASC
		OFFSET $` + itoa(len(args)+1) + ` LIMIT $` + itoa(len(args)+2)
	args = append(args, page.Offset(), page.Limit())

	rows, err := s.q(ctx).Query(ctx, selectQ, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []models.ListAccountReceivableResponse
	for rows.Next() {
		var item models.ListAccountReceivableResponse
		if err := rows.Scan(
			&item.ID, &item.CustomerID, &item.CustomerName, &item.Description, &item.ProcessingMonthID,
			&item.IssueDate, &item.DueDate, &item.Amount, &item.ReceivedAmount, &item.Status,
			&item.Notes, &item.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		item.Balance = item.Amount - item.ReceivedAmount
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (s *Store) CreateAccountReceivable(ctx context.Context, id, orgID, customerID, description string, processingMonthID *string, issueDate, dueDate time.Time, amount float64, notes *string, now time.Time) error {
	_, err := s.q(ctx).Exec(ctx, `
		INSERT INTO "AccountsReceivable" (
			"Id", "OrganizationId", "CustomerId", "Description", "ProcessingMonthId",
			"IssueDate", "DueDate", "Amount", "ReceivedAmount", "Status", "Notes", "CreatedAt", "UpdatedAt"
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 0, 'open'::financial_entry_status, $9, $10, $10)`,
		id, orgID, customerID, description, processingMonthID, issueDate, dueDate, amount, notes, now)
	return err
}

func (s *Store) UpdateAccountReceivable(ctx context.Context, orgID, id, description string, dueDate time.Time, amount float64, status string, notes *string, now time.Time) error {
	tag, err := s.q(ctx).Exec(ctx, `
		UPDATE "AccountsReceivable"
		SET "Description" = $3, "DueDate" = $4, "Amount" = $5,
			"Status" = $6::financial_entry_status, "Notes" = $7, "UpdatedAt" = $8
		WHERE "OrganizationId" = $1 AND "Id" = $2`,
		orgID, id, description, dueDate, amount, status, notes, now)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (s *Store) RegisterReceivablePayment(ctx context.Context, paymentID, orgID, accountID, userID string, amount float64, paymentDate time.Time, reference, notes *string, now time.Time) error {
	return s.WithTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		var totalAmount, receivedAmount float64
		var dueDate time.Time
		err := tx.QueryRow(ctx, `
			SELECT "Amount", "ReceivedAmount", "DueDate" FROM "AccountsReceivable"
			WHERE "OrganizationId" = $1 AND "Id" = $2 FOR UPDATE`, orgID, accountID).Scan(&totalAmount, &receivedAmount, &dueDate)
		if err != nil {
			return err
		}

		newReceived := receivedAmount + amount
		status := computeFinancialStatus(totalAmount, newReceived, dueDate)
		_, err = tx.Exec(ctx, `
			UPDATE "AccountsReceivable"
			SET "ReceivedAmount" = $3, "Status" = $4::financial_entry_status, "UpdatedAt" = $5
			WHERE "OrganizationId" = $1 AND "Id" = $2`,
			orgID, accountID, newReceived, status, now)
		if err != nil {
			return err
		}

		_, err = tx.Exec(ctx, `
			INSERT INTO "FinancialPayments" (
				"Id", "OrganizationId", "AccountType", "AccountId", "Amount",
				"PaymentDate", "Reference", "Notes", "CreatedByUserId", "CreatedAt"
			) VALUES ($1, $2, 'receivable', $3, $4, $5, $6, $7, $8, $9)`,
			paymentID, orgID, accountID, amount, paymentDate, reference, notes, userID, now)
		return err
	})
}

func (s *Store) GetPartnerCommissionPercent(ctx context.Context, orgID string) (float64, error) {
	var pct float64
	err := s.q(ctx).QueryRow(ctx, `
		SELECT "DefaultCommissionPercent" FROM "PartnerCommissionSettings"
		WHERE "OrganizationId" = $1`, orgID).Scan(&pct)
	if errors.Is(err, pgx.ErrNoRows) {
		return 10.0, nil
	}
	return pct, err
}

func (s *Store) UpsertPartnerCommissionSettings(ctx context.Context, orgID string, percent float64, now time.Time) error {
	_, err := s.q(ctx).Exec(ctx, `
		INSERT INTO "PartnerCommissionSettings" ("OrganizationId", "DefaultCommissionPercent", "UpdatedAt")
		VALUES ($1, $2, $3)
		ON CONFLICT ("OrganizationId") DO UPDATE
		SET "DefaultCommissionPercent" = EXCLUDED."DefaultCommissionPercent",
			"UpdatedAt" = EXCLUDED."UpdatedAt"`,
		orgID, percent, now)
	return err
}

func (s *Store) GetPartnerCommissionSettings(ctx context.Context, orgID string) (*models.PartnerCommissionSettingsResponse, error) {
	var item models.PartnerCommissionSettingsResponse
	err := s.q(ctx).QueryRow(ctx, `
		SELECT "DefaultCommissionPercent", "UpdatedAt"
		FROM "PartnerCommissionSettings" WHERE "OrganizationId" = $1`, orgID).Scan(
		&item.DefaultCommissionPercent, &item.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return &models.PartnerCommissionSettingsResponse{
			DefaultCommissionPercent: 10,
			UpdatedAt:                time.Now().UTC(),
		}, nil
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *Store) ListPartnerSales(ctx context.Context, orgID string, salespersonUserID, status *string, page httputil.PageSearch) ([]models.ListPartnerSaleResponse, int64, error) {
	base := `
		FROM "PartnerSalesRecords" ps
		JOIN "Customers" c ON c."Id" = ps."CustomerId"
		JOIN "PhoneLines" pl ON pl."Id" = ps."PhoneLineId"
		WHERE ps."OrganizationId" = $1`
	args := []any{orgID}
	if salespersonUserID != nil && *salespersonUserID != "" {
		base += ` AND ps."SalespersonUserId" = $` + itoa(len(args)+1)
		args = append(args, *salespersonUserID)
	}
	if status != nil && *status != "" {
		base += ` AND ps."Status" = $` + itoa(len(args)+1) + `::partner_sale_status`
		args = append(args, *status)
	}

	var total int64
	if err := s.q(ctx).QueryRow(ctx, `SELECT COUNT(*) `+base, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	selectQ := `
		SELECT ps."Id", ps."SalespersonUserId", ps."CustomerId", c."Name", ps."PhoneLineId", pl."Number",
			ps."ReferenceMonth", ps."GrossAmount", ps."CommissionPercent", ps."CommissionAmount",
			ps."Status"::text, ps."AccountPayableId", ps."CreatedAt"
		` + base + `
		ORDER BY ps."ReferenceMonth" DESC, c."Name"
		OFFSET $` + itoa(len(args)+1) + ` LIMIT $` + itoa(len(args)+2)
	args = append(args, page.Offset(), page.Limit())

	rows, err := s.q(ctx).Query(ctx, selectQ, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []models.ListPartnerSaleResponse
	for rows.Next() {
		var item models.ListPartnerSaleResponse
		if err := rows.Scan(
			&item.ID, &item.SalespersonUserID, &item.CustomerID, &item.CustomerName,
			&item.PhoneLineID, &item.PhoneLineNumber, &item.ReferenceMonth,
			&item.GrossAmount, &item.CommissionPercent, &item.CommissionAmount,
			&item.Status, &item.AccountPayableID, &item.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (s *Store) UpdatePartnerSaleStatus(ctx context.Context, orgID, id, status string, accountPayableID *string, now time.Time) error {
	tag, err := s.q(ctx).Exec(ctx, `
		UPDATE "PartnerSalesRecords"
		SET "Status" = $3::partner_sale_status, "AccountPayableId" = $4, "UpdatedAt" = $5
		WHERE "OrganizationId" = $1 AND "Id" = $2`,
		orgID, id, status, accountPayableID, now)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (s *Store) GetPartnerSaleByID(ctx context.Context, orgID, id string) (*models.ListPartnerSaleResponse, error) {
	var item models.ListPartnerSaleResponse
	err := s.q(ctx).QueryRow(ctx, `
		SELECT ps."Id", ps."SalespersonUserId", ps."CustomerId", c."Name", ps."PhoneLineId", pl."Number",
			ps."ReferenceMonth", ps."GrossAmount", ps."CommissionPercent", ps."CommissionAmount",
			ps."Status"::text, ps."AccountPayableId", ps."CreatedAt"
		FROM "PartnerSalesRecords" ps
		JOIN "Customers" c ON c."Id" = ps."CustomerId"
		JOIN "PhoneLines" pl ON pl."Id" = ps."PhoneLineId"
		WHERE ps."OrganizationId" = $1 AND ps."Id" = $2`, orgID, id).Scan(
		&item.ID, &item.SalespersonUserID, &item.CustomerID, &item.CustomerName,
		&item.PhoneLineID, &item.PhoneLineNumber, &item.ReferenceMonth,
		&item.GrossAmount, &item.CommissionPercent, &item.CommissionAmount,
		&item.Status, &item.AccountPayableID, &item.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *Store) InsertPartnerSaleRecord(ctx context.Context, id, orgID, salespersonID, customerID, phoneLineID string, refMonth time.Time, gross, commissionPercent, commission float64, now time.Time) (bool, error) {
	tag, err := s.q(ctx).Exec(ctx, `
		INSERT INTO "PartnerSalesRecords" (
			"Id", "OrganizationId", "SalespersonUserId", "CustomerId", "PhoneLineId",
			"ReferenceMonth", "GrossAmount", "CommissionPercent", "CommissionAmount",
			"Status", "CreatedAt", "UpdatedAt"
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, 'accrued', $10, $10)
		ON CONFLICT ("PhoneLineId", "ReferenceMonth") DO NOTHING`,
		id, orgID, salespersonID, customerID, phoneLineID, refMonth, gross, commissionPercent, commission, now)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

type partnerSaleCandidate struct {
	SalespersonID string
	CustomerID    string
	PhoneLineID   string
	Gross         float64
}

func (s *Store) ListPartnerSaleCandidates(ctx context.Context, orgID string) ([]partnerSaleCandidate, error) {
	rows, err := s.q(ctx).Query(ctx, `
		SELECT c."ResponsibleSalespersonUserId", c."Id", pl."Id",
			COALESCE(pl."CostWithConsumption", pl."BaseCost", 0)
		FROM "PhoneLines" pl
		JOIN "PhoneLineCustomerLinks" l ON l."PhoneLineId" = pl."Id" AND l."EndDate" IS NULL
		JOIN "Customers" c ON c."Id" = l."CustomerId"
		WHERE c."OrganizationId" = $1
		  AND c."ResponsibleSalespersonUserId" IS NOT NULL
		  AND c."ResponsibleSalespersonUserId" <> ''`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []partnerSaleCandidate
	for rows.Next() {
		var item partnerSaleCandidate
		if err := rows.Scan(&item.SalespersonID, &item.CustomerID, &item.PhoneLineID, &item.Gross); err != nil {
			return nil, err
		}
		if item.Gross > 0 {
			items = append(items, item)
		}
	}
	return items, rows.Err()
}

func (s *Store) SyncPartnerSalesFromLines(ctx context.Context, orgID string, refMonth time.Time, commissionPercent float64, now time.Time, newID func() string) (int, error) {
	candidates, err := s.ListPartnerSaleCandidates(ctx, orgID)
	if err != nil {
		return 0, err
	}
	inserted := 0
	for _, c := range candidates {
		commission := c.Gross * commissionPercent / 100
		ok, err := s.InsertPartnerSaleRecord(ctx, newID(), orgID, c.SalespersonID, c.CustomerID, c.PhoneLineID, refMonth, c.Gross, commissionPercent, commission, now)
		if err != nil {
			return inserted, err
		}
		if ok {
			inserted++
		}
	}
	return inserted, nil
}

func (s *Store) GetPartnerFinancialSummary(ctx context.Context, orgID, salespersonUserID string) (*models.PartnerFinancialSummaryResponse, error) {
	var summary models.PartnerFinancialSummaryResponse
	err := s.q(ctx).QueryRow(ctx, `
		SELECT
			COALESCE(SUM("GrossAmount"), 0),
			COALESCE(SUM("CommissionAmount") FILTER (WHERE "Status" = 'accrued'), 0),
			COALESCE(SUM("CommissionAmount") FILTER (WHERE "Status" = 'approved'), 0),
			COALESCE(SUM("CommissionAmount") FILTER (WHERE "Status" = 'paid'), 0),
			COALESCE(COUNT(*) FILTER (WHERE "Status" = 'accrued'), 0)::int
		FROM "PartnerSalesRecords"
		WHERE "OrganizationId" = $1 AND "SalespersonUserId" = $2`,
		orgID, salespersonUserID).Scan(
		&summary.TotalGrossSales,
		&summary.TotalCommissionAccrued,
		&summary.TotalCommissionApproved,
		&summary.TotalCommissionPaid,
		&summary.PendingSalesCount,
	)
	if err != nil {
		return nil, err
	}

	_ = s.q(ctx).QueryRow(ctx, `
		SELECT COALESCE(SUM(ar."Amount" - ar."ReceivedAmount"), 0)
		FROM "AccountsReceivable" ar
		JOIN "Customers" c ON c."Id" = ar."CustomerId"
		WHERE ar."OrganizationId" = $1
		  AND c."ResponsibleSalespersonUserId" = $2
		  AND ar."Status" IN ('open', 'partially_settled', 'overdue')`,
		orgID, salespersonUserID).Scan(&summary.TotalReceivableFromSales)

	return &summary, nil
}

func (s *Store) ProviderInvoiceExistsInOrg(ctx context.Context, orgID, invoiceID string) (bool, error) {
	var exists bool
	err := s.q(ctx).QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM "ProviderInvoices" i
			JOIN "ProviderAccounts" pa ON pa."Id" = i."ProviderAccountId"
			JOIN "ContractingCompanies" cc ON cc."Id" = pa."ContractingCompanyId"
			JOIN "Providers" p ON p."Id" = cc."ProviderId"
			WHERE p."OrganizationId" = $1 AND i."Id" = $2)`, orgID, invoiceID).Scan(&exists)
	return exists, err
}

func (s *Store) PayableExistsForProviderInvoice(ctx context.Context, orgID, invoiceID string) (bool, error) {
	var exists bool
	err := s.q(ctx).QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM "AccountsPayable"
			WHERE "OrganizationId" = $1 AND "ProviderInvoiceId" = $2)`, orgID, invoiceID).Scan(&exists)
	return exists, err
}

func (s *Store) GetProviderInvoiceForPayable(ctx context.Context, orgID, invoiceID string) (vendorName, description string, dueDate time.Time, amount float64, err error) {
	err = s.q(ctx).QueryRow(ctx, `
		SELECT p."Name" || ' - ' || pa."AccountNumber", 'Fatura operadora ' || COALESCE(NULLIF(i."Number", ''), i."Id"),
			i."DueDate", i."TotalAmount"
		FROM "ProviderInvoices" i
		JOIN "ProviderAccounts" pa ON pa."Id" = i."ProviderAccountId"
		JOIN "ContractingCompanies" cc ON cc."Id" = i."ContractingCompanyId"
		JOIN "Providers" p ON p."Id" = cc."ProviderId"
		WHERE p."OrganizationId" = $1 AND i."Id" = $2`, orgID, invoiceID).Scan(&vendorName, &description, &dueDate, &amount)
	return vendorName, description, dueDate, amount, err
}

func computeFinancialStatus(total, settled float64, dueDate time.Time) string {
	if settled >= total {
		return "settled"
	}
	if settled > 0 {
		return "partially_settled"
	}
	today := time.Now().UTC().Truncate(24 * time.Hour)
	if dueDate.Before(today) {
		return "overdue"
	}
	return "open"
}
