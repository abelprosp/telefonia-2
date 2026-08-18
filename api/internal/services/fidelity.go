package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/luxus-connect/telefonia/api/internal/billingcalc"
	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/models"
	"github.com/luxus-connect/telefonia/api/internal/notifications"
	"github.com/luxus-connect/telefonia/api/internal/store"
)

func (s *Service) GetLineFidelity(ctx context.Context, phoneLineID string) (*models.LineFidelityResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := s.GetPhoneLine(ctx, phoneLineID); err != nil {
		return nil, err
	}
	_ = s.applyDueFidelityRenewals(ctx, orgID)
	row, err := s.Store.GetLineFidelity(ctx, phoneLineID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if row == nil {
		return nil, httputil.NotFoundError(notifications.LineFidelityNotFound)
	}
	return s.toFidelityResponse(ctx, *row)
}

func (s *Service) UpsertLineFidelity(ctx context.Context, phoneLineID string, input models.UpsertLineFidelityInput) (*models.LineFidelityResponse, error) {
	if _, err := s.GetPhoneLine(ctx, phoneLineID); err != nil {
		return nil, err
	}
	if input.InitialMonths <= 0 {
		return nil, httputil.ValidationError(notifications.LineFidelityMonthsInvalid)
	}
	start, err := parseFinancialDate(input.StartDate)
	if err != nil {
		return nil, err
	}
	autoRenew := false
	if input.AutoRenew != nil {
		autoRenew = *input.AutoRenew
	}
	if autoRenew && (input.RenewalPeriodMonths == nil || *input.RenewalPeriodMonths <= 0) {
		return nil, httputil.ValidationError(notifications.LineFidelityRenewalInvalid)
	}
	now := time.Now().UTC()
	existing, err := s.Store.GetLineFidelity(ctx, phoneLineID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	id := uuid.New().String()
	if existing != nil {
		id = existing.ID
	}
	predicted := store.PredictedFidelityEnd(start, input.InitialMonths)
	status := "active"
	if predicted.Before(now.Truncate(24 * time.Hour)) {
		status = "expired"
	}
	row := store.LineFidelityRow{
		ID: id, PhoneLineID: phoneLineID, StartDate: start, InitialMonths: input.InitialMonths,
		PredictedEndDate: predicted, AutoRenew: autoRenew, RenewalPeriodMonths: input.RenewalPeriodMonths,
		Status: status, CreatedAt: now, UpdatedAt: now,
	}
	if existing != nil {
		row.CreatedAt = existing.CreatedAt
	}
	if err := s.Store.UpsertLineFidelity(ctx, row); err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if existing == nil {
		_ = s.Store.InsertLineFidelityEvent(ctx, uuid.New().String(), id, "created", now, userIDPtr(ctx), strPtr("Cadastro de fidelidade."))
	}
	s.auditLog(ctx, "Upsert", "LineFidelity", id, existing, row)
	return s.toFidelityResponse(ctx, row)
}

func (s *Service) DecideFidelityRenewal(ctx context.Context, phoneLineID string, input models.FidelityRenewalDecisionInput) (*models.LineFidelityResponse, error) {
	if _, err := s.GetPhoneLine(ctx, phoneLineID); err != nil {
		return nil, err
	}
	row, err := s.Store.GetLineFidelity(ctx, phoneLineID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if row == nil {
		return nil, httputil.NotFoundError(notifications.LineFidelityNotFound)
	}
	now := time.Now().UTC()
	notes := strings.TrimSpace(input.Notes)
	if input.Renew {
		predicted := store.PredictedFidelityEnd(now, row.InitialMonths)
		if err := s.Store.UpdateLineFidelityDates(ctx, row.ID, predicted, "active", now); err != nil {
			return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
		}
		if notes == "" {
			notes = "Renovação por alteração contratual."
		}
		_ = s.Store.InsertLineFidelityEvent(ctx, uuid.New().String(), row.ID, "contractual_change", now, userIDPtr(ctx), &notes)
		row.PredictedEndDate = predicted
		row.Status = "active"
	} else {
		if notes == "" {
			notes = "Operador optou por não renovar a fidelidade."
		}
		_ = s.Store.InsertLineFidelityEvent(ctx, uuid.New().String(), row.ID, "declined", now, userIDPtr(ctx), &notes)
	}
	return s.toFidelityResponse(ctx, *row)
}

func (s *Service) ListFidelityRenewalTriggers(ctx context.Context) ([]models.FidelityRenewalTriggerResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	items, err := s.Store.ListFidelityRenewalTriggers(ctx, orgID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	return items, nil
}

func (s *Service) UpdateFidelityRenewalTrigger(ctx context.Context, id string, input models.UpdateFidelityRenewalTriggerInput) (*models.FidelityRenewalTriggerResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	if input.PromptEnabled == nil {
		return nil, httputil.ValidationError(notifications.N("FIDELITY_TRIGGER_REQUIRED", "Informe prompt_enabled."))
	}
	if err := s.Store.UpdateFidelityRenewalTrigger(ctx, orgID, id, *input.PromptEnabled, time.Now().UTC()); err != nil {
		if isPgNoRows(err) {
			return nil, httputil.NotFoundError(notifications.N("FIDELITY_TRIGGER_NOT_FOUND", "Gatilho de renovação não encontrado."))
		}
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	items, err := s.Store.ListFidelityRenewalTriggers(ctx, orgID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	for i := range items {
		if items[i].ID == id {
			return &items[i], nil
		}
	}
	return nil, httputil.NotFoundError(notifications.N("FIDELITY_TRIGGER_NOT_FOUND", "Gatilho de renovação não encontrado."))
}

func (s *Service) applyDueFidelityRenewals(ctx context.Context, orgID string) error {
	now := time.Now().UTC()
	due, err := s.Store.ListDueAutoRenewFidelities(ctx, orgID, now)
	if err != nil {
		return err
	}
	for _, row := range due {
		period := 0
		if row.RenewalPeriodMonths != nil {
			period = *row.RenewalPeriodMonths
		}
		if period <= 0 {
			continue
		}
		predicted := row.PredictedEndDate
		for !predicted.After(now) {
			predicted = predicted.AddDate(0, period, 0)
		}
		if err := s.Store.UpdateLineFidelityDates(ctx, row.ID, predicted, "active", now); err != nil {
			continue
		}
		notes := "Renovação automática na data de término."
		_ = s.Store.InsertLineFidelityEvent(ctx, uuid.New().String(), row.ID, "automatic", now, nil, &notes)
		s.auditLog(ctx, "AutoRenew", "LineFidelity", row.ID, row.PredictedEndDate, predicted)
	}
	expired, err := s.Store.ListExpiredFidelitiesToMark(ctx, orgID, now)
	if err != nil {
		return err
	}
	for _, row := range expired {
		_ = s.Store.UpdateLineFidelityDates(ctx, row.ID, row.PredictedEndDate, "expired", now)
	}
	return nil
}

func (s *Service) maybePromptFidelityRenewal(ctx context.Context, orgID, phoneLineID, eventKey string, renew *bool) bool {
	enabled, err := s.Store.FidelityTriggerPromptEnabled(ctx, orgID, eventKey)
	if err != nil || !enabled {
		return false
	}
	row, err := s.Store.GetLineFidelity(ctx, phoneLineID)
	if err != nil || row == nil {
		return false
	}
	if renew != nil {
		_, _ = s.DecideFidelityRenewal(ctx, phoneLineID, models.FidelityRenewalDecisionInput{
			Renew: *renew, Trigger: eventKey,
		})
		return false
	}
	return true
}

func (s *Service) EstimateLineFidelityPenalty(ctx context.Context, phoneLineID string, cancelDateStr *string, penaltyPercent *float64) (*models.FidelityPenaltyEstimateResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	line, err := s.GetPhoneLine(ctx, phoneLineID)
	if err != nil {
		return nil, err
	}
	fid, err := s.Store.GetLineFidelity(ctx, phoneLineID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if fid == nil {
		return nil, httputil.NotFoundError(notifications.LineFidelityNotFound)
	}

	cancelDate := time.Now().UTC()
	if cancelDateStr != nil && strings.TrimSpace(*cancelDateStr) != "" {
		parsed, err := parseFinancialDate(*cancelDateStr)
		if err != nil {
			return nil, err
		}
		cancelDate = parsed
	}

	pct := 30.0
	if penaltyPercent != nil && *penaltyPercent > 0 {
		pct = *penaltyPercent
	}

	monthlyAmount := 0.0
	linkID, err := s.Store.GetActiveLinkIDForPhoneLine(ctx, orgID, phoneLineID)
	if err == nil && linkID != "" {
		processings, _ := s.Store.ListBillingProcessingsForLink(ctx, linkID)
		for _, p := range processings {
			if p.Perspective == perspectiveLuxusCustomer {
				tot, err := s.Store.SumBillingProcessingTotal(ctx, p.ID)
				if err == nil && tot > 0 {
					monthlyAmount = tot
				}
				break
			}
		}
	}
	if monthlyAmount == 0 && line.BaseCost != nil {
		monthlyAmount = *line.BaseCost
	}

	calc := billingcalc.CalculateFidelityPenalty(monthlyAmount, fid.StartDate, fid.PredictedEndDate, cancelDate, pct)

	return &models.FidelityPenaltyEstimateResponse{
		PhoneLineID:       phoneLineID,
		FidelityStartDate: calc.FidelityStartDate,
		PredictedEndDate:  calc.PredictedEndDate,
		CancelDate:        calc.CancelDate,
		TotalMonths:       calc.TotalMonths,
		MonthsServed:      calc.MonthsServed,
		MonthsRemaining:   calc.MonthsRemaining,
		MonthlyAmount:     calc.MonthlyAmount,
		PenaltyPercentage: calc.PenaltyPercentage,
		PenaltyAmount:     calc.PenaltyAmount,
		IsExempt:          calc.IsExempt,
	}, nil
}

func (s *Service) ApplyLineFidelityPenalty(ctx context.Context, phoneLineID string, input models.ApplyFidelityPenaltyInput) (*models.FidelityPenaltyEstimateResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	estimate, err := s.EstimateLineFidelityPenalty(ctx, phoneLineID, input.CancelDate, input.PenaltyPercentage)
	if err != nil {
		return nil, err
	}
	if estimate.IsExempt || estimate.PenaltyAmount <= 0 {
		return estimate, nil
	}

	linkID, err := s.Store.GetActiveLinkIDForPhoneLine(ctx, orgID, phoneLineID)
	if err != nil || linkID == "" {
		return nil, httputil.BusinessError(notifications.PhoneLineActiveCustomerLinkNotFound)
	}

	processings, err := s.Store.ListBillingProcessingsForLink(ctx, linkID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	var primaryProcID string
	for _, p := range processings {
		if p.Perspective == perspectiveLuxusCustomer {
			primaryProcID = p.ID
			break
		}
	}
	if primaryProcID == "" {
		return nil, httputil.BusinessError(notifications.N("NO_PRIMARY_PROCESSING", "Processamento primário de faturamento não encontrado."))
	}

	now := time.Now().UTC()
	penaltyDesc := fmt.Sprintf("Multa rescisória por quebra de fidelidade (%d meses restantes a %.1f%%)", estimate.MonthsRemaining, estimate.PenaltyPercentage)
	if strings.TrimSpace(input.Justification) != "" {
		penaltyDesc += " — " + strings.TrimSpace(input.Justification)
	}

	row := store.BillingCompositionItemRow{
		ID:           uuid.New().String(),
		ProcessingID: primaryProcID,
		ItemType:     "extra_charge",
		Description:  penaltyDesc,
		Amount:       estimate.PenaltyAmount,
		Quantity:     1,
		Active:       true,
		CreatedAt:    now,
		UpdatedAt:    now,
		Proportional: false,
	}
	if err := s.Store.CreateBillingCompositionItem(ctx, row); err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}

	s.auditLog(ctx, "ApplyFidelityPenalty", "LineBillingCompositionItem", row.ID, nil, map[string]any{
		"phone_line_id":      phoneLineID,
		"penalty_amount":     estimate.PenaltyAmount,
		"months_remaining":   estimate.MonthsRemaining,
		"penalty_percentage": estimate.PenaltyPercentage,
	})

	_ = s.Store.InsertLineFidelityEvent(ctx, uuid.New().String(), phoneLineID, "penalty_applied", now, userIDPtr(ctx), &penaltyDesc)

	return estimate, nil
}

func (s *Service) toFidelityResponse(ctx context.Context, row store.LineFidelityRow) (*models.LineFidelityResponse, error) {
	history, err := s.Store.ListLineFidelityEvents(ctx, row.ID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	return &models.LineFidelityResponse{
		ID: row.ID, PhoneLineID: row.PhoneLineID, StartDate: row.StartDate, InitialMonths: row.InitialMonths,
		PredictedEndDate: row.PredictedEndDate, AutoRenew: row.AutoRenew, RenewalPeriodMonths: row.RenewalPeriodMonths,
		Status: row.Status, History: history,
	}, nil
}

func userIDPtr(ctx context.Context) *string {
	u, err := userFrom(ctx)
	if err != nil || u == nil {
		return nil
	}
	id := u.ID
	return &id
}

func strPtr(v string) *string { return &v }

