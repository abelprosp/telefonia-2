package services

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/models"
	"github.com/luxus-connect/telefonia/api/internal/notifications"
)

func parseFinancialDate(value string) (time.Time, error) {
	v := strings.TrimSpace(value)
	if v == "" {
		return time.Time{}, httputil.ValidationError(notifications.FinancialDateRequired)
	}
	t, err := time.Parse("2006-01-02", v)
	if err != nil {
		return time.Time{}, httputil.ValidationError(notifications.FinancialDateInvalid)
	}
	return t, nil
}

func normalizeFinancialStatus(status string) string {
	return strings.ToLower(strings.TrimSpace(status))
}

func (s *Service) GetFinancialSummary(ctx context.Context) (*models.FinancialSummaryResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	summary, err := s.Store.GetFinancialSummary(ctx, orgID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	return summary, nil
}

func (s *Service) ListAccountsPayable(ctx context.Context, status *string, page httputil.PageSearch) ([]models.ListAccountPayableResponse, int64, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, 0, err
	}
	if status != nil && *status != "" {
		n := normalizeFinancialStatus(*status)
		status = &n
	}
	return s.Store.ListAccountsPayable(ctx, orgID, status, page)
}

func (s *Service) CreateAccountPayable(ctx context.Context, input models.CreateAccountPayableInput) (*models.ListAccountPayableResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(input.Description) == "" || strings.TrimSpace(input.VendorName) == "" {
		return nil, httputil.ValidationError(notifications.FinancialDescriptionRequired)
	}
	if input.Amount <= 0 {
		return nil, httputil.ValidationError(notifications.FinancialAmountInvalid)
	}
	issueDate, err := parseFinancialDate(input.IssueDate)
	if err != nil {
		return nil, err
	}
	dueDate, err := parseFinancialDate(input.DueDate)
	if err != nil {
		return nil, err
	}
	id := uuid.New().String()
	now := time.Now().UTC()
	if err := s.Store.CreateAccountPayable(ctx, id, orgID, strings.TrimSpace(input.Description), strings.TrimSpace(input.VendorName),
		input.ProviderInvoiceID, input.PartnerSalespersonUserID, issueDate, dueDate, input.Amount, input.Notes, now); err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	items, _, err := s.Store.ListAccountsPayable(ctx, orgID, nil, httputil.PageSearch{PageIndex: 0, PageSize: 1})
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	for _, item := range items {
		if item.ID == id {
			return &item, nil
		}
	}
	return &models.ListAccountPayableResponse{ID: id}, nil
}

func (s *Service) CreateAccountPayableFromInvoice(ctx context.Context, invoiceID string) (*models.CreateAccountPayableFromInvoiceResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	ok, err := s.Store.ProviderInvoiceExistsInOrg(ctx, orgID, invoiceID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if !ok {
		return nil, httputil.NotFoundError(notifications.InvoiceNotFound)
	}
	exists, err := s.Store.PayableExistsForProviderInvoice(ctx, orgID, invoiceID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if exists {
		return nil, httputil.BusinessError(notifications.FinancialPayableFromInvoiceExists)
	}
	vendor, desc, dueDate, amount, err := s.Store.GetProviderInvoiceForPayable(ctx, orgID, invoiceID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	id := uuid.New().String()
	now := time.Now().UTC()
	invID := invoiceID
	if err := s.Store.CreateAccountPayable(ctx, id, orgID, desc, vendor, &invID, nil, now, dueDate, amount, nil, now); err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	return &models.CreateAccountPayableFromInvoiceResponse{ID: id}, nil
}

func (s *Service) UpdateAccountPayable(ctx context.Context, id string, input models.UpdateAccountPayableInput) error {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return err
	}
	dueDate, err := parseFinancialDate(input.DueDate)
	if err != nil {
		return err
	}
	status := normalizeFinancialStatus(input.Status)
	if err := s.Store.UpdateAccountPayable(ctx, orgID, id, strings.TrimSpace(input.Description), strings.TrimSpace(input.VendorName),
		dueDate, input.Amount, status, input.Notes, time.Now().UTC()); err != nil {
		if isPgNoRows(err) {
			return httputil.NotFoundError(notifications.FinancialPayableNotFound)
		}
		return httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	return nil
}

func (s *Service) RegisterPayablePayment(ctx context.Context, id string, input models.RegisterFinancialPaymentInput) error {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return err
	}
	user, err := userFrom(ctx)
	if err != nil {
		return err
	}
	if input.Amount <= 0 {
		return httputil.ValidationError(notifications.FinancialAmountInvalid)
	}
	paymentDate, err := parseFinancialDate(input.PaymentDate)
	if err != nil {
		return err
	}
	if err := s.Store.RegisterPayablePayment(ctx, uuid.New().String(), orgID, id, user.ID, input.Amount, paymentDate, input.Reference, input.Notes, time.Now().UTC()); err != nil {
		if isPgNoRows(err) {
			return httputil.NotFoundError(notifications.FinancialPayableNotFound)
		}
		return httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	return nil
}

func (s *Service) ListAccountsReceivable(ctx context.Context, customerID, status *string, page httputil.PageSearch) ([]models.ListAccountReceivableResponse, int64, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, 0, err
	}
	if status != nil && *status != "" {
		n := normalizeFinancialStatus(*status)
		status = &n
	}
	return s.Store.ListAccountsReceivable(ctx, orgID, customerID, status, page)
}

func (s *Service) CreateAccountReceivable(ctx context.Context, input models.CreateAccountReceivableInput) (*models.ListAccountReceivableResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(input.CustomerID) == "" {
		return nil, httputil.ValidationError(notifications.CustomerNotFound)
	}
	if strings.TrimSpace(input.Description) == "" {
		return nil, httputil.ValidationError(notifications.FinancialDescriptionRequired)
	}
	if input.Amount <= 0 {
		return nil, httputil.ValidationError(notifications.FinancialAmountInvalid)
	}
	ok, err := s.Store.CustomerExistsInOrg(ctx, orgID, input.CustomerID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if !ok {
		return nil, httputil.NotFoundError(notifications.CustomerNotFound)
	}
	issueDate, err := parseFinancialDate(input.IssueDate)
	if err != nil {
		return nil, err
	}
	dueDate, err := parseFinancialDate(input.DueDate)
	if err != nil {
		return nil, err
	}
	id := uuid.New().String()
	now := time.Now().UTC()
	if err := s.Store.CreateAccountReceivable(ctx, id, orgID, input.CustomerID, strings.TrimSpace(input.Description),
		input.ProcessingMonthID, issueDate, dueDate, input.Amount, input.Notes, now); err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	items, _, err := s.Store.ListAccountsReceivable(ctx, orgID, &input.CustomerID, nil, httputil.PageSearch{PageIndex: 0, PageSize: 100})
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	for _, item := range items {
		if item.ID == id {
			return &item, nil
		}
	}
	return &models.ListAccountReceivableResponse{ID: id, CustomerID: input.CustomerID}, nil
}

func (s *Service) UpdateAccountReceivable(ctx context.Context, id string, input models.UpdateAccountReceivableInput) error {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return err
	}
	dueDate, err := parseFinancialDate(input.DueDate)
	if err != nil {
		return err
	}
	status := normalizeFinancialStatus(input.Status)
	if err := s.Store.UpdateAccountReceivable(ctx, orgID, id, strings.TrimSpace(input.Description), dueDate, input.Amount, status, input.Notes, time.Now().UTC()); err != nil {
		if isPgNoRows(err) {
			return httputil.NotFoundError(notifications.FinancialReceivableNotFound)
		}
		return httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	return nil
}

func (s *Service) RegisterReceivablePayment(ctx context.Context, id string, input models.RegisterFinancialPaymentInput) error {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return err
	}
	user, err := userFrom(ctx)
	if err != nil {
		return err
	}
	if input.Amount <= 0 {
		return httputil.ValidationError(notifications.FinancialAmountInvalid)
	}
	paymentDate, err := parseFinancialDate(input.PaymentDate)
	if err != nil {
		return err
	}
	if err := s.Store.RegisterReceivablePayment(ctx, uuid.New().String(), orgID, id, user.ID, input.Amount, paymentDate, input.Reference, input.Notes, time.Now().UTC()); err != nil {
		if isPgNoRows(err) {
			return httputil.NotFoundError(notifications.FinancialReceivableNotFound)
		}
		return httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	return nil
}

func (s *Service) ListPartnerSales(ctx context.Context, salespersonUserID, status *string, page httputil.PageSearch) ([]models.ListPartnerSaleResponse, int64, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, 0, err
	}
	if status != nil && *status != "" {
		n := strings.ToLower(strings.TrimSpace(*status))
		status = &n
	}
	return s.Store.ListPartnerSales(ctx, orgID, salespersonUserID, status, page)
}

func (s *Service) SyncPartnerSales(ctx context.Context, input models.SyncPartnerSalesInput) (*models.SyncPartnerSalesResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	refMonth, err := parseFinancialDate(input.ReferenceMonth)
	if err != nil {
		return nil, err
	}
	refMonth = time.Date(refMonth.Year(), refMonth.Month(), 1, 0, 0, 0, 0, time.UTC)
	pct, err := s.Store.GetPartnerCommissionPercent(ctx, orgID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	now := time.Now().UTC()
	count, err := s.Store.SyncPartnerSalesFromLines(ctx, orgID, refMonth, pct, now, func() string { return uuid.New().String() })
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	return &models.SyncPartnerSalesResponse{InsertedCount: count}, nil
}

func (s *Service) UpdatePartnerSaleStatus(ctx context.Context, id string, input models.UpdatePartnerSaleStatusInput) (*models.ListPartnerSaleResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	status := strings.ToLower(strings.TrimSpace(input.Status))
	if status != "approved" && status != "paid" && status != "cancelled" && status != "accrued" {
		return nil, httputil.ValidationError(notifications.FinancialPartnerSaleStatusInvalid)
	}

	sale, err := s.Store.GetPartnerSaleByID(ctx, orgID, id)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if sale == nil {
		return nil, httputil.NotFoundError(notifications.FinancialPartnerSaleNotFound)
	}

	now := time.Now().UTC()
	var accountPayableID *string

	if status == "approved" && sale.Status == "accrued" {
		apID := uuid.New().String()
		desc := "Comissão parceiro - " + sale.CustomerName + " - linha " + sale.PhoneLineNumber
		due := now.AddDate(0, 0, 30)
		partnerID := sale.SalespersonUserID
		if err := s.Store.CreateAccountPayable(ctx, apID, orgID, desc, "Parceiro comercial", nil, &partnerID, now, due, sale.CommissionAmount, nil, now); err != nil {
			return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
		}
		accountPayableID = &apID
	} else if status == "paid" && sale.AccountPayableID != nil {
		accountPayableID = sale.AccountPayableID
		_ = s.Store.RegisterPayablePayment(ctx, uuid.New().String(), orgID, *sale.AccountPayableID, sale.SalespersonUserID, sale.CommissionAmount, now, nil, nil, now)
	} else {
		accountPayableID = sale.AccountPayableID
	}

	if err := s.Store.UpdatePartnerSaleStatus(ctx, orgID, id, status, accountPayableID, now); err != nil {
		if isPgNoRows(err) {
			return nil, httputil.NotFoundError(notifications.FinancialPartnerSaleNotFound)
		}
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	return s.Store.GetPartnerSaleByID(ctx, orgID, id)
}

func (s *Service) GetPartnerCommissionSettings(ctx context.Context) (*models.PartnerCommissionSettingsResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	item, err := s.Store.GetPartnerCommissionSettings(ctx, orgID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	return item, nil
}

func (s *Service) UpdatePartnerCommissionSettings(ctx context.Context, input models.UpdatePartnerCommissionSettingsInput) (*models.PartnerCommissionSettingsResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	if input.DefaultCommissionPercent < 0 || input.DefaultCommissionPercent > 100 {
		return nil, httputil.ValidationError(notifications.FinancialCommissionPercentInvalid)
	}
	now := time.Now().UTC()
	if err := s.Store.UpsertPartnerCommissionSettings(ctx, orgID, input.DefaultCommissionPercent, now); err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	return s.Store.GetPartnerCommissionSettings(ctx, orgID)
}

func (s *Service) PartnerGetFinancialSummary(ctx context.Context) (*models.PartnerFinancialSummaryResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	user, err := partnerUserFrom(ctx)
	if err != nil {
		return nil, err
	}
	summary, err := s.Store.GetPartnerFinancialSummary(ctx, orgID, user.ID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	return summary, nil
}

func (s *Service) PartnerListSales(ctx context.Context, page httputil.PageSearch) ([]models.ListPartnerSaleResponse, int64, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, 0, err
	}
	user, err := partnerUserFrom(ctx)
	if err != nil {
		return nil, 0, err
	}
	return s.Store.ListPartnerSales(ctx, orgID, &user.ID, nil, page)
}
