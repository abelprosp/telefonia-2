package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

type ApprovalRequestRow struct {
	ID                  string
	OrganizationID      string
	ActionType          string
	EntityType          string
	EntityID            string
	Status              string
	RequesterUserID     string
	FirstApproverUserID *string
	SecondApproverUserID *string
	Justification       *string
	PayloadJSON         *string
	BeforeJSON          *string
	AfterJSON           *string
	RejectionReason     *string
	RejectedBy          *string
	CreatedAt           time.Time
	FirstApprovedAt     *time.Time
	SecondApprovedAt    *time.Time
	RejectedAt          *time.Time
	ExecutedAt          *time.Time
}

func (s *Store) InsertApprovalRequest(ctx context.Context, r ApprovalRequestRow) error {
	_, err := s.q(ctx).Exec(ctx, `
		INSERT INTO "ApprovalRequests" (
			"Id", "OrganizationId", "ActionType", "EntityType", "EntityId", "Status",
			"RequesterUserId", "Justification", "Payload", "BeforeSnapshot", "CreatedAt"
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9::jsonb,$10::jsonb,$11)`,
		r.ID, r.OrganizationID, r.ActionType, r.EntityType, r.EntityID, r.Status,
		r.RequesterUserID, r.Justification, r.PayloadJSON, r.BeforeJSON, r.CreatedAt)
	return err
}

func (s *Store) GetApprovalRequest(ctx context.Context, orgID, id string) (*ApprovalRequestRow, error) {
	return s.scanApproval(s.q(ctx).QueryRow(ctx, approvalSelect+` WHERE "OrganizationId"=$1 AND "Id"=$2`, orgID, id))
}

func (s *Store) FindOpenApproval(ctx context.Context, orgID, actionType, entityID string) (*ApprovalRequestRow, error) {
	return s.scanApproval(s.q(ctx).QueryRow(ctx, approvalSelect+`
		WHERE "OrganizationId"=$1 AND "ActionType"=$2 AND "EntityId"=$3
			AND "Status" IN ('pending_first','pending_second')
		ORDER BY "CreatedAt" DESC LIMIT 1`, orgID, actionType, entityID))
}

func (s *Store) ListApprovalRequests(ctx context.Context, orgID, status string, limit int) ([]ApprovalRequestRow, error) {
	if limit <= 0 {
		limit = 50
	}
	q := approvalSelect + ` WHERE "OrganizationId"=$1`
	args := []any{orgID}
	if status != "" {
		q += ` AND "Status"=$2`
		args = append(args, status)
	}
	q += ` ORDER BY "CreatedAt" DESC LIMIT ` + itoa(limit)
	rows, err := s.q(ctx).Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ApprovalRequestRow
	for rows.Next() {
		item, err := scanApprovalRow(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if items == nil {
		items = []ApprovalRequestRow{}
	}
	return items, rows.Err()
}

func (s *Store) UpdateApprovalRequest(ctx context.Context, r ApprovalRequestRow) error {
	_, err := s.q(ctx).Exec(ctx, `
		UPDATE "ApprovalRequests" SET
			"Status"=$2, "FirstApproverUserId"=$3, "SecondApproverUserId"=$4,
			"FirstApprovedAt"=$5, "SecondApprovedAt"=$6, "RejectedAt"=$7, "RejectedBy"=$8,
			"RejectionReason"=$9, "AfterSnapshot"=$10::jsonb, "ExecutedAt"=$11
		WHERE "Id"=$1`,
		r.ID, r.Status, r.FirstApproverUserID, r.SecondApproverUserID,
		r.FirstApprovedAt, r.SecondApprovedAt, r.RejectedAt, r.RejectedBy,
		r.RejectionReason, r.AfterJSON, r.ExecutedAt)
	return err
}

const approvalSelect = `
	SELECT "Id", "OrganizationId", "ActionType", "EntityType", "EntityId", "Status",
		"RequesterUserId", "FirstApproverUserId", "SecondApproverUserId", "Justification",
		"Payload"::text, "BeforeSnapshot"::text, "AfterSnapshot"::text, "RejectionReason", "RejectedBy",
		"CreatedAt", "FirstApprovedAt", "SecondApprovedAt", "RejectedAt", "ExecutedAt"
	FROM "ApprovalRequests"`

func (s *Store) scanApproval(row pgx.Row) (*ApprovalRequestRow, error) {
	item, err := scanApprovalRow(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

type approvalScanner interface {
	Scan(dest ...any) error
}

func scanApprovalRow(row approvalScanner) (ApprovalRequestRow, error) {
	var r ApprovalRequestRow
	err := row.Scan(&r.ID, &r.OrganizationID, &r.ActionType, &r.EntityType, &r.EntityID, &r.Status,
		&r.RequesterUserID, &r.FirstApproverUserID, &r.SecondApproverUserID, &r.Justification,
		&r.PayloadJSON, &r.BeforeJSON, &r.AfterJSON, &r.RejectionReason, &r.RejectedBy,
		&r.CreatedAt, &r.FirstApprovedAt, &r.SecondApprovedAt, &r.RejectedAt, &r.ExecutedAt)
	return r, err
}
