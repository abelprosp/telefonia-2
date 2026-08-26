package store

import (
	"context"
	"strings"
	"time"
	"unicode"

	"github.com/luxus-connect/telefonia/api/internal/models"
)

type AgentCustomerRow struct {
	ID           string
	Name         string
	LegalName    *string
	CpfCnpj      string
	Active       bool
	BillingEmail *string
	MatchReason  string
}

func (s *Store) SearchAgentCustomers(ctx context.Context, orgID string, input models.FinancialAgentLookupInput) ([]AgentCustomerRow, error) {
	query := strings.TrimSpace(firstNonEmptyStr(input.Query, input.Name, input.Document, input.Phone, input.LineNumber, input.InvoiceNumber, input.WhatsAppNumber))
	if query == "" {
		return nil, nil
	}
	digits := onlyDigitsAgent(query)
	like := "%" + query + "%"
	digitLike := ""
	if len(digits) >= 4 {
		suffix := digits
		if len(suffix) > 8 {
			suffix = suffix[len(suffix)-8:]
		}
		digitLike = "%" + suffix + "%"
	}

	q := `
		SELECT DISTINCT c."Id", c."Name", c."LegalName",
			COALESCE((SELECT cd."Number" FROM "CustomerDocuments" cd
				WHERE cd."CustomerId" = c."Id" AND cd."DocumentType" IN ('cpf','cnpj') LIMIT 1), ''),
			c."Active", c."BillingEmail",
			CASE
				WHEN $3 <> '' AND regexp_replace(COALESCE((
					SELECT cd."Number" FROM "CustomerDocuments" cd
					WHERE cd."CustomerId" = c."Id" AND cd."DocumentType" IN ('cpf','cnpj') LIMIT 1), ''), '[^0-9]', '', 'g') LIKE '%' || $3 || '%' THEN 'document'
				WHEN $4 <> '' AND EXISTS (
					SELECT 1 FROM "PhoneLineCustomerLinks" l
					JOIN "PhoneLines" pl ON pl."Id" = l."PhoneLineId"
					WHERE l."CustomerId" = c."Id" AND regexp_replace(pl."Number", '[^0-9]', '', 'g') LIKE $4
				) THEN 'phone_line'
				WHEN $5 <> '' AND EXISTS (
					SELECT 1 FROM "CustomerBillingDocuments" d
					WHERE d."CustomerId" = c."Id" AND d."OrganizationId" = c."OrganizationId"
						AND (d."InvoiceNumber" ILIKE $2 OR ($3 <> '' AND regexp_replace(COALESCE(d."SicrediNossoNumero",''), '[^0-9]', '', 'g') LIKE '%' || $3 || '%'))
				) THEN 'invoice'
				WHEN c."Name" ILIKE $2 OR COALESCE(c."LegalName",'') ILIKE $2 THEN 'name'
				WHEN COALESCE(c."BillingEmail",'') ILIKE $2 THEN 'email'
				ELSE 'query'
			END AS match_reason
		FROM "Customers" c
		WHERE c."OrganizationId" = $1
			AND (
				c."Name" ILIKE $2
				OR COALESCE(c."LegalName",'') ILIKE $2
				OR COALESCE(c."BillingEmail",'') ILIKE $2
				OR ($3 <> '' AND EXISTS (
					SELECT 1 FROM "CustomerDocuments" cd
					WHERE cd."CustomerId" = c."Id"
						AND regexp_replace(cd."Number", '[^0-9]', '', 'g') LIKE '%' || $3 || '%'
				))
				OR ($4 <> '' AND EXISTS (
					SELECT 1 FROM "PhoneLineCustomerLinks" l
					JOIN "PhoneLines" pl ON pl."Id" = l."PhoneLineId"
					WHERE l."CustomerId" = c."Id"
						AND regexp_replace(pl."Number", '[^0-9]', '', 'g') LIKE $4
				))
				OR EXISTS (
					SELECT 1 FROM "CustomerBillingDocuments" d
					WHERE d."CustomerId" = c."Id" AND d."OrganizationId" = c."OrganizationId"
						AND (d."InvoiceNumber" ILIKE $2
							OR ($3 <> '' AND regexp_replace(COALESCE(d."SicrediNossoNumero",''), '[^0-9]', '', 'g') LIKE '%' || $3 || '%')
							OR ($3 <> '' AND regexp_replace(COALESCE(d."SicrediLinhaDigitavel",''), '[^0-9]', '', 'g') LIKE '%' || $3 || '%'))
				)
			)
		ORDER BY c."Name"
		LIMIT 20`
	rows, err := s.q(ctx).Query(ctx, q, orgID, like, digits, digitLike, strings.TrimSpace(input.InvoiceNumber))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []AgentCustomerRow
	for rows.Next() {
		var item AgentCustomerRow
		if err := rows.Scan(&item.ID, &item.Name, &item.LegalName, &item.CpfCnpj, &item.Active, &item.BillingEmail, &item.MatchReason); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) ListAgentCustomerLineNumbers(ctx context.Context, orgID, customerID string) ([]string, error) {
	rows, err := s.q(ctx).Query(ctx, `
		SELECT DISTINCT pl."Number"
		FROM "PhoneLineCustomerLinks" l
		JOIN "PhoneLines" pl ON pl."Id" = l."PhoneLineId"
		JOIN "Customers" c ON c."Id" = l."CustomerId"
		WHERE c."OrganizationId" = $1 AND l."CustomerId" = $2 AND l."EndDate" IS NULL
		ORDER BY pl."Number"`, orgID, customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var numbers []string
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			return nil, err
		}
		numbers = append(numbers, n)
	}
	return numbers, rows.Err()
}

func (s *Store) CountAgentCustomerInvoiceStats(ctx context.Context, orgID, customerID string, asOf time.Time) (openCount, overdueCount int, overdueBalance float64, err error) {
	err = s.q(ctx).QueryRow(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE d."Status" <> 'cancelled' AND d."SicrediPaidAt" IS NULL)::int,
			COUNT(*) FILTER (WHERE d."Status" <> 'cancelled' AND d."SicrediPaidAt" IS NULL AND d."DueDate" < $3::date)::int,
			COALESCE(SUM(d."Amount") FILTER (WHERE d."Status" <> 'cancelled' AND d."SicrediPaidAt" IS NULL AND d."DueDate" < $3::date), 0)
		FROM "CustomerBillingDocuments" d
		WHERE d."OrganizationId" = $1 AND d."CustomerId" = $2`, orgID, customerID, asOf).
		Scan(&openCount, &overdueCount, &overdueBalance)
	return
}

func (s *Store) ListAgentInvoices(ctx context.Context, orgID string, customerID, invoiceNumber, status string, dueDate *time.Time, overdueOnly bool, asOf time.Time, limit int) ([]models.ListCustomerBillingDocumentResponse, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	base := `
		FROM "CustomerBillingDocuments" d
		JOIN "Customers" c ON c."Id" = d."CustomerId"
		LEFT JOIN "PhoneLines" pl ON pl."Id" = d."PhoneLineId"
		WHERE d."OrganizationId" = $1`
	args := []any{orgID}
	if customerID != "" {
		base += ` AND d."CustomerId" = $` + itoa(len(args)+1)
		args = append(args, customerID)
	}
	if invoiceNumber != "" {
		base += ` AND d."InvoiceNumber" ILIKE $` + itoa(len(args)+1)
		args = append(args, "%"+invoiceNumber+"%")
	}
	if status != "" {
		base += ` AND d."Status" = $` + itoa(len(args)+1) + `::customer_billing_document_status`
		args = append(args, status)
	}
	if dueDate != nil {
		base += ` AND d."DueDate" = $` + itoa(len(args)+1) + `::date`
		args = append(args, dueDate.Format("2006-01-02"))
	}
	if overdueOnly {
		base += ` AND d."DueDate" < $` + itoa(len(args)+1) + `::date
			AND d."Status" <> 'cancelled'
			AND d."SicrediPaidAt" IS NULL`
		args = append(args, asOf)
	}
	selectQ := `
		SELECT d."Id", d."CustomerId", c."Name", d."AccountsReceivableId", d."ProcessingMonthId",
			d."InvoiceNumber", d."IssueDate", d."DueDate", d."Amount", d."Status"::text,
			d."RecipientEmail", d."EmailSubject", d."SendCount", d."SentAt", d."LastSentAt", d."CreatedAt",
			d."SicrediNossoNumero", d."SicrediLinhaDigitavel", d."SicrediCodigoBarras",
			d."SicrediPixQrCode", d."SicrediPixTxId", d."SicrediBoletoStatus", d."SicrediBoletoError",
			d."SicrediPaidAt", d."PhoneLineId", d."BillingGroupType", pl."Number"
		` + base + `
		ORDER BY d."DueDate" DESC, d."CreatedAt" DESC
		LIMIT $` + itoa(len(args)+1)
	args = append(args, limit)
	rows, err := s.q(ctx).Query(ctx, selectQ, args...)
	if err != nil {
		return nil, err
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
			&item.SicrediPaidAt, &item.PhoneLineID, &item.BillingGroupType, &item.PhoneLineNumber,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func onlyDigitsAgent(s string) string {
	var b strings.Builder
	for _, r := range s {
		if unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func firstNonEmptyStr(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}
