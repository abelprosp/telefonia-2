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
			d."RecipientEmail", d."EmailSubject", d."SendCount", d."SentAt", d."LastSentAt", d."CreatedAt",
			d."SicrediNossoNumero", d."SicrediLinhaDigitavel", d."SicrediCodigoBarras",
			d."SicrediPixQrCode", d."SicrediPixTxId", d."SicrediBoletoStatus", d."SicrediBoletoError",
			d."SicrediPaidAt"
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
			&item.SicrediNossoNumero, &item.SicrediLinhaDigitavel, &item.SicrediCodigoBarras,
			&item.SicrediPixQrCode, &item.SicrediPixTxID, &item.SicrediBoletoStatus, &item.SicrediBoletoError,
			&item.SicrediPaidAt,
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
			d."CreatedAt", d."UpdatedAt",
			d."SicrediNossoNumero", d."SicrediLinhaDigitavel", d."SicrediCodigoBarras",
			d."SicrediPixQrCode", d."SicrediPixTxId", d."SicrediBoletoStatus", d."SicrediBoletoError",
			d."SicrediPaidAt"
		FROM "CustomerBillingDocuments" d
		JOIN "Customers" c ON c."Id" = d."CustomerId"
		WHERE d."OrganizationId" = $1 AND d."Id" = $2`, orgID, id).Scan(
		&item.ID, &item.CustomerID, &item.CustomerName, &item.AccountsReceivableID, &item.ProcessingMonthID,
		&item.InvoiceNumber, &item.IssueDate, &item.DueDate, &item.Amount, &item.Status,
		&item.RecipientEmail, &item.EmailSubject, &item.EmailBodyHtml, &item.SendCount, &item.SentAt, &item.LastSentAt,
		&item.CreatedAt, &item.UpdatedAt,
		&item.SicrediNossoNumero, &item.SicrediLinhaDigitavel, &item.SicrediCodigoBarras,
		&item.SicrediPixQrCode, &item.SicrediPixTxID, &item.SicrediBoletoStatus, &item.SicrediBoletoError,
		&item.SicrediPaidAt,
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

func (s *Store) UpdateCustomerBillingDocumentSicredi(ctx context.Context, orgID, id string,
	nossoNumero, linhaDigitavel, codigoBarras, pixQrCode, pixTxID, boletoStatus string,
	boletoError *string, emailBodyHTML string, now time.Time,
) error {
	tag, err := s.q(ctx).Exec(ctx, `
		UPDATE "CustomerBillingDocuments"
		SET "SicrediNossoNumero" = $3,
			"SicrediLinhaDigitavel" = $4,
			"SicrediCodigoBarras" = $5,
			"SicrediPixQrCode" = $6,
			"SicrediPixTxId" = $7,
			"SicrediBoletoStatus" = $8,
			"SicrediBoletoError" = $9,
			"EmailBodyHtml" = $10,
			"UpdatedAt" = $11
		WHERE "OrganizationId" = $1 AND "Id" = $2`,
		orgID, id, nullIfEmpty(nossoNumero), nullIfEmpty(linhaDigitavel), nullIfEmpty(codigoBarras),
		nullIfEmpty(pixQrCode), nullIfEmpty(pixTxID), nullIfEmpty(boletoStatus), boletoError, emailBodyHTML, now)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
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

func (s *Store) CountProviderInvoicesForMonth(ctx context.Context, orgID, processingMonthID string) (int, error) {
	var count int
	err := s.q(ctx).QueryRow(ctx, `
		SELECT COUNT(*)::int
		FROM "ProviderInvoices" i
		JOIN "ProviderAccounts" pa ON pa."Id" = i."ProviderAccountId"
		JOIN "ContractingCompanies" cc ON cc."Id" = pa."ContractingCompanyId"
		JOIN "Providers" p ON p."Id" = cc."ProviderId"
		WHERE p."OrganizationId" = $1 AND i."ProcessingMonthId" = $2`, orgID, processingMonthID).Scan(&count)
	return count, err
}

func (s *Store) ListBulkBillingCandidates(ctx context.Context, orgID, processingMonthID string, customerIDs []string) ([]models.BulkBillingPreviewItem, error) {
	args := []any{orgID, processingMonthID}
	customerFilter := ""
	if len(customerIDs) > 0 {
		args = append(args, customerIDs)
		customerFilter = ` AND c."Id" = ANY($` + itoa(len(args)) + `::text[])`
	}
	q := `
		WITH customer_lines AS (
			SELECT
				l."CustomerId" AS customer_id,
				COUNT(DISTINCT pl."Id")::int AS line_count,
				COALESCE(SUM(COALESCE(pl."CostWithConsumption", pl."BaseCost", 0)), 0) AS provider_cost,
				COALESCE(SUM(` + lineLuxusBillingAmountSQL + `), 0) AS line_amount
			FROM "PhoneLineCustomerLinks" l
			JOIN "PhoneLines" pl ON pl."Id" = l."PhoneLineId"
			JOIN "ProviderInvoicePhoneLines" j ON j."PhoneLinesId" = pl."Id"
			JOIN "ProviderInvoices" i ON i."Id" = j."ProviderInvoicesId" AND i."ProcessingMonthId" = $2
			WHERE l."EndDate" IS NULL
			GROUP BY l."CustomerId"
		),
		customer_devices AS (
			SELECT
				d."CustomerId" AS customer_id,
				COUNT(d."Id")::int AS device_count,
				COALESCE(SUM(d."MonthlyAmount"), 0) AS device_amount
			FROM "CustomerDeviceLinks" d
			WHERE d."EndDate" IS NULL
			GROUP BY d."CustomerId"
		)
		SELECT
			c."Id", c."Name",
			COALESCE((
				SELECT cd."Number" FROM "CustomerDocuments" cd
				WHERE cd."CustomerId" = c."Id" AND cd."DocumentType" IN ('cpf', 'cnpj')
				LIMIT 1
			), '') AS customer_document,
			COALESCE(c."BillingEmail", '') AS billing_email,
			COALESCE(cl.line_count, 0),
			COALESCE(cd.device_count, 0),
			COALESCE(cl.line_amount, 0) + COALESCE(cd.device_amount, 0) AS monthly_amount,
			COALESCE(cl.provider_cost, 0),
			EXISTS(
				SELECT 1 FROM "CustomerBillingDocuments" d
				WHERE d."OrganizationId" = $1 AND d."CustomerId" = c."Id"
					AND d."Status" != 'cancelled' AND d."ProcessingMonthId" = $2
			) AS already_billed
		FROM "Customers" c
		LEFT JOIN customer_lines cl ON cl.customer_id = c."Id"
		LEFT JOIN customer_devices cd ON cd.customer_id = c."Id"
		WHERE c."OrganizationId" = $1 AND c."Active" = true
			AND COALESCE(cl.line_count, 0) > 0` + customerFilter + `
		ORDER BY c."Name"`
	rows, err := s.q(ctx).Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []models.BulkBillingPreviewItem
	for rows.Next() {
		var item models.BulkBillingPreviewItem
		if err := rows.Scan(
			&item.CustomerID, &item.CustomerName, &item.CustomerDocument, &item.BillingEmail,
			&item.LineCount, &item.DeviceCount, &item.MonthlyAmount, &item.ProviderCost, &item.AlreadyBilled,
		); err != nil {
			return nil, err
		}
		item.Eligible, item.SkipReason = bulkBillingEligibility(item)
		items = append(items, item)
	}
	return items, rows.Err()
}

func bulkBillingEligibility(item models.BulkBillingPreviewItem) (bool, string) {
	if item.LineCount == 0 {
		return false, "no_lines_on_invoice"
	}
	if item.MonthlyAmount <= 0 {
		return false, "no_monthly_amount"
	}
	if item.AlreadyBilled {
		return false, "already_billed"
	}
	return true, ""
}

func (s *Store) ListManualBillingCandidates(ctx context.Context, orgID string, customerIDs []string) ([]models.BulkBillingPreviewItem, error) {
	args := []any{orgID}
	customerFilter := ""
	if len(customerIDs) > 0 {
		args = append(args, customerIDs)
		customerFilter = ` AND c."Id" = ANY($` + itoa(len(args)) + `::text[])`
	}
	q := `
		WITH customer_lines AS (
			SELECT
				l."CustomerId" AS customer_id,
				COUNT(DISTINCT pl."Id")::int AS line_count,
				COALESCE(SUM(` + lineLuxusBillingAmountSQL + `), 0) AS line_amount
			FROM "PhoneLineCustomerLinks" l
			JOIN "PhoneLines" pl ON pl."Id" = l."PhoneLineId"
			WHERE l."EndDate" IS NULL
			GROUP BY l."CustomerId"
		),
		customer_devices AS (
			SELECT
				d."CustomerId" AS customer_id,
				COUNT(d."Id")::int AS device_count,
				COALESCE(SUM(d."MonthlyAmount"), 0) AS device_amount
			FROM "CustomerDeviceLinks" d
			WHERE d."EndDate" IS NULL
			GROUP BY d."CustomerId"
		)
		SELECT
			c."Id", c."Name",
			COALESCE((
				SELECT cd."Number" FROM "CustomerDocuments" cd
				WHERE cd."CustomerId" = c."Id" AND cd."DocumentType" IN ('cpf', 'cnpj')
				LIMIT 1
			), '') AS customer_document,
			COALESCE(c."BillingEmail", '') AS billing_email,
			COALESCE(cl.line_count, 0),
			COALESCE(cd.device_count, 0),
			COALESCE(cl.line_amount, 0) + COALESCE(cd.device_amount, 0) AS monthly_amount,
			0::float8 AS provider_cost,
			false AS already_billed
		FROM "Customers" c
		LEFT JOIN customer_lines cl ON cl.customer_id = c."Id"
		LEFT JOIN customer_devices cd ON cd.customer_id = c."Id"
		WHERE c."OrganizationId" = $1 AND c."Active" = true
			AND (COALESCE(cl.line_count, 0) > 0 OR COALESCE(cd.device_count, 0) > 0)` + customerFilter + `
		ORDER BY c."Name"`
	rows, err := s.q(ctx).Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []models.BulkBillingPreviewItem
	for rows.Next() {
		var item models.BulkBillingPreviewItem
		if err := rows.Scan(
			&item.CustomerID, &item.CustomerName, &item.CustomerDocument, &item.BillingEmail,
			&item.LineCount, &item.DeviceCount, &item.MonthlyAmount, &item.ProviderCost, &item.AlreadyBilled,
		); err != nil {
			return nil, err
		}
		item.Eligible, item.SkipReason = manualBillingEligibility(item)
		items = append(items, item)
	}
	return items, rows.Err()
}

func manualBillingEligibility(item models.BulkBillingPreviewItem) (bool, string) {
	if item.LineCount == 0 && item.DeviceCount == 0 {
		return false, "no_active_lines"
	}
	if item.MonthlyAmount <= 0 {
		return false, "no_monthly_amount"
	}
	return true, ""
}

type CustomerProviderInvoiceLayoutContext struct {
	PeriodStart     time.Time
	PeriodEnd       time.Time
	IssueDate       time.Time
	DueDate         time.Time
	ServicesTotal   float64
	DiscountsTotal  float64
	ReferenceMonth  string
}

func (s *Store) ListCustomerBillingItemsForProcessingMonth(ctx context.Context, customerID, processingMonthID string) ([]CustomerBillingItemRow, error) {
	serviceRows, err := s.q(ctx).Query(ctx, `
		WITH customer_invoices AS (
			SELECT DISTINCT i."Id" AS invoice_id
			FROM "ProviderInvoices" i
			JOIN "ProviderInvoicePhoneLines" j ON j."ProviderInvoicesId" = i."Id"
			JOIN "PhoneLineCustomerLinks" l ON l."PhoneLineId" = j."PhoneLinesId" AND l."EndDate" IS NULL
			WHERE l."CustomerId" = $1 AND i."ProcessingMonthId" = $2
		),
		invoice_line_counts AS (
			SELECT j."ProviderInvoicesId" AS invoice_id, COUNT(*)::numeric AS total_lines
			FROM "ProviderInvoicePhoneLines" j
			GROUP BY j."ProviderInvoicesId"
		),
		customer_line_counts AS (
			SELECT j."ProviderInvoicesId" AS invoice_id, COUNT(*)::numeric AS customer_lines
			FROM "ProviderInvoicePhoneLines" j
			JOIN "PhoneLineCustomerLinks" l ON l."PhoneLineId" = j."PhoneLinesId" AND l."EndDate" IS NULL
			WHERE l."CustomerId" = $1
			GROUP BY j."ProviderInvoicesId"
		)
		SELECT s."Description", 'Mensal' AS item_type,
			ROUND(s."TotalPrice" * (clc.customer_lines / ilc.total_lines), 2) AS amount
		FROM "ProviderInvoiceServices" s
		JOIN customer_invoices ci ON ci.invoice_id = s."InvoiceId"
		JOIN invoice_line_counts ilc ON ilc.invoice_id = s."InvoiceId"
		JOIN customer_line_counts clc ON clc.invoice_id = s."InvoiceId"
		WHERE ilc.total_lines > 0`, customerID, processingMonthID)
	if err != nil {
		return nil, err
	}
	defer serviceRows.Close()

	var items []CustomerBillingItemRow
	hasServices := false
	for serviceRows.Next() {
		var item CustomerBillingItemRow
		if err := serviceRows.Scan(&item.Description, &item.ItemType, &item.Amount); err != nil {
			return nil, err
		}
		if item.Amount == 0 {
			continue
		}
		hasServices = true
		items = append(items, item)
	}
	if err := serviceRows.Err(); err != nil {
		return nil, err
	}

	if !hasServices {
		lineRows, err := s.q(ctx).Query(ctx, `
			SELECT
				COALESCE(pp."Name", pl."Number") || ' — ' || pl."Number" AS description,
				'Mensal' AS item_type,
				` + lineLuxusBillingAmountSQL + ` AS amount
			FROM "PhoneLineCustomerLinks" l
			JOIN "PhoneLines" pl ON pl."Id" = l."PhoneLineId"
			LEFT JOIN "ProviderPlans" pp ON pp."Id" = pl."ProviderPlanId"
			JOIN "ProviderInvoicePhoneLines" j ON j."PhoneLinesId" = pl."Id"
			JOIN "ProviderInvoices" i ON i."Id" = j."ProviderInvoicesId" AND i."ProcessingMonthId" = $2
			WHERE l."CustomerId" = $1 AND l."EndDate" IS NULL
			ORDER BY pl."Number"`, customerID, processingMonthID)
		if err != nil {
			return nil, err
		}
		defer lineRows.Close()
		for lineRows.Next() {
			var item CustomerBillingItemRow
			if err := lineRows.Scan(&item.Description, &item.ItemType, &item.Amount); err != nil {
				return nil, err
			}
			if item.Amount == 0 {
				continue
			}
			items = append(items, item)
		}
		if err := lineRows.Err(); err != nil {
			return nil, err
		}
	}

	deviceRows, err := s.q(ctx).Query(ctx, `
		SELECT d."Description", 'Aparelho', d."MonthlyAmount"
		FROM "CustomerDeviceLinks" d
		WHERE d."CustomerId" = $1 AND d."EndDate" IS NULL
		ORDER BY d."Description"`, customerID)
	if err != nil {
		return nil, err
	}
	defer deviceRows.Close()
	for deviceRows.Next() {
		var item CustomerBillingItemRow
		if err := deviceRows.Scan(&item.Description, &item.ItemType, &item.Amount); err != nil {
			return nil, err
		}
		if item.Amount <= 0 {
			continue
		}
		items = append(items, item)
	}
	return items, deviceRows.Err()
}

func (s *Store) GetCustomerProviderInvoiceLayoutContext(ctx context.Context, customerID, processingMonthID string) (*CustomerProviderInvoiceLayoutContext, error) {
	var ctxRow CustomerProviderInvoiceLayoutContext
	var periodStart, periodEnd, issueDate, dueDate time.Time
	err := s.q(ctx).QueryRow(ctx, `
		SELECT
			MIN(bc."StartDate"), MAX(bc."EndDate"),
			MIN(i."IssueDate"), MIN(i."DueDate"),
			COALESCE(SUM(i."SubtotalServices"), 0),
			COALESCE(SUM(i."SubtotalDiscounts"), 0)
		FROM "ProviderInvoices" i
		JOIN "BillingCycles" bc ON bc."Id" = i."BillingCycleId"
		JOIN "ProviderInvoicePhoneLines" j ON j."ProviderInvoicesId" = i."Id"
		JOIN "PhoneLineCustomerLinks" l ON l."PhoneLineId" = j."PhoneLinesId" AND l."EndDate" IS NULL
		WHERE l."CustomerId" = $1 AND i."ProcessingMonthId" = $2`, customerID, processingMonthID).Scan(
		&periodStart, &periodEnd, &issueDate, &dueDate, &ctxRow.ServicesTotal, &ctxRow.DiscountsTotal,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	ctxRow.PeriodStart = periodStart
	ctxRow.PeriodEnd = periodEnd
	ctxRow.IssueDate = issueDate
	ctxRow.DueDate = dueDate
	if !periodEnd.IsZero() {
		ctxRow.ReferenceMonth = periodEnd.Format("01/2006")
	} else if !issueDate.IsZero() {
		ctxRow.ReferenceMonth = issueDate.Format("01/2006")
	}
	return &ctxRow, nil
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

type UnpaidSicrediBillingDocument struct {
	ID                   string
	OrganizationID       string
	InvoiceNumber        string
	CustomerName         string
	AccountsReceivableID *string
	Amount               float64
	SicrediNossoNumero   string
}

func (s *Store) ListUnpaidSicrediBillingDocuments(ctx context.Context, orgID string) ([]UnpaidSicrediBillingDocument, error) {
	q := `
		SELECT d."Id", d."OrganizationId", d."InvoiceNumber", c."Name",
			d."AccountsReceivableId", d."Amount", d."SicrediNossoNumero"
		FROM "CustomerBillingDocuments" d
		JOIN "Customers" c ON c."Id" = d."CustomerId"
		WHERE d."SicrediNossoNumero" IS NOT NULL
			AND d."SicrediBoletoStatus" = 'issued'
			AND d."SicrediPaidAt" IS NULL`
	args := []any{}
	if orgID != "" {
		q += ` AND d."OrganizationId" = $1`
		args = append(args, orgID)
	}
	q += ` ORDER BY d."CreatedAt" ASC`

	rows, err := s.q(ctx).Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []UnpaidSicrediBillingDocument
	for rows.Next() {
		var item UnpaidSicrediBillingDocument
		if err := rows.Scan(
			&item.ID, &item.OrganizationID, &item.InvoiceNumber, &item.CustomerName,
			&item.AccountsReceivableID, &item.Amount, &item.SicrediNossoNumero,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) GetBillingDocumentBySicrediNossoNumero(ctx context.Context, orgID, nossoNumero string) (*UnpaidSicrediBillingDocument, error) {
	var item UnpaidSicrediBillingDocument
	err := s.q(ctx).QueryRow(ctx, `
		SELECT d."Id", d."OrganizationId", d."InvoiceNumber", c."Name",
			d."AccountsReceivableId", d."Amount", d."SicrediNossoNumero"
		FROM "CustomerBillingDocuments" d
		JOIN "Customers" c ON c."Id" = d."CustomerId"
		WHERE d."OrganizationId" = $1 AND d."SicrediNossoNumero" = $2
			AND d."SicrediPaidAt" IS NULL`, orgID, nossoNumero).Scan(
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

func (s *Store) MarkSicrediBoletoPaid(ctx context.Context, orgID, documentID string, paidAt time.Time) error {
	tag, err := s.q(ctx).Exec(ctx, `
		UPDATE "CustomerBillingDocuments"
		SET "SicrediBoletoStatus" = 'paid',
			"SicrediPaidAt" = $3,
			"UpdatedAt" = $3
		WHERE "OrganizationId" = $1 AND "Id" = $2 AND "SicrediPaidAt" IS NULL`,
		orgID, documentID, paidAt)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (s *Store) IsBillingDocumentSicrediPaid(ctx context.Context, orgID, documentID string) (bool, error) {
	var paid bool
	err := s.q(ctx).QueryRow(ctx, `
		SELECT "SicrediPaidAt" IS NOT NULL
		FROM "CustomerBillingDocuments"
		WHERE "OrganizationId" = $1 AND "Id" = $2`, orgID, documentID).Scan(&paid)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	return paid, err
}
