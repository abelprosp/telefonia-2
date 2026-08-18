package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/models"
)

func (s *Store) ListProviderInvoices(ctx context.Context, orgID string, processingMonthID *string, page httputil.PageSearch) ([]models.ListProviderInvoiceResponse, int64, error) {
	base := `
		FROM "ProviderInvoices" i
		JOIN "ProviderAccounts" pa ON pa."Id" = i."ProviderAccountId"
		JOIN "ContractingCompanies" cc ON cc."Id" = i."ContractingCompanyId"
		JOIN "Providers" p ON p."Id" = cc."ProviderId"
		LEFT JOIN "BillingCycles" bc ON bc."Id" = i."BillingCycleId"
		LEFT JOIN "AccountsPayable" ap ON ap."ProviderInvoiceId" = i."Id" AND ap."OrganizationId" = p."OrganizationId"
		WHERE p."OrganizationId" = $1`
	args := []any{orgID}
	if processingMonthID != nil && *processingMonthID != "" {
		base += ` AND i."ProcessingMonthId" = $2`
		args = append(args, *processingMonthID)
	}

	var total int64
	if err := s.q(ctx).QueryRow(ctx, `SELECT COUNT(*) `+base, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offsetParam := len(args) + 1
	limitParam := len(args) + 2
	selectQ := `
		SELECT i."Id", i."ProviderAccountId", pa."AccountNumber", i."ContractingCompanyId", cc."LegalName",
			p."Id", p."Name", COALESCE(i."BillingCycleId", ''), COALESCE(bc."Name", ''), i."ProcessingMonthId",
			i."CostCenterId", i."ParentInvoiceId", i."IssueDate", i."DueDate", i."TotalAmount",
			i."Status"::text, i."SubtotalServices", i."SubtotalUsage", i."SubtotalTaxes",
			i."SubtotalDiscounts", i."SubtotalInstallments",
			ap."Id", ap."Status"::text
		` + base + `
		ORDER BY i."IssueDate" DESC
		OFFSET $` + itoa(offsetParam) + ` LIMIT $` + itoa(limitParam)
	args = append(args, page.Offset(), page.Limit())

	rows, err := s.q(ctx).Query(ctx, selectQ, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []models.ListProviderInvoiceResponse
	for rows.Next() {
		var item models.ListProviderInvoiceResponse
		if err := rows.Scan(
			&item.ID, &item.ProviderAccountID, &item.ProviderAccountNumber,
			&item.ContractingCompanyID, &item.ContractingCompanyName,
			&item.ProviderID, &item.ProviderName, &item.BillingCycleID, &item.BillingCycleName,
			&item.ProcessingMonthID, &item.CostCenterID, &item.ParentInvoiceID,
			&item.IssueDate, &item.DueDate, &item.TotalAmount, &item.Status,
			&item.SubtotalServices, &item.SubtotalUsage, &item.SubtotalTaxes,
			&item.SubtotalDiscounts, &item.SubtotalInstallments,
			&item.AccountPayableID, &item.AccountPayableStatus); err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (s *Store) GetProviderInvoice(ctx context.Context, orgID, id string) (*models.GetProviderInvoiceResponse, error) {
	var item models.GetProviderInvoiceResponse
	err := s.q(ctx).QueryRow(ctx, `
		SELECT i."Id", i."ProviderAccountId", pa."AccountNumber", i."ContractingCompanyId", cc."LegalName",
			p."Id", p."Name", i."BillingCycleId", bc."Name", i."ProcessingMonthId",
			pm."DisplayName", i."CostCenterId", cst."Name", i."ParentInvoiceId",
			i."IssueDate", i."DueDate", i."TotalAmount", i."Status"::text,
			i."SubtotalServices", i."SubtotalUsage", i."SubtotalTaxes",
			i."SubtotalDiscounts", i."SubtotalInstallments", i."Number",
			ap."Id", ap."Status"::text
		FROM "ProviderInvoices" i
		JOIN "ProviderAccounts" pa ON pa."Id" = i."ProviderAccountId"
		JOIN "ContractingCompanies" cc ON cc."Id" = i."ContractingCompanyId"
		JOIN "Providers" p ON p."Id" = cc."ProviderId"
		JOIN "BillingCycles" bc ON bc."Id" = i."BillingCycleId"
		LEFT JOIN "ProcessingMonths" pm ON pm."Id" = i."ProcessingMonthId"
		LEFT JOIN "CostCenters" cst ON cst."Id" = i."CostCenterId"
		LEFT JOIN "AccountsPayable" ap ON ap."ProviderInvoiceId" = i."Id" AND ap."OrganizationId" = p."OrganizationId"
		WHERE p."OrganizationId" = $1 AND i."Id" = $2`, orgID, id).
		Scan(
			&item.ID, &item.ProviderAccountID, &item.ProviderAccountNumber,
			&item.ContractingCompanyID, &item.ContractingCompanyName,
			&item.ProviderID, &item.ProviderName, &item.BillingCycleID, &item.BillingCycleName,
			&item.ProcessingMonthID, &item.ProcessingMonthName, &item.CostCenterID, &item.CostCenterName,
			&item.ParentInvoiceID, &item.IssueDate, &item.DueDate, &item.TotalAmount, &item.Status,
			&item.SubtotalServices, &item.SubtotalUsage, &item.SubtotalTaxes,
			&item.SubtotalDiscounts, &item.SubtotalInstallments, &item.Number,
			&item.AccountPayableID, &item.AccountPayableStatus)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	item.ProviderInvoiceItems = []models.GetProviderInvoiceItemResponse{}
	item.ProviderInvoiceServices = []models.GetProviderInvoiceServiceResponse{}
	item.ProviderInvoiceQuotaSharing = []models.GetProviderInvoiceQuotaSharingResponse{}
	item.PhoneLines = []models.GetProviderPhoneLineResponse{}

	// Items (top-level only)
	itemRows, err := s.q(ctx).Query(ctx, `
		SELECT "Id", "InvoiceId", "ParentId", "Description", "Quantity", "TotalPrice",
			"ItemType"::text, "QuotaAmount", "ConsumedAmount", "Unit"::text
		FROM "ProviderInvoiceItems"
		WHERE "InvoiceId" = $1 AND "ParentId" IS NULL`, id)
	if err != nil {
		return nil, err
	}
	defer itemRows.Close()
	for itemRows.Next() {
		var it models.GetProviderInvoiceItemResponse
		if err := itemRows.Scan(&it.ID, &it.InvoiceID, &it.ParentID, &it.Description,
			&it.Quantity, &it.TotalPrice, &it.ItemType, &it.QuotaAmount, &it.ConsumedAmount, &it.Unit); err != nil {
			return nil, err
		}
		it.Children = []models.GetProviderInvoiceItemResponse{}
		item.ProviderInvoiceItems = append(item.ProviderInvoiceItems, it)
	}

	svcRows, err := s.q(ctx).Query(ctx, `
		SELECT s."Id", s."InvoiceId", s."PlanId", pp."Name", s."Description",
			s."Quantity", s."TotalPrice", s."QuotaAmount", s."ConsumedAmount", s."Unit"::text
		FROM "ProviderInvoiceServices" s
		JOIN "ProviderPlans" pp ON pp."Id" = s."PlanId"
		WHERE s."InvoiceId" = $1`, id)
	if err != nil {
		return nil, err
	}
	defer svcRows.Close()
	for svcRows.Next() {
		var svc models.GetProviderInvoiceServiceResponse
		if err := svcRows.Scan(&svc.ID, &svc.InvoiceID, &svc.PlanID, &svc.PlanName,
			&svc.Description, &svc.Quantity, &svc.TotalPrice, &svc.QuotaAmount,
			&svc.ConsumedAmount, &svc.Unit); err != nil {
			return nil, err
		}
		item.ProviderInvoiceServices = append(item.ProviderInvoiceServices, svc)
	}

	plRows, err := s.q(ctx).Query(ctx, `
		SELECT pl."Id", pl."ProviderPlanId", pp."Name", pl."ProviderAccountId", pa."AccountNumber",
			pl."CostCenterId", cst."Name", pl."LastInvoiceId", inv."Number",
			pl."TitularLineId", tit."Number", pl."Number", pl."LineClassification"::text,
			pl."Status"::text, pl."TransitionSubStatus"::text
		FROM "ProviderInvoicePhoneLines" j
		JOIN "PhoneLines" pl ON pl."Id" = j."PhoneLinesId"
		JOIN "ProviderPlans" pp ON pp."Id" = pl."ProviderPlanId"
		JOIN "ProviderAccounts" pa ON pa."Id" = pl."ProviderAccountId"
		LEFT JOIN "CostCenters" cst ON cst."Id" = pl."CostCenterId"
		LEFT JOIN "ProviderInvoices" inv ON inv."Id" = pl."LastInvoiceId"
		LEFT JOIN "PhoneLines" tit ON tit."Id" = pl."TitularLineId"
		WHERE j."ProviderInvoicesId" = $1`, id)
	if err != nil {
		return nil, err
	}
	defer plRows.Close()
	for plRows.Next() {
		var pl models.GetProviderPhoneLineResponse
		if err := plRows.Scan(&pl.ID, &pl.ProviderPlanID, &pl.ProviderPlanName, &pl.ProviderAccountID,
			&pl.ProviderAccountNumber, &pl.CostCenterID, &pl.CostCenterName, &pl.LastInvoiceID,
			&pl.LastInvoiceNumber, &pl.TitularLineID, &pl.TitularLineNumber, &pl.Number,
			&pl.LineClassification, &pl.Status, &pl.TransitionSubStatus); err != nil {
			return nil, err
		}
		item.PhoneLines = append(item.PhoneLines, pl)
	}

	return &item, nil
}

func (s *Store) InvoiceDuplicateExists(ctx context.Context, accountID, companyID, processingMonthID string, dueDate time.Time) (bool, error) {
	var exists bool
	err := s.q(ctx).QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM "ProviderInvoices"
			WHERE "ProviderAccountId" = $1 AND "ContractingCompanyId" = $2
				AND "ProcessingMonthId" = $3 AND "DueDate" = $4
				AND "Status" <> 'substituted'::provider_invoice_status)`,
		accountID, companyID, processingMonthID, dueDate).Scan(&exists)
	return exists, err
}

func (s *Store) InvoiceExistsInOtherProcessingMonth(ctx context.Context, accountID, companyID, processingMonthID string, dueDate time.Time) (bool, error) {
	var exists bool
	err := s.q(ctx).QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM "ProviderInvoices"
			WHERE "ProviderAccountId" = $1 AND "ContractingCompanyId" = $2
				AND "DueDate" = $3 AND "ProcessingMonthId" != $4
				AND "Status" <> 'substituted'::provider_invoice_status)`,
		accountID, companyID, dueDate, processingMonthID).Scan(&exists)
	return exists, err
}

func (s *Store) CreateProviderInvoice(ctx context.Context, inv ProviderInvoiceInsert) error {
	_, err := s.q(ctx).Exec(ctx, `
		INSERT INTO "ProviderInvoices" ("Id", "Number", "ProviderAccountId", "ContractingCompanyId",
			"BillingCycleId", "ProcessingMonthId", "IssueDate", "DueDate", "TotalAmount", "Status",
			"SubtotalServices", "SubtotalUsage", "SubtotalTaxes", "SubtotalDiscounts", "SubtotalInstallments",
			"ParentInvoiceId", "ContentSHA256", "SubstitutionImpact")
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, 'pending'::provider_invoice_status,
			$10, $11, $12, $13, $14, $15, $16, $17)`,
		inv.ID, inv.Number, inv.ProviderAccountID, inv.ContractingCompanyID,
		inv.BillingCycleID, inv.ProcessingMonthID, inv.IssueDate, inv.DueDate, inv.TotalAmount,
		inv.SubtotalServices, inv.SubtotalUsage, inv.SubtotalTaxes, inv.SubtotalDiscounts, inv.SubtotalInstallments,
		inv.ParentInvoiceID, inv.ContentSHA256, inv.SubstitutionImpact)
	return err
}

func (s *Store) LinkInvoicePhoneLine(ctx context.Context, invoiceID, phoneLineID string) error {
	_, err := s.q(ctx).Exec(ctx, `
		INSERT INTO "ProviderInvoicePhoneLines" ("PhoneLinesId", "ProviderInvoicesId")
		VALUES ($1, $2) ON CONFLICT DO NOTHING`, phoneLineID, invoiceID)
	return err
}

func (s *Store) CreateInvoiceItem(ctx context.Context, item InvoiceItemInsert) error {
	_, err := s.q(ctx).Exec(ctx, `
		INSERT INTO "ProviderInvoiceItems" ("Id", "InvoiceId", "ParentId", "Description",
			"Quantity", "TotalPrice", "ItemType", "QuotaAmount", "ConsumedAmount", "Unit")
		VALUES ($1, $2, $3, $4, $5, $6, $7::provider_invoice_item_type, $8, $9, $10::invoice_item_unit)`,
		item.ID, item.InvoiceID, item.ParentID, item.Description, item.Quantity, item.TotalPrice,
		item.ItemType, item.QuotaAmount, item.ConsumedAmount, item.Unit)
	return err
}

func (s *Store) CreateInvoiceService(ctx context.Context, svc InvoiceServiceInsert) error {
	_, err := s.q(ctx).Exec(ctx, `
		INSERT INTO "ProviderInvoiceServices" ("Id", "InvoiceId", "PlanId", "Description",
			"Quantity", "TotalPrice", "QuotaAmount", "ConsumedAmount", "Unit")
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9::invoice_item_unit)`,
		svc.ID, svc.InvoiceID, svc.PlanID, svc.Description, svc.Quantity, svc.TotalPrice,
		svc.QuotaAmount, svc.ConsumedAmount, svc.Unit)
	return err
}

type ProviderInvoiceInsert struct {
	ID                    string
	Number                string
	ProviderAccountID     string
	ContractingCompanyID  string
	BillingCycleID        string
	ProcessingMonthID     string
	IssueDate             time.Time
	DueDate               time.Time
	TotalAmount           float64
	SubtotalServices      float64
	SubtotalUsage         float64
	SubtotalTaxes         float64
	SubtotalDiscounts     float64
	SubtotalInstallments  float64
	ParentInvoiceID       *string
	ContentSHA256         *string
	SubstitutionImpact    *float64
}

type InvoiceItemInsert struct {
	ID             string
	InvoiceID      string
	ParentID       *string
	Description    string
	Quantity       float64
	TotalPrice     float64
	ItemType       string
	QuotaAmount    *float64
	ConsumedAmount *float64
	Unit           *string
}

type InvoiceServiceInsert struct {
	ID             string
	InvoiceID      string
	PlanID         string
	Description    string
	Quantity       float64
	TotalPrice     float64
	QuotaAmount    *float64
	ConsumedAmount *float64
	Unit           *string
}

func (s *Store) CountDashboardProviderInvoices(ctx context.Context, orgID string) (int32, error) {
	var n int32
	err := s.q(ctx).QueryRow(ctx, `
		SELECT COUNT(*)::int FROM "ProviderInvoices" i
		JOIN "ProviderAccounts" pa ON pa."Id" = i."ProviderAccountId"
		JOIN "ContractingCompanies" cc ON cc."Id" = pa."ContractingCompanyId"
		JOIN "Providers" p ON p."Id" = cc."ProviderId"
		WHERE p."OrganizationId" = $1`, orgID).Scan(&n)
	return n, err
}

func (s *Store) SumProviderInvoiceTotalsForMonth(ctx context.Context, orgID, processingMonthID string) (float64, error) {
	var total float64
	err := s.q(ctx).QueryRow(ctx, `
		SELECT COALESCE(SUM(i."TotalAmount"), 0)
		FROM "ProviderInvoices" i
		JOIN "ProviderAccounts" pa ON pa."Id" = i."ProviderAccountId"
		JOIN "ContractingCompanies" cc ON cc."Id" = pa."ContractingCompanyId"
		JOIN "Providers" p ON p."Id" = cc."ProviderId"
		WHERE p."OrganizationId" = $1 AND i."ProcessingMonthId" = $2
			AND i."Status" <> 'substituted'::provider_invoice_status
			AND i."Status" <> 'cancelled'::provider_invoice_status`, orgID, processingMonthID).Scan(&total)
	return total, err
}

type ProviderInvoiceIdentity struct {
	ID                    string
	TotalAmount           float64
	ProviderAccountID     string
	ProcessingMonthID     *string
	DueDate               time.Time
	Status                string
}

func (s *Store) FindActiveInvoiceByContentSHA256(ctx context.Context, hash string) (*ProviderInvoiceIdentity, error) {
	if hash == "" {
		return nil, nil
	}
	var row ProviderInvoiceIdentity
	err := s.q(ctx).QueryRow(ctx, `
		SELECT "Id", "TotalAmount", "ProviderAccountId", "ProcessingMonthId", "DueDate", "Status"::text
		FROM "ProviderInvoices"
		WHERE "ContentSHA256" = $1 AND "Status" <> 'substituted'::provider_invoice_status
		LIMIT 1`, hash).Scan(&row.ID, &row.TotalAmount, &row.ProviderAccountID, &row.ProcessingMonthID, &row.DueDate, &row.Status)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (s *Store) FindActiveInvoiceByBusinessKey(ctx context.Context, accountID, processingMonthID string, dueDate time.Time) (*ProviderInvoiceIdentity, error) {
	var row ProviderInvoiceIdentity
	err := s.q(ctx).QueryRow(ctx, `
		SELECT "Id", "TotalAmount", "ProviderAccountId", "ProcessingMonthId", "DueDate", "Status"::text
		FROM "ProviderInvoices"
		WHERE "ProviderAccountId" = $1 AND "ProcessingMonthId" = $2 AND "DueDate" = $3
			AND "Status" <> 'substituted'::provider_invoice_status
		LIMIT 1`, accountID, processingMonthID, dueDate).
		Scan(&row.ID, &row.TotalAmount, &row.ProviderAccountID, &row.ProcessingMonthID, &row.DueDate, &row.Status)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (s *Store) MarkInvoiceSubstituted(ctx context.Context, id string) error {
	_, err := s.q(ctx).Exec(ctx, `
		UPDATE "ProviderInvoices"
		SET "Status" = 'substituted'::provider_invoice_status
		WHERE "Id" = $1 AND "Status" <> 'substituted'::provider_invoice_status`, id)
	return err
}
