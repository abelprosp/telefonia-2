package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/models"
)

func (s *Store) ListBillingCycles(ctx context.Context, orgID string, page httputil.PageSearch) ([]models.ListBillingCycleResponse, int64, error) {
	q := s.q(ctx)
	var total int64
	if err := q.QueryRow(ctx, `SELECT COUNT(*) FROM "BillingCycles" WHERE "OrganizationId" = $1`, orgID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := q.Query(ctx, `
		SELECT "Id", "ProviderId", "Code", "Name", "StartDate", "EndDate",
			"Status"::text, "ClosedAt", "ClosedBy"
		FROM "BillingCycles"
		WHERE "OrganizationId" = $1
		ORDER BY "StartDate" DESC
		OFFSET $2 LIMIT $3`, orgID, page.Offset(), page.Limit())
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []models.ListBillingCycleResponse
	for rows.Next() {
		var item models.ListBillingCycleResponse
		if err := rows.Scan(&item.ID, &item.ProviderID, &item.Code, &item.Name,
			&item.StartDate, &item.EndDate, &item.Status, &item.ClosedAt, &item.ClosedBy); err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (s *Store) GetBillingCycle(ctx context.Context, orgID, id string) (*models.GetBillingCycleResponse, error) {
	var item models.GetBillingCycleResponse
	err := s.q(ctx).QueryRow(ctx, `
		SELECT "Id", "ProviderId", "Code", "Name", "StartDate", "EndDate",
			"Status"::text, "ClosedAt", "ClosedBy"
		FROM "BillingCycles"
		WHERE "OrganizationId" = $1 AND "Id" = $2`, orgID, id).
		Scan(&item.ID, &item.ProviderID, &item.Code, &item.Name,
			&item.StartDate, &item.EndDate, &item.Status, &item.ClosedAt, &item.ClosedBy)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *Store) GetBillingCycleByCode(ctx context.Context, providerID, code string) (*models.ListBillingCycleResponse, error) {
	var item models.ListBillingCycleResponse
	err := s.q(ctx).QueryRow(ctx, `
		SELECT "Id", "ProviderId", "Code", "Name", "StartDate", "EndDate",
			"Status"::text, "ClosedAt", "ClosedBy"
		FROM "BillingCycles"
		WHERE "ProviderId" = $1 AND "Code" = $2`, providerID, code).
		Scan(&item.ID, &item.ProviderID, &item.Code, &item.Name,
			&item.StartDate, &item.EndDate, &item.Status, &item.ClosedAt, &item.ClosedBy)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &item, err
}

func (s *Store) CreateBillingCycle(ctx context.Context, orgID, id, providerID, code, name string, start, end time.Time) error {
	_, err := s.q(ctx).Exec(ctx, `
		INSERT INTO "BillingCycles" ("Id", "OrganizationId", "ProviderId", "Code", "Name",
			"StartDate", "EndDate", "Status")
		VALUES ($1, $2, $3, $4, $5, $6, $7, 'open'::billing_cycle_status)`,
		id, orgID, providerID, code, name, start, end)
	return err
}

func (s *Store) UpdateBillingCycle(ctx context.Context, orgID, id, providerID, code, name string, start, end time.Time) error {
	tag, err := s.q(ctx).Exec(ctx, `
		UPDATE "BillingCycles" SET "ProviderId" = $3, "Code" = $4, "Name" = $5,
			"StartDate" = $6, "EndDate" = $7
		WHERE "OrganizationId" = $1 AND "Id" = $2 AND "Status" = 'open'::billing_cycle_status`,
		orgID, id, providerID, code, name, start, end)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (s *Store) ExistsClosedProcessingMonthIntersecting(ctx context.Context, orgID, providerID string, start, end time.Time) (bool, error) {
	var exists bool
	err := s.q(ctx).QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM "ProcessingMonths"
			WHERE "OrganizationId" = $1 AND "ProviderId" = $2
				AND "Status" = 'closed'::processing_month_status
				AND make_date("Year", "Month", 1) <= $4::date
				AND (make_date("Year", "Month", 1) + interval '1 month' - interval '1 day') >= $3::date)`,
		orgID, providerID, start, end).Scan(&exists)
	return exists, err
}

func (s *Store) CountDashboardBillingCycles(ctx context.Context, orgID string) (int32, error) {
	var n int32
	err := s.q(ctx).QueryRow(ctx, `SELECT COUNT(*)::int FROM "BillingCycles" WHERE "OrganizationId" = $1`, orgID).Scan(&n)
	return n, err
}
