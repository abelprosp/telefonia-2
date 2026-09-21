package store

import (
	"context"
	"encoding/json"
	"time"

	"github.com/luxus-connect/telefonia/api/internal/httputil"
)

type DomainAuditEventRow struct {
	ID             string
	OrganizationID string
	EntityType     string
	EntityID       string
	Action         string
	ActorUserID    *string
	ActorKind      string
	Source         string
	CorrelationID  *string
	BeforeJSON     *string
	AfterJSON      *string
	MetadataJSON   *string
	CreatedAt      time.Time
}

type DomainAuditFilter struct {
	EntityType    string
	EntityID      string
	Action        string
	ActorUserID   string
	From          *time.Time
	To            *time.Time
	PhoneLineID   string // resolved via entity when EntityType=phone_line
	CustomerID    string
}

func (s *Store) InsertDomainAuditEvent(ctx context.Context, row DomainAuditEventRow) error {
	if row.ActorKind == "" {
		row.ActorKind = "user"
	}
	if row.Source == "" {
		row.Source = "api"
	}
	if row.CreatedAt.IsZero() {
		row.CreatedAt = time.Now().UTC()
	}
	_, err := s.q(ctx).Exec(ctx, `
		INSERT INTO "DomainAuditEvents" (
			"Id", "OrganizationId", "EntityType", "EntityId", "Action",
			"ActorUserId", "ActorKind", "Source", "CorrelationId",
			"BeforeJson", "AfterJson", "MetadataJson", "CreatedAt"
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10::jsonb,$11::jsonb,$12::jsonb,$13)`,
		row.ID, row.OrganizationID, row.EntityType, row.EntityID, row.Action,
		row.ActorUserID, row.ActorKind, row.Source, row.CorrelationID,
		row.BeforeJSON, row.AfterJSON, row.MetadataJSON, row.CreatedAt)
	return err
}

func (s *Store) ListDomainAuditEvents(ctx context.Context, orgID string, f DomainAuditFilter, page httputil.PageSearch) ([]DomainAuditEventRow, int64, error) {
	base := `FROM "DomainAuditEvents" WHERE "OrganizationId" = $1`
	args := []any{orgID}
	if f.EntityType != "" {
		args = append(args, f.EntityType)
		base += ` AND "EntityType" = $` + itoa(len(args))
	}
	if f.EntityID != "" {
		args = append(args, f.EntityID)
		base += ` AND "EntityId" = $` + itoa(len(args))
	}
	if f.Action != "" {
		args = append(args, f.Action)
		base += ` AND "Action" = $` + itoa(len(args))
	}
	if f.ActorUserID != "" {
		args = append(args, f.ActorUserID)
		base += ` AND "ActorUserId" = $` + itoa(len(args))
	}
	if f.From != nil {
		args = append(args, *f.From)
		base += ` AND "CreatedAt" >= $` + itoa(len(args))
	}
	if f.To != nil {
		args = append(args, *f.To)
		base += ` AND "CreatedAt" <= $` + itoa(len(args))
	}
	if f.PhoneLineID != "" {
		args = append(args, f.PhoneLineID)
		base += ` AND (("EntityType" = 'phone_line' AND "EntityId" = $` + itoa(len(args)) + `)
			OR ("MetadataJson"->>'phone_line_id') = $` + itoa(len(args)) + `)`
	}
	if f.CustomerID != "" {
		args = append(args, f.CustomerID)
		base += ` AND (("EntityType" = 'customer' AND "EntityId" = $` + itoa(len(args)) + `)
			OR ("MetadataJson"->>'customer_id') = $` + itoa(len(args)) + `)`
	}

	var total int64
	if err := s.q(ctx).QueryRow(ctx, `SELECT COUNT(*) `+base, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	selectQ := `
		SELECT "Id", "OrganizationId", "EntityType", "EntityId", "Action",
			"ActorUserId", "ActorKind", "Source", "CorrelationId",
			"BeforeJson"::text, "AfterJson"::text, "MetadataJson"::text, "CreatedAt"
		` + base + `
		ORDER BY "CreatedAt" DESC
		OFFSET $` + itoa(len(args)+1) + ` LIMIT $` + itoa(len(args)+2)
	args = append(args, page.Offset(), page.Limit())

	rows, err := s.q(ctx).Query(ctx, selectQ, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []DomainAuditEventRow
	for rows.Next() {
		var item DomainAuditEventRow
		if err := rows.Scan(
			&item.ID, &item.OrganizationID, &item.EntityType, &item.EntityID, &item.Action,
			&item.ActorUserID, &item.ActorKind, &item.Source, &item.CorrelationID,
			&item.BeforeJSON, &item.AfterJSON, &item.MetadataJSON, &item.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	if items == nil {
		items = []DomainAuditEventRow{}
	}
	return items, total, rows.Err()
}

func AuditJSON(v any) *string {
	if v == nil {
		return nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	s := string(b)
	return &s
}
