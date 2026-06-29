package services

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/luxus-connect/telefonia/api/internal/httputil"
	"github.com/luxus-connect/telefonia/api/internal/models"
	"github.com/luxus-connect/telefonia/api/internal/notifications"
	"github.com/luxus-connect/telefonia/api/internal/store"
)

const (
	perspectiveLuxusCustomer   = "luxus_customer"
	perspectiveCustomerEndUser = "customer_end_user"
)

func (s *Service) ListLineBillingProcessings(ctx context.Context, phoneLineID string) (*models.ListLineBillingProcessingsResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := s.GetPhoneLine(ctx, phoneLineID); err != nil {
		return nil, err
	}
	linkID, err := s.Store.GetActiveLinkIDForPhoneLine(ctx, orgID, phoneLineID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if linkID == "" {
		return &models.ListLineBillingProcessingsResponse{Processings: []models.LineBillingProcessingResponse{}}, nil
	}
	rows, err := s.Store.ListBillingProcessingsForLink(ctx, linkID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	resp := &models.ListLineBillingProcessingsResponse{LinkID: linkID, Processings: make([]models.LineBillingProcessingResponse, 0, len(rows))}
	for _, row := range rows {
		item, err := s.Store.ToBillingProcessingResponse(ctx, row)
		if err != nil {
			return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
		}
		resp.Processings = append(resp.Processings, item)
	}
	return resp, nil
}

func (s *Service) EnsureBillingProcessingsForLink(ctx context.Context, linkID, customerID string, monthlyAmount *float64) error {
	now := time.Now().UTC()
	existing, err := s.Store.ListBillingProcessingsForLink(ctx, linkID)
	if err != nil {
		return err
	}
	hasPrimary := false
	hasSecondary := false
	for _, p := range existing {
		if p.Perspective == perspectiveLuxusCustomer {
			hasPrimary = true
		}
		if p.Perspective == perspectiveCustomerEndUser {
			hasSecondary = true
		}
	}
	if !hasPrimary {
		procID := uuid.New().String()
		if err := s.Store.CreateBillingProcessing(ctx, store.BillingProcessingRow{
			ID:                      procID,
			PhoneLineCustomerLinkID: linkID,
			Perspective:             perspectiveLuxusCustomer,
			Active:                  true,
			CreatedAt:               now,
			UpdatedAt:               now,
		}); err != nil {
			return err
		}
		amount := 0.0
		if monthlyAmount != nil && *monthlyAmount > 0 {
			amount = *monthlyAmount
		}
		if amount > 0 {
			if err := s.Store.CreateBillingCompositionItem(ctx, store.BillingCompositionItemRow{
				ID:           uuid.New().String(),
				ProcessingID: procID,
				ItemType:     "service",
				Description:  "Mensalidade",
				Amount:       amount,
				Quantity:     1,
				Active:       true,
				CreatedAt:    now,
				UpdatedAt:    now,
			}); err != nil {
				return err
			}
		}
	}
	isReseller, err := s.Store.CustomerIsReseller(ctx, customerID)
	if err != nil {
		return err
	}
	if isReseller && !hasSecondary {
		label := "Usuário final"
		if err := s.Store.CreateBillingProcessing(ctx, store.BillingProcessingRow{
			ID:                      uuid.New().String(),
			PhoneLineCustomerLinkID: linkID,
			Perspective:             perspectiveCustomerEndUser,
			Label:                   &label,
			Active:                  true,
			CreatedAt:               now,
			UpdatedAt:               now,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) EnableEndUserProcessing(ctx context.Context, phoneLineID string) (*models.LineBillingProcessingResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := s.GetPhoneLine(ctx, phoneLineID); err != nil {
		return nil, err
	}
	linkID, err := s.Store.GetActiveLinkIDForPhoneLine(ctx, orgID, phoneLineID)
	if err != nil || linkID == "" {
		return nil, httputil.BusinessError(notifications.PhoneLineActiveCustomerLinkNotFound)
	}
	_, customerID, err := s.Store.GetActivePhoneLineCustomerLink(ctx, phoneLineID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	isReseller, err := s.Store.CustomerIsReseller(ctx, customerID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if !isReseller {
		return nil, httputil.ValidationError(notifications.N("CUSTOMER_NOT_RESELLER", "Cliente não é PJ revendedor. Ative a flag no cadastro do cliente."))
	}
	existing, err := s.Store.ListBillingProcessingsForLink(ctx, linkID)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	for _, p := range existing {
		if p.Perspective == perspectiveCustomerEndUser {
			resp, err := s.Store.ToBillingProcessingResponse(ctx, p)
			if err != nil {
				return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
			}
			return &resp, nil
		}
	}
	now := time.Now().UTC()
	procID := uuid.New().String()
	label := "Usuário final"
	if err := s.Store.CreateBillingProcessing(ctx, store.BillingProcessingRow{
		ID:                      procID,
		PhoneLineCustomerLinkID: linkID,
		Perspective:             perspectiveCustomerEndUser,
		Label:                   &label,
		Active:                  true,
		CreatedAt:               now,
		UpdatedAt:               now,
	}); err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	s.auditLog(ctx, "Create", "LineBillingProcessing", procID, nil, map[string]any{"perspective": perspectiveCustomerEndUser})
	resp, err := s.Store.ToBillingProcessingResponse(ctx, store.BillingProcessingRow{
		ID: procID, PhoneLineCustomerLinkID: linkID, Perspective: perspectiveCustomerEndUser, Label: &label, Active: true,
	})
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	return &resp, nil
}

func (s *Service) UpdateLineBillingProcessing(ctx context.Context, phoneLineID, processingID string, input models.UpdateLineBillingProcessingInput) (*models.LineBillingProcessingResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := s.GetPhoneLine(ctx, phoneLineID); err != nil {
		return nil, err
	}
	before, err := s.Store.GetBillingProcessing(ctx, orgID, processingID)
	if err != nil || before == nil {
		return nil, httputil.NotFoundError(notifications.N("BILLING_PROCESSING_NOT_FOUND", "Processamento não encontrado."))
	}
	now := time.Now().UTC()
	if err := s.Store.UpdateBillingProcessing(ctx, processingID, input.Label, input.MirrorFromPrimary, now); err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	if input.MirrorFromPrimary != nil && *input.MirrorFromPrimary {
		if err := s.mirrorProcessingFromPrimary(ctx, orgID, before.PhoneLineCustomerLinkID, processingID); err != nil {
			return nil, err
		}
	}
	s.auditLog(ctx, "Update", "LineBillingProcessing", processingID,
		map[string]any{"label": before.Label, "mirror": before.MirrorFromPrimary},
		map[string]any{"label": input.Label, "mirror": input.MirrorFromPrimary})
	after, err := s.Store.GetBillingProcessing(ctx, orgID, processingID)
	if err != nil || after == nil {
		return nil, httputil.NotFoundError(notifications.N("BILLING_PROCESSING_NOT_FOUND", "Processamento não encontrado."))
	}
	resp, err := s.Store.ToBillingProcessingResponse(ctx, *after)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	return &resp, nil
}

func (s *Service) MirrorProcessingFromPrimary(ctx context.Context, phoneLineID, processingID string) (*models.LineBillingProcessingResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := s.GetPhoneLine(ctx, phoneLineID); err != nil {
		return nil, err
	}
	target, err := s.Store.GetBillingProcessing(ctx, orgID, processingID)
	if err != nil || target == nil {
		return nil, httputil.NotFoundError(notifications.N("BILLING_PROCESSING_NOT_FOUND", "Processamento não encontrado."))
	}
	if target.Perspective != perspectiveCustomerEndUser {
		return nil, httputil.ValidationError(notifications.N("BILLING_MIRROR_SECONDARY_ONLY", "Espelhamento só se aplica ao processamento cliente→usuário final."))
	}
	if err := s.mirrorProcessingFromPrimary(ctx, orgID, target.PhoneLineCustomerLinkID, processingID); err != nil {
		return nil, err
	}
	mirror := true
	now := time.Now().UTC()
	_ = s.Store.UpdateBillingProcessing(ctx, processingID, nil, &mirror, now)
	resp, err := s.Store.ToBillingProcessingResponse(ctx, *target)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	return &resp, nil
}

func (s *Service) mirrorProcessingFromPrimary(ctx context.Context, orgID, linkID, targetProcessingID string) error {
	processings, err := s.Store.ListBillingProcessingsForLink(ctx, linkID)
	if err != nil {
		return httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	var primaryID string
	for _, p := range processings {
		if p.Perspective == perspectiveLuxusCustomer {
			primaryID = p.ID
			break
		}
	}
	if primaryID == "" {
		return httputil.BusinessError(notifications.N("BILLING_PRIMARY_MISSING", "Processamento Luxus→Cliente não encontrado."))
	}
	primaryItems, err := s.Store.ListBillingCompositionItems(ctx, primaryID)
	if err != nil {
		return httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	now := time.Now().UTC()
	if err := s.Store.DeactivateBillingCompositionItemsForProcessing(ctx, targetProcessingID, now); err != nil {
		return httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	for _, it := range primaryItems {
		copy := it
		copy.ID = uuid.New().String()
		copy.ProcessingID = targetProcessingID
		copy.CreatedAt = now
		copy.UpdatedAt = now
		if err := s.Store.CreateBillingCompositionItem(ctx, copy); err != nil {
			return httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
		}
	}
	s.auditLog(ctx, "Mirror", "LineBillingProcessing", targetProcessingID, nil, map[string]any{"from": primaryID})
	return nil
}

func (s *Service) CreateLineBillingCompositionItem(ctx context.Context, phoneLineID, processingID string, input models.CreateLineBillingCompositionItemInput) (*models.LineBillingCompositionItemResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := s.GetPhoneLine(ctx, phoneLineID); err != nil {
		return nil, err
	}
	proc, err := s.Store.GetBillingProcessing(ctx, orgID, processingID)
	if err != nil || proc == nil {
		return nil, httputil.NotFoundError(notifications.N("BILLING_PROCESSING_NOT_FOUND", "Processamento não encontrado."))
	}
	if err := validateCompositionItemInput(input.ItemType, input.Description, input.Amount); err != nil {
		return nil, err
	}
	qty := 1.0
	if input.Quantity != nil && *input.Quantity > 0 {
		qty = *input.Quantity
	}
	now := time.Now().UTC()
	id := uuid.New().String()
	row := store.BillingCompositionItemRow{
		ID: id, ProcessingID: processingID, ItemType: strings.ToLower(strings.TrimSpace(input.ItemType)),
		Description: strings.TrimSpace(input.Description), Amount: input.Amount, Quantity: qty,
		InstallmentCount: input.InstallmentCount, InstallmentCurrent: input.InstallmentCurrent,
		Active: true, CreatedAt: now, UpdatedAt: now,
	}
	if input.StartDate != nil && strings.TrimSpace(*input.StartDate) != "" {
		t, err := parseFinancialDate(*input.StartDate)
		if err != nil {
			return nil, err
		}
		row.StartDate = &t
	}
	if input.EndDate != nil && strings.TrimSpace(*input.EndDate) != "" {
		t, err := parseFinancialDate(*input.EndDate)
		if err != nil {
			return nil, err
		}
		row.EndDate = &t
	}
	if err := s.Store.CreateBillingCompositionItem(ctx, row); err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	s.auditLog(ctx, "Create", "LineBillingCompositionItem", id, nil, compositionAuditMap(row))
	if proc.Perspective == perspectiveLuxusCustomer {
		_ = s.syncLinkMonthlyAmountFromPrimary(ctx, proc.PhoneLineCustomerLinkID, processingID)
	}
	resp := store.CompositionItemToModel(row)
	return &resp, nil
}

func (s *Service) UpdateLineBillingCompositionItem(ctx context.Context, phoneLineID, processingID, itemID string, input models.UpdateLineBillingCompositionItemInput) (*models.LineBillingCompositionItemResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := s.GetPhoneLine(ctx, phoneLineID); err != nil {
		return nil, err
	}
	before, err := s.Store.GetBillingCompositionItem(ctx, orgID, itemID)
	if err != nil || before == nil || before.ProcessingID != processingID {
		return nil, httputil.NotFoundError(notifications.N("BILLING_ITEM_NOT_FOUND", "Item de composição não encontrado."))
	}
	if input.Amount != nil && *input.Amount < 0 {
		return nil, httputil.ValidationError(notifications.N("BILLING_ITEM_AMOUNT_INVALID", "Valor não pode ser negativo."))
	}
	now := time.Now().UTC()
	var startDate, endDate *time.Time
	if input.StartDate != nil && strings.TrimSpace(*input.StartDate) != "" {
		t, err := parseFinancialDate(*input.StartDate)
		if err != nil {
			return nil, err
		}
		startDate = &t
	}
	if input.EndDate != nil && strings.TrimSpace(*input.EndDate) != "" {
		t, err := parseFinancialDate(*input.EndDate)
		if err != nil {
			return nil, err
		}
		endDate = &t
	}
	if err := s.Store.UpdateBillingCompositionItem(ctx, itemID, input.Description, input.Amount, input.Quantity,
		input.InstallmentCount, input.InstallmentCurrent, startDate, endDate, now); err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	s.auditLog(ctx, "Update", "LineBillingCompositionItem", itemID, compositionAuditMap(*before), input)
	proc, _ := s.Store.GetBillingProcessing(ctx, orgID, processingID)
	if proc != nil && proc.Perspective == perspectiveLuxusCustomer {
		_ = s.syncLinkMonthlyAmountFromPrimary(ctx, proc.PhoneLineCustomerLinkID, processingID)
	}
	after, err := s.Store.GetBillingCompositionItem(ctx, orgID, itemID)
	if err != nil || after == nil {
		return nil, httputil.NotFoundError(notifications.N("BILLING_ITEM_NOT_FOUND", "Item de composição não encontrado."))
	}
	resp := store.CompositionItemToModel(*after)
	return &resp, nil
}

func (s *Service) DeleteLineBillingCompositionItem(ctx context.Context, phoneLineID, processingID, itemID string) error {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return err
	}
	if _, err := s.GetPhoneLine(ctx, phoneLineID); err != nil {
		return err
	}
	before, err := s.Store.GetBillingCompositionItem(ctx, orgID, itemID)
	if err != nil || before == nil || before.ProcessingID != processingID {
		return httputil.NotFoundError(notifications.N("BILLING_ITEM_NOT_FOUND", "Item de composição não encontrado."))
	}
	items, err := s.Store.ListBillingCompositionItems(ctx, processingID)
	if err != nil {
		return httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	serviceCount := 0
	for _, it := range items {
		if it.ItemType == "service" {
			serviceCount++
		}
	}
	if before.ItemType == "service" && serviceCount <= 1 {
		return httputil.ValidationError(notifications.N("BILLING_MIN_ONE_SERVICE", "O processamento deve ter ao menos um serviço ativo."))
	}
	now := time.Now().UTC()
	if err := s.Store.DeactivateBillingCompositionItem(ctx, itemID, now); err != nil {
		return httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	s.auditLog(ctx, "Delete", "LineBillingCompositionItem", itemID, compositionAuditMap(*before), nil)
	proc, _ := s.Store.GetBillingProcessing(ctx, orgID, processingID)
	if proc != nil && proc.Perspective == perspectiveLuxusCustomer {
		_ = s.syncLinkMonthlyAmountFromPrimary(ctx, proc.PhoneLineCustomerLinkID, processingID)
	}
	return nil
}

func (s *Service) ListLineBillingProcessingAudit(ctx context.Context, phoneLineID, processingID string) ([]models.AuditLogResponse, error) {
	orgID, err := orgFrom(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := s.GetPhoneLine(ctx, phoneLineID); err != nil {
		return nil, err
	}
	if proc, err := s.Store.GetBillingProcessing(ctx, orgID, processingID); err != nil || proc == nil {
		return nil, httputil.NotFoundError(notifications.N("BILLING_PROCESSING_NOT_FOUND", "Processamento não encontrado."))
	}
	return s.listAuditForProcessing(ctx, processingID)
}

func (s *Service) listAuditForProcessing(ctx context.Context, processingID string) ([]models.AuditLogResponse, error) {
	rows, err := s.Store.ListAuditLogsForEntity(ctx, "LineBillingProcessing", processingID, 100)
	if err != nil {
		return nil, httputil.InternalError(notifications.SharedUnexpectedError(err.Error()))
	}
	itemRows, _ := s.Store.ListBillingCompositionItems(ctx, processingID)
	ids := make([]string, 0, len(itemRows))
	for _, it := range itemRows {
		ids = append(ids, it.ID)
	}
	for _, id := range ids {
		more, err := s.Store.ListAuditLogsForEntity(ctx, "LineBillingCompositionItem", id, 50)
		if err == nil {
			rows = append(rows, more...)
		}
	}
	out := make([]models.AuditLogResponse, 0, len(rows))
	for _, r := range rows {
		out = append(out, models.AuditLogResponse{
			ID: r.ID, ChangeType: r.ChangeType, EntityName: r.EntityName, KeyValues: r.KeyValues,
			ChangedBy: r.ChangedBy, OldValues: r.OldValues, NewValues: r.NewValues, Timestamp: r.Timestamp,
		})
	}
	return out, nil
}

func (s *Service) syncLinkMonthlyAmountFromPrimary(ctx context.Context, linkID, processingID string) error {
	total, err := s.Store.SumBillingProcessingTotal(ctx, processingID)
	if err != nil {
		return err
	}
	return s.Store.UpdateActivePhoneLineCustomerLinkAmountByLinkID(ctx, linkID, &total)
}

func validateCompositionItemInput(itemType, description string, amount float64) error {
	t := strings.ToLower(strings.TrimSpace(itemType))
	switch t {
	case "service", "discount", "extra_charge", "installment":
	default:
		return httputil.ValidationError(notifications.N("BILLING_ITEM_TYPE_INVALID", "Tipo inválido. Use: service, discount, extra_charge, installment."))
	}
	if strings.TrimSpace(description) == "" {
		return httputil.ValidationError(notifications.N("BILLING_ITEM_DESCRIPTION_REQUIRED", "Descrição obrigatória."))
	}
	if amount < 0 {
		return httputil.ValidationError(notifications.N("BILLING_ITEM_AMOUNT_INVALID", "Valor não pode ser negativo."))
	}
	return nil
}

func (s *Service) auditLog(ctx context.Context, changeType, entityName, key string, oldVal, newVal any) {
	var oldStr, newStr *string
	if oldVal != nil {
		if b, err := json.Marshal(oldVal); err == nil {
			s := string(b)
			oldStr = &s
		}
	}
	if newVal != nil {
		if b, err := json.Marshal(newVal); err == nil {
			s := string(b)
			newStr = &s
		}
	}
	var changedBy *string
	if u, err := userFrom(ctx); err == nil && u != nil {
		changedBy = &u.ID
	}
	_ = s.Store.InsertAuditLog(ctx, uuid.New().String(), changeType, entityName, key, changedBy, oldStr, newStr, time.Now().UTC())
}

func compositionAuditMap(row store.BillingCompositionItemRow) map[string]any {
	return map[string]any{
		"item_type": row.ItemType, "description": row.Description, "amount": row.Amount, "quantity": row.Quantity,
	}
}
