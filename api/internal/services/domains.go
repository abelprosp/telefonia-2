package services

import (
	"context"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/models"
	"github.com/luxus-connect/telefonia/api/internal/notifications"
	"github.com/luxus-connect/telefonia/api/internal/store"
)

func (s *Service) ListPhoneLines(ctx context.Context, status *string, page httputil.PageSearch) ([]models.ListPhoneLineResponse, int64, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, 0, err
	}
	if status != nil && *status != "" {
		normalized := httputil.NormalizePhoneLineStatus(*status)
		status = &normalized
	}
	return s.Store.ListPhoneLines(ctx, orgID, status, page)
}

func (s *Service) GetPhoneLine(ctx context.Context, id string) (*models.GetPhoneLineResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	pl, err := s.Store.GetPhoneLine(ctx, orgID, id)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if pl == nil {
		return nil, httputil.NotFoundError(notifications.PhoneLineNotFound)
	}
	return pl, nil
}

func (s *Service) CreateStockPhoneLine(ctx context.Context, input models.CreateStockPhoneLineInput) (*models.GetPhoneLineResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}

	number := httputil.NormalizeDigits(strings.TrimSpace(input.Number))
	if number == "" {
		return nil, httputil.ValidationError(notifications.PhoneLineNumberRequired)
	}

	providerID := strings.TrimSpace(input.ProviderID)
	planID := strings.TrimSpace(input.ProviderPlanID)
	accountNumber := strings.TrimSpace(input.ProviderAccountNumber)
	if providerID == "" || planID == "" || accountNumber == "" {
		return nil, httputil.ValidationError(notifications.N("REQUEST_VALIDATION", "Provider, account and plan are required."))
	}

	provider, err := s.Store.GetProvider(ctx, orgID, providerID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if provider == nil {
		return nil, httputil.NotFoundError(notifications.ProviderNotFound)
	}

	planOK, err := s.Store.ProviderPlanExistsForProvider(ctx, orgID, providerID, planID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if !planOK {
		return nil, httputil.ValidationError(notifications.PhoneLineProviderPlanInvalid)
	}

	account, err := s.Store.GetProviderAccountByProviderAndNumber(ctx, orgID, providerID, accountNumber)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if account == nil {
		return nil, httputil.ValidationError(notifications.PhoneLineProviderAccountNotFound)
	}

	existing, err := s.Store.GetPhoneLineByNumber(ctx, number)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if existing != nil {
		if existing.ProviderAccountID != account.ID {
			return nil, httputil.BusinessError(notifications.PhoneLineNumberDuplicated)
		}
		if existing.Status != "in_stock" {
			_ = s.Store.UpdatePhoneLineStatus(ctx, existing.ID, "in_stock")
		}
		return s.GetPhoneLine(ctx, existing.ID)
	}

	id := uuid.New().String()
	if err := s.Store.CreatePhoneLine(ctx, id, planID, account.ID, number); err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	return s.GetPhoneLine(ctx, id)
}

func (s *Service) ListPhoneLineCustomerLinks(ctx context.Context, phoneLineID string) ([]models.PhoneLineCustomerLinkResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	return s.Store.ListPhoneLineCustomerLinks(ctx, orgID, phoneLineID)
}

func (s *Service) AssignPhoneLineCustomer(ctx context.Context, phoneLineID string, input models.AssignPhoneLineCustomerInput) (*models.PhoneLineCustomerLinkResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	start := time.Now().UTC()
	if input.StartDate != nil {
		start = input.StartDate.UTC()
	}
	if _, err := s.GetPhoneLine(ctx, phoneLineID); err != nil {
		return nil, err
	}
	ok, err := s.Store.CustomerExistsInOrg(ctx, orgID, input.CustomerID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if !ok {
		return nil, httputil.NotFoundError(notifications.CustomerNotFound)
	}
	providerID, err := s.Store.GetPhoneLineProviderID(ctx, orgID, phoneLineID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	hasProvider, err := s.Store.CustomerHasActiveProvider(ctx, input.CustomerID, providerID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if !hasProvider {
		_ = s.Store.AddCustomerProviderLink(ctx, input.CustomerID, providerID, start)
	}
	_, prevCustomerID, _ := s.Store.GetActivePhoneLineCustomerLink(ctx, phoneLineID)
	if err := s.Store.AssignPhoneLineCustomer(ctx, phoneLineID, input.CustomerID, start); err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	_ = s.Store.ReactivateCustomer(ctx, input.CustomerID)
	if prevCustomerID != "" && prevCustomerID != input.CustomerID {
		hasOther, _ := s.Store.CustomerHasOtherActivePhoneLines(ctx, orgID, prevCustomerID, phoneLineID)
		if !hasOther {
			_ = s.Store.InactivateCustomer(ctx, prevCustomerID)
		}
	}
	return s.activePhoneLineCustomerLink(ctx, orgID, phoneLineID)
}

func (s *Service) TransferPhoneLineCustomer(ctx context.Context, phoneLineID string, input models.TransferPhoneLineCustomerInput) (*models.PhoneLineCustomerLinkResponse, error) {
	if _, err := orgFrom(ctx); err != nil {
		return nil, err
	}
	transferDate := time.Now().UTC()
	if input.TransferDate != nil {
		transferDate = input.TransferDate.UTC()
	}
	if _, err := s.GetPhoneLine(ctx, phoneLineID); err != nil {
		return nil, err
	}
	_, activeCustomerID, err := s.Store.GetActivePhoneLineCustomerLink(ctx, phoneLineID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if activeCustomerID == "" {
		return nil, httputil.BusinessError(notifications.PhoneLineActiveCustomerLinkNotFound)
	}
	if activeCustomerID == input.CustomerID {
		return nil, httputil.BusinessError(notifications.PhoneLineCustomerTransferSame)
	}
	assignInput := models.AssignPhoneLineCustomerInput{CustomerID: input.CustomerID, StartDate: &transferDate}
	return s.AssignPhoneLineCustomer(ctx, phoneLineID, assignInput)
}

func (s *Service) UnassignPhoneLineCustomer(ctx context.Context, phoneLineID string, input models.UnassignPhoneLineCustomerInput) error {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return err
	}
	end := time.Now().UTC()
	if input.EndDate != nil {
		end = input.EndDate.UTC()
	}
	_, activeCustomerID, err := s.Store.GetActivePhoneLineCustomerLink(ctx, phoneLineID)
	if err != nil {
		return httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if activeCustomerID == "" {
		return httputil.BusinessError(notifications.PhoneLineActiveCustomerLinkNotFound)
	}
	if err := s.Store.UnassignPhoneLineCustomer(ctx, phoneLineID, end); err != nil {
		if isPgNoRows(err) {
			return httputil.BusinessError(notifications.PhoneLineActiveCustomerLinkNotFound)
		}
		return httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	hasOther, _ := s.Store.CustomerHasOtherActivePhoneLines(ctx, orgID, activeCustomerID, phoneLineID)
	if !hasOther {
		_ = s.Store.InactivateCustomer(ctx, activeCustomerID)
	}
	return nil
}

func (s *Service) activePhoneLineCustomerLink(ctx context.Context, orgID, phoneLineID string) (*models.PhoneLineCustomerLinkResponse, error) {
	links, err := s.Store.ListPhoneLineCustomerLinks(ctx, orgID, phoneLineID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	for _, l := range links {
		if l.IsActive {
			return &l, nil
		}
	}
	return nil, httputil.BusinessError(notifications.PhoneLineActiveCustomerLinkNotFound)
}

func (s *Service) ListBillingCycles(ctx context.Context, page httputil.PageSearch) ([]models.ListBillingCycleResponse, int64, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, 0, err
	}
	return s.Store.ListBillingCycles(ctx, orgID, page)
}

func (s *Service) GetBillingCycle(ctx context.Context, id string) (*models.GetBillingCycleResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	bc, err := s.Store.GetBillingCycle(ctx, orgID, id)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if bc == nil {
		return nil, httputil.NotFoundError(notifications.BillingCycleNotFound)
	}
	return bc, nil
}

func (s *Service) CreateBillingCycle(ctx context.Context, input models.CreateBillingCycleInput) (*models.CreateBillingCycleResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(input.Code) == "" {
		return nil, httputil.ValidationError(notifications.BillingCycleCodeRequired)
	}
	if strings.TrimSpace(input.Name) == "" {
		return nil, httputil.ValidationError(notifications.BillingCycleNameRequired)
	}
	if input.StartDate.IsZero() {
		return nil, httputil.ValidationError(notifications.N("BILLING_CYCLE_START_DATE_REQUIRED", "Start date is required."))
	}
	if input.EndDate.IsZero() {
		return nil, httputil.ValidationError(notifications.N("BILLING_CYCLE_END_DATE_REQUIRED", "End date is required."))
	}
	if input.EndDate.Before(input.StartDate.Time) {
		return nil, httputil.ValidationError(notifications.N("BILLING_CYCLE_DATE_RANGE_INVALID", "End date must be on or after start date."))
	}
	ok, err := s.Store.ProviderExists(ctx, orgID, input.ProviderID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if !ok {
		return nil, httputil.BusinessError(notifications.ProviderNotFound)
	}
	blocked, err := s.Store.ExistsClosedProcessingMonthIntersecting(ctx, orgID, input.ProviderID, input.StartDate.Time, input.EndDate.Time)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if blocked {
		return nil, httputil.BusinessError(notifications.ProcessingMonthRetroactiveBlocked)
	}
	id := uuid.New().String()
	if err := s.Store.CreateBillingCycle(ctx, orgID, id, input.ProviderID, input.Code, input.Name, input.StartDate.Time, input.EndDate.Time); err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	return s.GetBillingCycle(ctx, id)
}

func (s *Service) UpdateBillingCycle(ctx context.Context, id string, input models.UpdateBillingCycleInput) error {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return err
	}
	existing, err := s.GetBillingCycle(ctx, id)
	if err != nil {
		return err
	}
	if existing.Status == "closed" {
		return httputil.BusinessError(notifications.BillingCycleConsolidated)
	}
	blocked, err := s.Store.ExistsClosedProcessingMonthIntersecting(ctx, orgID, input.ProviderID, input.StartDate.Time, input.EndDate.Time)
	if err != nil {
		return httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if blocked {
		return httputil.BusinessError(notifications.ProcessingMonthRetroactiveBlocked)
	}
	if err := s.Store.UpdateBillingCycle(ctx, orgID, id, input.ProviderID, input.Code, input.Name, input.StartDate.Time, input.EndDate.Time); err != nil {
		if isPgNoRows(err) {
			return httputil.BusinessError(notifications.BillingCycleConsolidated)
		}
		return httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	return nil
}

func (s *Service) ListProcessingMonths(ctx context.Context, page httputil.PageSearch) ([]models.ListProcessingMonthResponse, int64, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, 0, err
	}
	return s.Store.ListProcessingMonths(ctx, orgID, page)
}

func (s *Service) GetProcessingMonth(ctx context.Context, id string) (*models.GetProcessingMonthResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	m, err := s.Store.GetProcessingMonth(ctx, orgID, id)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if m == nil {
		return nil, httputil.NotFoundError(notifications.ProcessingMonthNotFound)
	}
	resp := store.ToProcessingMonthResponse(m)
	return &resp, nil
}

func (s *Service) CreateProcessingMonth(ctx context.Context, input models.CreateProcessingMonthInput) (*models.GetProcessingMonthResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	if input.Year < 2000 || input.Year > 2100 {
		return nil, httputil.ValidationError(notifications.ProcessingMonthYearInvalid)
	}
	if input.Month < 1 || input.Month > 12 {
		return nil, httputil.ValidationError(notifications.ProcessingMonthMonthInvalid)
	}
	if strings.TrimSpace(input.DisplayName) == "" {
		return nil, httputil.ValidationError(notifications.ProcessingMonthDisplayNameRequired)
	}
	if utf8.RuneCountInString(input.DisplayName) > 128 {
		return nil, httputil.ValidationError(notifications.ProcessingMonthDisplayNameMaxLength)
	}
	ok, err := s.Store.ProviderExists(ctx, orgID, input.ProviderID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if !ok {
		return nil, httputil.BusinessError(notifications.ProviderNotFound)
	}
	dup, err := s.Store.ProcessingMonthDuplicateExists(ctx, orgID, input.ProviderID, input.Year, input.Month)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if dup {
		return nil, httputil.BusinessError(notifications.ProcessingMonthDuplicate)
	}
	id := uuid.New().String()
	if err := s.Store.CreateProcessingMonth(ctx, orgID, id, input.ProviderID, input.DisplayName, input.Year, input.Month); err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	return s.GetProcessingMonth(ctx, id)
}

func (s *Service) CloseProcessingMonth(ctx context.Context, id string) (*models.GetProcessingMonthResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	user, err := userFrom(ctx)
	if err != nil {
		return nil, err
	}
	m, err := s.Store.GetProcessingMonth(ctx, orgID, id)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if m == nil {
		return nil, httputil.NotFoundError(notifications.ProcessingMonthNotFound)
	}
	if m.Status == "closed" {
		return nil, httputil.BusinessError(notifications.ProcessingMonthAlreadyClosed)
	}
	if err := s.Store.CloseProcessingMonth(ctx, orgID, id, user.ID, false, nil); err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	return s.GetProcessingMonth(ctx, id)
}

func (s *Service) CloseProcessingMonthContingency(ctx context.Context, id string, input models.CloseProcessingMonthContingencyInput) (*models.GetProcessingMonthResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	user, err := userFrom(ctx)
	if err != nil {
		return nil, err
	}
	j := strings.TrimSpace(input.Justification)
	if utf8.RuneCountInString(j) < 10 {
		return nil, httputil.ValidationError(notifications.ProcessingMonthContingencyJustMin)
	}
	if utf8.RuneCountInString(j) > 4000 {
		return nil, httputil.ValidationError(notifications.ProcessingMonthContingencyJustMax)
	}
	m, err := s.Store.GetProcessingMonth(ctx, orgID, id)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if m == nil {
		return nil, httputil.NotFoundError(notifications.ProcessingMonthNotFound)
	}
	if m.Status == "closed" {
		return nil, httputil.BusinessError(notifications.ProcessingMonthAlreadyClosed)
	}
	if err := s.Store.CloseProcessingMonth(ctx, orgID, id, user.ID, true, &j); err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	return s.GetProcessingMonth(ctx, id)
}

func (s *Service) ListProviderInvoices(ctx context.Context, processingMonthID *string, page httputil.PageSearch) ([]models.ListProviderInvoiceResponse, int64, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, 0, err
	}
	return s.Store.ListProviderInvoices(ctx, orgID, processingMonthID, page)
}

func (s *Service) GetProviderInvoice(ctx context.Context, id string) (*models.GetProviderInvoiceResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	inv, err := s.Store.GetProviderInvoice(ctx, orgID, id)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if inv == nil {
		return nil, httputil.NotFoundError(notifications.InvoiceNotFound)
	}
	return inv, nil
}

func (s *Service) RequestProviderInvoiceImport(ctx context.Context, input models.ProviderInvoiceImportRequestInput) (*models.RequestProviderInvoiceImportResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	user, err := userFrom(ctx)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(input.ProviderID) == "" {
		return nil, httputil.ValidationError(notifications.ImportProviderIDRequired)
	}
	if strings.TrimSpace(input.ProcessingMonthID) == "" {
		return nil, httputil.ValidationError(notifications.ImportProcessingMonthIDRequired)
	}
	if strings.TrimSpace(input.StorageBucket) == "" {
		return nil, httputil.ValidationError(notifications.ImportStorageBucketRequired)
	}
	if strings.TrimSpace(input.StorageObjectKey) == "" {
		return nil, httputil.ValidationError(notifications.ImportStorageObjectKeyRequired)
	}
	ok, err := s.Store.ProviderExists(ctx, orgID, input.ProviderID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if !ok {
		return nil, httputil.BusinessError(notifications.ProviderNotFound)
	}
	month, err := s.Store.GetProcessingMonth(ctx, orgID, input.ProcessingMonthID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if month == nil {
		return nil, httputil.BusinessError(notifications.ProcessingMonthNotFound)
	}
	if month.ProviderID != input.ProviderID {
		return nil, httputil.BusinessError(notifications.ProcessingMonthProviderMismatch)
	}
	if month.Status != "open" {
		return nil, httputil.BusinessError(notifications.ProcessingMonthNotOpen)
	}
	id := uuid.New().String()
	row := store.ImportRequestRow{
		ID: id, OrganizationID: orgID, ProviderID: input.ProviderID,
		ProcessingMonthID: input.ProcessingMonthID, StorageBucket: input.StorageBucket,
		StorageObjectKey: input.StorageObjectKey, OriginalFileName: input.OriginalFileName,
		CreatedBy: user.ID,
	}
	if err := s.Store.CreateImportRequest(ctx, row); err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if s.Publisher != nil {
		_ = s.Publisher.PublishInvoiceImportRequested(ctx, id, input.StorageBucket, input.StorageObjectKey, input.OriginalFileName, user.ID)
	}
	return &models.RequestProviderInvoiceImportResponse{
		ID: id, ProcessingMonthID: input.ProcessingMonthID,
		Status: httputil.ImportRequestStatusString(0),
	}, nil
}

func (s *Service) ListCostCenters(ctx context.Context, page httputil.PageSearch) ([]models.ListCostCenterResponse, int64, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, 0, err
	}
	return s.Store.ListCostCenters(ctx, orgID, page)
}

func (s *Service) GetDashboardStats(ctx context.Context) (*models.DashboardStatsResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	return s.Store.GetDashboardStats(ctx, orgID)
}
