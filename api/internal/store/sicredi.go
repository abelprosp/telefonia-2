package store

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

type SicrediWebhookEventRow struct {
	ID              string
	OrganizationID  *string
	NossoNumero     *string
	SeuNumero       *string
	IdTituloEmpresa *string
	EventType       string
	Payload         json.RawMessage
	Processed       bool
	ProcessError    *string
	CreatedAt       time.Time
	ProcessedAt     *time.Time
}

func (s *Store) InsertSicrediWebhookEvent(ctx context.Context, row SicrediWebhookEventRow) (bool, error) {
	tag, err := s.q(ctx).Exec(ctx, `
		INSERT INTO "SicrediWebhookEvents" (
			"Id", "OrganizationId", "NossoNumero", "SeuNumero", "IdTituloEmpresa",
			"EventType", "Payload", "Processed", "ProcessError", "CreatedAt", "ProcessedAt"
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT ("NossoNumero", "EventType") WHERE "NossoNumero" IS NOT NULL AND "EventType" <> ''
		DO NOTHING`,
		row.ID, row.OrganizationID, row.NossoNumero, row.SeuNumero, row.IdTituloEmpresa,
		row.EventType, row.Payload, row.Processed, row.ProcessError, row.CreatedAt, row.ProcessedAt,
	)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

func (s *Store) MarkSicrediWebhookEventProcessed(ctx context.Context, id string, processError *string, processedAt time.Time) error {
	_, err := s.q(ctx).Exec(ctx, `
		UPDATE "SicrediWebhookEvents"
		SET "Processed" = true, "ProcessError" = $2, "ProcessedAt" = $3
		WHERE "Id" = $1`, id, processError, processedAt)
	return err
}

func (s *Store) GetBillingDocumentByIDGlobal(ctx context.Context, documentID string) (*UnpaidSicrediBillingDocument, error) {
	var item UnpaidSicrediBillingDocument
	err := s.q(ctx).QueryRow(ctx, `
		SELECT d."Id", d."OrganizationId", d."InvoiceNumber", c."Name",
			d."AccountsReceivableId", d."Amount", d."SicrediNossoNumero"
		FROM "CustomerBillingDocuments" d
		JOIN "Customers" c ON c."Id" = d."CustomerId"
		WHERE d."Id" = $1`, documentID).Scan(
		&item.ID, &item.OrganizationID, &item.InvoiceNumber, &item.CustomerName,
		&item.AccountsReceivableID, &item.Amount, &item.SicrediNossoNumero,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *Store) GetBillingDocumentByNossoNumeroGlobal(ctx context.Context, nossoNumero string) (*UnpaidSicrediBillingDocument, error) {
	var item UnpaidSicrediBillingDocument
	err := s.q(ctx).QueryRow(ctx, `
		SELECT d."Id", d."OrganizationId", d."InvoiceNumber", c."Name",
			d."AccountsReceivableId", d."Amount", d."SicrediNossoNumero"
		FROM "CustomerBillingDocuments" d
		JOIN "Customers" c ON c."Id" = d."CustomerId"
		WHERE d."SicrediNossoNumero" = $1
		ORDER BY d."CreatedAt" DESC
		LIMIT 1`, nossoNumero).Scan(
		&item.ID, &item.OrganizationID, &item.InvoiceNumber, &item.CustomerName,
		&item.AccountsReceivableID, &item.Amount, &item.SicrediNossoNumero,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *Store) UpdateCustomerBillingDocumentSicrediStatus(ctx context.Context, orgID, id, status string, now time.Time) error {
	tag, err := s.q(ctx).Exec(ctx, `
		UPDATE "CustomerBillingDocuments"
		SET "SicrediBoletoStatus" = $3, "UpdatedAt" = $4
		WHERE "OrganizationId" = $1 AND "Id" = $2`, orgID, id, status, now)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}
