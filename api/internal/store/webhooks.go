package store

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

type WebhookSubscriptionRow struct {
	ID             string
	OrganizationID string
	URL            string
	Events         []string
	Secret         string
	IsActive       bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (s *Store) ListWebhookSubscriptions(ctx context.Context, orgID string) ([]WebhookSubscriptionRow, error) {
	rows, err := s.q(ctx).Query(ctx, `
		SELECT "Id", "OrganizationId", "Url", COALESCE("Events"::text, '[]'), "Secret", "IsActive", "CreatedAt", "UpdatedAt"
		FROM "WebhookSubscriptions"
		WHERE "OrganizationId" = $1
		ORDER BY "CreatedAt" DESC`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []WebhookSubscriptionRow
	for rows.Next() {
		var r WebhookSubscriptionRow
		var eventsJSON string
		if err := rows.Scan(&r.ID, &r.OrganizationID, &r.URL, &eventsJSON, &r.Secret, &r.IsActive, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal([]byte(eventsJSON), &r.Events)
		if r.Events == nil {
			r.Events = []string{}
		}
		items = append(items, r)
	}
	if items == nil {
		items = []WebhookSubscriptionRow{}
	}
	return items, rows.Err()
}

func (s *Store) GetWebhookSubscription(ctx context.Context, orgID, id string) (*WebhookSubscriptionRow, error) {
	var r WebhookSubscriptionRow
	var eventsJSON string
	err := s.q(ctx).QueryRow(ctx, `
		SELECT "Id", "OrganizationId", "Url", COALESCE("Events"::text, '[]'), "Secret", "IsActive", "CreatedAt", "UpdatedAt"
		FROM "WebhookSubscriptions"
		WHERE "OrganizationId" = $1 AND "Id" = $2`, orgID, id).
		Scan(&r.ID, &r.OrganizationID, &r.URL, &eventsJSON, &r.Secret, &r.IsActive, &r.CreatedAt, &r.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	_ = json.Unmarshal([]byte(eventsJSON), &r.Events)
	if r.Events == nil {
		r.Events = []string{}
	}
	return &r, nil
}

func (s *Store) InsertWebhookSubscription(ctx context.Context, r WebhookSubscriptionRow) error {
	events, err := json.Marshal(r.Events)
	if err != nil {
		return err
	}
	_, err = s.q(ctx).Exec(ctx, `
		INSERT INTO "WebhookSubscriptions" (
			"Id", "OrganizationId", "Url", "Events", "Secret", "IsActive", "CreatedAt", "UpdatedAt"
		) VALUES ($1, $2, $3, $4::jsonb, $5, $6, $7, $8)`,
		r.ID, r.OrganizationID, r.URL, string(events), r.Secret, r.IsActive, r.CreatedAt, r.UpdatedAt)
	return err
}

func (s *Store) DeleteWebhookSubscription(ctx context.Context, orgID, id string) error {
	tag, err := s.q(ctx).Exec(ctx, `
		DELETE FROM "WebhookSubscriptions" WHERE "OrganizationId" = $1 AND "Id" = $2`, orgID, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (s *Store) ListActiveWebhooksForEvent(ctx context.Context, orgID, event string) ([]WebhookSubscriptionRow, error) {
	rows, err := s.q(ctx).Query(ctx, `
		SELECT "Id", "OrganizationId", "Url", COALESCE("Events"::text, '[]'), "Secret", "IsActive", "CreatedAt", "UpdatedAt"
		FROM "WebhookSubscriptions"
		WHERE "OrganizationId" = $1 AND "IsActive" = true
			AND "Events" @> jsonb_build_array($2::text)`, orgID, event)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []WebhookSubscriptionRow
	for rows.Next() {
		var r WebhookSubscriptionRow
		var eventsJSON string
		if err := rows.Scan(&r.ID, &r.OrganizationID, &r.URL, &eventsJSON, &r.Secret, &r.IsActive, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal([]byte(eventsJSON), &r.Events)
		items = append(items, r)
	}
	if items == nil {
		items = []WebhookSubscriptionRow{}
	}
	return items, rows.Err()
}

func (s *Store) InsertWebhookDelivery(ctx context.Context, id, subscriptionID, eventType string, statusCode *int, success bool, errMsg *string, ts time.Time) error {
	_, err := s.q(ctx).Exec(ctx, `
		INSERT INTO "WebhookDeliveries" ("Id", "SubscriptionId", "EventType", "StatusCode", "Success", "Error", "CreatedAt")
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		id, subscriptionID, eventType, statusCode, success, errMsg, ts)
	return err
}

func (s *Store) InsertOrganizationDataExport(ctx context.Context, id, orgID, exportedBy, checksum string, payloadBytes int, summaryJSON string, ts time.Time) error {
	_, err := s.q(ctx).Exec(ctx, `
		INSERT INTO "OrganizationDataExports" (
			"Id", "OrganizationId", "ExportedBy", "ChecksumSHA256", "PayloadBytes", "Summary", "CreatedAt"
		) VALUES ($1, $2, $3, $4, $5, $6::jsonb, $7)`,
		id, orgID, exportedBy, checksum, payloadBytes, summaryJSON, ts)
	return err
}
