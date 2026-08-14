package store

import (
	"context"
	"time"

	"github.com/luxus-connect/telefonia/api/internal/models"
)

func (s *Store) ListMovementReports(ctx context.Context, orgID string, month *ProcessingMonthRow) (*models.MovementReportsResponse, error) {
	prevYear, prevMonth := month.Year, month.Month-1
	if prevMonth < 1 {
		prevMonth = 12
		prevYear--
	}

	var prevID *string
	var id string
	err := s.q(ctx).QueryRow(ctx, `
		SELECT "Id" FROM "ProcessingMonths"
		WHERE "OrganizationId" = $1 AND "ProviderId" = $2 AND "Year" = $3 AND "Month" = $4`,
		orgID, month.ProviderID, prevYear, prevMonth).Scan(&id)
	if err == nil {
		prevID = &id
	}

	entries, err := s.listInvoiceLinesForMonth(ctx, orgID, month.ID, prevID, true)
	if err != nil {
		return nil, err
	}
	exits, err := s.listInvoiceLinesForMonth(ctx, orgID, month.ID, prevID, false)
	if err != nil {
		return nil, err
	}
	pending, err := s.listActivationPending(ctx, orgID, month.ID, month.ProviderID)
	if err != nil {
		return nil, err
	}

	return &models.MovementReportsResponse{
		ProcessingMonthID:   month.ID,
		ProcessingMonthName: month.DisplayName,
		Entries:             entries,
		Exits:               exits,
		ActivationPending:   pending,
	}, nil
}

func (s *Store) listInvoiceLinesForMonth(ctx context.Context, orgID, currentMonthID string, prevMonthID *string, entries bool) ([]models.MovementReportItem, error) {
	var q string
	var args []any
	if entries {
		q = `
			SELECT DISTINCT pl."Id", pl."Number", pl."Status"::text, pl."LineClassification"::text,
				l."CustomerId", c."Name", pl."LastInvoiceId", inv."Number",
				pl."TransitionSubStatus"::text, pl."TransitionStartedAt"
			FROM "ProviderInvoicePhoneLines" j
			JOIN "ProviderInvoices" i ON i."Id" = j."ProviderInvoicesId"
			JOIN "PhoneLines" pl ON pl."Id" = j."PhoneLinesId"
			JOIN "ProviderAccounts" pa ON pa."Id" = pl."ProviderAccountId"
			JOIN "ContractingCompanies" cc ON cc."Id" = pa."ContractingCompanyId"
			JOIN "Providers" p ON p."Id" = cc."ProviderId"
			LEFT JOIN "PhoneLineCustomerLinks" l ON l."PhoneLineId" = pl."Id" AND l."EndDate" IS NULL
			LEFT JOIN "Customers" c ON c."Id" = l."CustomerId"
			LEFT JOIN "ProviderInvoices" inv ON inv."Id" = pl."LastInvoiceId"
			WHERE p."OrganizationId" = $1 AND i."ProcessingMonthId" = $2`
		args = []any{orgID, currentMonthID}
		if prevMonthID != nil {
			q += `
				AND NOT EXISTS (
					SELECT 1 FROM "ProviderInvoicePhoneLines" pj
					JOIN "ProviderInvoices" pi ON pi."Id" = pj."ProviderInvoicesId"
					WHERE pj."PhoneLinesId" = pl."Id" AND pi."ProcessingMonthId" = $3
				)`
			args = append(args, *prevMonthID)
		}
		q += ` ORDER BY pl."Number"`
	} else {
		if prevMonthID == nil {
			return []models.MovementReportItem{}, nil
		}
		q = `
			SELECT DISTINCT pl."Id", pl."Number", pl."Status"::text, pl."LineClassification"::text,
				l."CustomerId", c."Name", pl."LastInvoiceId", inv."Number",
				pl."TransitionSubStatus"::text, pl."TransitionStartedAt"
			FROM "ProviderInvoicePhoneLines" j
			JOIN "ProviderInvoices" i ON i."Id" = j."ProviderInvoicesId"
			JOIN "PhoneLines" pl ON pl."Id" = j."PhoneLinesId"
			JOIN "ProviderAccounts" pa ON pa."Id" = pl."ProviderAccountId"
			JOIN "ContractingCompanies" cc ON cc."Id" = pa."ContractingCompanyId"
			JOIN "Providers" p ON p."Id" = cc."ProviderId"
			LEFT JOIN "PhoneLineCustomerLinks" l ON l."PhoneLineId" = pl."Id" AND l."EndDate" IS NULL
			LEFT JOIN "Customers" c ON c."Id" = l."CustomerId"
			LEFT JOIN "ProviderInvoices" inv ON inv."Id" = pl."LastInvoiceId"
			WHERE p."OrganizationId" = $1 AND i."ProcessingMonthId" = $2
				AND NOT EXISTS (
					SELECT 1 FROM "ProviderInvoicePhoneLines" cj
					JOIN "ProviderInvoices" ci ON ci."Id" = cj."ProviderInvoicesId"
					WHERE cj."PhoneLinesId" = pl."Id" AND ci."ProcessingMonthId" = $3
				)
			ORDER BY pl."Number"`
		args = []any{orgID, *prevMonthID, currentMonthID}
	}

	rows, err := s.q(ctx).Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMovementItems(rows)
}

func (s *Store) listActivationPending(ctx context.Context, orgID, monthID, providerID string) ([]models.MovementReportItem, error) {
	rows, err := s.q(ctx).Query(ctx, `
		SELECT pl."Id", pl."Number", pl."Status"::text, pl."LineClassification"::text,
			l."CustomerId", c."Name", pl."LastInvoiceId", inv."Number",
			pl."TransitionSubStatus"::text, pl."TransitionStartedAt"
		FROM "PhoneLines" pl
		JOIN "ProviderAccounts" pa ON pa."Id" = pl."ProviderAccountId"
		JOIN "ContractingCompanies" cc ON cc."Id" = pa."ContractingCompanyId"
		JOIN "Providers" p ON p."Id" = cc."ProviderId"
		LEFT JOIN "PhoneLineCustomerLinks" l ON l."PhoneLineId" = pl."Id" AND l."EndDate" IS NULL
		LEFT JOIN "Customers" c ON c."Id" = l."CustomerId"
		LEFT JOIN "ProviderInvoices" inv ON inv."Id" = pl."LastInvoiceId"
		WHERE p."OrganizationId" = $1 AND p."Id" = $2
			AND pl."Status" = 'in_transition'::phone_line_status
			AND NOT EXISTS (
				SELECT 1 FROM "ProviderInvoicePhoneLines" j
				JOIN "ProviderInvoices" i ON i."Id" = j."ProviderInvoicesId"
				WHERE j."PhoneLinesId" = pl."Id" AND i."ProcessingMonthId" = $3
			)
		ORDER BY pl."TransitionStartedAt" NULLS LAST, pl."Number"`, orgID, providerID, monthID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items, err := scanMovementItems(rows)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	for i := range items {
		if items[i].TransitionStartedAt != nil {
			months := (now.Year()-items[i].TransitionStartedAt.Year())*12 + int(now.Month()-items[i].TransitionStartedAt.Month())
			if months < 1 {
				months = 1
			}
			items[i].PendingCycles = months
		} else {
			items[i].PendingCycles = 1
		}
	}
	return items, nil
}

func scanMovementItems(rows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
}) ([]models.MovementReportItem, error) {
	var items []models.MovementReportItem
	for rows.Next() {
		var item models.MovementReportItem
		if err := rows.Scan(&item.PhoneLineID, &item.Number, &item.Status, &item.LineClassification,
			&item.CustomerID, &item.CustomerName, &item.LastInvoiceID, &item.LastInvoiceNumber,
			&item.TransitionSubStatus, &item.TransitionStartedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if items == nil {
		items = []models.MovementReportItem{}
	}
	return items, rows.Err()
}
