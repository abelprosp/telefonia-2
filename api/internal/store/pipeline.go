package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

type ProcessingMonthRunRow struct {
	ID                string
	OrganizationID    string
	ProcessingMonthID string
	Version           int
	Status            string
	TriggeredBy       *string
	SummaryJSON       *string
	CreatedAt         time.Time
	CompletedAt       *time.Time
}

type ProcessingMonthRunStepRow struct {
	ID         string
	RunID      string
	StepKey    string
	StepOrder  int
	Label      string
	Status     string
	StartedAt  *time.Time
	CompletedAt *time.Time
	DurationMs *int
	Error      *string
	SummaryJSON *string
}

func (s *Store) NextProcessingMonthRunVersion(ctx context.Context, monthID string) (int, error) {
	var n int
	err := s.q(ctx).QueryRow(ctx, `SELECT COALESCE(MAX("Version"), 0) + 1 FROM "ProcessingMonthRuns" WHERE "ProcessingMonthId"=$1`, monthID).Scan(&n)
	return n, err
}

func (s *Store) InsertProcessingMonthRun(ctx context.Context, r ProcessingMonthRunRow) error {
	_, err := s.q(ctx).Exec(ctx, `
		INSERT INTO "ProcessingMonthRuns" ("Id","OrganizationId","ProcessingMonthId","Version","Status","TriggeredBy","CreatedAt")
		VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		r.ID, r.OrganizationID, r.ProcessingMonthID, r.Version, r.Status, r.TriggeredBy, r.CreatedAt)
	return err
}

func (s *Store) UpdateProcessingMonthRun(ctx context.Context, id, status string, summary *string, completedAt *time.Time) error {
	_, err := s.q(ctx).Exec(ctx, `
		UPDATE "ProcessingMonthRuns" SET "Status"=$2, "Summary"=$3::jsonb, "CompletedAt"=$4 WHERE "Id"=$1`,
		id, status, summary, completedAt)
	return err
}

func (s *Store) InsertProcessingMonthRunStep(ctx context.Context, st ProcessingMonthRunStepRow) error {
	_, err := s.q(ctx).Exec(ctx, `
		INSERT INTO "ProcessingMonthRunSteps" ("Id","RunId","StepKey","StepOrder","Label","Status")
		VALUES ($1,$2,$3,$4,$5,$6)`,
		st.ID, st.RunID, st.StepKey, st.StepOrder, st.Label, st.Status)
	return err
}

func (s *Store) UpdateProcessingMonthRunStep(ctx context.Context, st ProcessingMonthRunStepRow) error {
	_, err := s.q(ctx).Exec(ctx, `
		UPDATE "ProcessingMonthRunSteps" SET
			"Status"=$2, "StartedAt"=$3, "CompletedAt"=$4, "DurationMs"=$5, "Error"=$6, "Summary"=$7::jsonb
		WHERE "Id"=$1`,
		st.ID, st.Status, st.StartedAt, st.CompletedAt, st.DurationMs, st.Error, st.SummaryJSON)
	return err
}

func (s *Store) GetProcessingMonthRun(ctx context.Context, orgID, id string) (*ProcessingMonthRunRow, error) {
	var r ProcessingMonthRunRow
	err := s.q(ctx).QueryRow(ctx, `
		SELECT "Id","OrganizationId","ProcessingMonthId","Version","Status","TriggeredBy","Summary"::text,"CreatedAt","CompletedAt"
		FROM "ProcessingMonthRuns" WHERE "OrganizationId"=$1 AND "Id"=$2`, orgID, id).
		Scan(&r.ID, &r.OrganizationID, &r.ProcessingMonthID, &r.Version, &r.Status, &r.TriggeredBy, &r.SummaryJSON, &r.CreatedAt, &r.CompletedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &r, err
}

func (s *Store) ListProcessingMonthRuns(ctx context.Context, orgID, monthID string) ([]ProcessingMonthRunRow, error) {
	rows, err := s.q(ctx).Query(ctx, `
		SELECT "Id","OrganizationId","ProcessingMonthId","Version","Status","TriggeredBy","Summary"::text,"CreatedAt","CompletedAt"
		FROM "ProcessingMonthRuns"
		WHERE "OrganizationId"=$1 AND "ProcessingMonthId"=$2
		ORDER BY "Version" DESC`, orgID, monthID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ProcessingMonthRunRow
	for rows.Next() {
		var r ProcessingMonthRunRow
		if err := rows.Scan(&r.ID, &r.OrganizationID, &r.ProcessingMonthID, &r.Version, &r.Status, &r.TriggeredBy, &r.SummaryJSON, &r.CreatedAt, &r.CompletedAt); err != nil {
			return nil, err
		}
		items = append(items, r)
	}
	if items == nil {
		items = []ProcessingMonthRunRow{}
	}
	return items, rows.Err()
}

func (s *Store) ListProcessingMonthRunSteps(ctx context.Context, runID string) ([]ProcessingMonthRunStepRow, error) {
	rows, err := s.q(ctx).Query(ctx, `
		SELECT "Id","RunId","StepKey","StepOrder","Label","Status","StartedAt","CompletedAt","DurationMs","Error","Summary"::text
		FROM "ProcessingMonthRunSteps" WHERE "RunId"=$1 ORDER BY "StepOrder"`, runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ProcessingMonthRunStepRow
	for rows.Next() {
		var st ProcessingMonthRunStepRow
		if err := rows.Scan(&st.ID, &st.RunID, &st.StepKey, &st.StepOrder, &st.Label, &st.Status, &st.StartedAt, &st.CompletedAt, &st.DurationMs, &st.Error, &st.SummaryJSON); err != nil {
			return nil, err
		}
		items = append(items, st)
	}
	if items == nil {
		items = []ProcessingMonthRunStepRow{}
	}
	return items, rows.Err()
}

type PortalCustomerLinkRow struct {
	ID             string
	OrganizationID string
	UserID         string
	CustomerID     string
	Document       string
	CreatedAt      time.Time
}

func (s *Store) GetPortalLinkByUser(ctx context.Context, orgID, userID string) (*PortalCustomerLinkRow, error) {
	var r PortalCustomerLinkRow
	err := s.q(ctx).QueryRow(ctx, `
		SELECT "Id","OrganizationId","UserId","CustomerId","Document","CreatedAt"
		FROM "PortalCustomerLinks" WHERE "OrganizationId"=$1 AND "UserId"=$2`, orgID, userID).
		Scan(&r.ID, &r.OrganizationID, &r.UserID, &r.CustomerID, &r.Document, &r.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &r, err
}

func (s *Store) UpsertPortalCustomerLink(ctx context.Context, r PortalCustomerLinkRow) error {
	_, err := s.q(ctx).Exec(ctx, `
		INSERT INTO "PortalCustomerLinks" ("Id","OrganizationId","UserId","CustomerId","Document","CreatedAt")
		VALUES ($1,$2,$3,$4,$5,$6)
		ON CONFLICT ("OrganizationId","UserId") DO UPDATE SET "CustomerId"=EXCLUDED."CustomerId", "Document"=EXCLUDED."Document"`,
		r.ID, r.OrganizationID, r.UserID, r.CustomerID, r.Document, r.CreatedAt)
	return err
}

func (s *Store) InsertOperationMetric(ctx context.Context, id, orgID, operation string, durationMs int, success bool, meta *string) error {
	_, err := s.q(ctx).Exec(ctx, `
		INSERT INTO "OperationMetrics" ("Id","OrganizationId","Operation","DurationMs","Success","Metadata","CreatedAt")
		VALUES ($1,$2,$3,$4,$5,$6::jsonb,now())`, id, orgID, operation, durationMs, success, meta)
	return err
}
