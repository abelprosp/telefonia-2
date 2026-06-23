package services

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/luxus-connect/telefonia/api/internal/auth"
	"github.com/luxus-connect/telefonia/api/internal/email"
	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/keycloak"
	"github.com/luxus-connect/telefonia/api/internal/models"
	"github.com/luxus-connect/telefonia/api/internal/notifications"
	"github.com/luxus-connect/telefonia/api/internal/store"
)

type Service struct {
	Store     *store.Store
	Publisher EventPublisher
	Keycloak  *keycloak.AdminClient
	Mailer    *email.Sender
}

type EventPublisher interface {
	PublishInvoiceImportRequested(ctx context.Context, importRequestID, bucket, key string, originalFileName *string, userID string) error
}

func orgFrom(ctx context.Context) (string, error) {
	org := auth.OrganizationFromContext(ctx)
	if org == nil || org.ID == "" {
		return "", httputil.BusinessError(notifications.SharedOrganizationRequired)
	}
	return org.ID, nil
}

func userFrom(ctx context.Context) (*auth.User, error) {
	u := auth.UserFromContext(ctx)
	if u == nil || u.ID == "" {
		return nil, httputil.BusinessError(notifications.SharedDomainViolation)
	}
	return u, nil
}

func isPgNoRows(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}

// --- Providers ---

func (s *Service) ListProviders(ctx context.Context, page httputil.PageSearch) ([]models.ListProvidersResponse, int64, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, 0, err
	}
	return s.Store.ListProviders(ctx, orgID, page)
}

func (s *Service) GetProvider(ctx context.Context, id string) (*models.GetProviderResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	p, err := s.Store.GetProvider(ctx, orgID, id)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if p == nil {
		return nil, httputil.NotFoundError(notifications.ProviderNotFound)
	}
	return p, nil
}

func (s *Service) CreateProvider(ctx context.Context, input models.CreateProviderInput) (*models.CreateProviderResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	if err := validateProviderInput(input.Name, input.Slug); err != nil {
		return nil, err
	}
	dup, err := s.Store.ProviderSlugExists(ctx, orgID, strings.TrimSpace(input.Slug), "")
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if dup {
		return nil, httputil.BusinessError(notifications.ProviderSlugDuplicated)
	}
	id := uuid.New().String()
	if err := s.Store.CreateProvider(ctx, orgID, id, strings.TrimSpace(input.Name), strings.TrimSpace(input.Slug)); err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	return &models.CreateProviderResponse{ID: id, Name: input.Name, Slug: input.Slug, Active: true}, nil
}

func (s *Service) UpdateProvider(ctx context.Context, id string, input models.UpdateProviderInput) error {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return err
	}
	if err := validateProviderInput(input.Name, input.Slug); err != nil {
		return err
	}
	dup, err := s.Store.ProviderSlugExists(ctx, orgID, strings.TrimSpace(input.Slug), id)
	if err != nil {
		return httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if dup {
		return httputil.BusinessError(notifications.ProviderSlugDuplicated)
	}
	if err := s.Store.UpdateProvider(ctx, orgID, id, strings.TrimSpace(input.Name), strings.TrimSpace(input.Slug)); err != nil {
		if isPgNoRows(err) {
			return httputil.NotFoundError(notifications.ProviderNotFound)
		}
		return httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	return nil
}

func (s *Service) InactivateProvider(ctx context.Context, id string) error {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return err
	}
	if err := s.Store.InactivateProvider(ctx, orgID, id); err != nil {
		if isPgNoRows(err) {
			return httputil.NotFoundError(notifications.ProviderNotFound)
		}
		return httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	return nil
}

func validateProviderInput(name, slug string) error {
	name = strings.TrimSpace(name)
	slug = strings.TrimSpace(slug)
	if name == "" {
		return httputil.ValidationError(notifications.ProviderNameRequired)
	}
	if utf8.RuneCountInString(name) > 100 {
		return httputil.ValidationError(notifications.ProviderNameMaxLength)
	}
	if slug == "" {
		return httputil.ValidationError(notifications.ProviderSlugRequired)
	}
	if utf8.RuneCountInString(slug) > 50 {
		return httputil.ValidationError(notifications.ProviderSlugMaxLength)
	}
	return nil
}

// --- Customers ---

func (s *Service) ListCustomers(ctx context.Context, providerID *string, page httputil.PageSearch) ([]models.ListCustomerResponse, int64, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, 0, err
	}
	return s.Store.ListCustomers(ctx, orgID, providerID, nil, page)
}

func (s *Service) GetCustomer(ctx context.Context, id string) (*models.ListCustomerResponse, error) {
	c, err := s.Store.GetCustomer(ctx, id)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if c == nil {
		return nil, httputil.NotFoundError(notifications.CustomerNotFound)
	}
	return c, nil
}

func (s *Service) CreateCustomer(ctx context.Context, input models.CreateCustomerInput) (*models.CreateCustomerResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	if err := validateCreateCustomer(input); err != nil {
		return nil, err
	}
	birthDate, err := httputil.ParseOptionalDate(input.BirthOrOpeningDate)
	if err != nil {
		return nil, httputil.ValidationError(notifications.N("INVALID_DATE", "Invalid birth or opening date."))
	}
	doc := httputil.NormalizeDigits(input.Document)
	dup, err := s.Store.CustomerDocumentExists(ctx, orgID, doc)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if dup {
		return nil, httputil.BusinessError(notifications.CustomerDocumentDuplicated)
	}
	ok, err := s.Store.ProviderExists(ctx, orgID, input.ProviderID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if !ok {
		return nil, httputil.NotFoundError(notifications.ProviderNotFound)
	}
	customerType := httputil.CustomerTypeFromInput(input.Type)
	docType := httputil.DocumentTypeForCustomer(customerType)
	id := uuid.New().String()
	if err := s.Store.CreateCustomer(ctx, orgID, id, input.ProviderID, customerType,
		strings.TrimSpace(input.Name), doc, docType, input.LegalName, input.StateRegistration,
		input.ResponsibleSalespersonUserID, birthDate); err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	for _, addr := range input.Addresses {
		if err := s.Store.CreateCustomerAddress(ctx, id, addr); err != nil {
			return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
		}
	}
	return s.GetCustomer(ctx, id)
}

func validateCreateCustomer(input models.CreateCustomerInput) error {
	if strings.TrimSpace(input.Name) == "" {
		return httputil.ValidationError(notifications.CustomerNameRequired)
	}
	if utf8.RuneCountInString(input.Name) > 256 {
		return httputil.ValidationError(notifications.CustomerNameMaxLength)
	}
	if strings.TrimSpace(input.Document) == "" {
		return httputil.ValidationError(notifications.CustomerDocumentRequired)
	}
	if utf8.RuneCountInString(input.Document) > 20 {
		return httputil.ValidationError(notifications.CustomerDocumentMaxLength)
	}
	if httputil.CustomerTypeFromInput(input.Type) == "pj" && (input.LegalName == nil || strings.TrimSpace(*input.LegalName) == "") {
		return httputil.ValidationError(notifications.CustomerLegalNameRequiredForPJ)
	}
	return nil
}

func (s *Service) UpdateCustomer(ctx context.Context, id string, input models.UpdateCustomerInput) error {
	if strings.TrimSpace(input.Name) == "" {
		return httputil.ValidationError(notifications.CustomerNameRequired)
	}
	birthDate, err := httputil.ParseOptionalDate(input.BirthOrOpeningDate)
	if err != nil {
		return httputil.ValidationError(notifications.N("INVALID_DATE", "Invalid birth or opening date."))
	}
	if err := s.Store.UpdateCustomer(ctx, id, strings.TrimSpace(input.Name), input.LegalName,
		input.StateRegistration, input.ResponsibleSalespersonUserID, birthDate); err != nil {
		if isPgNoRows(err) {
			return httputil.NotFoundError(notifications.CustomerNotFound)
		}
		return httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if input.BillingEmail != nil {
		orgID, err := orgFrom(ctx)
		if err != nil {
			return err
		}
		if err := s.Store.UpdateCustomerBillingEmail(ctx, orgID, id, strings.TrimSpace(*input.BillingEmail)); err != nil && !isPgNoRows(err) {
			return httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
		}
	}
	return nil
}

func (s *Service) InactivateCustomer(ctx context.Context, id string) error {
	if _, err := userFrom(ctx); err != nil {
		return err
	}
	if err := s.Store.InactivateCustomer(ctx, id); err != nil {
		if isPgNoRows(err) {
			return httputil.NotFoundError(notifications.CustomerNotFound)
		}
		return httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	return nil
}

func (s *Service) ListCustomerProviderLinks(ctx context.Context, customerID string) ([]models.CustomerProviderLinkResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	return s.Store.ListCustomerProviderLinks(ctx, orgID, customerID)
}

func (s *Service) ListCustomerPhoneLines(ctx context.Context, customerID string, page httputil.PageSearch) ([]models.CustomerPhoneLineLinkResponse, int64, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, 0, err
	}
	return s.Store.ListCustomerPhoneLines(ctx, orgID, customerID, page)
}

func (s *Service) GetBillingReadiness(ctx context.Context, customerID, processingMonthID string) (*models.GetCustomerBillingReadinessResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	resp, err := s.Store.GetBillingReadiness(ctx, orgID, customerID, processingMonthID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if resp == nil {
		return nil, httputil.NotFoundError(notifications.CustomerBillingReadinessNotFound)
	}
	return resp, nil
}

func (s *Service) ManualReleaseCustomer(ctx context.Context, customerID, processingMonthID string, input models.ManuallyReleaseCustomerInput) (*models.GetCustomerBillingReadinessResponse, error) {
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
		return nil, httputil.ValidationError(notifications.CustomerManualReleaseJustMin)
	}
	if utf8.RuneCountInString(j) > 4000 {
		return nil, httputil.ValidationError(notifications.CustomerManualReleaseJustMax)
	}
	ok, err := s.Store.CustomerExistsInOrg(ctx, orgID, customerID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if !ok {
		return nil, httputil.NotFoundError(notifications.CustomerNotFound)
	}
	month, err := s.Store.GetProcessingMonth(ctx, orgID, processingMonthID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if month == nil {
		return nil, httputil.NotFoundError(notifications.ProcessingMonthNotFound)
	}
	hasProvider, err := s.Store.CustomerHasActiveProvider(ctx, customerID, month.ProviderID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if !hasProvider {
		return nil, httputil.BusinessError(notifications.CustomerProcessingMonthMismatch)
	}
	if month.Status != "open" {
		return nil, httputil.BusinessError(notifications.ProcessingMonthNotOpen)
	}
	exists, err := s.Store.ManualReleaseExists(ctx, orgID, customerID, processingMonthID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if exists {
		return nil, httputil.BusinessError(notifications.CustomerManualReleaseAlready)
	}
	if err := s.Store.CreateManualRelease(ctx, orgID, customerID, processingMonthID, uuid.New().String(), j, user.ID); err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	return s.GetBillingReadiness(ctx, customerID, processingMonthID)
}

func (s *Service) ListCustomerAttachments(ctx context.Context, customerID string) ([]models.CustomerAttachmentResponse, error) {
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
	items, err := s.Store.ListCustomerAttachments(ctx, orgID, customerID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	return items, nil
}

func (s *Service) CreateCustomerAttachment(ctx context.Context, customerID string, input models.RegisterCustomerAttachmentInput) (*models.CustomerAttachmentResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(input.OriginalFileName) == "" {
		return nil, httputil.ValidationError(notifications.CustomerAttachmentOriginalRequired)
	}
	ok, err := s.Store.CustomerExistsInOrg(ctx, orgID, customerID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if !ok {
		return nil, httputil.NotFoundError(notifications.CustomerNotFound)
	}
	return s.Store.CreateCustomerAttachment(ctx, orgID, customerID, uuid.New().String(), input)
}

func (s *Service) DeleteCustomerAttachment(ctx context.Context, customerID, attachmentID string) error {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return err
	}
	if err := s.Store.DeleteCustomerAttachment(ctx, orgID, customerID, attachmentID); err != nil {
		if isPgNoRows(err) {
			return httputil.NotFoundError(notifications.CustomerNotFound)
		}
		return httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	return nil
}
