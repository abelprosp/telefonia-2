package services

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/models"
	"github.com/luxus-connect/telefonia/api/internal/notifications"
	"github.com/luxus-connect/telefonia/api/internal/store"
)

func (s *Service) ListExceedanceTerms(ctx context.Context) ([]models.ExceedanceTermResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	items, err := s.Store.ListExceedanceTerms(ctx, orgID, false)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	return items, nil
}

func (s *Service) CreateExceedanceTerm(ctx context.Context, input models.CreateExceedanceTermInput) (*models.ExceedanceTermResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	term := strings.TrimSpace(input.Term)
	if term == "" {
		return nil, httputil.ValidationError(notifications.ExceedanceTermRequired)
	}
	chargeType := store.NormalizeExceedanceChargeType(input.ChargeType)
	if chargeType == "" {
		chargeType = "mirroed"
	}
	if input.ChargeType != "" && store.NormalizeExceedanceChargeType(input.ChargeType) == "" {
		return nil, httputil.ValidationError(notifications.ExceedanceChargeTypeInvalid)
	}
	active := true
	if input.Active != nil {
		active = *input.Active
	}
	now := time.Now().UTC()
	id := uuid.New().String()
	if err := s.Store.CreateExceedanceTerm(ctx, store.ExceedanceTermRow{
		ID: id, OrganizationID: orgID, Term: term, ChargeType: chargeType,
		TabulatedAmount: input.TabulatedAmount, Active: active, CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	item, err := s.Store.GetExceedanceTerm(ctx, orgID, id)
	if err != nil || item == nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError("termo criado mas não encontrado"))
	}
	return item, nil
}

func (s *Service) UpdateExceedanceTerm(ctx context.Context, id string, input models.UpdateExceedanceTermInput) (*models.ExceedanceTermResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	chargeType := ""
	if input.ChargeType != nil {
		chargeType = store.NormalizeExceedanceChargeType(*input.ChargeType)
		if chargeType == "" {
			return nil, httputil.ValidationError(notifications.ExceedanceChargeTypeInvalid)
		}
	}
	term := ""
	if input.Term != nil {
		term = strings.TrimSpace(*input.Term)
	}
	if err := s.Store.UpdateExceedanceTerm(ctx, orgID, id, term, chargeType, input.TabulatedAmount, input.Active, time.Now().UTC()); err != nil {
		if isPgNoRows(err) {
			return nil, httputil.NotFoundError(notifications.ExceedanceTermNotFound)
		}
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	item, err := s.Store.GetExceedanceTerm(ctx, orgID, id)
	if err != nil || item == nil {
		return nil, httputil.NotFoundError(notifications.ExceedanceTermNotFound)
	}
	return item, nil
}

func (s *Service) UpdatePhoneLineExceedanceSettings(ctx context.Context, phoneLineID string, input models.UpdatePhoneLineExceedanceInput) (*models.GetPhoneLineResponse, error) {
	if _, err := s.GetPhoneLine(ctx, phoneLineID); err != nil {
		return nil, err
	}
	var chargeType *string
	if input.ExceedanceChargeType != nil {
		n := store.NormalizeExceedanceChargeType(*input.ExceedanceChargeType)
		if n == "" {
			return nil, httputil.ValidationError(notifications.ExceedanceChargeTypeInvalid)
		}
		chargeType = &n
	}
	if err := s.Store.UpdatePhoneLineExceedanceSettings(ctx, phoneLineID, input.ChargeExceedances, chargeType); err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	return s.GetPhoneLine(ctx, phoneLineID)
}

func MatchExceedanceTerm(description string, terms []models.ExceedanceTermResponse) *models.ExceedanceTermResponse {
	hay := strings.ToLower(strings.TrimSpace(description))
	if hay == "" {
		return nil
	}
	var best *models.ExceedanceTermResponse
	bestLen := 0
	for i := range terms {
		needle := strings.ToLower(strings.TrimSpace(terms[i].Term))
		if needle == "" || !strings.Contains(hay, needle) {
			continue
		}
		if len(needle) > bestLen {
			best = &terms[i]
			bestLen = len(needle)
		}
	}
	return best
}

func ChargedExceedanceAmount(invoiceAmount float64, lineType string, term *models.ExceedanceTermResponse) (amount float64, chargeType string) {
	effective := strings.ToLower(strings.TrimSpace(lineType))
	if term != nil && term.ChargeType == "tabulated" {
		effective = "tabulated"
	}
	if effective == "tabulated" {
		if term != nil && term.TabulatedAmount != nil {
			return *term.TabulatedAmount, "tabulated"
		}
		return invoiceAmount, "tabulated"
	}
	return invoiceAmount, "mirrored"
}
