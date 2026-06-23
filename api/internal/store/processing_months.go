package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/models"
)

type ProcessingMonthRow struct {
	ID                      string
	OrganizationID          string
	ProviderID              string
	Year                    int
	Month                   int
	DisplayName             string
	Status                  string
	ClosedAt                *time.Time
	ClosedBy                *string
	ClosedInContingency     bool
	ContingencyJustification *string
}

func (s *Store) ListProcessingMonths(ctx context.Context, orgID string, page httputil.PageSearch) ([]models.ListProcessingMonthResponse, int64, error) {
	q := s.q(ctx)
	var total int64
	if err := q.QueryRow(ctx, `SELECT COUNT(*) FROM "ProcessingMonths" WHERE "OrganizationId" = $1`, orgID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := q.Query(ctx, `
		SELECT "Id", "ProviderId", "Year", "Month", "DisplayName", "Status"::text,
			"ClosedAt", "ClosedBy", "ClosedInContingency"
		FROM "ProcessingMonths"
		WHERE "OrganizationId" = $1
		ORDER BY "Year" DESC, "Month" DESC
		OFFSET $2 LIMIT $3`, orgID, page.Offset(), page.Limit())
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []models.ListProcessingMonthResponse
	for rows.Next() {
		var item models.ListProcessingMonthResponse
		if err := rows.Scan(&item.ID, &item.ProviderID, &item.Year, &item.Month, &item.DisplayName,
			&item.Status, &item.ClosedAt, &item.ClosedBy, &item.ClosedInContingency); err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (s *Store) GetProcessingMonth(ctx context.Context, orgID, id string) (*ProcessingMonthRow, error) {
	var m ProcessingMonthRow
	err := s.q(ctx).QueryRow(ctx, `
		SELECT "Id", "OrganizationId", "ProviderId", "Year", "Month", "DisplayName",
			"Status"::text, "ClosedAt", "ClosedBy", "ClosedInContingency", "ContingencyJustification"
		FROM "ProcessingMonths"
		WHERE "OrganizationId" = $1 AND "Id" = $2`, orgID, id).
		Scan(&m.ID, &m.OrganizationID, &m.ProviderID, &m.Year, &m.Month, &m.DisplayName,
			&m.Status, &m.ClosedAt, &m.ClosedBy, &m.ClosedInContingency, &m.ContingencyJustification)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &m, err
}

func (s *Store) ProcessingMonthDuplicateExists(ctx context.Context, orgID, providerID string, year, month int) (bool, error) {
	var exists bool
	err := s.q(ctx).QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM "ProcessingMonths"
			WHERE "OrganizationId" = $1 AND "ProviderId" = $2 AND "Year" = $3 AND "Month" = $4)`,
		orgID, providerID, year, month).Scan(&exists)
	return exists, err
}

func (s *Store) CreateProcessingMonth(ctx context.Context, orgID, id, providerID, displayName string, year, month int) error {
	_, err := s.q(ctx).Exec(ctx, `
		INSERT INTO "ProcessingMonths" ("Id", "OrganizationId", "ProviderId", "Year", "Month",
			"DisplayName", "Status", "ClosedInContingency")
		VALUES ($1, $2, $3, $4, $5, $6, 'open'::processing_month_status, false)`,
		id, orgID, providerID, year, month, displayName)
	return err
}

func (s *Store) CloseProcessingMonth(ctx context.Context, orgID, id, userID string, contingency bool, justification *string) error {
	tag, err := s.q(ctx).Exec(ctx, `
		UPDATE "ProcessingMonths"
		SET "Status" = 'closed'::processing_month_status, "ClosedAt" = $3,
			"ClosedBy" = $4, "ClosedInContingency" = $5, "ContingencyJustification" = $6
		WHERE "OrganizationId" = $1 AND "Id" = $2 AND "Status" = 'open'::processing_month_status`,
		orgID, id, time.Now().UTC(), userID, contingency, justification)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func ToProcessingMonthResponse(m *ProcessingMonthRow) models.GetProcessingMonthResponse {
	return models.GetProcessingMonthResponse{
		ListProcessingMonthResponse: models.ListProcessingMonthResponse{
			ID: m.ID, ProviderID: m.ProviderID, Year: m.Year, Month: m.Month,
			DisplayName: m.DisplayName, Status: m.Status, ClosedAt: m.ClosedAt,
			ClosedBy: m.ClosedBy, ClosedInContingency: m.ClosedInContingency,
		},
		ContingencyJustification: m.ContingencyJustification,
	}
}
