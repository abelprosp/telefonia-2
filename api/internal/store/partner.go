package store

import (
	"context"

	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/models"
)

func (s *Store) GetPartnerDashboardStats(ctx context.Context, orgID, salespersonUserID string) (*models.PartnerDashboardStatsResponse, error) {
	q := s.q(ctx)
	var stats models.PartnerDashboardStatsResponse
	err := q.QueryRow(ctx, `
		SELECT
			(SELECT COUNT(*)::int FROM "Customers"
			 WHERE "OrganizationId" = $1 AND "ResponsibleSalespersonUserId" = $2),
			(SELECT COUNT(DISTINCT pl."Id")::int
			 FROM "PhoneLines" pl
			 JOIN "PhoneLineCustomerLinks" l ON l."PhoneLineId" = pl."Id" AND l."EndDate" IS NULL
			 JOIN "Customers" c ON c."Id" = l."CustomerId"
			 WHERE c."OrganizationId" = $1 AND c."ResponsibleSalespersonUserId" = $2),
			(SELECT COUNT(*)::int FROM "PhoneLineOperationRequests"
			 WHERE "OrganizationId" = $1 AND "RequestedByUserId" = $2 AND "Status" = 'pending'),
			COALESCE((SELECT SUM(pl."BaseCost")
			 FROM "PhoneLines" pl
			 JOIN "PhoneLineCustomerLinks" l ON l."PhoneLineId" = pl."Id" AND l."EndDate" IS NULL
			 JOIN "Customers" c ON c."Id" = l."CustomerId"
			 WHERE c."OrganizationId" = $1 AND c."ResponsibleSalespersonUserId" = $2), 0),
			COALESCE((SELECT SUM(pl."CostWithConsumption")
			 FROM "PhoneLines" pl
			 JOIN "PhoneLineCustomerLinks" l ON l."PhoneLineId" = pl."Id" AND l."EndDate" IS NULL
			 JOIN "Customers" c ON c."Id" = l."CustomerId"
			 WHERE c."OrganizationId" = $1 AND c."ResponsibleSalespersonUserId" = $2), 0)`,
		orgID, salespersonUserID).Scan(
		&stats.CustomersCount,
		&stats.PhoneLinesCount,
		&stats.PendingOperationRequests,
		&stats.TotalBaseCost,
		&stats.TotalCostWithConsumption,
	)
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

func (s *Store) ListPartnerPhoneLines(ctx context.Context, orgID, salespersonUserID string, page httputil.PageSearch) ([]models.PartnerPhoneLineResponse, int64, error) {
	base := `
		FROM "PhoneLines" pl
		JOIN "ProviderPlans" pp ON pp."Id" = pl."ProviderPlanId"
		JOIN "PhoneLineCustomerLinks" l ON l."PhoneLineId" = pl."Id" AND l."EndDate" IS NULL
		JOIN "Customers" c ON c."Id" = l."CustomerId"
		WHERE c."OrganizationId" = $1 AND c."ResponsibleSalespersonUserId" = $2`
	args := []any{orgID, salespersonUserID}

	var total int64
	if err := s.q(ctx).QueryRow(ctx, `SELECT COUNT(*) `+base, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	selectQ := `
		SELECT pl."Id", pl."Number", pl."Status"::text, pl."TransitionSubStatus"::text,
			c."Id", c."Name", pp."Name", pl."BaseCost", pl."CostWithConsumption"
		` + base + `
		ORDER BY pl."Number"
		OFFSET $3 LIMIT $4`
	args = append(args, page.Offset(), page.Limit())

	rows, err := s.q(ctx).Query(ctx, selectQ, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []models.PartnerPhoneLineResponse
	for rows.Next() {
		var item models.PartnerPhoneLineResponse
		if err := rows.Scan(
			&item.ID, &item.Number, &item.Status, &item.TransitionSubStatus,
			&item.CustomerID, &item.CustomerName, &item.ProviderPlanName,
			&item.BaseCost, &item.CostWithConsumption,
		); err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (s *Store) PhoneLineOwnedBySalesperson(ctx context.Context, orgID, phoneLineID, salespersonUserID string) (bool, error) {
	var exists bool
	err := s.q(ctx).QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1
			FROM "PhoneLines" pl
			JOIN "PhoneLineCustomerLinks" l ON l."PhoneLineId" = pl."Id" AND l."EndDate" IS NULL
			JOIN "Customers" c ON c."Id" = l."CustomerId"
			WHERE c."OrganizationId" = $1 AND pl."Id" = $2 AND c."ResponsibleSalespersonUserId" = $3)`,
		orgID, phoneLineID, salespersonUserID).Scan(&exists)
	return exists, err
}
