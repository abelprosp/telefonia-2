package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/models"
)

func (s *Store) ListInvoiceEmailTemplates(ctx context.Context, orgID string, kind *string, page httputil.PageSearch) ([]models.ListInvoiceEmailTemplateResponse, int64, error) {
	base := ` FROM "InvoiceEmailTemplates" WHERE "OrganizationId" = $1`
	args := []any{orgID}
	if kind != nil && *kind != "" {
		base += ` AND "Kind" = $2::invoice_email_template_kind`
		args = append(args, *kind)
	}
	var total int64
	if err := s.q(ctx).QueryRow(ctx, `SELECT COUNT(*) `+base, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	selectQ := `
		SELECT "Id", "Name", "Code", "Kind"::text, "SubjectTemplate", "Active", "CreatedAt", "UpdatedAt"
		` + base + `
		ORDER BY "Name"
		OFFSET $` + itoa(len(args)+1) + ` LIMIT $` + itoa(len(args)+2)
	args = append(args, page.Offset(), page.Limit())
	rows, err := s.q(ctx).Query(ctx, selectQ, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []models.ListInvoiceEmailTemplateResponse
	for rows.Next() {
		var item models.ListInvoiceEmailTemplateResponse
		if err := rows.Scan(&item.ID, &item.Name, &item.Code, &item.Kind, &item.SubjectTemplate, &item.Active, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (s *Store) GetInvoiceEmailTemplate(ctx context.Context, orgID, id string) (*models.GetInvoiceEmailTemplateResponse, error) {
	var item models.GetInvoiceEmailTemplateResponse
	err := s.q(ctx).QueryRow(ctx, `
		SELECT "Id", "Name", "Code", "Kind"::text, "SubjectTemplate", "BodyTemplateHtml", "Active", "CreatedAt", "UpdatedAt"
		FROM "InvoiceEmailTemplates"
		WHERE "OrganizationId" = $1 AND "Id" = $2`, orgID, id).Scan(
		&item.ID, &item.Name, &item.Code, &item.Kind, &item.SubjectTemplate, &item.BodyTemplateHtml,
		&item.Active, &item.CreatedAt, &item.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *Store) GetInvoiceEmailTemplateByCode(ctx context.Context, orgID, code string) (*models.GetInvoiceEmailTemplateResponse, error) {
	var item models.GetInvoiceEmailTemplateResponse
	err := s.q(ctx).QueryRow(ctx, `
		SELECT "Id", "Name", "Code", "Kind"::text, "SubjectTemplate", "BodyTemplateHtml", "Active", "CreatedAt", "UpdatedAt"
		FROM "InvoiceEmailTemplates"
		WHERE "OrganizationId" = $1 AND "Code" = $2 AND "Active" = true`, orgID, code).Scan(
		&item.ID, &item.Name, &item.Code, &item.Kind, &item.SubjectTemplate, &item.BodyTemplateHtml,
		&item.Active, &item.CreatedAt, &item.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *Store) InvoiceEmailTemplateCodeExists(ctx context.Context, orgID, code string, excludeID *string) (bool, error) {
	q := `SELECT EXISTS(SELECT 1 FROM "InvoiceEmailTemplates" WHERE "OrganizationId" = $1 AND "Code" = $2`
	args := []any{orgID, code}
	if excludeID != nil && *excludeID != "" {
		q += ` AND "Id" <> $3`
		args = append(args, *excludeID)
	}
	q += `)`
	var exists bool
	err := s.q(ctx).QueryRow(ctx, q, args...).Scan(&exists)
	return exists, err
}

func (s *Store) CreateInvoiceEmailTemplate(ctx context.Context, id, orgID, name, code, kind, subject, body string, active bool, now time.Time) error {
	_, err := s.q(ctx).Exec(ctx, `
		INSERT INTO "InvoiceEmailTemplates" (
			"Id", "OrganizationId", "Name", "Code", "Kind", "SubjectTemplate", "BodyTemplateHtml", "Active", "CreatedAt", "UpdatedAt"
		) VALUES ($1, $2, $3, $4, $5::invoice_email_template_kind, $6, $7, $8, $9, $9)`,
		id, orgID, name, code, kind, subject, body, active, now)
	return err
}

func (s *Store) UpdateInvoiceEmailTemplate(ctx context.Context, orgID, id, name, subject, body string, active bool, now time.Time) error {
	tag, err := s.q(ctx).Exec(ctx, `
		UPDATE "InvoiceEmailTemplates"
		SET "Name" = $3, "SubjectTemplate" = $4, "BodyTemplateHtml" = $5, "Active" = $6, "UpdatedAt" = $7
		WHERE "OrganizationId" = $1 AND "Id" = $2`, orgID, id, name, subject, body, active, now)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (s *Store) ListCustomerBillingDocuments(ctx context.Context, orgID string, status, customerID *string, page httputil.PageSearch) ([]models.ListCustomerBillingDocumentResponse, int64, error) {
	base := `
		FROM "CustomerBillingDocuments" d
		JOIN "Customers" c ON c."Id" = d."CustomerId"
		WHERE d."OrganizationId" = $1`
	args := []any{orgID}
	if status != nil && *status != "" {
		base += ` AND d."Status" = $` + itoa(len(args)+1) + `::customer_billing_document_status`
		args = append(args, *status)
	}
	if customerID != nil && *customerID != "" {
		base += ` AND d."CustomerId" = $` + itoa(len(args)+1)
		args = append(args, *customerID)
	}
	var total int64
	if err := s.q(ctx).QueryRow(ctx, `SELECT COUNT(*) `+base, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	selectQ := `
		SELECT d."Id", d."CustomerId", c."Name", d."AccountsReceivableId", d."ProcessingMonthId",
			d."InvoiceNumber", d."IssueDate", d."DueDate", d."Amount", d."Status"::text,
			d."RecipientEmail", d."EmailSubject", d."SendCount", d."SentAt", d."LastSentAt", d."CreatedAt"
		` + base + `
		ORDER BY d."CreatedAt" DESC
		OFFSET $` + itoa(len(args)+1) + ` LIMIT $` + itoa(len(args)+2)
	args = append(args, page.Offset(), page.Limit())
	rows, err := s.q(ctx).Query(ctx, selectQ, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []models.ListCustomerBillingDocumentResponse
	for rows.Next() {
		var item models.ListCustomerBillingDocumentResponse
		if err := rows.Scan(
			&item.ID, &item.CustomerID, &item.CustomerName, &item.AccountsReceivableID, &item.ProcessingMonthID,
			&item.InvoiceNumber, &item.IssueDate, &item.DueDate, &item.Amount, &item.Status,
			&item.RecipientEmail, &item.EmailSubject, &item.SendCount, &item.SentAt, &item.LastSentAt, &item.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (s *Store) GetCustomerBillingDocument(ctx context.Context, orgID, id string) (*models.GetCustomerBillingDocumentResponse, error) {
	var item models.GetCustomerBillingDocumentResponse
	err := s.q(ctx).QueryRow(ctx, `
		SELECT d."Id", d."CustomerId", c."Name", d."AccountsReceivableId", d."ProcessingMonthId",
			d."InvoiceNumber", d."IssueDate", d."DueDate", d."Amount", d."Status"::text,
			d."RecipientEmail", d."EmailSubject", d."EmailBodyHtml", d."SendCount", d."SentAt", d."LastSentAt",
			d."CreatedAt", d."UpdatedAt"
		FROM "CustomerBillingDocuments" d
		JOIN "Customers" c ON c."Id" = d."CustomerId"
		WHERE d."OrganizationId" = $1 AND d."Id" = $2`, orgID, id).Scan(
		&item.ID, &item.CustomerID, &item.CustomerName, &item.AccountsReceivableID, &item.ProcessingMonthID,
		&item.InvoiceNumber, &item.IssueDate, &item.DueDate, &item.Amount, &item.Status,
		&item.RecipientEmail, &item.EmailSubject, &item.EmailBodyHtml, &item.SendCount, &item.SentAt, &item.LastSentAt,
		&item.CreatedAt, &item.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *Store) CreateCustomerBillingDocument(ctx context.Context, doc models.CustomerBillingDocumentRow) error {
	_, err := s.q(ctx).Exec(ctx, `
		INSERT INTO "CustomerBillingDocuments" (
			"Id", "OrganizationId", "CustomerId", "AccountsReceivableId", "ProcessingMonthId",
			"InvoiceNumber", "IssueDate", "DueDate", "Amount", "Status",
			"RecipientEmail", "EmailSubject", "EmailBodyHtml", "CreatedAt", "UpdatedAt"
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10::customer_billing_document_status, $11, $12, $13, $14, $14)`,
		doc.ID, doc.OrganizationID, doc.CustomerID, doc.AccountsReceivableID, doc.ProcessingMonthID,
		doc.InvoiceNumber, doc.IssueDate, doc.DueDate, doc.Amount, doc.Status,
		doc.RecipientEmail, doc.EmailSubject, doc.EmailBodyHTML, doc.CreatedAt)
	return err
}

func (s *Store) UpdateCustomerBillingDocument(ctx context.Context, orgID, id, recipient, subject, body, status string, now time.Time) error {
	tag, err := s.q(ctx).Exec(ctx, `
		UPDATE "CustomerBillingDocuments"
		SET "RecipientEmail" = $3, "EmailSubject" = $4, "EmailBodyHtml" = $5,
			"Status" = $6::customer_billing_document_status, "UpdatedAt" = $7
		WHERE "OrganizationId" = $1 AND "Id" = $2`, orgID, id, recipient, subject, body, status, now)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (s *Store) MarkCustomerBillingDocumentSent(ctx context.Context, orgID, id string, now time.Time) error {
	_, err := s.q(ctx).Exec(ctx, `
		UPDATE "CustomerBillingDocuments"
		SET "Status" = 'sent'::customer_billing_document_status,
			"SendCount" = "SendCount" + 1,
			"SentAt" = COALESCE("SentAt", $3),
			"LastSentAt" = $3,
			"UpdatedAt" = $3
		WHERE "OrganizationId" = $1 AND "Id" = $2`, orgID, id, now)
	return err
}

func (s *Store) InsertCustomerBillingSendLog(ctx context.Context, id, orgID, documentID, recipient, subject string, success bool, errMsg, userID string, sentAt time.Time) error {
	var errPtr *string
	if errMsg != "" {
		errPtr = &errMsg
	}
	_, err := s.q(ctx).Exec(ctx, `
		INSERT INTO "CustomerBillingSendLog" (
			"Id", "OrganizationId", "DocumentId", "RecipientEmail", "Subject", "Success", "ErrorMessage", "SentByUserId", "SentAt"
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		id, orgID, documentID, recipient, subject, success, errPtr, userID, sentAt)
	return err
}

func (s *Store) ListCustomerBillingSendLog(ctx context.Context, orgID, documentID string) ([]models.CustomerBillingSendLogResponse, error) {
	rows, err := s.q(ctx).Query(ctx, `
		SELECT "Id", "RecipientEmail", "Subject", "Success", "ErrorMessage", "SentByUserId", "SentAt"
		FROM "CustomerBillingSendLog"
		WHERE "OrganizationId" = $1 AND "DocumentId" = $2
		ORDER BY "SentAt" DESC`, orgID, documentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []models.CustomerBillingSendLogResponse
	for rows.Next() {
		var item models.CustomerBillingSendLogResponse
		if err := rows.Scan(&item.ID, &item.RecipientEmail, &item.Subject, &item.Success, &item.ErrorMessage, &item.SentByUserID, &item.SentAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) NextBillingInvoiceNumber(ctx context.Context, orgID string) (string, error) {
	var seq int64
	err := s.q(ctx).QueryRow(ctx, `
		SELECT COUNT(*) + 1 FROM "CustomerBillingDocuments" WHERE "OrganizationId" = $1`, orgID).Scan(&seq)
	if err != nil {
		return "", err
	}
	return formatInvoiceNumber(seq), nil
}

func formatInvoiceNumber(seq int64) string {
	return "FAT-" + itoa(int(seq))
}

func (s *Store) GetCustomerBillingEmail(ctx context.Context, orgID, customerID string) (string, error) {
	var email *string
	err := s.q(ctx).QueryRow(ctx, `
		SELECT "BillingEmail" FROM "Customers" WHERE "OrganizationId" = $1 AND "Id" = $2`, orgID, customerID).Scan(&email)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	if email == nil {
		return "", nil
	}
	return *email, nil
}

func (s *Store) UpdateCustomerBillingEmail(ctx context.Context, orgID, customerID, email string) error {
	tag, err := s.q(ctx).Exec(ctx, `
		UPDATE "Customers" SET "BillingEmail" = $3 WHERE "OrganizationId" = $1 AND "Id" = $2`,
		orgID, customerID, nullIfEmpty(email))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func nullIfEmpty(v string) any {
	if v == "" {
		return nil
	}
	return v
}

type ReceivableForBilling struct {
	ID                string
	CustomerID        string
	CustomerName      string
	CustomerDocument  string
	BillingEmail      string
	Description       string
	ProcessingMonthID *string
	IssueDate         time.Time
	DueDate           time.Time
	Amount            float64
	Balance           float64
	Status            string
}

func (s *Store) GetReceivableForBilling(ctx context.Context, orgID, receivableID string) (*ReceivableForBilling, error) {
	var item ReceivableForBilling
	err := s.q(ctx).QueryRow(ctx, `
		SELECT ar."Id", ar."CustomerId", c."Name",
			COALESCE((SELECT cd."Number" FROM "CustomerDocuments" cd
				WHERE cd."CustomerId" = c."Id" AND cd."DocumentType" IN ('cpf','cnpj') LIMIT 1), ''),
			COALESCE(c."BillingEmail", ''),
			ar."Description", ar."ProcessingMonthId", ar."IssueDate", ar."DueDate",
			ar."Amount", ar."Amount" - ar."ReceivedAmount", ar."Status"::text
		FROM "AccountsReceivable" ar
		JOIN "Customers" c ON c."Id" = ar."CustomerId"
		WHERE ar."OrganizationId" = $1 AND ar."Id" = $2`, orgID, receivableID).Scan(
		&item.ID, &item.CustomerID, &item.CustomerName, &item.CustomerDocument, &item.BillingEmail,
		&item.Description, &item.ProcessingMonthID, &item.IssueDate, &item.DueDate,
		&item.Amount, &item.Balance, &item.Status,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *Store) ListOverdueReceivables(ctx context.Context, orgID string, page httputil.PageSearch) ([]models.OverdueReceivableResponse, int64, error) {
	base := `
		FROM "AccountsReceivable" ar
		JOIN "Customers" c ON c."Id" = ar."CustomerId"
		WHERE ar."OrganizationId" = $1 AND ar."Status" = 'overdue'::financial_entry_status`
	args := []any{orgID}
	var total int64
	if err := s.q(ctx).QueryRow(ctx, `SELECT COUNT(*) `+base, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	selectQ := `
		SELECT ar."Id", ar."CustomerId", c."Name", COALESCE(c."BillingEmail", ''),
			ar."Description", ar."DueDate", ar."Amount" - ar."ReceivedAmount",
			(SELECT COUNT(*)::int FROM "CollectionReminders" cr WHERE cr."AccountsReceivableId" = ar."Id" AND cr."Status" = 'sent')
		` + base + `
		ORDER BY ar."DueDate" ASC
		OFFSET $` + itoa(len(args)+1) + ` LIMIT $` + itoa(len(args)+2)
	args = append(args, page.Offset(), page.Limit())
	rows, err := s.q(ctx).Query(ctx, selectQ, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []models.OverdueReceivableResponse
	for rows.Next() {
		var item models.OverdueReceivableResponse
		if err := rows.Scan(
			&item.ID, &item.CustomerID, &item.CustomerName, &item.BillingEmail,
			&item.Description, &item.DueDate, &item.Balance, &item.RemindersSent,
		); err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (s *Store) InsertCollectionReminder(ctx context.Context, id, orgID, receivableID string, level int, recipient, subject, body, userID string, status string, errMsg *string, sentAt *time.Time, createdAt time.Time) error {
	_, err := s.q(ctx).Exec(ctx, `
		INSERT INTO "CollectionReminders" (
			"Id", "OrganizationId", "AccountsReceivableId", "ReminderLevel",
			"RecipientEmail", "Subject", "BodyHtml", "Status", "ErrorMessage", "SentByUserId", "SentAt", "CreatedAt"
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8::collection_reminder_status, $9, $10, $11, $12)`,
		id, orgID, receivableID, level, recipient, subject, body, status, errMsg, userID, sentAt, createdAt)
	return err
}

func (s *Store) CountBillingDocumentsByStatus(ctx context.Context, orgID string) (draft, ready, sent int32, err error) {
	err = s.q(ctx).QueryRow(ctx, `
		SELECT
			COALESCE(SUM(CASE WHEN "Status" = 'draft' THEN 1 ELSE 0 END), 0)::int,
			COALESCE(SUM(CASE WHEN "Status" = 'ready' THEN 1 ELSE 0 END), 0)::int,
			COALESCE(SUM(CASE WHEN "Status" = 'sent' THEN 1 ELSE 0 END), 0)::int
		FROM "CustomerBillingDocuments" WHERE "OrganizationId" = $1`, orgID).Scan(&draft, &ready, &sent)
	return
}
