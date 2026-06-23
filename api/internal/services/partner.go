package services

import (
	"context"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/luxus-connect/telefonia/api/internal/auth"
	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/models"
	"github.com/luxus-connect/telefonia/api/internal/notifications"
)

func partnerUserFrom(ctx context.Context) (*auth.User, error) {
	u, err := userFrom(ctx)
	if err != nil {
		return nil, err
	}
	if !auth.IsPartner(ctx) {
		return nil, httputil.BusinessError(notifications.N("FORBIDDEN", "Partner role required"))
	}
	return u, nil
}

func (s *Service) PartnerGetDashboardStats(ctx context.Context) (*models.PartnerDashboardStatsResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	user, err := partnerUserFrom(ctx)
	if err != nil {
		return nil, err
	}
	stats, err := s.Store.GetPartnerDashboardStats(ctx, orgID, user.ID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	return stats, nil
}

func (s *Service) PartnerListProviders(ctx context.Context) ([]models.ListProvidersResponse, int64, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, 0, err
	}
	if _, err := partnerUserFrom(ctx); err != nil {
		return nil, 0, err
	}
	return s.Store.ListProviders(ctx, orgID, httputil.PageSearch{PageIndex: 0, PageSize: 100})
}

func (s *Service) PartnerListCustomers(ctx context.Context, providerID *string, page httputil.PageSearch) ([]models.ListCustomerResponse, int64, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, 0, err
	}
	user, err := partnerUserFrom(ctx)
	if err != nil {
		return nil, 0, err
	}
	return s.Store.ListCustomers(ctx, orgID, providerID, &user.ID, page)
}

func (s *Service) PartnerGetCustomer(ctx context.Context, id string) (*models.ListCustomerResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	user, err := partnerUserFrom(ctx)
	if err != nil {
		return nil, err
	}
	c, err := s.Store.GetCustomerInOrg(ctx, orgID, id, &user.ID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if c == nil {
		return nil, httputil.NotFoundError(notifications.CustomerNotFound)
	}
	return c, nil
}

func (s *Service) PartnerCreateCustomer(ctx context.Context, input models.CreateCustomerInput) (*models.CreateCustomerResponse, error) {
	user, err := partnerUserFrom(ctx)
	if err != nil {
		return nil, err
	}
	input.ResponsibleSalespersonUserID = &user.ID
	return s.CreateCustomer(ctx, input)
}

func (s *Service) PartnerUpdateCustomer(ctx context.Context, id string, input models.UpdateCustomerInput) error {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return err
	}
	user, err := partnerUserFrom(ctx)
	if err != nil {
		return err
	}
	ok, err := s.Store.CustomerOwnedBySalesperson(ctx, orgID, id, user.ID)
	if err != nil {
		return httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if !ok {
		return httputil.ForbiddenError(notifications.PartnerCustomerAccessDenied)
	}
	input.ResponsibleSalespersonUserID = &user.ID
	return s.UpdateCustomer(ctx, id, input)
}

func (s *Service) PartnerListCustomerPhoneLines(ctx context.Context, customerID string, page httputil.PageSearch) ([]models.CustomerPhoneLineLinkResponse, int64, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, 0, err
	}
	user, err := partnerUserFrom(ctx)
	if err != nil {
		return nil, 0, err
	}
	ok, err := s.Store.CustomerOwnedBySalesperson(ctx, orgID, customerID, user.ID)
	if err != nil {
		return nil, 0, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if !ok {
		return nil, 0, httputil.ForbiddenError(notifications.PartnerCustomerAccessDenied)
	}
	return s.Store.ListCustomerPhoneLines(ctx, orgID, customerID, page)
}

func (s *Service) PartnerListPhoneLines(ctx context.Context, page httputil.PageSearch) ([]models.PartnerPhoneLineResponse, int64, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, 0, err
	}
	user, err := partnerUserFrom(ctx)
	if err != nil {
		return nil, 0, err
	}
	return s.Store.ListPartnerPhoneLines(ctx, orgID, user.ID, page)
}

func (s *Service) PartnerCreateLineOperationRequest(ctx context.Context, input models.CreatePhoneLineOperationRequestInput) (*models.PhoneLineOperationRequestResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	user, err := partnerUserFrom(ctx)
	if err != nil {
		return nil, err
	}

	opType := strings.ToLower(strings.TrimSpace(input.OperationType))
	if opType != "activation" && opType != "deactivation" {
		return nil, httputil.ValidationError(notifications.N("PHONE_LINE_OPERATION_TYPE_INVALID", "Operation type must be activation or deactivation."))
	}

	ok, err := s.Store.PhoneLineOwnedBySalesperson(ctx, orgID, input.PhoneLineID, user.ID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if !ok {
		return nil, httputil.ForbiddenError(notifications.PartnerPhoneLineAccessDenied)
	}

	ok, err = s.Store.CustomerOwnedBySalesperson(ctx, orgID, input.CustomerID, user.ID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if !ok {
		return nil, httputil.ForbiddenError(notifications.PartnerCustomerAccessDenied)
	}

	linkID, activeCustomerID, err := s.Store.GetActivePhoneLineCustomerLink(ctx, input.PhoneLineID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if linkID == "" || activeCustomerID != input.CustomerID {
		return nil, httputil.BusinessError(notifications.N("PHONE_LINE_CUSTOMER_MISMATCH", "Phone line is not linked to the selected customer."))
	}

	pending, err := s.Store.PendingPhoneLineOperationExists(ctx, orgID, input.PhoneLineID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if pending {
		return nil, httputil.BusinessError(notifications.PhoneLineOperationPendingExists)
	}

	if input.Justification != nil {
		j := strings.TrimSpace(*input.Justification)
		if utf8.RuneCountInString(j) > 4000 {
			return nil, httputil.ValidationError(notifications.CustomerManualReleaseJustMax)
		}
		if j == "" {
			input.Justification = nil
		} else {
			input.Justification = &j
		}
	}

	id := uuid.New().String()
	now := time.Now().UTC()
	if err := s.Store.CreatePhoneLineOperationRequest(ctx, id, orgID, input.PhoneLineID, input.CustomerID, user.ID, opType, input.Justification, now); err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	return s.Store.GetPhoneLineOperationRequest(ctx, orgID, id, &user.ID)
}

func (s *Service) PartnerListLineOperationRequests(ctx context.Context, page httputil.PageSearch) ([]models.PhoneLineOperationRequestResponse, int64, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, 0, err
	}
	user, err := partnerUserFrom(ctx)
	if err != nil {
		return nil, 0, err
	}
	return s.Store.ListPhoneLineOperationRequests(ctx, orgID, &user.ID, page)
}

func (s *Service) ListLineOperationRequests(ctx context.Context, page httputil.PageSearch) ([]models.PhoneLineOperationRequestResponse, int64, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, 0, err
	}
	return s.Store.ListPhoneLineOperationRequests(ctx, orgID, nil, page)
}

func (s *Service) ReviewLineOperationRequest(ctx context.Context, id string, input models.ReviewPhoneLineOperationRequestInput) (*models.PhoneLineOperationRequestResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	user, err := userFrom(ctx)
	if err != nil {
		return nil, err
	}

	status := strings.ToLower(strings.TrimSpace(input.Status))
	if status != "approved" && status != "rejected" {
		return nil, httputil.ValidationError(notifications.N("PHONE_LINE_OPERATION_REVIEW_INVALID", "Review status must be approved or rejected."))
	}

	phoneLineID, opType, err := s.Store.GetPhoneLineOperationRequestPhoneLineID(ctx, orgID, id)
	if err != nil {
		if isPgNoRows(err) {
			return nil, httputil.NotFoundError(notifications.PhoneLineOperationNotFound)
		}
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}

	now := time.Now().UTC()
	if err := s.Store.ReviewPhoneLineOperationRequest(ctx, orgID, id, status, user.ID, input.AdminNotes, now); err != nil {
		if isPgNoRows(err) {
			return nil, httputil.BusinessError(notifications.PhoneLineOperationAlreadyReviewed)
		}
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}

	if status == "approved" {
		subStatus := "pending_activation"
		if opType == "deactivation" {
			subStatus = "pending_cancellation"
		}
		if err := s.Store.UpdatePhoneLineTransition(ctx, phoneLineID, "in_transition", subStatus, now); err != nil {
			return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
		}
	}

	return s.Store.GetPhoneLineOperationRequest(ctx, orgID, id, nil)
}
