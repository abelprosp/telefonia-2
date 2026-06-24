package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/luxus-connect/telefonia/api/internal/email"
	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/invoicelayout"
	"github.com/luxus-connect/telefonia/api/internal/models"
	"github.com/luxus-connect/telefonia/api/internal/notifications"
	"github.com/luxus-connect/telefonia/api/internal/store"
)

type invoiceTemplateData struct {
	CustomerName     string
	CustomerDocument string
	BillingEmail     string
	InvoiceNumber    string
	InvoiceAmount    string
	InvoiceDueDate   string
	InvoiceIssueDate string
	Description      string
}

func renderInvoiceTemplate(template string, data invoiceTemplateData) string {
	replacer := strings.NewReplacer(
		"{{customer.name}}", data.CustomerName,
		"{{customer.document}}", data.CustomerDocument,
		"{{customer.billing_email}}", data.BillingEmail,
		"{{invoice.number}}", data.InvoiceNumber,
		"{{invoice.amount}}", data.InvoiceAmount,
		"{{invoice.due_date}}", data.InvoiceDueDate,
		"{{invoice.issue_date}}", data.InvoiceIssueDate,
		"{{invoice.description}}", data.Description,
	)
	return replacer.Replace(template)
}

func formatInvoiceMoney(value float64) string {
	return formatMoneyBR(value)
}

func formatInvoiceDate(t time.Time) string {
	return t.Format("02/01/2006")
}

func (s *Service) ListInvoiceEmailTemplates(ctx context.Context, kind *string, page httputil.PageSearch) ([]models.ListInvoiceEmailTemplateResponse, int64, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, 0, err
	}
	if kind != nil && *kind != "" {
		k := strings.ToLower(strings.TrimSpace(*kind))
		kind = &k
	}
	return s.Store.ListInvoiceEmailTemplates(ctx, orgID, kind, page)
}

func (s *Service) GetInvoiceEmailTemplate(ctx context.Context, id string) (*models.GetInvoiceEmailTemplateResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	t, err := s.Store.GetInvoiceEmailTemplate(ctx, orgID, id)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if t == nil {
		return nil, httputil.NotFoundError(notifications.BillingEmailTemplateNotFound)
	}
	return t, nil
}

func (s *Service) CreateInvoiceEmailTemplate(ctx context.Context, input models.CreateInvoiceEmailTemplateInput) (*models.GetInvoiceEmailTemplateResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	name := strings.TrimSpace(input.Name)
	code := strings.ToLower(strings.TrimSpace(input.Code))
	kind := strings.ToLower(strings.TrimSpace(input.Kind))
	if kind == "" {
		kind = "billing_invoice"
	}
	subject := strings.TrimSpace(input.SubjectTemplate)
	body := strings.TrimSpace(input.BodyTemplateHtml)
	if name == "" || code == "" || subject == "" || body == "" {
		return nil, httputil.ValidationError(notifications.BillingEmailTemplateFieldsRequired)
	}
	exists, err := s.Store.InvoiceEmailTemplateCodeExists(ctx, orgID, code, nil)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if exists {
		return nil, httputil.BusinessError(notifications.BillingEmailTemplateCodeDuplicated)
	}
	active := true
	if input.Active != nil {
		active = *input.Active
	}
	id := uuid.New().String()
	now := time.Now().UTC()
	if err := s.Store.CreateInvoiceEmailTemplate(ctx, id, orgID, name, code, kind, subject, body, active, now); err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	return s.GetInvoiceEmailTemplate(ctx, id)
}

func (s *Service) UpdateInvoiceEmailTemplate(ctx context.Context, id string, input models.UpdateInvoiceEmailTemplateInput) (*models.GetInvoiceEmailTemplateResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	name := strings.TrimSpace(input.Name)
	subject := strings.TrimSpace(input.SubjectTemplate)
	body := strings.TrimSpace(input.BodyTemplateHtml)
	if name == "" || subject == "" || body == "" {
		return nil, httputil.ValidationError(notifications.BillingEmailTemplateFieldsRequired)
	}
	active := true
	if input.Active != nil {
		active = *input.Active
	}
	now := time.Now().UTC()
	if err := s.Store.UpdateInvoiceEmailTemplate(ctx, orgID, id, name, subject, body, active, now); err != nil {
		if isPgNoRows(err) {
			return nil, httputil.NotFoundError(notifications.BillingEmailTemplateNotFound)
		}
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	return s.GetInvoiceEmailTemplate(ctx, id)
}

func (s *Service) ListCustomerBillingDocuments(ctx context.Context, status, customerID *string, page httputil.PageSearch) ([]models.ListCustomerBillingDocumentResponse, int64, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, 0, err
	}
	if status != nil && *status != "" {
		s := strings.ToLower(strings.TrimSpace(*status))
		status = &s
	}
	return s.Store.ListCustomerBillingDocuments(ctx, orgID, status, customerID, page)
}

func (s *Service) GetCustomerBillingDocument(ctx context.Context, id string) (*models.GetCustomerBillingDocumentResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	doc, err := s.Store.GetCustomerBillingDocument(ctx, orgID, id)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if doc == nil {
		return nil, httputil.NotFoundError(notifications.BillingDocumentNotFound)
	}
	return doc, nil
}

func (s *Service) CreateCustomerBillingDocumentFromReceivable(ctx context.Context, receivableID, templateCode, layoutTemplateCode string) (*models.CreateCustomerBillingDocumentFromReceivableResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	rec, err := s.Store.GetReceivableForBilling(ctx, orgID, receivableID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if rec == nil {
		return nil, httputil.NotFoundError(notifications.FinancialReceivableNotFound)
	}
	id, err := s.createBillingDocumentFromReceivable(ctx, orgID, rec, templateCode, layoutTemplateCode)
	if err != nil {
		return nil, err
	}
	return &models.CreateCustomerBillingDocumentFromReceivableResponse{ID: id}, nil
}

func (s *Service) createBillingDocumentFromReceivable(ctx context.Context, orgID string, rec *store.ReceivableForBilling, templateCode, layoutTemplateCode string) (string, error) {
	if strings.TrimSpace(templateCode) == "" {
		templateCode = "default-billing-invoice"
	}
	tmpl, err := s.Store.GetInvoiceEmailTemplateByCode(ctx, orgID, templateCode)
	if err != nil {
		return "", httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if tmpl == nil {
		return "", httputil.NotFoundError(notifications.BillingEmailTemplateNotFound)
	}

	invoiceNumber, err := s.Store.NextBillingInvoiceNumber(ctx, orgID)
	if err != nil {
		return "", httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}

	data := invoiceTemplateData{
		CustomerName:     rec.CustomerName,
		CustomerDocument: rec.CustomerDocument,
		BillingEmail:     rec.BillingEmail,
		InvoiceNumber:    invoiceNumber,
		InvoiceAmount:    formatInvoiceMoney(rec.Amount),
		InvoiceDueDate:   formatInvoiceDate(rec.DueDate),
		InvoiceIssueDate: formatInvoiceDate(rec.IssueDate),
		Description:      rec.Description,
	}
	subject := renderInvoiceTemplate(tmpl.SubjectTemplate, data)
	body := renderInvoiceTemplate(tmpl.BodyTemplateHtml, data)
	if strings.TrimSpace(layoutTemplateCode) != "" {
		layoutBody, err := renderInvoiceLayoutBody(ctx, s, orgID, layoutTemplateCode, invoiceNumber, rec)
		if err != nil {
			return "", err
		}
		body = layoutBody
	}
	recipient := strings.TrimSpace(rec.BillingEmail)
	// E-mail opcional na criação; obrigatório apenas no envio.

	id := uuid.New().String()
	now := time.Now().UTC()
	recID := rec.ID
	row := models.CustomerBillingDocumentRow{
		ID:                   id,
		OrganizationID:       orgID,
		CustomerID:           rec.CustomerID,
		AccountsReceivableID: &recID,
		ProcessingMonthID:    rec.ProcessingMonthID,
		InvoiceNumber:        invoiceNumber,
		IssueDate:            rec.IssueDate,
		DueDate:              rec.DueDate,
		Amount:               rec.Amount,
		Status:               "draft",
		RecipientEmail:       recipient,
		EmailSubject:         subject,
		EmailBodyHTML:        body,
		CreatedAt:            now,
	}
	if err := s.Store.CreateCustomerBillingDocument(ctx, row); err != nil {
		return "", httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	s.tryAttachSicrediBoleto(ctx, orgID, id)
	return id, nil
}

func (s *Service) UpdateCustomerBillingDocument(ctx context.Context, id string, input models.UpdateCustomerBillingDocumentInput) (*models.GetCustomerBillingDocumentResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	recipient := strings.TrimSpace(input.RecipientEmail)
	subject := strings.TrimSpace(input.EmailSubject)
	body := strings.TrimSpace(input.EmailBodyHtml)
	status := strings.ToLower(strings.TrimSpace(input.Status))
	if subject == "" || body == "" {
		return nil, httputil.ValidationError(notifications.BillingDocumentFieldsRequired)
	}
	if status == "" {
		status = "draft"
	}
	now := time.Now().UTC()
	if err := s.Store.UpdateCustomerBillingDocument(ctx, orgID, id, recipient, subject, body, status, now); err != nil {
		if isPgNoRows(err) {
			return nil, httputil.NotFoundError(notifications.BillingDocumentNotFound)
		}
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	return s.GetCustomerBillingDocument(ctx, id)
}

func (s *Service) GetCustomerBillingDocumentDownload(ctx context.Context, documentID string) ([]byte, string, error) {
	doc, err := s.GetCustomerBillingDocument(ctx, documentID)
	if err != nil {
		return nil, "", err
	}
	body := strings.TrimSpace(doc.EmailBodyHtml)
	if body == "" {
		return nil, "", httputil.ValidationError(notifications.N("BILLING_DOCUMENT_EMPTY", "Esta fatura ainda não possui conteúdo."))
	}
	html := invoicelayout.EnsureHTMLDocument(body)
	filename := "fatura-" + doc.InvoiceNumber + ".html"
	return []byte(html), filename, nil
}

func (s *Service) SendCustomerBillingDocument(ctx context.Context, id string) (*models.SendCustomerBillingDocumentResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	user, err := userFrom(ctx)
	if err != nil {
		return nil, err
	}
	doc, err := s.GetCustomerBillingDocument(ctx, id)
	if err != nil {
		return nil, err
	}
	if doc.Status == "cancelled" {
		return nil, httputil.BusinessError(notifications.BillingDocumentCancelled)
	}
	if strings.TrimSpace(doc.RecipientEmail) == "" {
		return nil, httputil.ValidationError(notifications.BillingCustomerEmailRequired)
	}
	if s.Mailer == nil || !s.Mailer.Enabled() {
		return nil, httputil.BusinessError(notifications.BillingEmailNotConfigured)
	}

	now := time.Now().UTC()
	sendErr := s.Mailer.Send(ctx, email.Message{
		To:      doc.RecipientEmail,
		Subject: doc.EmailSubject,
		HTML:    doc.EmailBodyHtml,
	})

	logID := uuid.New().String()
	errMsg := ""
	success := sendErr == nil
	if sendErr != nil {
		errMsg = sendErr.Error()
	}
	_ = s.Store.InsertCustomerBillingSendLog(ctx, logID, orgID, id, doc.RecipientEmail, doc.EmailSubject, success, errMsg, user.ID, now)
	if !success {
		return nil, httputil.BusinessError(notifications.N("BILLING_EMAIL_SEND_FAILED", errMsg))
	}
	_ = s.Store.MarkCustomerBillingDocumentSent(ctx, orgID, id, now)
	return &models.SendCustomerBillingDocumentResponse{Success: true, Message: "E-mail enviado com sucesso."}, nil
}

func (s *Service) ListCustomerBillingSendLog(ctx context.Context, documentID string) ([]models.CustomerBillingSendLogResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	return s.Store.ListCustomerBillingSendLog(ctx, orgID, documentID)
}

func (s *Service) ListOverdueReceivables(ctx context.Context, page httputil.PageSearch) ([]models.OverdueReceivableResponse, int64, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, 0, err
	}
	return s.Store.ListOverdueReceivables(ctx, orgID, page)
}

func (s *Service) SendCollectionReminder(ctx context.Context, input models.SendCollectionReminderInput) (*models.SendCollectionReminderResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	user, err := userFrom(ctx)
	if err != nil {
		return nil, err
	}
	rec, err := s.Store.GetReceivableForBilling(ctx, orgID, strings.TrimSpace(input.AccountsReceivableID))
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if rec == nil {
		return nil, httputil.NotFoundError(notifications.FinancialReceivableNotFound)
	}
	if rec.Status != "overdue" {
		return nil, httputil.BusinessError(notifications.BillingReceivableNotOverdue)
	}
	recipient := strings.TrimSpace(rec.BillingEmail)
	if recipient == "" {
		return nil, httputil.ValidationError(notifications.BillingCustomerEmailRequired)
	}
	templateCode := strings.TrimSpace(input.TemplateCode)
	if templateCode == "" {
		templateCode = "default-collection-reminder"
	}
	tmpl, err := s.Store.GetInvoiceEmailTemplateByCode(ctx, orgID, templateCode)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if tmpl == nil {
		return nil, httputil.NotFoundError(notifications.BillingEmailTemplateNotFound)
	}
	if s.Mailer == nil || !s.Mailer.Enabled() {
		return nil, httputil.BusinessError(notifications.BillingEmailNotConfigured)
	}

	level := input.ReminderLevel
	if level <= 0 {
		level = 1
	}
	data := invoiceTemplateData{
		CustomerName:     rec.CustomerName,
		CustomerDocument: rec.CustomerDocument,
		BillingEmail:     rec.BillingEmail,
		InvoiceNumber:    fmt.Sprintf("CR-%s", rec.ID[:8]),
		InvoiceAmount:    formatInvoiceMoney(rec.Balance),
		InvoiceDueDate:   formatInvoiceDate(rec.DueDate),
		InvoiceIssueDate: formatInvoiceDate(rec.IssueDate),
		Description:      rec.Description,
	}
	subject := renderInvoiceTemplate(tmpl.SubjectTemplate, data)
	body := renderInvoiceTemplate(tmpl.BodyTemplateHtml, data)

	now := time.Now().UTC()
	sendErr := s.Mailer.Send(ctx, email.Message{To: recipient, Subject: subject, HTML: body})
	reminderID := uuid.New().String()
	status := "sent"
	var errPtr *string
	var sentAt *time.Time
	if sendErr != nil {
		status = "failed"
		msg := sendErr.Error()
		errPtr = &msg
	} else {
		sentAt = &now
	}
	_ = s.Store.InsertCollectionReminder(ctx, reminderID, orgID, rec.ID, level, recipient, subject, body, user.ID, status, errPtr, sentAt, now)
	if sendErr != nil {
		return nil, httputil.BusinessError(notifications.N("COLLECTION_REMINDER_SEND_FAILED", sendErr.Error()))
	}
	return &models.SendCollectionReminderResponse{Success: true, Message: "Lembrete de cobrança enviado."}, nil
}

func (s *Service) BulkBillingPreview(ctx context.Context, processingMonthID string, customerIDs []string) (*models.BulkBillingPreviewResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	monthID := strings.TrimSpace(processingMonthID)
	if monthID == "" {
		return nil, httputil.ValidationError(notifications.ProcessingMonthNotFound)
	}
	pm, err := s.Store.GetProcessingMonth(ctx, orgID, monthID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if pm == nil {
		return nil, httputil.NotFoundError(notifications.ProcessingMonthNotFound)
	}
	items, err := s.Store.ListBulkBillingCandidates(ctx, orgID, monthID, customerIDs)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	providerInvoices, err := s.Store.CountProviderInvoicesForMonth(ctx, orgID, monthID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	eligible := 0
	if providerInvoices == 0 {
		for i := range items {
			items[i].Eligible = false
			items[i].SkipReason = "no_provider_invoice"
		}
	} else {
		for _, item := range items {
			if item.Eligible {
				eligible++
			}
		}
	}
	return &models.BulkBillingPreviewResponse{
		ProcessingMonthID:     monthID,
		ProcessingMonthName:   pm.DisplayName,
		ProviderInvoicesCount: providerInvoices,
		Items:                 items,
		EligibleCount:         eligible,
	}, nil
}

func (s *Service) BulkGenerateBillingDocuments(ctx context.Context, input models.BulkGenerateBillingDocumentsInput) (*models.BulkGenerateBillingDocumentsResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	monthID := strings.TrimSpace(input.ProcessingMonthID)
	if monthID == "" {
		return nil, httputil.ValidationError(notifications.ProcessingMonthNotFound)
	}
	pm, err := s.Store.GetProcessingMonth(ctx, orgID, monthID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if pm == nil {
		return nil, httputil.NotFoundError(notifications.ProcessingMonthNotFound)
	}
	issueDate, err := parseFinancialDate(input.IssueDate)
	if err != nil {
		return nil, err
	}
	dueDate, err := parseFinancialDate(input.DueDate)
	if err != nil {
		return nil, err
	}
	templateCode := strings.TrimSpace(input.TemplateCode)
	if templateCode == "" {
		templateCode = "default-billing-invoice"
	}
	layoutCode := strings.TrimSpace(input.LayoutTemplateCode)
	if layoutCode == "" {
		layoutCode = "default-invoice-layout"
	}
	descriptionTemplate := strings.TrimSpace(input.Description)
	if descriptionTemplate == "" {
		descriptionTemplate = "Mensalidade telefonia — " + pm.DisplayName
	}

	candidates, err := s.Store.ListBulkBillingCandidates(ctx, orgID, monthID, input.CustomerIDs)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	providerInvoices, err := s.Store.CountProviderInvoicesForMonth(ctx, orgID, monthID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if providerInvoices == 0 {
		return nil, httputil.BusinessError(notifications.N("BILLING_NO_PROVIDER_INVOICE", "Nenhuma fatura da operadora importada para este mês de processamento."))
	}

	procMonthID := monthID
	return s.generateBillingDocumentsForCandidates(ctx, orgID, candidates, &procMonthID, issueDate, dueDate, descriptionTemplate, templateCode, layoutCode)
}

func bulkSkipReasonMessage(reason string) string {
	switch reason {
	case "no_billing_email":
		return "Cliente sem e-mail de cobrança."
	case "no_monthly_amount":
		return "Sem valor nas linhas da fatura da operadora ou aparelhos vinculados."
	case "no_lines_on_invoice":
		return "Nenhuma linha do cliente consta na fatura importada deste mês."
	case "no_provider_invoice":
		return "Nenhuma fatura da operadora importada para este mês."
	case "already_billed":
		return "Já existe fatura para este mês de processamento."
	case "no_active_lines":
		return "Cliente sem linhas ou aparelhos ativos vinculados."
	default:
		return reason
	}
}
