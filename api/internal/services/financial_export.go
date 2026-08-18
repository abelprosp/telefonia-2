package services

import (
	"bytes"
	"context"
	"encoding/csv"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/models"
	"github.com/luxus-connect/telefonia/api/internal/notifications"
)

func (s *Service) ExportFinancialBilling(ctx context.Context, processingMonthID string) (*models.FinancialExportResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	pm, err := s.Store.GetProcessingMonth(ctx, orgID, processingMonthID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if pm == nil {
		return nil, httputil.NotFoundError(notifications.ProcessingMonthNotFound)
	}
	candidates, err := s.Store.ListBulkBillingCandidates(ctx, orgID, processingMonthID, nil)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	s.applyMonthlyBillingRules(ctx, orgID, processingMonthID, pm, candidates)
	items := make([]models.FinancialExportItem, 0, len(candidates))
	var total float64
	for _, c := range candidates {
		item := models.FinancialExportItem{
			BillingGroupID:   c.BillingGroupID,
			BillingGroupType: c.BillingGroupType,
			GroupLabel:       c.GroupLabel,
			CustomerID:       c.CustomerID,
			CustomerName:     c.CustomerName,
			CustomerDocument: c.CustomerDocument,
			PhoneLineID:      c.PhoneLineID,
			PhoneLineNumber:  c.PhoneLineNumber,
			Amount:           c.MonthlyAmount,
			LineCount:        c.LineCount,
			DeviceCount:      c.DeviceCount,
			AlreadyBilled:    c.AlreadyBilled,
		}
		total += c.MonthlyAmount
		items = append(items, item)
	}
	return &models.FinancialExportResponse{
		ProcessingMonthID:   pm.ID,
		ProcessingMonthName: pm.DisplayName,
		Year:                pm.Year,
		Month:               pm.Month,
		GeneratedAt:         time.Now().UTC(),
		Items:               items,
		TotalAmount:         total,
		ItemCount:           len(items),
	}, nil
}

func FinancialExportCSV(data *models.FinancialExportResponse) []byte {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	_ = w.Write([]string{
		"processing_month_id", "processing_month_name", "billing_group_id", "billing_group_type",
		"group_label", "customer_id", "customer_name", "customer_document",
		"phone_line_id", "phone_line_number", "amount", "line_count", "device_count", "already_billed",
	})
	for _, item := range data.Items {
		_ = w.Write([]string{
			data.ProcessingMonthID,
			data.ProcessingMonthName,
			item.BillingGroupID,
			item.BillingGroupType,
			item.GroupLabel,
			item.CustomerID,
			item.CustomerName,
			item.CustomerDocument,
			deref(item.PhoneLineID),
			deref(item.PhoneLineNumber),
			strconv.FormatFloat(item.Amount, 'f', 2, 64),
			strconv.Itoa(item.LineCount),
			strconv.Itoa(item.DeviceCount),
			strconv.FormatBool(item.AlreadyBilled),
		})
	}
	w.Flush()
	return buf.Bytes()
}

func (s *Service) PushFinancialExportSFTP(ctx context.Context, processingMonthID string) (*models.FinancialSFTPPushResponse, error) {
	host := strings.TrimSpace(os.Getenv("FINANCIAL_SFTP_HOST"))
	user := strings.TrimSpace(os.Getenv("FINANCIAL_SFTP_USER"))
	pass := os.Getenv("FINANCIAL_SFTP_PASSWORD")
	if host == "" || user == "" || pass == "" {
		return &models.FinancialSFTPPushResponse{
			Status:  "not_configured",
			Message: "SFTP não configurado. Defina FINANCIAL_SFTP_HOST, FINANCIAL_SFTP_USER e FINANCIAL_SFTP_PASSWORD. Exportação disponível via API e CSV.",
		}, nil
	}
	return &models.FinancialSFTPPushResponse{
		Status:  "unavailable",
		Message: "Credenciais SFTP presentes, mas o envio SFTP não está habilitado neste ambiente (não simula sucesso). Use CSV ou a API JSON.",
	}, nil
}

func (s *Service) ListGeneratedContractsForCustomer(ctx context.Context, customerID string) ([]models.GeneratedContractResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	ok, err := s.Store.CustomerExistsInOrg(ctx, orgID, customerID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if !ok {
		return nil, httputil.NotFoundError(notifications.CustomerNotFound)
	}
	items, err := s.Store.ListGeneratedContractsForCustomer(ctx, orgID, customerID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	return items, nil
}

func (s *Service) ListGeneratedContractsForPhoneLine(ctx context.Context, phoneLineID string) ([]models.GeneratedContractResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := s.GetPhoneLine(ctx, phoneLineID); err != nil {
		return nil, err
	}
	items, err := s.Store.ListGeneratedContractsForPhoneLine(ctx, orgID, phoneLineID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	return items, nil
}

func deref(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func (s *Service) maybeGenerateAutomaticContract(ctx context.Context, orgID, customerID, phoneLineID, trigger, extraDesc string) {
	templateID, err := s.Store.GetFirstActiveContractTemplateID(ctx, orgID)
	if err != nil || templateID == "" {
		return
	}
	template, err := s.Store.GetContractTemplate(ctx, orgID, templateID)
	if err != nil || template == nil {
		return
	}
	customer, err := s.Store.GetCustomerContractData(ctx, orgID, customerID)
	if err != nil || customer == nil {
		return
	}
	tags := map[string]string{
		"{{customer.name}}":                 customer.Name,
		"{{customer.legal_name}}":           derefString(customer.LegalName, customer.Name),
		"{{customer.document}}":             customer.Document,
		"{{customer.type}}":                 customer.Type,
		"{{customer.address.full}}":         formatFullAddress(customer),
		"{{customer.address.street}}":       customer.Street,
		"{{customer.address.number}}":       customer.Number,
		"{{customer.address.neighborhood}}": customer.Neighborhood,
		"{{customer.address.city}}":         customer.City,
		"{{customer.address.state}}":        customer.State,
		"{{customer.address.zip_code}}":     customer.ZipCode,
		"{{customer.address.country}}":      customer.Country,
		"{{sale.sale_number}}":              "AUTO-" + strings.ToUpper(trigger),
		"{{sale.total_amount}}":             "—",
		"{{sale.sold_at}}":                  time.Now().UTC().Format("02/01/2006"),
		"{{sale.items_table}}":              "<p>" + htmlEscape(extraDesc) + "</p>",
		"{{salesperson.name}}":              salespersonDisplayName(userFromSafe(ctx)),
		"{{line.number}}":                   phoneLineID,
		"{{trigger}}":                       trigger,
	}
	rendered := template.BodyTemplate
	for k, v := range tags {
		rendered = strings.ReplaceAll(rendered, k, v)
	}
	now := time.Now().UTC()
	cid := customerID
	var pid *string
	if strings.TrimSpace(phoneLineID) != "" {
		pid = &phoneLineID
	}
	_ = s.Store.SaveGeneratedContractWithTrigger(ctx, uuid.New().String(), orgID, nil, &cid, pid, templateID, trigger, rendered, now)
}
