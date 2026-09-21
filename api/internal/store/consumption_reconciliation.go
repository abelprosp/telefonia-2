package store

import (
	"context"
	"fmt"
	"time"

	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/models"
	"github.com/luxus-connect/telefonia/api/internal/phone"
)

type ConsumptionReportFilter struct {
	ProcessingMonthID string
	CustomerID        string
	PhoneLineID       string
	ProviderID        string
	Status            string
	From              *time.Time
	To                *time.Time
}

func (s *Store) ListLineConsumptionReport(ctx context.Context, orgID string, f ConsumptionReportFilter, page httputil.PageSearch) ([]models.LineConsumptionReportItem, int64, float64, error) {
	base := `
		FROM "ProviderInvoicePhoneLines" j
		JOIN "ProviderInvoices" i ON i."Id" = j."ProviderInvoicesId"
		JOIN "PhoneLines" pl ON pl."Id" = j."PhoneLinesId"
		JOIN "ProviderAccounts" pa ON pa."Id" = pl."ProviderAccountId"
		JOIN "ContractingCompanies" cc ON cc."Id" = pa."ContractingCompanyId"
		JOIN "Providers" p ON p."Id" = cc."ProviderId"
		JOIN "ProcessingMonths" pm ON pm."Id" = i."ProcessingMonthId"
		LEFT JOIN "PhoneLineCustomerLinks" l ON l."PhoneLineId" = pl."Id"
			AND l."StartDate" <= COALESCE(i."IssueDate", CURRENT_DATE)
			AND (l."EndDate" IS NULL OR l."EndDate" >= COALESCE(i."IssueDate", CURRENT_DATE))
		LEFT JOIN "Customers" c ON c."Id" = l."CustomerId"
		WHERE p."OrganizationId" = $1
			AND i."Status" NOT IN ('substituted', 'cancelled')`
	args := []any{orgID}
	if f.ProcessingMonthID != "" {
		args = append(args, f.ProcessingMonthID)
		base += ` AND i."ProcessingMonthId" = $` + itoa(len(args))
	}
	if f.CustomerID != "" {
		args = append(args, f.CustomerID)
		base += ` AND l."CustomerId" = $` + itoa(len(args))
	}
	if f.PhoneLineID != "" {
		args = append(args, f.PhoneLineID)
		base += ` AND pl."Id" = $` + itoa(len(args))
	}
	if f.ProviderID != "" {
		args = append(args, f.ProviderID)
		base += ` AND p."Id" = $` + itoa(len(args))
	}
	if f.Status != "" {
		args = append(args, f.Status)
		base += ` AND pl."Status" = $` + itoa(len(args)) + `::phone_line_status`
	}
	if page.Search != "" {
		digits := phone.Digits(page.Search)
		args = append(args, "%"+page.Search+"%")
		idx := len(args)
		base += ` AND (pl."Number" ILIKE $` + itoa(idx) + ` OR COALESCE(c."Name",'') ILIKE $` + itoa(idx)
		if digits != "" {
			args = append(args, "%"+digits+"%")
			base += ` OR COALESCE(pl."NormalizedNumber", pl."Number") LIKE $` + itoa(len(args))
		}
		base += `)`
	}

	var total int64
	if err := s.q(ctx).QueryRow(ctx, `SELECT COUNT(*) `+base, args...).Scan(&total); err != nil {
		return nil, 0, 0, err
	}
	var sum float64
	_ = s.q(ctx).QueryRow(ctx, `
		SELECT COALESCE(SUM(COALESCE(pl."CostWithConsumption", pl."BaseCost", 0)), 0) `+base, args...).Scan(&sum)

	selectQ := `
		SELECT pl."Id", pl."Number", p."Id", p."Name",
			l."CustomerId", c."Name",
			pm."Id", pm."DisplayName", pm."Year", pm."Month",
			pl."Status"::text,
			pl."BaseCost", pl."CostWithConsumption",
			i."Id", i."Number", i."TotalAmount"
		` + base + `
		ORDER BY pm."Year" DESC, pm."Month" DESC, pl."Number"
		OFFSET $` + itoa(len(args)+1) + ` LIMIT $` + itoa(len(args)+2)
	args = append(args, page.Offset(), page.Limit())

	rows, err := s.q(ctx).Query(ctx, selectQ, args...)
	if err != nil {
		return nil, 0, 0, err
	}
	defer rows.Close()
	var items []models.LineConsumptionReportItem
	for rows.Next() {
		var item models.LineConsumptionReportItem
		if err := rows.Scan(
			&item.PhoneLineID, &item.PhoneNumber, &item.ProviderID, &item.ProviderName,
			&item.CustomerID, &item.CustomerName,
			&item.ProcessingMonthID, &item.ProcessingMonthName, &item.Year, &item.Month,
			&item.Status, &item.BaseCost, &item.CostWithConsumption,
			&item.InvoiceID, &item.InvoiceNumber, &item.InvoiceTotal,
		); err != nil {
			return nil, 0, 0, err
		}
		items = append(items, item)
	}
	if items == nil {
		items = []models.LineConsumptionReportItem{}
	}
	return items, total, sum, rows.Err()
}

func (s *Store) ListLinesMissingFromExternalSnapshot(ctx context.Context, orgID, source, providerID string, jobID string, isPartial bool) ([]models.ReconciliationFindingResponse, error) {
	q := `
		SELECT pl."Id", pl."Number", pl."Status"::text, l."CustomerId", c."Name", p."Name"
		FROM "PhoneLines" pl
		JOIN "ProviderAccounts" pa ON pa."Id" = pl."ProviderAccountId"
		JOIN "ContractingCompanies" cc ON cc."Id" = pa."ContractingCompanyId"
		JOIN "Providers" p ON p."Id" = cc."ProviderId"
		LEFT JOIN "PhoneLineCustomerLinks" l ON l."PhoneLineId" = pl."Id" AND l."EndDate" IS NULL
		LEFT JOIN "Customers" c ON c."Id" = l."CustomerId"
		WHERE p."OrganizationId" = $1
			AND ($2 = '' OR p."Id" = $2)
			AND pl."Status"::text NOT IN ('cancelled')
			AND NOT EXISTS (
				SELECT 1 FROM "ExternalLineSnapshots" s
				WHERE s."OrganizationId" = $1
					AND s."Source" = $3
					AND ($4 = '' OR s."ImportJobId" = $4)
					AND s."NormalizedNumber" = COALESCE(pl."NormalizedNumber", regexp_replace(pl."Number", '[^0-9]', '', 'g'))
			)
		ORDER BY pl."Number"
		LIMIT 5000`
	rows, err := s.q(ctx).Query(ctx, q, orgID, providerID, source, jobID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []models.ReconciliationFindingResponse
	for rows.Next() {
		var lineID, number, status string
		var customerID, customerName, providerName *string
		if err := rows.Scan(&lineID, &number, &status, &customerID, &customerName, &providerName); err != nil {
			return nil, err
		}
		item := models.ReconciliationFindingResponse{
			PhoneLineID:      &lineID,
			NormalizedNumber: phone.Digits(number),
			InternalStatus:   &status,
			CustomerID:       customerID,
			CustomerName:     customerName,
			ProviderName:     providerName,
			FindingType:      "missing_in_operator",
			IsInconclusive:   isPartial,
		}
		if isPartial {
			item.Evidence = "Fonte externa marcada como parcial; divergência inconclusiva."
		} else {
			item.Evidence = fmt.Sprintf("Linha presente no sistema e ausente na base %s.", source)
		}
		items = append(items, item)
	}
	if items == nil {
		items = []models.ReconciliationFindingResponse{}
	}
	return items, rows.Err()
}

func (s *Store) ListCancelledExternallyActiveInternally(ctx context.Context, orgID, source, jobID string) ([]models.ReconciliationFindingResponse, error) {
	q := `
		SELECT pl."Id", pl."Number", pl."Status"::text, s."StatusExternal",
			l."CustomerId", c."Name", p."Name", s."ImportJobId"
		FROM "ExternalLineSnapshots" s
		JOIN "PhoneLines" pl ON COALESCE(pl."NormalizedNumber", regexp_replace(pl."Number", '[^0-9]', '', 'g')) = s."NormalizedNumber"
		JOIN "ProviderAccounts" pa ON pa."Id" = pl."ProviderAccountId"
		JOIN "ContractingCompanies" cc ON cc."Id" = pa."ContractingCompanyId"
		JOIN "Providers" p ON p."Id" = cc."ProviderId"
		LEFT JOIN "PhoneLineCustomerLinks" l ON l."PhoneLineId" = pl."Id" AND l."EndDate" IS NULL
		LEFT JOIN "Customers" c ON c."Id" = l."CustomerId"
		WHERE s."OrganizationId" = $1 AND p."OrganizationId" = $1
			AND s."Source" = $2
			AND ($3 = '' OR s."ImportJobId" = $3)
			AND LOWER(COALESCE(s."StatusExternal", '')) IN ('cancelled', 'cancelado', 'canceled', 'inactive', 'inativo')
			AND pl."Status"::text IN ('active', 'in_stock', 'suspended')
		ORDER BY pl."Number"
		LIMIT 5000`
	rows, err := s.q(ctx).Query(ctx, q, orgID, source, jobID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []models.ReconciliationFindingResponse
	for rows.Next() {
		var lineID, number, internal, external, job string
		var customerID, customerName, providerName *string
		if err := rows.Scan(&lineID, &number, &internal, &external, &customerID, &customerName, &providerName, &job); err != nil {
			return nil, err
		}
		items = append(items, models.ReconciliationFindingResponse{
			PhoneLineID:      &lineID,
			NormalizedNumber: phone.Digits(number),
			InternalStatus:   &internal,
			ExternalStatus:   &external,
			CustomerID:       customerID,
			CustomerName:     customerName,
			ProviderName:     providerName,
			ImportJobID:      &job,
			FindingType:      "cancelled_externally_active_internally",
			Evidence:         "Status externo indica cancelamento; linha ainda ativa/em estoque internamente.",
		})
	}
	if items == nil {
		items = []models.ReconciliationFindingResponse{}
	}
	return items, rows.Err()
}
