package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/models"
)

func (s *Store) CreatePhoneLineOperationRequest(ctx context.Context, id, orgID, phoneLineID, customerID, requestedBy, operationType string, justification *string, createdAt time.Time) error {
	_, err := s.q(ctx).Exec(ctx, `
		INSERT INTO "PhoneLineOperationRequests" (
			"Id", "OrganizationId", "PhoneLineId", "CustomerId", "RequestedByUserId",
			"OperationType", "Status", "Justification", "CreatedAt"
		) VALUES ($1, $2, $3, $4, $5, $6::phone_line_operation_type, 'pending'::phone_line_operation_status, $7, $8)`,
		id, orgID, phoneLineID, customerID, requestedBy, operationType, justification, createdAt)
	return err
}

func (s *Store) PendingPhoneLineOperationExists(ctx context.Context, orgID, phoneLineID string) (bool, error) {
	var exists bool
	err := s.q(ctx).QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM "PhoneLineOperationRequests"
			WHERE "OrganizationId" = $1 AND "PhoneLineId" = $2 AND "Status" = 'pending')`,
		orgID, phoneLineID).Scan(&exists)
	return exists, err
}

func (s *Store) ListPhoneLineOperationRequests(ctx context.Context, orgID string, requestedBy *string, page httputil.PageSearch) ([]models.PhoneLineOperationRequestResponse, int64, error) {
	base := `
		FROM "PhoneLineOperationRequests" r
		JOIN "PhoneLines" pl ON pl."Id" = r."PhoneLineId"
		JOIN "Customers" c ON c."Id" = r."CustomerId"
		WHERE r."OrganizationId" = $1`
	args := []any{orgID}
	if requestedBy != nil && *requestedBy != "" {
		base += ` AND r."RequestedByUserId" = $` + itoa(len(args)+1)
		args = append(args, *requestedBy)
	}

	var total int64
	if err := s.q(ctx).QueryRow(ctx, `SELECT COUNT(*) `+base, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	selectQ := `
		SELECT r."Id", r."PhoneLineId", pl."Number", r."CustomerId", c."Name",
			r."OperationType"::text, r."Status"::text, r."Justification", r."AdminNotes",
			r."RequestedByUserId", r."ReviewedByUserId", r."ReviewedAt", r."CreatedAt"
		` + base + `
		ORDER BY r."CreatedAt" DESC
		OFFSET $` + itoa(len(args)+1) + ` LIMIT $` + itoa(len(args)+2)
	args = append(args, page.Offset(), page.Limit())

	rows, err := s.q(ctx).Query(ctx, selectQ, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	return scanPhoneLineOperationRequests(rows, total)
}

func (s *Store) GetPhoneLineOperationRequest(ctx context.Context, orgID, id string, requestedBy *string) (*models.PhoneLineOperationRequestResponse, error) {
	query := `
		SELECT r."Id", r."PhoneLineId", pl."Number", r."CustomerId", c."Name",
			r."OperationType"::text, r."Status"::text, r."Justification", r."AdminNotes",
			r."RequestedByUserId", r."ReviewedByUserId", r."ReviewedAt", r."CreatedAt"
		FROM "PhoneLineOperationRequests" r
		JOIN "PhoneLines" pl ON pl."Id" = r."PhoneLineId"
		JOIN "Customers" c ON c."Id" = r."CustomerId"
		WHERE r."OrganizationId" = $1 AND r."Id" = $2`
	args := []any{orgID, id}
	if requestedBy != nil && *requestedBy != "" {
		query += ` AND r."RequestedByUserId" = $3`
		args = append(args, *requestedBy)
	}
	var item models.PhoneLineOperationRequestResponse
	err := s.q(ctx).QueryRow(ctx, query, args...).Scan(
		&item.ID, &item.PhoneLineID, &item.PhoneLineNumber, &item.CustomerID, &item.CustomerName,
		&item.OperationType, &item.Status, &item.Justification, &item.AdminNotes,
		&item.RequestedByUserID, &item.ReviewedByUserID, &item.ReviewedAt, &item.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *Store) ReviewPhoneLineOperationRequest(ctx context.Context, orgID, id, status, reviewedBy string, adminNotes *string, reviewedAt time.Time) error {
	tag, err := s.q(ctx).Exec(ctx, `
		UPDATE "PhoneLineOperationRequests"
		SET "Status" = $3::phone_line_operation_status,
			"AdminNotes" = $4,
			"ReviewedByUserId" = $5,
			"ReviewedAt" = $6
		WHERE "OrganizationId" = $1 AND "Id" = $2 AND "Status" = 'pending'`,
		orgID, id, status, adminNotes, reviewedBy, reviewedAt)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (s *Store) GetPhoneLineOperationRequestPhoneLineID(ctx context.Context, orgID, id string) (phoneLineID, operationType string, err error) {
	err = s.q(ctx).QueryRow(ctx, `
		SELECT "PhoneLineId", "OperationType"::text
		FROM "PhoneLineOperationRequests"
		WHERE "OrganizationId" = $1 AND "Id" = $2`, orgID, id).Scan(&phoneLineID, &operationType)
	return phoneLineID, operationType, err
}

func scanPhoneLineOperationRequests(rows pgx.Rows, total int64) ([]models.PhoneLineOperationRequestResponse, int64, error) {
	var items []models.PhoneLineOperationRequestResponse
	for rows.Next() {
		var item models.PhoneLineOperationRequestResponse
		if err := rows.Scan(
			&item.ID, &item.PhoneLineID, &item.PhoneLineNumber, &item.CustomerID, &item.CustomerName,
			&item.OperationType, &item.Status, &item.Justification, &item.AdminNotes,
			&item.RequestedByUserID, &item.ReviewedByUserID, &item.ReviewedAt, &item.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}
