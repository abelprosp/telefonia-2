package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/luxus-connect/telefonia/api/internal/httputil"
)

type BillingDivergenceRow struct {
	ID                string
	OrganizationID    string
	ProcessingMonthID *string
	DivergenceType    string
	Severity          string
	Competence        *string
	OperatorName      *string
	AccountNumber     *string
	CustomerID        *string
	PhoneLineID       *string
	PhoneNumber       *string
	OwnerUserID       *string
	FinancialImpact   float64
	Status            string
	Cause             *string
	RecommendedAction *string
	Evidence          *string
	ResolutionNotes   *string
	ResolvedBy        *string
	CreatedAt         time.Time
	UpdatedAt         time.Time
	ResolvedAt        *time.Time
}

type BillingDivergenceCommentRow struct {
	ID           string
	DivergenceID string
	AuthorUserID *string
	Body         string
	CreatedAt    time.Time
}

type BillingDivergenceHistoryRow struct {
	ID           string
	DivergenceID string
	ActorUserID  *string
	EventType    string
	FromValue    *string
	ToValue      *string
	Notes        *string
	CreatedAt    time.Time
}

func (s *Store) InsertBillingDivergence(ctx context.Context, r BillingDivergenceRow) error {
	_, err := s.q(ctx).Exec(ctx, `
		INSERT INTO "BillingDivergences" (
			"Id","OrganizationId","ProcessingMonthId","DivergenceType","Severity","Competence",
			"OperatorName","AccountNumber","CustomerId","PhoneLineId","PhoneNumber","OwnerUserId",
			"FinancialImpact","Status","Cause","RecommendedAction","Evidence","CreatedAt","UpdatedAt"
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19)`,
		r.ID, r.OrganizationID, r.ProcessingMonthID, r.DivergenceType, r.Severity, r.Competence,
		r.OperatorName, r.AccountNumber, r.CustomerID, r.PhoneLineID, r.PhoneNumber, r.OwnerUserID,
		r.FinancialImpact, r.Status, r.Cause, r.RecommendedAction, r.Evidence, r.CreatedAt, r.UpdatedAt)
	return err
}

func (s *Store) GetBillingDivergence(ctx context.Context, orgID, id string) (*BillingDivergenceRow, error) {
	item, err := scanDivergence(s.q(ctx).QueryRow(ctx, divergenceSelect+` WHERE d."OrganizationId"=$1 AND d."Id"=$2`, orgID, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *Store) UpdateBillingDivergence(ctx context.Context, r BillingDivergenceRow) error {
	_, err := s.q(ctx).Exec(ctx, `
		UPDATE "BillingDivergences" SET
			"Status"=$2, "OwnerUserId"=$3, "Cause"=$4, "RecommendedAction"=$5, "Evidence"=$6,
			"ResolutionNotes"=$7, "ResolvedBy"=$8, "UpdatedAt"=$9, "ResolvedAt"=$10, "FinancialImpact"=$11
		WHERE "Id"=$1 AND "OrganizationId"=$12`,
		r.ID, r.Status, r.OwnerUserID, r.Cause, r.RecommendedAction, r.Evidence,
		r.ResolutionNotes, r.ResolvedBy, r.UpdatedAt, r.ResolvedAt, r.FinancialImpact, r.OrganizationID)
	return err
}

func (s *Store) ListBillingDivergences(ctx context.Context, orgID string, monthID, status string, page httputil.PageSearch) ([]BillingDivergenceRow, int64, error) {
	where := ` WHERE d."OrganizationId"=$1`
	args := []any{orgID}
	n := 2
	if monthID != "" {
		where += ` AND d."ProcessingMonthId"=$` + itoa(n)
		args = append(args, monthID)
		n++
	}
	if status != "" {
		where += ` AND d."Status"=$` + itoa(n)
		args = append(args, status)
		n++
	}
	var total int64
	if err := s.q(ctx).QueryRow(ctx, `SELECT COUNT(*) FROM "BillingDivergences" d`+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	args = append(args, page.Offset(), page.Limit())
	rows, err := s.q(ctx).Query(ctx, divergenceSelect+where+` ORDER BY d."CreatedAt" DESC OFFSET $`+itoa(n)+` LIMIT $`+itoa(n+1), args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []BillingDivergenceRow
	for rows.Next() {
		item, err := scanDivergence(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	if items == nil {
		items = []BillingDivergenceRow{}
	}
	return items, total, rows.Err()
}

func (s *Store) CountOpenDivergences(ctx context.Context, orgID string) (int, error) {
	var n int
	err := s.q(ctx).QueryRow(ctx, `
		SELECT COUNT(*)::int FROM "BillingDivergences"
		WHERE "OrganizationId" = $1 AND "Status" IN ('open', 'pending')`, orgID).Scan(&n)
	return n, err
}

func (s *Store) InsertDivergenceComment(ctx context.Context, c BillingDivergenceCommentRow) error {
	_, err := s.q(ctx).Exec(ctx, `
		INSERT INTO "BillingDivergenceComments" ("Id","DivergenceId","AuthorUserId","Body","CreatedAt")
		VALUES ($1,$2,$3,$4,$5)`, c.ID, c.DivergenceID, c.AuthorUserID, c.Body, c.CreatedAt)
	return err
}

func (s *Store) ListDivergenceComments(ctx context.Context, id string) ([]BillingDivergenceCommentRow, error) {
	rows, err := s.q(ctx).Query(ctx, `
		SELECT "Id","DivergenceId","AuthorUserId","Body","CreatedAt"
		FROM "BillingDivergenceComments" WHERE "DivergenceId"=$1 ORDER BY "CreatedAt"`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []BillingDivergenceCommentRow
	for rows.Next() {
		var c BillingDivergenceCommentRow
		if err := rows.Scan(&c.ID, &c.DivergenceID, &c.AuthorUserID, &c.Body, &c.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, c)
	}
	if items == nil {
		items = []BillingDivergenceCommentRow{}
	}
	return items, rows.Err()
}

func (s *Store) InsertDivergenceHistory(ctx context.Context, h BillingDivergenceHistoryRow) error {
	_, err := s.q(ctx).Exec(ctx, `
		INSERT INTO "BillingDivergenceHistory" ("Id","DivergenceId","ActorUserId","EventType","FromValue","ToValue","Notes","CreatedAt")
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		h.ID, h.DivergenceID, h.ActorUserID, h.EventType, h.FromValue, h.ToValue, h.Notes, h.CreatedAt)
	return err
}

func (s *Store) ListDivergenceHistory(ctx context.Context, id string) ([]BillingDivergenceHistoryRow, error) {
	rows, err := s.q(ctx).Query(ctx, `
		SELECT "Id","DivergenceId","ActorUserId","EventType","FromValue","ToValue","Notes","CreatedAt"
		FROM "BillingDivergenceHistory" WHERE "DivergenceId"=$1 ORDER BY "CreatedAt"`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []BillingDivergenceHistoryRow
	for rows.Next() {
		var h BillingDivergenceHistoryRow
		if err := rows.Scan(&h.ID, &h.DivergenceID, &h.ActorUserID, &h.EventType, &h.FromValue, &h.ToValue, &h.Notes, &h.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, h)
	}
	if items == nil {
		items = []BillingDivergenceHistoryRow{}
	}
	return items, rows.Err()
}

const divergenceSelect = `
	SELECT d."Id", d."OrganizationId", d."ProcessingMonthId", d."DivergenceType", d."Severity",
		d."Competence", d."OperatorName", d."AccountNumber", d."CustomerId", d."PhoneLineId",
		d."PhoneNumber", d."OwnerUserId", d."FinancialImpact", d."Status", d."Cause",
		d."RecommendedAction", d."Evidence", d."ResolutionNotes", d."ResolvedBy",
		d."CreatedAt", d."UpdatedAt", d."ResolvedAt"
	FROM "BillingDivergences" d`

func scanDivergence(row approvalScanner) (BillingDivergenceRow, error) {
	var r BillingDivergenceRow
	err := row.Scan(&r.ID, &r.OrganizationID, &r.ProcessingMonthID, &r.DivergenceType, &r.Severity,
		&r.Competence, &r.OperatorName, &r.AccountNumber, &r.CustomerID, &r.PhoneLineID,
		&r.PhoneNumber, &r.OwnerUserID, &r.FinancialImpact, &r.Status, &r.Cause,
		&r.RecommendedAction, &r.Evidence, &r.ResolutionNotes, &r.ResolvedBy,
		&r.CreatedAt, &r.UpdatedAt, &r.ResolvedAt)
	return r, err
}
