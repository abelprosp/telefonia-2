package store

import (
	"context"
	"time"
)

func (s *Store) InsertAuditLog(ctx context.Context, id, changeType, entityName, keyValues string, changedBy *string, oldValues, newValues *string, ts time.Time) error {
	_, err := s.q(ctx).Exec(ctx, `
		INSERT INTO "AuditLogs" (
			"Id", "ChangeType", "EntityName", "KeyValues", "ChangedBy", "OldValues", "NewValues", "Timestamp"
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		id, changeType, entityName, keyValues, changedBy, oldValues, newValues, ts)
	return err
}

func (s *Store) ListAuditLogsForEntity(ctx context.Context, entityName, keyValues string, limit int) ([]AuditLogRow, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	rows, err := s.q(ctx).Query(ctx, `
		SELECT "Id", "ChangeType", "EntityName", "KeyValues", "ChangedBy", "OldValues", "NewValues", "Timestamp"
		FROM "AuditLogs"
		WHERE "EntityName" = $1 AND "KeyValues" = $2
		ORDER BY "Timestamp" DESC
		LIMIT $3`, entityName, keyValues, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []AuditLogRow
	for rows.Next() {
		var item AuditLogRow
		if err := rows.Scan(&item.ID, &item.ChangeType, &item.EntityName, &item.KeyValues,
			&item.ChangedBy, &item.OldValues, &item.NewValues, &item.Timestamp); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if items == nil {
		items = []AuditLogRow{}
	}
	return items, rows.Err()
}

type AuditLogRow struct {
	ID         string
	ChangeType string
	EntityName string
	KeyValues  string
	ChangedBy  *string
	OldValues  *string
	NewValues  *string
	Timestamp  time.Time
}
