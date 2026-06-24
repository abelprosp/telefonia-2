package services

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/invoicelayout"
	"github.com/luxus-connect/telefonia/api/internal/models"
	"github.com/luxus-connect/telefonia/api/internal/notifications"
	"github.com/luxus-connect/telefonia/api/internal/store"
)

func (s *Service) ListInvoiceLayoutTemplates(ctx context.Context, page httputil.PageSearch) ([]models.ListInvoiceLayoutTemplateResponse, int64, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, 0, err
	}
	return s.Store.ListInvoiceLayoutTemplates(ctx, orgID, page)
}

func (s *Service) GetInvoiceLayoutTemplate(ctx context.Context, id string) (*models.GetInvoiceLayoutTemplateResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	t, err := s.Store.GetInvoiceLayoutTemplate(ctx, orgID, id)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if t == nil {
		return nil, httputil.NotFoundError(notifications.BillingLayoutTemplateNotFound)
	}
	return t, nil
}

func (s *Service) CreateInvoiceLayoutTemplate(ctx context.Context, input models.CreateInvoiceLayoutTemplateInput) (*models.GetInvoiceLayoutTemplateResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	name := strings.TrimSpace(input.Name)
	code := strings.ToLower(strings.TrimSpace(input.Code))
	if name == "" || code == "" || len(input.ConfigJson) == 0 {
		return nil, httputil.ValidationError(notifications.BillingLayoutTemplateFieldsRequired)
	}
	if _, err := invoicelayout.ParseConfig(input.ConfigJson); err != nil {
		return nil, httputil.ValidationError(notifications.N("BILLING_LAYOUT_INVALID_CONFIG", "Invalid layout configuration JSON."))
	}
	exists, err := s.Store.InvoiceLayoutTemplateCodeExists(ctx, orgID, code, nil)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if exists {
		return nil, httputil.BusinessError(notifications.BillingLayoutTemplateCodeDuplicated)
	}
	active := true
	if input.Active != nil {
		active = *input.Active
	}
	id := uuid.New().String()
	now := time.Now().UTC()
	if err := s.Store.CreateInvoiceLayoutTemplate(ctx, id, orgID, name, code, input.ConfigJson, active, now); err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	return s.GetInvoiceLayoutTemplate(ctx, id)
}

func (s *Service) UpdateInvoiceLayoutTemplate(ctx context.Context, id string, input models.UpdateInvoiceLayoutTemplateInput) (*models.GetInvoiceLayoutTemplateResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	name := strings.TrimSpace(input.Name)
	if name == "" || len(input.ConfigJson) == 0 {
		return nil, httputil.ValidationError(notifications.BillingLayoutTemplateFieldsRequired)
	}
	if _, err := invoicelayout.ParseConfig(input.ConfigJson); err != nil {
		return nil, httputil.ValidationError(notifications.N("BILLING_LAYOUT_INVALID_CONFIG", "Invalid layout configuration JSON."))
	}
	active := true
	if input.Active != nil {
		active = *input.Active
	}
	now := time.Now().UTC()
	if err := s.Store.UpdateInvoiceLayoutTemplate(ctx, orgID, id, name, input.ConfigJson, active, now); err != nil {
		if isPgNoRows(err) {
			return nil, httputil.NotFoundError(notifications.BillingLayoutTemplateNotFound)
		}
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	return s.GetInvoiceLayoutTemplate(ctx, id)
}

func (s *Service) PreviewInvoiceLayout(ctx context.Context, input models.PreviewInvoiceLayoutInput) (*models.PreviewInvoiceLayoutResponse, error) {
	cfg, err := invoicelayout.ParseConfig(input.ConfigJson)
	if err != nil {
		return nil, httputil.ValidationError(notifications.N("BILLING_LAYOUT_INVALID_CONFIG", "Invalid layout configuration JSON."))
	}
	html := invoicelayout.Render(cfg, sampleLayoutRenderData())
	return &models.PreviewInvoiceLayoutResponse{Html: html}, nil
}

func sampleLayoutRenderData() invoicelayout.RenderData {
	return invoicelayout.RenderData{
		CustomerName:     "REDOBRAI ARTUR ABEL SCHMITZ",
		CustomerAddress:  "Rua Venancio Aires, 400, Igrejinha - Lajeado - RS - CEP: 95.910-674",
		CustomerPhone:    "(51) 99962-0231",
		InvoiceNumber:    "FAT-001",
		InvoiceAmount:    "R$ 0,00",
		InvoiceDueDate:   "10/06/2026",
		InvoiceIssueDate: "16/04/2026",
		ReferenceMonth:   "05/2026",
		PeriodStart:      "16/04/2026",
		PeriodEnd:        "15/05/2026",
		ServicesTotal:    "R$ 19,99",
		Discounts:        "R$ -24,99",
		LineItems: []invoicelayout.LineItem{
			{Description: "SMART ILIMITADO 3GB", Quantity: "1", Type: "Mensal", UnitPrice: "R$ 19,99", Total: "R$ 19,99"},
			{Description: "PACOTE DE TORPEDOS 800 torpedos", Quantity: "1", Type: "Mensal", UnitPrice: "R$ 0,00", Total: "R$ 0,00"},
			{Description: "Consumo", Quantity: "1", Type: "Mensal", UnitPrice: "R$ 0,00", Total: "R$ 0,00"},
			{Description: "DESCONTO DE PLANO", Quantity: "1", Type: "Mensal", UnitPrice: "R$ -24,99", Total: "R$ -24,99"},
		},
	}
}

func buildLayoutRenderDataFromReceivable(rec *store.ReceivableForBilling, invoiceNumber, address, phone string, providerCtx *store.CustomerProviderInvoiceLayoutContext) invoicelayout.RenderData {
	refMonth := rec.IssueDate.Format("01/2006")
	periodStart := rec.IssueDate
	periodEnd := rec.DueDate
	issueDate := rec.IssueDate
	dueDate := rec.DueDate
	servicesTotal := rec.Amount
	discounts := 0.0
	if providerCtx != nil {
		if !providerCtx.PeriodStart.IsZero() {
			periodStart = providerCtx.PeriodStart
		}
		if !providerCtx.PeriodEnd.IsZero() {
			periodEnd = providerCtx.PeriodEnd
		}
		if !providerCtx.IssueDate.IsZero() {
			issueDate = providerCtx.IssueDate
		}
		if !providerCtx.DueDate.IsZero() {
			dueDate = providerCtx.DueDate
		}
		if providerCtx.ReferenceMonth != "" {
			refMonth = providerCtx.ReferenceMonth
		}
		if providerCtx.ServicesTotal > 0 {
			servicesTotal = providerCtx.ServicesTotal
		}
		discounts = providerCtx.DiscountsTotal
	}
	amount := formatInvoiceMoney(rec.Amount)
	return invoicelayout.RenderData{
		CustomerName:     rec.CustomerName,
		CustomerAddress:  address,
		CustomerPhone:    phone,
		CustomerDocument: rec.CustomerDocument,
		InvoiceNumber:    invoiceNumber,
		InvoiceAmount:    amount,
		InvoiceDueDate:   formatInvoiceDate(dueDate),
		InvoiceIssueDate: formatInvoiceDate(issueDate),
		ReferenceMonth:   refMonth,
		PeriodStart:      formatInvoiceDate(periodStart),
		PeriodEnd:        formatInvoiceDate(periodEnd),
		ServicesTotal:    formatInvoiceMoney(servicesTotal),
		Discounts:        formatInvoiceMoney(discounts),
		Description:      rec.Description,
	}
}

func renderInvoiceLayoutBody(ctx context.Context, s *Service, orgID, layoutCode, invoiceNumber string, rec *store.ReceivableForBilling) (string, error) {
	layout, err := s.Store.GetInvoiceLayoutTemplateByCode(ctx, orgID, layoutCode)
	if err != nil {
		return "", err
	}
	if layout == nil {
		return "", httputil.NotFoundError(notifications.BillingLayoutTemplateNotFound)
	}
	cfg, err := invoicelayout.ParseConfig(layout.ConfigJson)
	if err != nil {
		return "", httputil.ValidationError(notifications.N("BILLING_LAYOUT_INVALID_CONFIG", "Invalid layout configuration JSON."))
	}
	address, _ := s.Store.GetCustomerAddressForBilling(ctx, rec.CustomerID)
	phone, _ := s.Store.GetCustomerPhoneForBilling(ctx, rec.CustomerID)

	var providerCtx *store.CustomerProviderInvoiceLayoutContext
	var billingItems []store.CustomerBillingItemRow
	if rec.ProcessingMonthID != nil && strings.TrimSpace(*rec.ProcessingMonthID) != "" {
		providerCtx, _ = s.Store.GetCustomerProviderInvoiceLayoutContext(ctx, rec.CustomerID, *rec.ProcessingMonthID)
		billingItems, err = s.Store.ListCustomerBillingItemsForProcessingMonth(ctx, rec.CustomerID, *rec.ProcessingMonthID)
	} else {
		billingItems, err = s.Store.ListCustomerBillingItems(ctx, rec.CustomerID)
	}
	if err != nil {
		return "", httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}

	data := buildLayoutRenderDataFromReceivable(rec, invoiceNumber, address, phone, providerCtx)
	var servicesSum, discountsSum float64
	for _, item := range billingItems {
		if item.Amount < 0 {
			discountsSum += item.Amount
		} else if item.Amount > 0 {
			servicesSum += item.Amount
		}
		if item.Amount == 0 {
			continue
		}
		amount := formatInvoiceMoney(item.Amount)
		data.LineItems = append(data.LineItems, invoicelayout.LineItem{
			Description: item.Description,
			Quantity:    "1",
			Type:        item.ItemType,
			UnitPrice:   amount,
			Total:       amount,
		})
	}
	if servicesSum > 0 || discountsSum < 0 {
		data.ServicesTotal = formatInvoiceMoney(servicesSum)
		data.Discounts = formatInvoiceMoney(discountsSum)
	}
	return invoicelayout.Render(cfg, data), nil
}
