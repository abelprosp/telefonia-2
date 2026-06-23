package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/models"
)

func (s *Store) ListCostCenters(ctx context.Context, orgID string, page httputil.PageSearch) ([]models.ListCostCenterResponse, int64, error) {
	q := s.q(ctx)
	var total int64
	if err := q.QueryRow(ctx, `SELECT COUNT(*) FROM "CostCenters" WHERE "OrganizationId" = $1`, orgID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := q.Query(ctx, `
		SELECT "Id", "Name", "Description"
		FROM "CostCenters"
		WHERE "OrganizationId" = $1
		ORDER BY "Name"
		OFFSET $2 LIMIT $3`, orgID, page.Offset(), page.Limit())
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []models.ListCostCenterResponse
	for rows.Next() {
		var item models.ListCostCenterResponse
		if err := rows.Scan(&item.ID, &item.Name, &item.Description); err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

type ImportRequestRow struct {
	ID                string
	OrganizationID    string
	ProviderID        string
	ProcessingMonthID string
	StorageBucket     string
	StorageObjectKey  string
	OriginalFileName  *string
	Status            int
	Error             *string
	CompletedAt       *time.Time
	CreatedBy         string
}

func (s *Store) GetImportRequest(ctx context.Context, id string) (*ImportRequestRow, error) {
	var r ImportRequestRow
	err := s.q(ctx).QueryRow(ctx, `
		SELECT "Id", "OrganizationId", "ProviderId", "ProcessingMonthId",
			"StorageBucket", "StorageObjectKey", "OriginalFileName", "Status",
			"Error", "CompletedAt", "CreatedBy"
		FROM "ProviderInvoiceImportRequests"
		WHERE "Id" = $1`, id).
		Scan(&r.ID, &r.OrganizationID, &r.ProviderID, &r.ProcessingMonthID,
			&r.StorageBucket, &r.StorageObjectKey, &r.OriginalFileName, &r.Status,
			&r.Error, &r.CompletedAt, &r.CreatedBy)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &r, err
}

func (s *Store) CreateImportRequest(ctx context.Context, r ImportRequestRow) error {
	_, err := s.q(ctx).Exec(ctx, `
		INSERT INTO "ProviderInvoiceImportRequests"
		("Id", "OrganizationId", "ProviderId", "ProcessingMonthId", "StorageBucket",
			"StorageObjectKey", "OriginalFileName", "Status", "CreatedBy")
		VALUES ($1, $2, $3, $4, $5, $6, $7, 0, $8)`,
		r.ID, r.OrganizationID, r.ProviderID, r.ProcessingMonthID,
		r.StorageBucket, r.StorageObjectKey, r.OriginalFileName, r.CreatedBy)
	return err
}

func (s *Store) UpdateImportRequestStatus(ctx context.Context, id string, status int, errMsg *string, completedAt *time.Time) error {
	_, err := s.q(ctx).Exec(ctx, `
		UPDATE "ProviderInvoiceImportRequests"
		SET "Status" = $2, "Error" = $3, "CompletedAt" = $4
		WHERE "Id" = $1`, id, status, errMsg, completedAt)
	return err
}

func (s *Store) GetDashboardStats(ctx context.Context, orgID string) (*models.DashboardStatsResponse, error) {
	providers, err := s.CountDashboardProviders(ctx, orgID)
	if err != nil {
		return nil, err
	}
	customers, err := s.CountDashboardCustomers(ctx, orgID)
	if err != nil {
		return nil, err
	}
	cycles, err := s.CountDashboardBillingCycles(ctx, orgID)
	if err != nil {
		return nil, err
	}
	invoices, err := s.CountDashboardProviderInvoices(ctx, orgID)
	if err != nil {
		return nil, err
	}
	lines, err := s.CountDashboardPhoneLines(ctx, orgID)
	if err != nil {
		return nil, err
	}
	return &models.DashboardStatsResponse{
		ProvidersCount: providers, CustomersCount: customers,
		BillingCyclesCount: cycles, ProviderInvoicesCount: invoices, PhoneLinesCount: lines,
	}, nil
}
