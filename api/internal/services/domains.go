package services

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/luxus-connect/telefonia/api/internal/auth"
	"github.com/luxus-connect/telefonia/api/internal/billingcalc"
	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/models"
	"github.com/luxus-connect/telefonia/api/internal/notifications"
	"github.com/luxus-connect/telefonia/api/internal/statemachine"
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
		company, err := s.Store.GetFirstContractingCompanyForProvider(ctx, providerID)
		if err != nil {
			return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
		}
		if company == nil {
			companyID := uuid.New().String()
			legalName := provider.Name
			if err := s.Store.CreateContractingCompany(ctx, companyID, providerID, legalName, "00000000000000"); err != nil {
				return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
			}
			company = &store.ContractingCompanyRow{ID: companyID, ProviderID: providerID, LegalName: legalName, TaxID: "00000000000000"}
		}
		accountID := uuid.New().String()
		if err := s.Store.CreateProviderAccount(ctx, accountID, company.ID, accountNumber); err != nil {
			return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
		}
		account = &store.ProviderAccountRow{ID: accountID, ContractingCompanyID: company.ID, AccountNumber: accountNumber}
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
	if input.StartDate != nil && strings.TrimSpace(*input.StartDate) != "" {
		parsed, err := parseFinancialDate(*input.StartDate)
		if err != nil {
			return nil, err
		}
		start = parsed
	}
	line, err := s.GetPhoneLine(ctx, phoneLineID)
	if err != nil {
		return nil, err
	}
	if err := s.SM().ValidateTransition(statemachine.EntityPhoneLine, line.Status, "active", userRoles(ctx)); err != nil {
		return nil, err
	}
	if err := s.blockRetroactiveIfClosed(ctx, orgID, phoneLineID, start); err != nil {
		return nil, err
	}
	if err := s.requireActiveServiceToLink(ctx, phoneLineID, input.MonthlyAmount); err != nil {
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
	if err := s.Store.AssignPhoneLineCustomer(ctx, phoneLineID, input.CustomerID, start, input.MonthlyAmount); err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	linkID, _, _ := s.Store.GetActivePhoneLineCustomerLink(ctx, phoneLineID)
	if linkID != "" {
		_ = s.EnsureBillingProcessingsForLink(ctx, linkID, input.CustomerID, input.MonthlyAmount)
	}
	_ = s.Store.UpdatePhoneLineStatus(ctx, phoneLineID, "active")
	_ = s.Store.ReactivateCustomer(ctx, input.CustomerID)

	var actorUserID *string
	if u := auth.UserFromContext(ctx); u != nil && u.ID != "" {
		actorUserID = &u.ID
	}
	_ = s.SM().RecordTransition(ctx, orgID, statemachine.EntityPhoneLine, phoneLineID, line.Status, "active", "assign_customer", nil, actorUserID, nil)

	s.auditLog(ctx, "Assign", "PhoneLineCustomerLink", phoneLineID, map[string]any{"previous_customer_id": prevCustomerID},
		map[string]any{"customer_id": input.CustomerID, "start_date": start.Format("2006-01-02")})
	if prevCustomerID != "" && prevCustomerID != input.CustomerID {
		hasOther, _ := s.Store.CustomerHasOtherActivePhoneLines(ctx, orgID, prevCustomerID, phoneLineID)
		if !hasOther {
			_ = s.Store.InactivateCustomer(ctx, prevCustomerID)
		}
	}
	s.maybeGenerateAutomaticContract(ctx, orgID, input.CustomerID, phoneLineID, "line_assign", "Vínculo de linha ao cliente.")
	return s.activePhoneLineCustomerLink(ctx, orgID, phoneLineID)
}

func (s *Service) TransferPhoneLineCustomer(ctx context.Context, phoneLineID string, input models.TransferPhoneLineCustomerInput) (*models.PhoneLineCustomerLinkResponse, error) {
	if _, err := orgFrom(ctx); err != nil {
		return nil, err
	}
	transferDate := time.Now().UTC()
	if input.TransferDate != nil && strings.TrimSpace(*input.TransferDate) != "" {
		parsed, err := parseFinancialDate(*input.TransferDate)
		if err != nil {
			return nil, err
		}
		transferDate = parsed
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
	transferStr := transferDate.Format("2006-01-02")
	assignInput := models.AssignPhoneLineCustomerInput{
		CustomerID:    input.CustomerID,
		StartDate:     &transferStr,
		MonthlyAmount: input.MonthlyAmount,
	}
	return s.AssignPhoneLineCustomer(ctx, phoneLineID, assignInput)
}

func (s *Service) UpdateActivePhoneLineCustomerLink(ctx context.Context, phoneLineID string, input models.UpdateActivePhoneLineCustomerLinkInput) (*models.PhoneLineCustomerLinkResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := s.GetPhoneLine(ctx, phoneLineID); err != nil {
		return nil, err
	}
	if input.MonthlyAmount != nil && *input.MonthlyAmount < 0 {
		return nil, httputil.ValidationError(notifications.N("PHONE_LINE_MONTHLY_AMOUNT_INVALID", "Monthly amount cannot be negative."))
	}
	if err := s.Store.UpdateActivePhoneLineCustomerLinkAmount(ctx, phoneLineID, input.MonthlyAmount); err != nil {
		if isPgNoRows(err) {
			return nil, httputil.NotFoundError(notifications.PhoneLineActiveCustomerLinkNotFound)
		}
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	linkID, err := s.Store.GetActiveLinkIDForPhoneLine(ctx, orgID, phoneLineID)
	if err == nil && linkID != "" && input.MonthlyAmount != nil {
		processings, _ := s.Store.ListBillingProcessingsForLink(ctx, linkID)
		for _, p := range processings {
			if p.Perspective == perspectiveLuxusCustomer {
				items, _ := s.Store.ListBillingCompositionItems(ctx, p.ID)
				for _, it := range items {
					if it.ItemType == "service" && strings.Contains(strings.ToLower(it.Description), "mensalidade") {
						now := time.Now().UTC()
						amt := *input.MonthlyAmount
						_ = s.Store.UpdateBillingCompositionItem(ctx, it.ID, nil, &amt, nil, nil, nil, nil, nil, now, nil, nil)
						break
					}
				}
			}
		}
	}
	return s.activePhoneLineCustomerLink(ctx, orgID, phoneLineID)
}

func (s *Service) UnassignPhoneLineCustomer(ctx context.Context, phoneLineID string, input models.UnassignPhoneLineCustomerInput) error {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return err
	}
	end := time.Now().UTC()
	if input.EndDate != nil && strings.TrimSpace(*input.EndDate) != "" {
		parsed, err := parseFinancialDate(*input.EndDate)
		if err != nil {
			return err
		}
		end = parsed
	}
	_, activeCustomerID, err := s.Store.GetActivePhoneLineCustomerLink(ctx, phoneLineID)
	if err != nil {
		return httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if activeCustomerID == "" {
		return httputil.BusinessError(notifications.PhoneLineActiveCustomerLinkNotFound)
	}
	line, err := s.GetPhoneLine(ctx, phoneLineID)
	if err != nil {
		return err
	}
	if err := s.SM().ValidateTransition(statemachine.EntityPhoneLine, line.Status, "in_stock", userRoles(ctx)); err != nil {
		return err
	}
	if err := s.blockRetroactiveIfClosed(ctx, orgID, phoneLineID, end); err != nil {
		return err
	}
	class, _, _, _ := s.Store.GetPhoneLineClassification(ctx, phoneLineID)
	if class == "titular" {
		_ = s.reclassifyDependentsOf(ctx, phoneLineID, "titular unlinked")
	}
	if err := s.Store.UnassignPhoneLineCustomer(ctx, phoneLineID, end); err != nil {
		if isPgNoRows(err) {
			return httputil.BusinessError(notifications.PhoneLineActiveCustomerLinkNotFound)
		}
		return httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	_ = s.Store.UpdatePhoneLineStatus(ctx, phoneLineID, "in_stock")

	var actorUserID *string
	if u := auth.UserFromContext(ctx); u != nil && u.ID != "" {
		actorUserID = &u.ID
	}
	_ = s.SM().RecordTransition(ctx, orgID, statemachine.EntityPhoneLine, phoneLineID, line.Status, "in_stock", "unassign_customer", nil, actorUserID, nil)

	s.auditLog(ctx, "Unassign", "PhoneLineCustomerLink", phoneLineID,
		map[string]any{"customer_id": activeCustomerID}, map[string]any{"end_date": end.Format("2006-01-02")})
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
	if err := s.queueTwoLevel(ctx, "cycle_impact", "billing_cycle", id, "Alteração com impacto no ciclo", map[string]any{
		"id": id, "input": input,
	}); err != nil {
		return err
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
	if err := s.queueTwoLevel(ctx, "consolidation", "processing_month", id, "Consolidação e fechamento da competência", map[string]any{"processing_month_id": id}); err != nil {
		return nil, err
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
	if err := s.SM().ValidateTransition(statemachine.EntityProcessingMonth, m.Status, "closed", userRoles(ctx)); err != nil {
		return nil, err
	}
	if err := s.Store.CloseProcessingMonth(ctx, orgID, id, user.ID, true, &j); err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	_ = s.SM().RecordTransition(ctx, orgID, statemachine.EntityProcessingMonth, id, m.Status, "closed", "close_month_contingency", &j, &user.ID, nil)
	s.auditLog(ctx, "CloseContingency", "ProcessingMonth", id, nil, map[string]any{"justification": j, "closed_by": user.ID})
	s.DispatchWebhookEvent(orgID, WebhookEventBillingClosed, map[string]any{
		"processing_month_id": id, "contingency": true,
	})
	return s.GetProcessingMonth(ctx, id)
}

func (s *Service) ListProviderInvoices(ctx context.Context, processingMonthID *string, page httputil.PageSearch) ([]models.ListProviderInvoiceResponse, int64, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, 0, err
	}
	return s.Store.ListProviderInvoices(ctx, orgID, processingMonthID, page)
}

func (s *Service) ListStateTransitionLogs(ctx context.Context, entityType, entityID string, limit int) ([]models.StateTransitionLogResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := s.Store.ListStateTransitionLogs(ctx, orgID, entityType, entityID, limit)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	var res []models.StateTransitionLogResponse
	for _, r := range rows {
		res = append(res, models.StateTransitionLogResponse{
			ID:             r.ID,
			OrganizationID: r.OrganizationID,
			EntityType:     r.EntityType,
			EntityID:       r.EntityID,
			FromState:      r.FromState,
			ToState:        r.ToState,
			TriggerEvent:   r.TriggerEvent,
			Justification:  r.Justification,
			ActorUserID:    r.ActorUserID,
			MetadataJSON:   r.MetadataJSON,
			CreatedAt:      r.CreatedAt,
		})
	}
	if res == nil {
		res = []models.StateTransitionLogResponse{}
	}
	return res, nil
}

func (s *Service) ApportionProviderInvoiceDiscount(ctx context.Context, invoiceID string, input models.ApportionGlobalDiscountInput) (*models.ApportionGlobalDiscountResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	if input.GlobalDiscountAmount <= 0 {
		return nil, httputil.ValidationError(notifications.N("INVALID_DISCOUNT_AMOUNT", "O valor do desconto global deve ser maior que zero."))
	}

	inv, err := s.Store.GetProviderInvoice(ctx, orgID, invoiceID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if inv == nil {
		return nil, httputil.NotFoundError(notifications.InvoiceNotFound)
	}

	if len(inv.PhoneLines) == 0 {
		return nil, httputil.BusinessError(notifications.N("NO_LINES_IN_INVOICE", "Nenhuma linha vinculada a esta fatura da operadora."))
	}

	// Coleta os alvos de rateio com base no processamento ativo de cada linha
	targets := make([]billingcalc.ApportionmentTarget, 0, len(inv.PhoneLines))
	lineProcessingMap := make(map[string]string) // phoneLineID -> processingID
	lineObjMap := make(map[string]models.GetProviderPhoneLineResponse)

	for _, line := range inv.PhoneLines {
		lineObjMap[line.ID] = line
		linkID, err := s.Store.GetActiveLinkIDForPhoneLine(ctx, orgID, line.ID)
		if err != nil || linkID == "" {
			continue
		}
		processings, err := s.Store.ListBillingProcessingsForLink(ctx, linkID)
		if err != nil {
			continue
		}
		for _, p := range processings {
			if p.Perspective == perspectiveLuxusCustomer {
				tot, err := s.Store.SumBillingProcessingTotal(ctx, p.ID)
				if err == nil && tot > 0 {
					targets = append(targets, billingcalc.ApportionmentTarget{
						ID:     line.ID,
						Amount: tot,
					})
					lineProcessingMap[line.ID] = p.ID
				}
				break
			}
		}
	}

	if len(targets) == 0 {
		return nil, httputil.BusinessError(notifications.N("NO_ELIGIBLE_LINES", "Nenhuma linha com processamento de faturamento ativo e valor positivo nesta fatura."))
	}

	apportioned, err := billingcalc.ApportionGlobalDiscount(input.GlobalDiscountAmount, targets)
	if err != nil {
		return nil, httputil.BusinessError(notifications.N("APPORTIONMENT_FAILED", err.Error()))
	}

	desc := strings.TrimSpace(input.Description)
	if desc == "" {
		desc = fmt.Sprintf("Rateio de desconto da fatura %s", inv.Number)
	}

	now := time.Now().UTC()
	var outItems []models.ApportionedLineItem

	for _, app := range apportioned {
		procID := lineProcessingMap[app.ID]
		lineObj := lineObjMap[app.ID]

		if app.AllocatedDiscount > 0 && procID != "" {
			row := store.BillingCompositionItemRow{
				ID:           uuid.New().String(),
				ProcessingID: procID,
				ItemType:     "discount",
				Description:  desc,
				Amount:       app.AllocatedDiscount,
				Quantity:     1,
				Active:       true,
				CreatedAt:    now,
				UpdatedAt:    now,
				Proportional: true,
			}
			if err := s.Store.CreateBillingCompositionItem(ctx, row); err == nil {
				s.auditLog(ctx, "ApportionDiscount", "LineBillingCompositionItem", row.ID, nil, map[string]any{
					"invoice_id":         invoiceID,
					"phone_line_id":      app.ID,
					"allocated_discount": app.AllocatedDiscount,
				})
			}
		}

		outItems = append(outItems, models.ApportionedLineItem{
			PhoneLineID:       app.ID,
			PhoneNumber:       lineObj.Number,
			OriginalAmount:    app.OriginalAmount,
			AllocatedDiscount: app.AllocatedDiscount,
			FinalAmount:       app.FinalAmount,
		})
	}

	return &models.ApportionGlobalDiscountResponse{
		InvoiceID:            invoiceID,
		GlobalDiscountAmount: input.GlobalDiscountAmount,
		LinesCount:           len(outItems),
		Items:                outItems,
	}, nil
}

func (s *Service) GetImportRequestStatus(ctx context.Context, id string) (*models.RequestProviderInvoiceImportResponse, error) {
	orgID, _ := orgFrom(ctx)
	var row *store.ImportRequestRow
	var err error
	if orgID != "" {
		row, err = s.Store.GetImportRequestForOrg(ctx, orgID, id)
	}
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if row == nil {
		row, err = s.Store.GetImportRequest(ctx, id)
		if err != nil {
			return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
		}
		if row == nil {
			return nil, httputil.NotFoundError(notifications.ImportRequestNotFound)
		}
	}
	return &models.RequestProviderInvoiceImportResponse{
		ID:                row.ID,
		ProcessingMonthID: row.ProcessingMonthID,
		Status:            httputil.ImportRequestStatusString(row.Status),
		Error:             row.Error,
		CompletedAt:       row.CompletedAt,
	}, nil
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
		CreatedBy: user.ID, AllowSubstitute: input.AllowSubstitute,
	}
	if err := s.Store.CreateImportRequest(ctx, row); err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if s.Processor != nil {
		_ = s.Processor.ProcessImport(ctx, id)
	} else if s.Publisher != nil {
		_ = s.Publisher.PublishInvoiceImportRequested(ctx, id, input.StorageBucket, input.StorageObjectKey, input.OriginalFileName, user.ID)
	}

	status := 0
	var errMsg *string
	var completedAt *time.Time
	if req, err := s.Store.GetImportRequest(ctx, id); err == nil && req != nil {
		status = req.Status
		errMsg = req.Error
		completedAt = req.CompletedAt
	}

	return &models.RequestProviderInvoiceImportResponse{
		ID:                id,
		ProcessingMonthID: input.ProcessingMonthID,
		Status:            httputil.ImportRequestStatusString(status),
		Error:             errMsg,
		CompletedAt:       completedAt,
	}, nil
}

type InvoicePreviewer interface {
	PreviewImport(ctx context.Context, orgID string, input models.ProviderInvoiceImportRequestInput) (*models.ImportPreviewResponse, error)
}

func (s *Service) PreviewProviderInvoiceImport(ctx context.Context, input models.ProviderInvoiceImportRequestInput) (*models.ImportPreviewResponse, error) {
	orgID, err := orgFrom(ctx)
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
	previewer, ok := s.Processor.(InvoicePreviewer)
	if !ok || previewer == nil {
		return nil, httputil.BusinessError(notifications.ObjectStorageUnavailable)
	}
	return previewer.PreviewImport(ctx, orgID, input)
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

func (s *Service) ListExpiringContracts(ctx context.Context, daysAhead int) ([]models.ExpiringContractResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	if daysAhead <= 0 {
		daysAhead = 30
	}
	return s.Store.ListExpiringContracts(ctx, orgID, daysAhead)
}

func (s *Service) GetPreClosingAlerts(ctx context.Context, processingMonthID string) (*models.PreClosingAlertsResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	month, err := s.Store.GetProcessingMonth(ctx, orgID, processingMonthID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if month == nil {
		return nil, httputil.NotFoundError(notifications.ProcessingMonthNotFound)
	}

	var alerts []PreClosingAlert
	canClose := true

	// 1. Checa faturas de provedor vinculadas ao mês de processamento
	invoices, _, err := s.Store.ListProviderInvoices(ctx, orgID, &processingMonthID, httputil.PageSearch{PageSize: 100})
	if err == nil {
		for _, inv := range invoices {
			if inv.Status != "processed" && inv.Status != "paid" {
				alerts = append(alerts, PreClosingAlert{
					Code:        "INVOICE_PENDING_CONCILIATION",
					Severity:    "WARNING",
					Category:    "INVOICES",
					Message:     fmt.Sprintf("Fatura de operadora %s pendente de conciliação (status: %s).", inv.ProviderAccountNumber, inv.Status),
					EntityID:    inv.ID,
					EntityLabel: inv.ProviderAccountNumber,
				})
			}
		}
	}

	// 2. Checa contratos de fidelidade expirando ou expirados nesta competência
	expiring, err := s.Store.ListExpiringContracts(ctx, orgID, 30)
	if err == nil {
		for _, exp := range expiring {
			phoneNum := ""
			if exp.PhoneNumber != nil {
				phoneNum = *exp.PhoneNumber
			}
			alerts = append(alerts, PreClosingAlert{
				Code:        "FIDELITY_EXPIRING",
				Severity:    "WARNING",
				Category:    "FIDELITY",
				Message:     fmt.Sprintf("Contrato de fidelidade da linha %s vence em %d dias (%s).", phoneNum, exp.DaysRemaining, exp.PredictedEndDate.Format("02/01/2006")),
				EntityID:    exp.ContractID,
				EntityLabel: phoneNum,
			})
		}
	}

	criticalCount := 0
	warningCount := 0
	for _, a := range alerts {
		if a.Severity == "CRITICAL" {
			criticalCount++
			canClose = false
		} else if a.Severity == "WARNING" {
			warningCount++
		}
	}

	if alerts == nil {
		alerts = []PreClosingAlert{}
	}

	var alertDtos []models.PreClosingAlert
	for _, a := range alerts {
		alertDtos = append(alertDtos, models.PreClosingAlert{
			Code:        a.Code,
			Severity:    a.Severity,
			Category:    a.Category,
			Message:     a.Message,
			EntityID:    a.EntityID,
			EntityLabel: a.EntityLabel,
		})
	}

	return &models.PreClosingAlertsResponse{
		ProcessingMonthID: processingMonthID,
		CanClose:          canClose,
		CriticalCount:     criticalCount,
		WarningCount:      warningCount,
		Alerts:            alertDtos,
	}, nil
}

type PreClosingAlert struct {
	Code        string
	Severity    string
	Category    string
	Message     string
	EntityID    string
	EntityLabel string
}

func (s *Service) GetPhoneLineTimeline(ctx context.Context, phoneLineID string) (*models.PhoneLineTimelineResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	line, err := s.GetPhoneLine(ctx, phoneLineID)
	if err != nil {
		return nil, err
	}

	var events []models.PhoneLineTimelineEvent

	// 1. Transições de Máquina de Estado
	transitions, err := s.Store.ListStateTransitionLogs(ctx, orgID, "PhoneLine", phoneLineID, 100)
	if err == nil {
		for _, t := range transitions {
			desc := fmt.Sprintf("Estado alterado de '%s' para '%s'", t.FromState, t.ToState)
			if t.Justification != nil && *t.Justification != "" {
				desc += ": " + *t.Justification
			}
			events = append(events, models.PhoneLineTimelineEvent{
				EventID:     t.ID,
				EventType:   "STATE_TRANSITION",
				Title:       fmt.Sprintf("Transição de Estado: %s -> %s", t.FromState, t.ToState),
				Description: desc,
				ActorUserID: t.ActorUserID,
				Timestamp:   t.CreatedAt,
			})
		}
	}

	// 2. Histórico de Vínculos com Cliente
	customerLinks, err := s.Store.ListPhoneLineCustomerLinks(ctx, orgID, phoneLineID)
	if err == nil {
		for _, cl := range customerLinks {
			statusStr := "Ativo"
			if !cl.IsActive {
				statusStr = "Encerrado"
			}
			amtStr := "R$ 0,00"
			if cl.MonthlyAmount != nil {
				amtStr = fmt.Sprintf("R$ %.2f", *cl.MonthlyAmount)
			}
			events = append(events, models.PhoneLineTimelineEvent{
				EventID:     cl.ID,
				EventType:   "CUSTOMER_LINK",
				Title:       fmt.Sprintf("Vínculo de Cliente: %s (%s)", cl.CustomerName, statusStr),
				Description: fmt.Sprintf("Cliente: %s | Mensalidade: %s | Início: %s", cl.CustomerName, amtStr, cl.StartDate.Format("02/01/2006")),
				Timestamp:   cl.StartDate,
			})
		}
	}

	// 3. Eventos de Fidelidade
	fid, err := s.Store.GetLineFidelity(ctx, phoneLineID)
	if err == nil && fid != nil {
		fidEvents, err := s.Store.ListLineFidelityEvents(ctx, fid.ID)
		if err == nil {
			for _, fe := range fidEvents {
				desc := fe.EventType
				if fe.Notes != nil && *fe.Notes != "" {
					desc += " — " + *fe.Notes
				}
				events = append(events, models.PhoneLineTimelineEvent{
					EventID:     fe.ID,
					EventType:   "FIDELITY",
					Title:       fmt.Sprintf("Evento de Fidelidade: %s", fe.EventType),
					Description: desc,
					ActorUserID: fe.UserID,
					Timestamp:   fe.OccurredAt,
				})
			}
		}
	}

	// 4. Criação da Linha
	actDate := time.Now().UTC()
	if line.ActivationDate != nil {
		actDate = *line.ActivationDate
	}
	events = append(events, models.PhoneLineTimelineEvent{
		EventID:     line.ID,
		EventType:   "AUDIT",
		Title:       "Linha Telefônica Cadastrada",
		Description: fmt.Sprintf("Linha %s cadastrada com operadora %s e plano %s", line.Number, line.ProviderName, line.ProviderPlanName),
		Timestamp:   actDate,
	})

	// Ordena por data decrescente
	for i := 0; i < len(events)-1; i++ {
		for j := i + 1; j < len(events); j++ {
			if events[i].Timestamp.Before(events[j].Timestamp) {
				events[i], events[j] = events[j], events[i]
			}
		}
	}

	if events == nil {
		events = []models.PhoneLineTimelineEvent{}
	}

	return &models.PhoneLineTimelineResponse{
		PhoneLineID: line.ID,
		PhoneNumber: line.Number,
		Events:      events,
	}, nil
}
