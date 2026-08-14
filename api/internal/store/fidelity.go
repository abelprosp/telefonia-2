package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/luxus-connect/telefonia/api/internal/models"
)

type LineFidelityRow struct {
	ID                  string
	PhoneLineID         string
	StartDate           time.Time
	InitialMonths       int
	PredictedEndDate    time.Time
	AutoRenew           bool
	RenewalPeriodMonths *int
	Status              string
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

func PredictedFidelityEnd(start time.Time, months int) time.Time {
	return start.AddDate(0, months, 0)
}

func (s *Store) GetLineFidelity(ctx context.Context, phoneLineID string) (*LineFidelityRow, error) {
	var row LineFidelityRow
	err := s.q(ctx).QueryRow(ctx, `
		SELECT "Id", "PhoneLineId", "StartDate", "InitialMonths", "PredictedEndDate",
			"AutoRenew", "RenewalPeriodMonths", "Status", "CreatedAt", "UpdatedAt"
		FROM "LineFidelities" WHERE "PhoneLineId" = $1`, phoneLineID).
		Scan(&row.ID, &row.PhoneLineID, &row.StartDate, &row.InitialMonths, &row.PredictedEndDate,
			&row.AutoRenew, &row.RenewalPeriodMonths, &row.Status, &row.CreatedAt, &row.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return &row, err
}

func (s *Store) UpsertLineFidelity(ctx context.Context, row LineFidelityRow) error {
	_, err := s.q(ctx).Exec(ctx, `
		INSERT INTO "LineFidelities" (
			"Id", "PhoneLineId", "StartDate", "InitialMonths", "PredictedEndDate",
			"AutoRenew", "RenewalPeriodMonths", "Status", "CreatedAt", "UpdatedAt"
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT ("PhoneLineId") DO UPDATE SET
			"StartDate" = EXCLUDED."StartDate",
			"InitialMonths" = EXCLUDED."InitialMonths",
			"PredictedEndDate" = EXCLUDED."PredictedEndDate",
			"AutoRenew" = EXCLUDED."AutoRenew",
			"RenewalPeriodMonths" = EXCLUDED."RenewalPeriodMonths",
			"Status" = EXCLUDED."Status",
			"UpdatedAt" = EXCLUDED."UpdatedAt"`,
		row.ID, row.PhoneLineID, row.StartDate, row.InitialMonths, row.PredictedEndDate,
		row.AutoRenew, row.RenewalPeriodMonths, row.Status, row.CreatedAt, row.UpdatedAt)
	return err
}

func (s *Store) UpdateLineFidelityDates(ctx context.Context, id string, predictedEnd time.Time, status string, now time.Time) error {
	_, err := s.q(ctx).Exec(ctx, `
		UPDATE "LineFidelities"
		SET "PredictedEndDate" = $2, "Status" = $3, "UpdatedAt" = $4
		WHERE "Id" = $1`, id, predictedEnd, status, now)
	return err
}

func (s *Store) InsertLineFidelityEvent(ctx context.Context, id, fidelityID, eventType string, occurredAt time.Time, userID, notes *string) error {
	_, err := s.q(ctx).Exec(ctx, `
		INSERT INTO "LineFidelityEvents" ("Id", "FidelityId", "EventType", "OccurredAt", "UserId", "Notes")
		VALUES ($1, $2, $3, $4, $5, $6)`, id, fidelityID, eventType, occurredAt, userID, notes)
	return err
}

func (s *Store) ListLineFidelityEvents(ctx context.Context, fidelityID string) ([]models.LineFidelityEventResponse, error) {
	rows, err := s.q(ctx).Query(ctx, `
		SELECT "Id", "EventType", "OccurredAt", "UserId", "Notes"
		FROM "LineFidelityEvents"
		WHERE "FidelityId" = $1
		ORDER BY "OccurredAt" DESC`, fidelityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []models.LineFidelityEventResponse
	for rows.Next() {
		var item models.LineFidelityEventResponse
		if err := rows.Scan(&item.ID, &item.EventType, &item.OccurredAt, &item.UserID, &item.Notes); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if items == nil {
		items = []models.LineFidelityEventResponse{}
	}
	return items, rows.Err()
}

func (s *Store) ListDueAutoRenewFidelities(ctx context.Context, orgID string, asOf time.Time) ([]LineFidelityRow, error) {
	rows, err := s.q(ctx).Query(ctx, `
		SELECT f."Id", f."PhoneLineId", f."StartDate", f."InitialMonths", f."PredictedEndDate",
			f."AutoRenew", f."RenewalPeriodMonths", f."Status", f."CreatedAt", f."UpdatedAt"
		FROM "LineFidelities" f
		JOIN "PhoneLines" pl ON pl."Id" = f."PhoneLineId"
		JOIN "ProviderAccounts" pa ON pa."Id" = pl."ProviderAccountId"
		JOIN "ContractingCompanies" cc ON cc."Id" = pa."ContractingCompanyId"
		JOIN "Providers" p ON p."Id" = cc."ProviderId"
		WHERE p."OrganizationId" = $1
			AND f."AutoRenew" = true
			AND f."RenewalPeriodMonths" IS NOT NULL
			AND f."PredictedEndDate" <= $2::date
			AND f."Status" = 'active'`, orgID, asOf)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []LineFidelityRow
	for rows.Next() {
		var row LineFidelityRow
		if err := rows.Scan(&row.ID, &row.PhoneLineID, &row.StartDate, &row.InitialMonths, &row.PredictedEndDate,
			&row.AutoRenew, &row.RenewalPeriodMonths, &row.Status, &row.CreatedAt, &row.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, row)
	}
	return items, rows.Err()
}

func (s *Store) ListExpiredFidelitiesToMark(ctx context.Context, orgID string, asOf time.Time) ([]LineFidelityRow, error) {
	rows, err := s.q(ctx).Query(ctx, `
		SELECT f."Id", f."PhoneLineId", f."StartDate", f."InitialMonths", f."PredictedEndDate",
			f."AutoRenew", f."RenewalPeriodMonths", f."Status", f."CreatedAt", f."UpdatedAt"
		FROM "LineFidelities" f
		JOIN "PhoneLines" pl ON pl."Id" = f."PhoneLineId"
		JOIN "ProviderAccounts" pa ON pa."Id" = pl."ProviderAccountId"
		JOIN "ContractingCompanies" cc ON cc."Id" = pa."ContractingCompanyId"
		JOIN "Providers" p ON p."Id" = cc."ProviderId"
		WHERE p."OrganizationId" = $1
			AND f."Status" = 'active'
			AND f."PredictedEndDate" < $2::date
			AND (f."AutoRenew" = false OR f."RenewalPeriodMonths" IS NULL)`, orgID, asOf)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []LineFidelityRow
	for rows.Next() {
		var row LineFidelityRow
		if err := rows.Scan(&row.ID, &row.PhoneLineID, &row.StartDate, &row.InitialMonths, &row.PredictedEndDate,
			&row.AutoRenew, &row.RenewalPeriodMonths, &row.Status, &row.CreatedAt, &row.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, row)
	}
	return items, rows.Err()
}

func (s *Store) ListFidelityRenewalTriggers(ctx context.Context, orgID string) ([]models.FidelityRenewalTriggerResponse, error) {
	rows, err := s.q(ctx).Query(ctx, `
		SELECT "Id", "EventKey", "Label", "PromptEnabled"
		FROM "FidelityRenewalTriggers"
		WHERE "OrganizationId" = $1
		ORDER BY "Label"`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []models.FidelityRenewalTriggerResponse
	for rows.Next() {
		var item models.FidelityRenewalTriggerResponse
		if err := rows.Scan(&item.ID, &item.EventKey, &item.Label, &item.PromptEnabled); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if items == nil {
		items = []models.FidelityRenewalTriggerResponse{}
	}
	return items, rows.Err()
}

func (s *Store) FidelityTriggerPromptEnabled(ctx context.Context, orgID, eventKey string) (bool, error) {
	var enabled bool
	err := s.q(ctx).QueryRow(ctx, `
		SELECT "PromptEnabled" FROM "FidelityRenewalTriggers"
		WHERE "OrganizationId" = $1 AND "EventKey" = $2`, orgID, eventKey).Scan(&enabled)
	if errors.Is(err, pgx.ErrNoRows) {
		return true, nil
	}
	return enabled, err
}

func (s *Store) UpdateFidelityRenewalTrigger(ctx context.Context, orgID, id string, enabled bool, now time.Time) error {
	tag, err := s.q(ctx).Exec(ctx, `
		UPDATE "FidelityRenewalTriggers"
		SET "PromptEnabled" = $3, "UpdatedAt" = $4
		WHERE "OrganizationId" = $1 AND "Id" = $2`, orgID, id, enabled, now)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (s *Store) GetFirstActiveContractTemplateID(ctx context.Context, orgID string) (string, error) {
	var id string
	err := s.q(ctx).QueryRow(ctx, `
		SELECT "Id" FROM "ContractTemplates"
		WHERE "OrganizationId" = $1 AND "Active" = true
		ORDER BY CASE WHEN "Code" = 'default_service' THEN 0 ELSE 1 END, "Name"
		LIMIT 1`, orgID).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}
	return id, err
}

func (s *Store) SaveGeneratedContractWithTrigger(ctx context.Context, id, orgID string, saleID, customerID, phoneLineID *string, templateID, trigger, renderedHTML string, now time.Time) error {
	_, err := s.q(ctx).Exec(ctx, `
		INSERT INTO "GeneratedContracts" (
			"Id", "OrganizationId", "SaleId", "CustomerId", "PhoneLineId", "ContractTemplateId",
			"Trigger", "Status", "RenderedHtml", "GeneratedAt", "CreatedAt"
		) VALUES ($1, $2, $3, $4, $5, $6, $7, 'generated'::generated_contract_status, $8, $9, $9)`,
		id, orgID, saleID, customerID, phoneLineID, templateID, trigger, renderedHTML, now)
	return err
}

func (s *Store) ListGeneratedContractsForCustomer(ctx context.Context, orgID, customerID string) ([]models.GeneratedContractResponse, error) {
	rows, err := s.q(ctx).Query(ctx, `
		SELECT g."Id", g."CustomerId", g."PhoneLineId", g."SaleId", g."ContractTemplateId", t."Name",
			g."Trigger", g."Status"::text, g."RenderedHtml", g."GeneratedAt", g."CreatedAt"
		FROM "GeneratedContracts" g
		JOIN "ContractTemplates" t ON t."Id" = g."ContractTemplateId"
		WHERE g."OrganizationId" = $1 AND (
			g."CustomerId" = $2
			OR g."SaleId" IN (SELECT s."Id" FROM "Sales" s WHERE s."CustomerId" = $2)
		)
		ORDER BY g."CreatedAt" DESC`, orgID, customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanGeneratedContracts(rows)
}

func (s *Store) ListGeneratedContractsForPhoneLine(ctx context.Context, orgID, phoneLineID string) ([]models.GeneratedContractResponse, error) {
	rows, err := s.q(ctx).Query(ctx, `
		SELECT g."Id", g."CustomerId", g."PhoneLineId", g."SaleId", g."ContractTemplateId", t."Name",
			g."Trigger", g."Status"::text, g."RenderedHtml", g."GeneratedAt", g."CreatedAt"
		FROM "GeneratedContracts" g
		JOIN "ContractTemplates" t ON t."Id" = g."ContractTemplateId"
		WHERE g."OrganizationId" = $1 AND g."PhoneLineId" = $2
		ORDER BY g."CreatedAt" DESC`, orgID, phoneLineID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanGeneratedContracts(rows)
}

func scanGeneratedContracts(rows pgx.Rows) ([]models.GeneratedContractResponse, error) {
	var items []models.GeneratedContractResponse
	for rows.Next() {
		var item models.GeneratedContractResponse
		if err := rows.Scan(&item.ID, &item.CustomerID, &item.PhoneLineID, &item.SaleID, &item.ContractTemplateID,
			&item.ContractTemplateName, &item.Trigger, &item.Status, &item.RenderedHTML, &item.GeneratedAt, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if items == nil {
		items = []models.GeneratedContractResponse{}
	}
	return items, rows.Err()
}
