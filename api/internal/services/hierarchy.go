package services

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/luxus-connect/telefonia/api/internal/auth"
	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/models"
	"github.com/luxus-connect/telefonia/api/internal/notifications"
	"github.com/luxus-connect/telefonia/api/internal/statemachine"
)

func (s *Service) UpdatePhoneLineClassification(ctx context.Context, phoneLineID string, input models.UpdatePhoneLineClassificationInput) (*models.GetPhoneLineResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	line, err := s.GetPhoneLine(ctx, phoneLineID)
	if err != nil {
		return nil, err
	}
	if err := s.blockRetroactiveIfClosed(ctx, orgID, line.ID, time.Now().UTC()); err != nil {
		return nil, err
	}

	class := strings.ToLower(strings.TrimSpace(input.LineClassification))
	switch class {
	case "normal", "titular", "dependent":
	default:
		return nil, httputil.ValidationError(notifications.PhoneLineClassificationInvalid)
	}

	var titularID *string
	if class == "dependent" {
		if input.TitularLineID == nil || strings.TrimSpace(*input.TitularLineID) == "" {
			return nil, httputil.ValidationError(notifications.PhoneLineTitularRequired)
		}
		tid := strings.TrimSpace(*input.TitularLineID)
		if tid == phoneLineID {
			return nil, httputil.ValidationError(notifications.PhoneLineTitularSelf)
		}
		titular, err := s.GetPhoneLine(ctx, tid)
		if err != nil {
			return nil, httputil.NotFoundError(notifications.PhoneLineTitularNotFound)
		}
		if titular.LineClassification != "titular" {
			if err := s.Store.UpdatePhoneLineClassification(ctx, titular.ID, "titular", nil); err != nil {
				return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
			}
			s.auditLog(ctx, "Reclassify", "PhoneLine", titular.ID, map[string]any{"classification": titular.LineClassification},
				map[string]any{"classification": "titular", "reason": "promoted_to_host_dependent"})
		}
		titularID = &tid
	}

	prevClass := line.LineClassification
	prevTitular := line.TitularLineID
	if err := s.Store.UpdatePhoneLineClassification(ctx, phoneLineID, class, titularID); err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	s.auditLog(ctx, "Update", "PhoneLine", phoneLineID,
		map[string]any{"classification": prevClass, "titular_line_id": prevTitular},
		map[string]any{"classification": class, "titular_line_id": titularID})

	if prevClass == "titular" && class != "titular" {
		if err := s.reclassifyDependentsOf(ctx, phoneLineID, "titular cancelled / placed as Normal"); err != nil {
			return nil, err
		}
	}
	if prevTitular != nil && (titularID == nil || *prevTitular != *titularID) {
		if err := s.normalizeOrphanTitular(ctx, *prevTitular); err != nil {
			return nil, err
		}
	}
	if class == "titular" {
		if err := s.normalizeOrphanTitular(ctx, phoneLineID); err != nil {
			return nil, err
		}
	}
	return s.GetPhoneLine(ctx, phoneLineID)
}

func (s *Service) reclassifyDependentsOf(ctx context.Context, titularID, reason string) error {
	ids, err := s.Store.ListDependentLineIDs(ctx, titularID)
	if err != nil {
		return httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if len(ids) == 0 {
		return nil
	}
	if err := s.Store.ReclassifyDependentsToNormal(ctx, titularID); err != nil {
		return httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	s.auditLog(ctx, "Reclassify", "PhoneLine", titularID, nil, map[string]any{
		"message": "Dependentes reclassificadas por ausência de Titular.",
		"reason":  reason,
		"lines":   ids,
	})
	return nil
}

func (s *Service) normalizeOrphanTitular(ctx context.Context, titularID string) error {
	n, err := s.Store.CountDependents(ctx, titularID)
	if err != nil {
		return httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if n > 0 {
		return nil
	}
	class, _, _, err := s.Store.GetPhoneLineClassification(ctx, titularID)
	if err != nil || class != "titular" {
		return err
	}
	if err := s.Store.UpdatePhoneLineClassification(ctx, titularID, "normal", nil); err != nil {
		return httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	s.auditLog(ctx, "Reclassify", "PhoneLine", titularID, map[string]any{"classification": "titular"},
		map[string]any{"classification": "normal", "message": "Titular reclassificado por ausência de Dependentes."})
	return nil
}

func (s *Service) PutPhoneLineTransition(ctx context.Context, phoneLineID string, input models.PutPhoneLineTransitionInput) (*models.GetPhoneLineResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	line, err := s.GetPhoneLine(ctx, phoneLineID)
	if err != nil {
		return nil, err
	}

	if err := s.SM().ValidateTransition(statemachine.EntityPhoneLine, line.Status, "in_transition", userRoles(ctx)); err != nil {
		return nil, err
	}

	sub := strings.ToLower(strings.TrimSpace(input.TransitionSubStatus))
	switch sub {
	case "pending_activation", "pending_cancellation", "pending_portability", "pending_pp", "pending_tt":
	default:
		return nil, httputil.ValidationError(notifications.PhoneLineTransitionSubStatusInvalid)
	}
	start := time.Now().UTC()
	if input.StartDate != nil && strings.TrimSpace(*input.StartDate) != "" {
		parsed, err := parseFinancialDate(*input.StartDate)
		if err != nil {
			return nil, err
		}
		start = parsed
	}
	if err := s.Store.UpdatePhoneLineTransition(ctx, phoneLineID, "in_transition", sub, start); err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}

	var actorUserID *string
	if u := auth.UserFromContext(ctx); u != nil && u.ID != "" {
		actorUserID = &u.ID
	}
	_ = s.SM().RecordTransition(ctx, orgID, statemachine.EntityPhoneLine, phoneLineID, line.Status, "in_transition", "put_transition", nil, actorUserID, nil)

	s.auditLog(ctx, "Transition", "PhoneLine", phoneLineID, nil, map[string]any{
		"status": "in_transition", "sub_status": sub, "started_at": start.Format("2006-01-02"),
	})
	s.DispatchWebhookEvent(orgID, WebhookEventLineTransition, map[string]any{
		"phone_line_id": phoneLineID, "from": line.Status, "to": "in_transition", "sub_status": sub,
	})
	return s.GetPhoneLine(ctx, phoneLineID)
}

func (s *Service) CreatePhoneLineService(ctx context.Context, phoneLineID string, input models.CreatePhoneLineServiceInput) (*models.GetPhoneLineServiceResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := s.GetPhoneLine(ctx, phoneLineID); err != nil {
		return nil, err
	}
	planSvcID := strings.TrimSpace(input.ProviderPlanServiceID)
	if planSvcID == "" {
		return nil, httputil.ValidationError(notifications.N("PHONE_LINE_PLAN_SERVICE_REQUIRED", "Serviço do portfólio é obrigatório."))
	}
	name, code, svcType, recurring, price, err := s.Store.GetProviderPlanService(ctx, orgID, planSvcID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if name == "" {
		return nil, httputil.NotFoundError(notifications.N("PROVIDER_PLAN_SERVICE_NOT_FOUND", "Serviço do portfólio não encontrado."))
	}
	if input.Name != "" {
		name = strings.TrimSpace(input.Name)
	}
	if input.Price != nil {
		price = input.Price
	}
	if input.ServiceType != nil && strings.TrimSpace(*input.ServiceType) != "" {
		svcType = strings.ToLower(strings.TrimSpace(*input.ServiceType))
	}
	dup, err := s.Store.ActiveServiceTypeExists(ctx, phoneLineID, svcType, nil)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if dup {
		return nil, httputil.BusinessError(notifications.PhoneLineServiceTypeDuplicated)
	}
	var start, end *time.Time
	if input.StartDate != nil && strings.TrimSpace(*input.StartDate) != "" {
		t, err := parseFinancialDate(*input.StartDate)
		if err != nil {
			return nil, err
		}
		start = &t
		if err := s.blockRetroactiveIfClosed(ctx, orgID, phoneLineID, t); err != nil {
			return nil, err
		}
	}
	if input.EndDate != nil && strings.TrimSpace(*input.EndDate) != "" {
		t, err := parseFinancialDate(*input.EndDate)
		if err != nil {
			return nil, err
		}
		end = &t
	}
	id := uuid.New().String()
	if err := s.Store.CreatePhoneLineService(ctx, id, phoneLineID, planSvcID, name, code, recurring, price, &svcType, start, end); err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	s.auditLog(ctx, "Create", "PhoneLineService", id, nil, map[string]any{
		"phone_line_id": phoneLineID, "name": name, "service_type": svcType,
	})
	_, customerID, _ := s.Store.GetActivePhoneLineCustomerLink(ctx, phoneLineID)
	if customerID != "" {
		s.maybeGenerateAutomaticContract(ctx, orgID, customerID, phoneLineID, "service_add", "Inclusão de serviço: "+name)
	}
	prompt := s.maybePromptFidelityRenewal(ctx, orgID, phoneLineID, "new_service", input.RenewFidelity)
	line, err := s.GetPhoneLine(ctx, phoneLineID)
	if err != nil {
		return nil, err
	}
	for _, svc := range line.Services {
		if svc.ID == id {
			svc.FidelityRenewalPrompt = prompt
			return &svc, nil
		}
	}
	return &models.GetPhoneLineServiceResponse{ID: id, PhoneLineID: phoneLineID, Name: name, Code: code, Recurring: recurring, Price: price, Active: true, ServiceType: &svcType, FidelityRenewalPrompt: prompt}, nil
}

func (s *Service) DeletePhoneLineService(ctx context.Context, phoneLineID, serviceID string) error {
	if _, err := s.GetPhoneLine(ctx, phoneLineID); err != nil {
		return err
	}
	if err := s.Store.DeactivatePhoneLineService(ctx, phoneLineID, serviceID); err != nil {
		if isPgNoRows(err) {
			return httputil.NotFoundError(notifications.PhoneLineServiceNotFound)
		}
		return httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	s.auditLog(ctx, "Delete", "PhoneLineService", serviceID, map[string]any{"phone_line_id": phoneLineID}, nil)
	return nil
}

func (s *Service) blockRetroactiveIfClosed(ctx context.Context, orgID, phoneLineID string, at time.Time) error {
	if auth.IsMaster(ctx) {
		return nil
	}
	providerID, err := s.Store.GetPhoneLineProviderID(ctx, orgID, phoneLineID)
	if err != nil || providerID == "" {
		return err
	}
	blocked, err := s.Store.ExistsClosedProcessingMonthIntersecting(ctx, orgID, providerID, at, at)
	if err != nil {
		return httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if blocked {
		return httputil.BusinessError(notifications.ProcessingMonthRetroactiveBlocked)
	}
	return nil
}

func (s *Service) requireActiveServiceToLink(ctx context.Context, phoneLineID string, monthlyAmount *float64) error {
	if monthlyAmount != nil && *monthlyAmount > 0 {
		return nil
	}
	ok, err := s.Store.PhoneLineHasActiveService(ctx, phoneLineID)
	if err != nil {
		return httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if !ok {
		return httputil.BusinessError(notifications.PhoneLineServiceRequiredToLink)
	}
	return nil
}
