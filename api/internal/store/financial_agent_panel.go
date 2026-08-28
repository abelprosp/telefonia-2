package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

type FinancialAgentSettingsRow struct {
	OrganizationID    string
	Enabled           bool
	EvolutionAPIURL   string
	EvolutionAPIKey   string
	EvolutionInstance string
	N8nWebhookURL     string
	UpdatedAt         time.Time
	UpdatedBy         *string
}

type FinancialAgentEventRow struct {
	ID             string
	OrganizationID string
	EventType      string
	WhatsAppNumber *string
	CustomerID     *string
	CustomerName   *string
	InvoiceID      *string
	InvoiceNumber  *string
	Success        bool
	Summary        string
	CreatedAt      time.Time
}

type FinancialAgentEventStats struct {
	TotalToday     int
	SuccessToday   int
	FailedToday    int
	LookupsToday   int
	ReceiptsToday  int
	PaymentsToday  int
}

func (s *Store) GetFinancialAgentSettings(ctx context.Context, orgID string) (*FinancialAgentSettingsRow, error) {
	var row FinancialAgentSettingsRow
	err := s.q(ctx).QueryRow(ctx, `
		SELECT "OrganizationId", "Enabled", "EvolutionApiUrl", "EvolutionApiKey",
			"EvolutionInstance", "N8nWebhookUrl", "UpdatedAt", "UpdatedBy"
		FROM "FinancialAgentSettings" WHERE "OrganizationId" = $1`, orgID).
		Scan(&row.OrganizationID, &row.Enabled, &row.EvolutionAPIURL, &row.EvolutionAPIKey,
			&row.EvolutionInstance, &row.N8nWebhookURL, &row.UpdatedAt, &row.UpdatedBy)
	if errors.Is(err, pgx.ErrNoRows) {
		return &FinancialAgentSettingsRow{
			OrganizationID:    orgID,
			Enabled:           true,
			EvolutionInstance: "luxus",
		}, nil
	}
	return &row, err
}

func (s *Store) UpsertFinancialAgentSettings(ctx context.Context, row FinancialAgentSettingsRow) error {
	if row.EvolutionInstance == "" {
		row.EvolutionInstance = "luxus"
	}
	_, err := s.q(ctx).Exec(ctx, `
		INSERT INTO "FinancialAgentSettings" (
			"OrganizationId", "Enabled", "EvolutionApiUrl", "EvolutionApiKey",
			"EvolutionInstance", "N8nWebhookUrl", "UpdatedAt", "UpdatedBy"
		) VALUES ($1,$2,$3,$4,$5,$6,now(),$7)
		ON CONFLICT ("OrganizationId") DO UPDATE SET
			"Enabled" = EXCLUDED."Enabled",
			"EvolutionApiUrl" = EXCLUDED."EvolutionApiUrl",
			"EvolutionApiKey" = CASE
				WHEN EXCLUDED."EvolutionApiKey" = '' THEN "FinancialAgentSettings"."EvolutionApiKey"
				ELSE EXCLUDED."EvolutionApiKey"
			END,
			"EvolutionInstance" = EXCLUDED."EvolutionInstance",
			"N8nWebhookUrl" = EXCLUDED."N8nWebhookUrl",
			"UpdatedAt" = now(),
			"UpdatedBy" = EXCLUDED."UpdatedBy"`,
		row.OrganizationID, row.Enabled, row.EvolutionAPIURL, row.EvolutionAPIKey,
		row.EvolutionInstance, row.N8nWebhookURL, row.UpdatedBy)
	return err
}

func (s *Store) InsertFinancialAgentEvent(ctx context.Context, row FinancialAgentEventRow) error {
	_, err := s.q(ctx).Exec(ctx, `
		INSERT INTO "FinancialAgentEvents" (
			"Id", "OrganizationId", "EventType", "WhatsAppNumber", "CustomerId", "CustomerName",
			"InvoiceId", "InvoiceNumber", "Success", "Summary", "CreatedAt"
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		row.ID, row.OrganizationID, row.EventType, row.WhatsAppNumber, row.CustomerID, row.CustomerName,
		row.InvoiceID, row.InvoiceNumber, row.Success, row.Summary, row.CreatedAt)
	return err
}

func (s *Store) ListFinancialAgentEvents(ctx context.Context, orgID string, limit int) ([]FinancialAgentEventRow, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	rows, err := s.q(ctx).Query(ctx, `
		SELECT "Id", "OrganizationId", "EventType", "WhatsAppNumber", "CustomerId", "CustomerName",
			"InvoiceId", "InvoiceNumber", "Success", "Summary", "CreatedAt"
		FROM "FinancialAgentEvents"
		WHERE "OrganizationId" = $1
		ORDER BY "CreatedAt" DESC
		LIMIT $2`, orgID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []FinancialAgentEventRow
	for rows.Next() {
		var item FinancialAgentEventRow
		if err := rows.Scan(&item.ID, &item.OrganizationID, &item.EventType, &item.WhatsAppNumber,
			&item.CustomerID, &item.CustomerName, &item.InvoiceID, &item.InvoiceNumber,
			&item.Success, &item.Summary, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if items == nil {
		items = []FinancialAgentEventRow{}
	}
	return items, rows.Err()
}

func (s *Store) CountFinancialAgentEventStats(ctx context.Context, orgID string) (FinancialAgentEventStats, error) {
	var st FinancialAgentEventStats
	err := s.q(ctx).QueryRow(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE "CreatedAt" >= date_trunc('day', now()))::int,
			COUNT(*) FILTER (WHERE "CreatedAt" >= date_trunc('day', now()) AND "Success")::int,
			COUNT(*) FILTER (WHERE "CreatedAt" >= date_trunc('day', now()) AND NOT "Success")::int,
			COUNT(*) FILTER (WHERE "CreatedAt" >= date_trunc('day', now()) AND "EventType" = 'lookup')::int,
			COUNT(*) FILTER (WHERE "CreatedAt" >= date_trunc('day', now()) AND "EventType" = 'verify_receipt')::int,
			COUNT(*) FILTER (WHERE "CreatedAt" >= date_trunc('day', now()) AND "EventType" = 'confirm_payment')::int
		FROM "FinancialAgentEvents"
		WHERE "OrganizationId" = $1`, orgID).
		Scan(&st.TotalToday, &st.SuccessToday, &st.FailedToday, &st.LookupsToday, &st.ReceiptsToday, &st.PaymentsToday)
	return st, err
}
