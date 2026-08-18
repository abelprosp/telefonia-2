package store

import (
	"context"
	"time"
)

type StateTransitionLogRow struct {
	ID            string
	OrganizationID string
	EntityType    string
	EntityID      string
	FromState     string
	ToState       string
	TriggerEvent  string
	Justification *string
	ActorUserID   *string
	MetadataJSON  *string
	CreatedAt     time.Time
}

func (s *Store) InsertStateTransitionLog(ctx context.Context, id, orgID, entityType, entityID, fromState, toState, triggerEvent string, justification, actorUserID *string, metadataJSON *string, ts time.Time) error {
	_, err := s.q(ctx).Exec(ctx, `
		INSERT INTO "StateTransitionLogs" (
			"Id", "OrganizationId", "EntityType", "EntityId", "FromState", "ToState", "TriggerEvent", "Justification", "ActorUserId", "Metadata", "CreatedAt"
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10::jsonb, $11)`,
		id, orgID, entityType, entityID, fromState, toState, triggerEvent, justification, actorUserID, metadataJSON, ts)
	return err
}

func (s *Store) ListStateTransitionLogs(ctx context.Context, orgID, entityType, entityID string, limit int) ([]StateTransitionLogRow, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	rows, err := s.q(ctx).Query(ctx, `
		SELECT "Id", "OrganizationId", "EntityType", "EntityId", "FromState", "ToState", "TriggerEvent", "Justification", "ActorUserId", "Metadata"::text, "CreatedAt"
		FROM "StateTransitionLogs"
		WHERE "OrganizationId" = $1 AND "EntityType" = $2 AND "EntityId" = $3
		ORDER BY "CreatedAt" DESC
		LIMIT $4`, orgID, entityType, entityID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []StateTransitionLogRow
	for rows.Next() {
		var item StateTransitionLogRow
		if err := rows.Scan(
			&item.ID, &item.OrganizationID, &item.EntityType, &item.EntityID,
			&item.FromState, &item.ToState, &item.TriggerEvent, &item.Justification,
			&item.ActorUserID, &item.MetadataJSON, &item.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if items == nil {
		items = []StateTransitionLogRow{}
	}
	return items, rows.Err()
}
